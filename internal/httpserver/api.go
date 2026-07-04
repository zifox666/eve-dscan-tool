package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zifox666/eve-dscan-tool/internal/model"
	"github.com/zifox666/eve-dscan-tool/internal/service/dscan"
	"github.com/zifox666/eve-dscan-tool/internal/store/postgres"
)

type submitDScanRequest struct {
	Data           string              `json:"data" form:"data"`
	DScanData      string              `json:"dscan_data" form:"dscan_data"`
	FilterDistance bool                `json:"filter_distance" form:"filter_distance"`
	UILang         string              `json:"ui_lang" form:"ui_lang"`
	GameLang       string              `json:"game_lang" form:"game_lang"`
	ManualShips    []manualShipRequest `json:"manual_ships"`
}

type manualShipRequest struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

type apiResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type createdDScanResponse struct {
	Type    string `json:"type"`
	ShortID string `json:"short_id"`
	ViewURL string `json:"view_url"`
}

func (s *Server) handleSubmitDScan(ctx *gin.Context) {
	var req submitDScanRequest
	if err := bindRequest(ctx, &req); err != nil {
		writeAPIError(ctx, http.StatusBadRequest, "请求格式不正确")
		return
	}

	req.applyLegacyForm(ctx)
	rawData := strings.TrimSpace(firstNonEmpty(req.Data, req.DScanData))
	rawData = appendManualShips(rawData, req.ManualShips)

	if rawData == "" {
		writeAPIError(ctx, http.StatusBadRequest, "DScan 数据不能为空")
		return
	}

	if detectDScanType(rawData) == "local" && len(req.ManualShips) == 0 {
		s.processLocal(ctx, rawData)
		return
	}

	s.processShip(ctx, rawData, req.FilterDistance)
}

func (s *Server) handleProcessLocalDScan(ctx *gin.Context) {
	var req submitDScanRequest
	if err := bindRequest(ctx, &req); err != nil {
		writeAPIError(ctx, http.StatusBadRequest, "请求格式不正确")
		return
	}
	req.applyLegacyForm(ctx)
	rawData := strings.TrimSpace(firstNonEmpty(req.Data, req.DScanData))
	if rawData == "" {
		writeAPIError(ctx, http.StatusBadRequest, "DScan 数据不能为空")
		return
	}

	s.processLocal(ctx, rawData)
}

func (s *Server) handleProcessShipDScan(ctx *gin.Context) {
	var req submitDScanRequest
	if err := bindRequest(ctx, &req); err != nil {
		writeAPIError(ctx, http.StatusBadRequest, "请求格式不正确")
		return
	}
	req.applyLegacyForm(ctx)
	rawData := strings.TrimSpace(firstNonEmpty(req.Data, req.DScanData))
	rawData = appendManualShips(rawData, req.ManualShips)
	if rawData == "" {
		writeAPIError(ctx, http.StatusBadRequest, "DScan 数据不能为空")
		return
	}

	s.processShip(ctx, rawData, req.FilterDistance)
}

func (s *Server) processLocal(ctx *gin.Context, rawData string) {
	result, err := dscan.AnalyzeLocal(ctx.Request.Context(), s.esi, rawData)
	if err != nil {
		s.logger.Warn("analyze local dscan", "error", err)
		writeAPIError(ctx, http.StatusBadGateway, "处理本地扫描失败")
		return
	}

	processed, err := json.Marshal(result)
	if err != nil {
		writeAPIError(ctx, http.StatusInternalServerError, "结果序列化失败")
		return
	}

	record, err := s.dscans.CreateLocal(ctx.Request.Context(), dscan.CreateInput{
		RawData:       rawData,
		ProcessedData: processed,
		ClientIP:      ctx.ClientIP(),
	})
	if err != nil {
		s.logger.Warn("create local dscan", "error", err)
		writeAPIError(ctx, http.StatusInternalServerError, "保存本地扫描失败")
		return
	}

	_ = s.cache.SetLocal(ctx.Request.Context(), cachedLocal(record))
	writeAPISuccess(ctx, http.StatusCreated, createdDScanResponse{
		Type:    "local",
		ShortID: record.ShortID,
		ViewURL: "/c/" + record.ShortID,
	})
}

