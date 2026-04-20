import { getToken, setToken, removeToken } from '../utils/auth';

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';

function emitApiError(message, path, status) {
  if (typeof window === 'undefined') return;

  window.dispatchEvent(
    new CustomEvent('app:api-error', {
      detail: { message, path, status },
    })
  );
}

async function request(path, options = {}, needsAuth = false, timeoutMs = 0) {
  const headers = {
    'Content-Type': 'application/json',
    ...(options.headers || {}),
  };

  const controller = new AbortController();
  const timeoutId = timeoutMs > 0 ? setTimeout(() => controller.abort(), timeoutMs) : null;

  if (needsAuth) {
    const token = getToken();
    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }
  }

  let response;
  try {
    response = await fetch(`${API_BASE}${path}`, {
      ...options,
      headers,
      signal: controller.signal,
    });
  } catch (err) {
    if (timeoutId) clearTimeout(timeoutId);

    const message = err?.name === 'AbortError' ? 'Request timed out. Please try again.' : 'Network error. Please check your connection.';
    emitApiError(message, path, 0);
    throw new Error(message);
  }

  if (timeoutId) clearTimeout(timeoutId);

  let payload = null;
  try {
    payload = await response.json();
  } catch {
    payload = null;
  }

  if (!response.ok) {
    const message = payload?.error?.message || payload?.error || 'Request failed';
    emitApiError(message, path, response.status);
    throw new Error(message);
  }

  return payload?.data ?? payload;
}

export const api = {
  auth: {
    register: async (email, password) => {
      return request('/auth/register', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      });
    },
    login: async (email, password) => {
      const data = await request('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      });

      if (data?.token) {
        setToken(data.token);
      }

      return data;
    },
    logout: () => {
      removeToken();
    },
  },

  courses: {
    list: async () => {
      return request('/courses');
    },
    listLessons: async (courseSlug) => {
      return request(`/courses/${courseSlug}/lessons`);
    },
    getProgress: async (courseSlug) => {
      return request(`/courses/${courseSlug}/progress`, {}, true);
    },
    markLessonProgress: async (courseSlug, lessonSlug, completed) => {
      return request(
        `/courses/${courseSlug}/lessons/${lessonSlug}/progress`,
        {
          method: 'PUT',
          body: JSON.stringify({ completed }),
        },
        true
      );
    },
  },

  progress: {
    getAll: async () => {
      return request('/progress/all', {}, true);
    },
  },

  execute: {
    runCode: async (language, code) => {
      const result = await request(
        '/execute',
        {
          method: 'POST',
          body: JSON.stringify({ language, code }),
        },
        false,
        20000
      );

      if (result?.timed_out) {
        throw new Error('Execution timed out. Try smaller input or simpler code.');
      }

      return result;
    },
    submitLesson: async (courseSlug, lessonSlug, language, code, expectedOutput) => {
      const result = await request(
        `/courses/${courseSlug}/lessons/${lessonSlug}/submit`,
        {
          method: 'POST',
          body: JSON.stringify({
            language,
            code,
            expected_output: expectedOutput || '',
          }),
        },
        true,
        30000
      );

      if (result?.timed_out) {
        throw new Error('Execution timed out. Try smaller input or simpler code.');
      }

      return result;
    },
  },
};