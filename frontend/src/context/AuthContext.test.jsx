import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { AuthProvider, useAuth } from './AuthContext';

vi.mock('../services/api', () => ({
  api: {
    auth: {
      login: vi.fn(),
      register: vi.fn(),
    },
  },
}));

function AuthProbe() {
  const { user, isAuthenticated, refreshAuthState } = useAuth();

  return (
    <div>
      <div data-testid="authed">{String(isAuthenticated)}</div>
      <div data-testid="email">{user?.email || ''}</div>
      <button type="button" onClick={refreshAuthState}>Refresh auth</button>
    </div>
  );
}

describe('AuthContext refreshAuthState', () => {
  it('refreshes from localStorage when token and email are added', async () => {
    render(
      <AuthProvider>
        <AuthProbe />
      </AuthProvider>
    );

    expect(screen.getByTestId('authed')).toHaveTextContent('false');

    localStorage.setItem('token', 'token-value');
    localStorage.setItem('user_email', 'student@example.com');

    fireEvent.click(screen.getByRole('button', { name: 'Refresh auth' }));

    await waitFor(() => {
      expect(screen.getByTestId('authed')).toHaveTextContent('true');
      expect(screen.getByTestId('email')).toHaveTextContent('student@example.com');
    });
  });

  it('clears auth state when token is removed and refreshAuthState is called', async () => {
    localStorage.setItem('token', 'token-value');
    localStorage.setItem('user_email', 'student@example.com');

    render(
      <AuthProvider>
        <AuthProbe />
      </AuthProvider>
    );

    await waitFor(() => {
      expect(screen.getByTestId('authed')).toHaveTextContent('true');
      expect(screen.getByTestId('email')).toHaveTextContent('student@example.com');
    });

    localStorage.removeItem('token');
    localStorage.removeItem('user_email');

    fireEvent.click(screen.getByRole('button', { name: 'Refresh auth' }));

    await waitFor(() => {
      expect(screen.getByTestId('authed')).toHaveTextContent('false');
      expect(screen.getByTestId('email')).toHaveTextContent('');
    });
  });
});
