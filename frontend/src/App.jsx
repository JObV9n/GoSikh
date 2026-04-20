import { lazy, Suspense } from 'react';
import { Routes, Route, Navigate, Outlet } from 'react-router-dom';
import { useAuth } from './context/AuthContext';
import Loader from './components/Loader';
import GlobalApiErrorBanner from './components/GlobalApiErrorBanner';
import ErrorBoundary from './components/ErrorBoundary';

const Login = lazy(() => import('./pages/Login'));
const Register = lazy(() => import('./pages/Register'));
const Dashboard = lazy(() => import('./pages/Dashboard'));
const Courses = lazy(() => import('./pages/Courses'));
const CourseView = lazy(() => import('./pages/CourseView'));
const LessonView = lazy(() => import('./pages/LessonView'));
const Playground = lazy(() => import('./pages/Playground'));
const Profile = lazy(() => import('./pages/Profile'));
const CourseProgressView = lazy(() => import('./pages/CourseProgressView'));

const guestRoutes = [
  { path: '/login', Component: Login },
  { path: '/register', Component: Register },
];

const protectedRoutes = [
  { path: '/dashboard', Component: Dashboard },
  { path: '/courses', Component: Courses },
  { path: '/course/:id', Component: CourseView },
  { path: '/course/:id/progress', Component: CourseProgressView },
  { path: '/lesson/:id', Component: LessonView },
  { path: '/playground', Component: Playground },
  { path: '/profile', Component: Profile },
];

function ProtectedLayout({ isAuthenticated }) {
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  return <Outlet />;
}

function GuestLayout({ isAuthenticated }) {
  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />;
  }
  return <Outlet />;
}

function RouteLoadError({ onRetry }) {
  return (
    <div className="grid place-items-center px-6 py-12 text-center">
      <div>
        <p className="text-xs font-semibold tracking-[0.2em] uppercase text-red-700">Page Error</p>
        <h2 className="text-2xl font-bold text-slate-900 mt-2">This page failed to load</h2>
        <p className="mt-2 text-sm text-slate-600">Please try again. If the issue persists, reload the app.</p>
        <button
          type="button"
          onClick={onRetry}
          className="mt-5 rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700"
        >
          Try again
        </button>
      </div>
    </div>
  );
}

function RoutedPage({ Component }) {
  return (
    <ErrorBoundary
      fullScreen={false}
      fallback={({ retry }) => (
        <RouteLoadError onRetry={retry} />
      )}
    >
      <Component />
    </ErrorBoundary>
  );
}

function App() {
  const { isAuthenticated, loading } = useAuth();

  if (loading) {
    return (
      <div className="min-h-screen grid place-items-center">
        <p className="text-slate-600">Loading platform...</p>
      </div>
    );
  }

  return (
    <>
      <GlobalApiErrorBanner />
      <Suspense fallback={<Loader text="Loading page..." />}>
        <Routes>
          <Route path="/" element={<Navigate to={isAuthenticated ? '/dashboard' : '/login'} replace />} />
          <Route element={<GuestLayout isAuthenticated={isAuthenticated} />}>
            {guestRoutes.map(({ path, Component }) => (
              <Route
                key={path}
                path={path}
                element={<RoutedPage Component={Component} />}
              />
            ))}
          </Route>
          <Route element={<ProtectedLayout isAuthenticated={isAuthenticated} />}>
            {protectedRoutes.map(({ path, Component }) => (
              <Route
                key={path}
                path={path}
                element={<RoutedPage Component={Component} />}
              />
            ))}
          </Route>
          <Route
            path="*"
            element={
              <div className="min-h-screen grid place-items-center px-6 text-center">
                <div>
                  <p className="text-xs font-semibold tracking-[0.25em] uppercase text-slate-500">404</p>
                  <h1 className="text-3xl font-bold text-slate-900 mt-2">Page not found</h1>
                </div>
              </div>
            }
          />
        </Routes>
      </Suspense>
    </>
  );
}

export default App;