import React from 'react';

export default class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error) {
    return { hasError: true, error };
  }

  componentDidCatch(error) {
    // Keep logging for debugging while presenting a safe fallback to users.
    console.error('Unhandled render error:', error);
  }

  handleReload = () => {
    window.location.reload();
  };

  handleRetry = () => {
    this.setState({ hasError: false, error: null });
  };

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        if (typeof this.props.fallback === 'function') {
          return this.props.fallback({
            error: this.state.error,
            retry: this.handleRetry,
          });
        }

        return this.props.fallback;
      }

      const wrapperClass = this.props.fullScreen === false
        ? 'grid place-items-center px-4 py-8'
        : 'min-h-screen grid place-items-center px-4';

      return (
        <div className={wrapperClass}>
          <div className="w-full max-w-lg rounded-2xl border border-red-200 bg-white p-6 text-center shadow-sm">
            <p className="text-xs font-semibold uppercase tracking-[0.2em] text-red-700">Something went wrong</p>
            <h1 className="mt-2 text-2xl font-bold text-slate-900">Application Error</h1>
            <p className="mt-3 text-sm text-slate-600">
              An unexpected issue occurred while rendering the page.
            </p>
            <button
              onClick={this.handleReload}
              className="mt-5 rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700"
            >
              Reload
            </button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}