func (s *Server) processShip(ctx *gin.Context, rawData string, filterDistance bool) {
	resultZH, err := dscan.AnalyzeShip(ctx.Request.Context(), s.sde, rawData, "zh", filterDistance)
	if err != nil {
		s.logger.Warn("analyze ship dscan zh", "error", err)
		writeAPIError(ctx, http.StatusBadGateway, "处理舰船扫描失败")
		return
	}
	resultEN, err := dscan.AnalyzeShip(ctx.Request.Context(), s.sde, rawData, "en", filterDistance)
	if err != nil {
		s.logger.Warn("analyze ship dscan en", "error", err)
		writeAPIError(ctx, http.StatusBadGateway, "处理舰船扫描失败")
		return
	}

	processedByLang := map[string]*dscan.ShipResult{
		"zh": resultZH,
		"en": resultEN,
	}
	processed, err := json.Marshal(processedByLang)
	if err != nil {
		writeAPIError(ctx, http.StatusInternalServerError, "结果序列化失败")
		return
	}

	record, err := s.dscans.CreateShip(ctx.Request.Context(), dscan.CreateInput{
		RawData:       rawData,
		ProcessedData: processed,
		ClientIP:      ctx.ClientIP(),
	})
	if err != nil {
		s.logger.Warn("create ship dscan", "error", err)
		writeAPIError(ctx, http.StatusInternalServerError, "保存舰船扫描失败")
		return
	}

	_ = s.cache.SetShip(ctx.Request.Context(), "zh", cachedShip(record, resultZH, "zh"))
	_ = s.cache.SetShip(ctx.Request.Context(), "en", cachedShip(record, resultEN, "en"))
	writeAPISuccess(ctx, http.StatusCreated, createdDScanResponse{
		Type:    "ship",
		ShortID: record.ShortID,
		ViewURL: "/v/" + record.ShortID,
	})
}

func (s *Server) handleViewLocalDScan(ctx *gin.Context) {
	shortID := ctx.Param("short_id")
	record, err := s.dscans.Repository().IncrementLocalViewCount(ctx.Request.Context(), shortID)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			writeAPIError(ctx, http.StatusNotFound, "找不到指定的本地扫描数据")
			return
		}
		writeAPIError(ctx, http.StatusInternalServerError, "读取本地扫描失败")
		return
	}

	result := cachedLocal(record)
	_ = s.cache.SetLocal(ctx.Request.Context(), result)
	writeAPISuccess(ctx, http.StatusOK, result)
}

func (s *Server) handleViewShipDScan(ctx *gin.Context) {
	shortID := ctx.Param("short_id")
	gameLang := normalizeGameLang(firstNonEmpty(ctx.Query("game_lang"), ctx.Query("lang"), readCookie(ctx, "game_lang"), readCookie(ctx, "lang")))
	record, err := s.dscans.Repository().IncrementShipViewCount(ctx.Request.Context(), shortID)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			writeAPIError(ctx, http.StatusNotFound, "找不到指定的舰船扫描数据")
			return
		}
		writeAPIError(ctx, http.StatusInternalServerError, "读取舰船扫描失败")
		return
	}

	result, err := cachedShipForLang(record, gameLang)
	if err != nil {
		writeAPIError(ctx, http.StatusInternalServerError, "读取舰船扫描失败")
		return
	}
	_ = s.cache.SetShip(ctx.Request.Context(), gameLang, result)
	writeAPISuccess(ctx, http.StatusOK, result)
}

func bindRequest(ctx *gin.Context, dest any) error {
	contentType := ctx.GetHeader("Content-Type")
	if strings.Contains(contentType, "application/json") {
		return ctx.ShouldBindJSON(dest)
	}
	return ctx.ShouldBind(dest)
}

func writeAPISuccess(ctx *gin.Context, status int, data interface{}) {
	ctx.JSON(status, apiResponse{Code: status, Msg: "成功", Data: data})
}

func writeAPIError(ctx *gin.Context, status int, msg string) {
	ctx.JSON(status, apiResponse{Code: status, Msg: msg, Data: gin.H{}})
}

