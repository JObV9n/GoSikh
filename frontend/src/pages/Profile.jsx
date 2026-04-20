import { Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export default function Profile() {
  const { user, logout } = useAuth();

  return (
    <div className="min-h-screen px-4 py-8">
      <div className="mx-auto max-w-2xl rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <div className="flex items-start justify-between gap-4">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">Profile</p>
            <h1 className="mt-2 text-2xl font-bold text-slate-900">{user?.email || 'Learner'}</h1>
          </div>
          <button
            onClick={logout}
            className="rounded-lg border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700"
          >
            Logout
          </button>
        </div>

        <div className="mt-6 grid grid-cols-1 gap-3 sm:grid-cols-2">
          <Link to="/dashboard" className="rounded-lg bg-sky-600 px-4 py-2 text-center text-sm font-medium text-white">
            Dashboard
          </Link>
          <Link to="/courses" className="rounded-lg border border-slate-300 px-4 py-2 text-center text-sm font-medium text-slate-700">
            Courses
          </Link>
        </div>
      </div>
    </div>
  );
}
