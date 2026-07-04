import { useState } from "react";
import { useAsyncData } from "../shared/hooks.js";
import { Loading, ErrorBox } from "../shared/Loading.jsx";
import ResultShell from "../shared/ResultShell.jsx";
import { fetchShip } from "../api.js";

export default function ShipScanResult({ t, shortId, gameLang, navigate }) {
  const { data, error, loading } = useAsyncData(
    () => fetchShip(shortId, gameLang),
    [shortId, gameLang],
  );
  if (loading) return <Loading t={t} />;
  if (error) return <ErrorBox t={t} error={error} navigate={navigate} />;

  const processed = data.processed_data;
  const system = processed.system_info || {};

  return (
    <ResultShell t={t} createdAt={data.created_at} onBack={() => navigate("/")}>
      <div className="mx-auto grid max-w-6xl grid-cols-1 gap-4 md:grid-cols-8">
        <div className="md:col-span-5">
          <EntityCategoryWrapper
            groups={processed.ship_types}
            typeLabel={t.ships || "类型"}
            nameLabel=""
          />
          <EntityCategoryWrapper
            groups={processed.structure_types}
            typeLabel={t.structures || "建筑类型"}
            nameLabel=""
          />
          <EntityCategoryWrapper
            groups={processed.misc_types}
            typeLabel={t.other || "Other"}
            nameLabel=""
          />
        </div>

        <div className="md:col-span-3">
          {system.name && (
            <div className="mb-3">
              <span className="inline-flex items-center rounded-full bg-primary-100 px-3 py-1 text-sm font-semibold text-primary-800 dark:bg-primary-900/30 dark:text-primary-300">
                {system.name}
                {system.security ? ` (${system.security})` : ""}
                {system.regionName ? ` « ${system.regionName}` : ""}
              </span>
            </div>
          )}

          {processed.filter_distance && (
            <div className="mb-2 text-sm text-gray-600 dark:text-gray-400">
              <span className="rounded-md bg-blue-100 px-2 py-1 dark:bg-blue-900">
                {t.only_have_distance || "仅显示有距离信息的扫描结果"}
              </span>
            </div>
          )}

          <div className="mb-6 grid grid-cols-2 gap-4 md:grid-cols-4">
            <div className="stat-card">
              <div className="stat-value">
                {processed.stats?.total_count || 0}
              </div>
              <div className="stat-label">{t.total}</div>
            </div>
            <div className="stat-card">
              <div className="stat-value">
                {processed.stats?.ship_count || 0}
              </div>
              <div className="stat-label">{t.ships}</div>
            </div>
            <div className="stat-card">
              <div className="stat-value">
                {processed.stats?.capital_count || 0}
              </div>
              <div className="stat-label">{t.capitals}</div>
            </div>
            <div className="stat-card">
              <div className="stat-value">
                {processed.stats?.structure_count || 0}
              </div>
              <div className="stat-label">{t.structures}</div>
            </div>
          </div>

          <EntityCategoryWrapper
            groups={processed.capital_types}
            typeLabel={t.capitals || "旗舰"}
            nameLabel=""
            small
          />
        </div>
      </div>
    </ResultShell>
  );
}

function hasEntries(obj) {
  return obj && typeof obj === "object" && Object.keys(obj).length > 0;
}

function computeTypeCounts(groups) {
  const counts = {};
  Object.entries(groups || {}).forEach(([type, items]) => {
    counts[type] = items.reduce((sum, item) => sum + (item.count || 0), 0);
  });
  return counts;
}

function EntityCategoryWrapper({ groups, typeLabel, nameLabel, small }) {
  const [hoveredType, setHoveredType] = useState(null);
  if (!hasEntries(groups)) return null;

  const typeCounts = computeTypeCounts(groups);

  const sortedTypes = Object.entries(groups).sort(
    (a, b) => (typeCounts[b[0]] || 0) - (typeCounts[a[0]] || 0),
  );

  const allItems = sortedTypes
    .flatMap(([typeName, items]) =>
      items.map((item) => ({ ...item, typeName })),
    )
    .sort((a, b) => (b.count || 0) - (a.count || 0));

  const textSize = small ? "text-xs" : "";

  return (
    <div className={`mb-4 ${textSize} entity-column`}>
      <div className="entity-header">{typeLabel}</div>
      <div className="grid grid-cols-2 divide-x divide-gray-300 dark:divide-gray-600">
        <div>
          {sortedTypes.map(([typeName]) => (
            <div
              key={typeName}
              className={`entity-item ${hoveredType === typeName ? "highlighted" : ""}`}
              onMouseEnter={() => setHoveredType(typeName)}
              onMouseLeave={() => setHoveredType(null)}
            >
              <span className="entity-name">{typeName}</span>
              <span className="entity-distance">{typeCounts[typeName]}</span>
            </div>
          ))}
        </div>

        <div>
          {allItems.map((item, idx) => {
            const hidden =
              hoveredType !== null && hoveredType !== item.typeName;
            return (
              <div
                key={`${item.typeName}-${item.name}-${idx}`}
                className={`entity-item `}
                style={hidden ? { display: "none" } : undefined}
              >
                <img src={`https://images.newdoublex.space/types/${item.id}/render`} alt={item.name} className="entity-img" />
                <span className="entity-name">{item.name}</span>
                <span className="entity-distance">{item.count}</span>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
