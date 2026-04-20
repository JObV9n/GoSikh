import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { courseCatalog } from './courses/catalog';

const STORAGE_KEY = 'learner-course-progress-v1';
const CourseContentContext = createContext(null);

function slugify(value) {
  return value
    .toLowerCase()
    .replace(/[^a-z0-9\s-]/g, '')
    .trim()
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-');
}

function plainText(value) {
  return value
    .replace(/```[\s\S]*?```/g, ' ')
    .replace(/`([^`]+)`/g, '$1')
    .replace(/\*\*([^*]+)\*\*/g, '$1')
    .replace(/\*([^*]+)\*/g, '$1')
    .replace(/\[[^\]]+\]\([^\)]+\)/g, ' ')
    .replace(/!\[[^\]]*\]\([^\)]+\)/g, ' ')
    .replace(/[#>-]/g, ' ')
    .replace(/\s+/g, ' ')
    .trim();
}

function parseLessons(markdown) {
  const lines = markdown.split('\n');
  const sections = [];
  let current = null;

  lines.forEach((line) => {
    const trimmed = line.trimStart();
    const headingMatch = trimmed.match(/^#\s+(.+)$/);
    if (headingMatch) {
      if (current) {
        sections.push(current);
      }

      current = {
        title: headingMatch[1].trim(),
        lines: [],
      };
      return;
    }

    if (!current) {
      if (trimmed.length > 0) {
        current = {
          title: trimmed,
          lines: [],
        };
        return;
      }

      current = {
        title: 'Introduction',
        lines: [],
      };
    }

    current.lines.push(line);
  });

  if (current) {
    sections.push(current);
  }

  return sections
    .map((section, index) => {
      const content = section.lines.join('\n').trim();
      if (!content) {
        return null;
      }

      const clean = plainText(content);
      return {
        id: `${index + 1}`,
        order: index + 1,
        title: section.title,
        slug: `${slugify(section.title)}-${index + 1}`,
        content,
        excerpt: clean.slice(0, 180),
      };
    })
    .filter(Boolean);
}

function buildCourse(course) {
  const lessons = parseLessons(course.markdown);
  return {
    ...course,
    lessons,
    totalLessons: lessons.length,
  };
}

function readStoredProgress() {
  if (typeof window === 'undefined') {
    return {};
  }

  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) {
      return {};
    }

    const parsed = JSON.parse(raw);
    return parsed && typeof parsed === 'object' ? parsed : {};
  } catch {
    return {};
  }
}

export function CourseContentProvider({ children }) {
  const [progress, setProgress] = useState(() => readStoredProgress());

  const courses = useMemo(() => courseCatalog.map(buildCourse), []);
  const courseMap = useMemo(() => {
    return courses.reduce((acc, course) => {
      acc[course.slug] = course;
      return acc;
    }, {});
  }, [courses]);

  useEffect(() => {
    if (typeof window === 'undefined') {
      return;
    }

    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(progress));
  }, [progress]);

  const getCourseBySlug = useCallback(
    (slug) => {
      return courseMap[slug] || null;
    },
    [courseMap]
  );

  const getLessonBySlug = useCallback(
    (courseSlug, lessonSlug) => {
      const course = courseMap[courseSlug];
      if (!course) {
        return null;
      }

      return course.lessons.find((lesson) => lesson.slug === lessonSlug) || null;
    },
    [courseMap]
  );

  const isLessonCompleted = useCallback(
    (courseSlug, lessonSlug) => {
      return Boolean(progress[courseSlug]?.[lessonSlug]);
    },
    [progress]
  );

  const markLessonComplete = useCallback((courseSlug, lessonSlug, completed = true) => {
    setProgress((prev) => {
      const nextCourseProgress = {
        ...(prev[courseSlug] || {}),
      };

      if (completed) {
        nextCourseProgress[lessonSlug] = true;
      } else {
        delete nextCourseProgress[lessonSlug];
      }

      return {
        ...prev,
        [courseSlug]: nextCourseProgress,
      };
    });
  }, []);

  const getCourseProgress = useCallback(
    (courseSlug) => {
      const course = courseMap[courseSlug];
      if (!course) {
        return {
          completedLessons: 0,
          totalLessons: 0,
          completionPercentage: 0,
        };
      }

      const completedLessons = course.lessons.filter((lesson) => {
        return Boolean(progress[courseSlug]?.[lesson.slug]);
      }).length;

      const totalLessons = course.lessons.length;
      const completionPercentage = totalLessons
        ? Math.round((completedLessons / totalLessons) * 100)
        : 0;

      return {
        completedLessons,
        totalLessons,
        completionPercentage,
      };
    },
    [courseMap, progress]
  );

  const value = useMemo(
    () => ({
      courses,
      getCourseBySlug,
      getLessonBySlug,
      isLessonCompleted,
      markLessonComplete,
      getCourseProgress,
    }),
    [
      courses,
      getCourseBySlug,
      getLessonBySlug,
      isLessonCompleted,
      markLessonComplete,
      getCourseProgress,
    ]
  );

  return <CourseContentContext.Provider value={value}>{children}</CourseContentContext.Provider>;
}

export function useCourseContent() {
  const value = useContext(CourseContentContext);
  if (!value) {
    throw new Error('useCourseContent must be used inside CourseContentProvider');
  }

  return value;
}
