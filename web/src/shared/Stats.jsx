export function Stats({ items }) {
  return (
    <div className="mb-4 grid grid-cols-2 gap-3 md:grid-cols-4">
      {items.map(([label, value]) => (
        <div className="rounded bg-gray-100 p-3 text-center dark:bg-gray-800" key={label}>
          <div className="text-xl font-bold text-primary-600 dark:text-primary-300">{value}</div>
          <div className="text-xs text-gray-500">{label}</div>
        </div>
      ))}
    </div>
  )
}
