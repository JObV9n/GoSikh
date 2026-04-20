export default function Button({
  children,
  type = 'button',
  variant = 'primary',
  className = '',
  disabled = false,
  ...props
}) {
  const safeType = ['button', 'submit', 'reset'].includes(type) ? type : 'button';
  const base = 'rounded-lg px-4 py-2 text-sm font-medium transition disabled:opacity-60 disabled:cursor-not-allowed';

  const variants = {
    primary: 'bg-sky-600 text-white hover:bg-sky-700',
    secondary: 'border border-slate-300 text-slate-700 hover:bg-slate-100',
    success: 'bg-emerald-600 text-white hover:bg-emerald-700',
    warning: 'bg-amber-500 text-white hover:bg-amber-600',
    ghost: 'text-slate-700 hover:bg-slate-100',
  };

  return (
    <button
      type={safeType}
      disabled={disabled}
      className={`${base} ${variants[variant] || variants.primary} ${className}`.trim()}
      {...props}
    >
      {children}
    </button>
  );
}
