import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { api } from '../services/api';
import { getToken, removeToken } from '../utils/auth';

export const AuthContext = createContext();

export const useAuth = () => {
  return useContext(AuthContext);
};

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [authed, setAuthed] = useState(false);
  const [loading, setLoading] = useState(true);

  const refreshAuthState = useCallback(() => {
    const token = getToken();
    const hasToken = Boolean(token);

    setAuthed(hasToken);

    if (!hasToken) {
      setUser(null);
      return false;
    }

    const savedEmail = localStorage.getItem('user_email');
    setUser((prevUser) => {
      if (prevUser) {
        return prevUser;
      }

      if (savedEmail) {
        return { email: savedEmail };
      }

      return null;
    });

    return true;
  }, []);

  useEffect(() => {
    refreshAuthState();
    setLoading(false);
  }, [refreshAuthState]);

  useEffect(() => {
    const handleStorageChange = (event) => {
      if (!event.key || event.key === 'token' || event.key === 'user_email') {
        refreshAuthState();
      }
    };

    window.addEventListener('storage', handleStorageChange);

    return () => {
      window.removeEventListener('storage', handleStorageChange);
    };
  }, [refreshAuthState]);

  const login = async (email, password) => {
    const data = await api.auth.login(email, password);

    if (!data?.token) {
      throw new Error('Authentication token not returned by server.');
    }

    setAuthed(true);

    const nextUser = data?.user || { email };
    setUser(nextUser);
    if (nextUser.email) {
      localStorage.setItem('user_email', nextUser.email);
    }

    return nextUser;
  };

  const register = async (email, password) => {
    return api.auth.register(email, password);
  };

  const logout = () => {
    removeToken();
    localStorage.removeItem('user_email');
    setUser(null);
    setAuthed(false);
  };

  const value = useMemo(() => ({
    user,
    login,
    register,
    logout,
    refreshAuthState,
    isAuthenticated: authed,
    loading,
  }), [authed, loading, refreshAuthState, user]);

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};