func cachedLocal(record *model.LocalDScan) dscan.CachedResult {
	return dscan.CachedResult{
		ID:            record.ID,
		ShortID:       record.ShortID,
		ProcessedData: json.RawMessage(record.ProcessedData),
		ViewCount:     record.ViewCount,
		CreatedAt:     record.CreatedAt,
		TimeAgo:       formatTimeAgo(record.CreatedAt),
	}
}

func cachedShip(record *model.ShipDScan, result *dscan.ShipResult, lang string) dscan.CachedResult {
	payload, _ := json.Marshal(result)
	return dscan.CachedResult{
		ID:            record.ID,
		ShortID:       record.ShortID,
		ProcessedData: payload,
		ViewCount:     record.ViewCount,
		CreatedAt:     record.CreatedAt,
		TimeAgo:       formatTimeAgo(record.CreatedAt),
		Lang:          lang,
	}
}

func cachedShipForLang(record *model.ShipDScan, lang string) (dscan.CachedResult, error) {
	var byLang map[string]json.RawMessage
	if err := json.Unmarshal(record.ProcessedData, &byLang); err != nil {
		return dscan.CachedResult{}, err
	}

	payload := byLang[lang]
	if len(payload) == 0 {
		payload = byLang["zh"]
	}

	return dscan.CachedResult{
		ID:            record.ID,
		ShortID:       record.ShortID,
		ProcessedData: payload,
		ViewCount:     record.ViewCount,
		CreatedAt:     record.CreatedAt,
		TimeAgo:       formatTimeAgo(record.CreatedAt),
		Lang:          lang,
	}, nil
}

func formatTimeAgo(createdAt time.Time) string {
	diff := time.Since(createdAt)
	if diff < time.Minute {
		return "Recently"
	}
	if diff < time.Hour {
		return plural(int(diff.Minutes()), "min ago")
	}
	if diff < 24*time.Hour {
		return plural(int(diff.Hours()), "hours ago")
	}
	return plural(int(diff.Hours()/24), "days ago")
}

func plural(value int, suffix string) string {
	return strconv.Itoa(value) + " " + suffix
}

func detectDScanType(rawData string) string {
	lines := strings.Split(strings.TrimSpace(rawData), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if len(strings.Split(line, "\t")) != 1 {
			return "ship"
		}
	}
	return "local"
}

func appendManualShips(rawData string, ships []manualShipRequest) string {
	var additions []string
	for _, ship := range ships {
		recon, ok := reconShips[strings.ToLower(ship.Key)]
		if !ok || ship.Count <= 0 {
			continue
		}
		for i := 0; i < ship.Count; i++ {
			additions = append(additions, recon.line())
		}
	}
	if len(additions) == 0 {
		return rawData
	}
	if strings.TrimSpace(rawData) != "" && !strings.HasSuffix(rawData, "\n") {
		rawData += "\n"
	}
	return rawData + strings.Join(additions, "\n")
}

func (r *submitDScanRequest) applyLegacyForm(ctx *gin.Context) {
	if len(r.ManualShips) > 0 {
		return
	}
	for key := range reconShips {
		count := parseFormInt(ctx, key+"_count")
		if count > 0 {
			r.ManualShips = append(r.ManualShips, manualShipRequest{Key: key, Count: count})
		}
	}
}

func parseFormInt(ctx *gin.Context, key string) int {
	value := strings.TrimSpace(ctx.PostForm(key))
	if value == "" {
		return 0
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return n
}

func normalizeGameLang(lang string) string {
	if lang == "en" {
		return "en"
	}
	return "zh"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func readCookie(ctx *gin.Context, name string) string {
	value, err := ctx.Cookie(name)
	if err != nil {
		return ""
	}
	return value
}

type reconShip struct {
	TypeID  int64
	ZHName  string
	Display string
}

func (r reconShip) line() string {
	return strconv.FormatInt(r.TypeID, 10) + "\t" + r.ZHName + "\t" + r.Display + "\t1km"
}

var reconShips = map[string]reconShip{
	"huginn":   {TypeID: 11961, ZHName: "休津级", Display: "Huginn"},
	"lachesis": {TypeID: 11971, ZHName: "拉克希斯级", Display: "Lachesis"},
	"rook":     {TypeID: 11959, ZHName: "白嘴鸦级", Display: "Rook"},
	"curse":    {TypeID: 20125, ZHName: "诅咒级", Display: "Curse"},
}
