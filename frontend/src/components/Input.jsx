import { useId } from 'react';

export default function Input({ label, id, className = '', ...props }) {
  const generatedId = useId();
  const inputId = id || generatedId;

  return (
    <div>
      {label && (
        <label htmlFor={inputId} className="mb-2 block text-sm font-medium text-slate-700">
          {label}
        </label>
      )}
      <input
        id={inputId}
        className={`w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-sky-500 ${className}`.trim()}
        {...props}
      />
    </div>
  );
}
