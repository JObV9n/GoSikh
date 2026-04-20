import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import App from './App';

const mockUseAuth = vi.fn();

vi.mock('./context/AuthContext', () => ({
  useAuth: () => mockUseAuth(),
}));

vi.mock('./components/GlobalApiErrorBanner', () => ({
  default: () => null,
}));

vi.mock('./pages/Login', () => ({
  default: () => <div>Login Page</div>,
}));

vi.mock('./pages/Register', () => ({
  default: () => <div>Register Page</div>,
}));

vi.mock('./pages/Dashboard', () => ({
  default: () => <div>Dashboard Page</div>,
}));

vi.mock('./pages/Courses', () => ({
  default: () => <div>Courses Page</div>,
}));

vi.mock('./pages/CourseView', () => ({
  default: () => <div>Course View Page</div>,
}));

vi.mock('./pages/LessonView', () => ({
  default: () => <div>Lesson View Page</div>,
}));

vi.mock('./pages/Playground', () => ({
  default: () => <div>Playground Page</div>,
}));

vi.mock('./pages/Profile', () => ({
  default: () => <div>Profile Page</div>,
}));

vi.mock('./pages/CourseProgressView', () => ({
  default: () => <div>Course Progress Page</div>,
}));

function renderAt(path) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <App />
    </MemoryRouter>
  );
}

describe('App route guards', () => {
  beforeEach(() => {
    mockUseAuth.mockReset();
  });

  it('redirects unauthenticated users away from protected routes', async () => {
    mockUseAuth.mockReturnValue({ isAuthenticated: false, loading: false });

    renderAt('/dashboard');

    expect(await screen.findByText('Login Page')).toBeInTheDocument();
  });

  it('redirects authenticated users away from guest routes', async () => {
    mockUseAuth.mockReturnValue({ isAuthenticated: true, loading: false });

    renderAt('/login');

    expect(await screen.findByText('Dashboard Page')).toBeInTheDocument();
  });
});
