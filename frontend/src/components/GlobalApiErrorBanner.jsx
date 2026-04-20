import { useEffect, useState } from 'react';

export default function GlobalApiErrorBanner() {
  const [error, setError] = useState(null);

  useEffect(() => {
    const onApiError = (event) => {
      const nextMessage = event?.detail?.message || 'Request failed';
      setError(nextMessage);
    };

    window.addEventListener('app:api-error', onApiError);
    return () => window.removeEventListener('app:api-error', onApiError);
  }, []);

  useEffect(() => {
    if (!error) return;
    const timerId = setTimeout(() => setError(null), 5000);
    return () => clearTimeout(timerId);
  }, [error]);

  if (!error) return null;

  return (
    <div className="sticky top-0 z-50 px-4 pt-3">
      <div className="mx-auto max-w-6xl rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 shadow-sm">
        <div className="flex items-center justify-between gap-3">
          <p>{error}</p>
          <button
            onClick={() => setError(null)}
            className="rounded border border-red-300 px-2 py-1 text-xs font-medium text-red-700 hover:bg-red-100"
          >
            Dismiss
          </button>
        </div>
      </div>
    </div>
  );
}