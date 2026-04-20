export default function ProgressBar({ value = 0, className = '' }) {
  const pct = Math.max(0, Math.min(100, Math.round(value)));

  return (
    <div className={`h-2 w-full overflow-hidden rounded-full bg-slate-200 ${className}`.trim()}>
      <div className="h-full rounded-full bg-emerald-500" style={{ width: `${pct}%` }} />
    </div>
  );
}
