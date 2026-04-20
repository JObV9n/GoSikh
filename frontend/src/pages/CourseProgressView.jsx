import { Link, useParams } from 'react-router-dom';
import { useCourseContent } from '../context/CourseContentContext';
import Card from '../components/Card';
import ProgressBar from '../components/ProgressBar';
import Button from '../components/Button';

export default function CourseProgressView() {
  const { id: courseSlug } = useParams();
  const {
    getCourseBySlug,
    getCourseProgress,
    isLessonCompleted,
    markLessonComplete,
  } = useCourseContent();

  const course = getCourseBySlug(courseSlug);
  const lessons = course?.lessons || [];

  if (!course) {
    return (
      <div className="min-h-screen grid place-items-center">
        <p className="text-slate-600">Course not found.</p>
      </div>
    );
  }

  const { completedLessons, totalLessons, completionPercentage } = getCourseProgress(courseSlug);

  return (
    <div className="min-h-screen px-4 py-8">
      <div className="mx-auto max-w-6xl">
        <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h1 className="text-3xl font-bold text-slate-900">{course.title} Progress</h1>
            <p className="mt-2 text-slate-600">{completedLessons}/{totalLessons} lessons complete ({completionPercentage}%)</p>
          </div>
          <Link
            to={`/course/${course.slug}`}
            className="rounded-lg border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700"
          >
            Back to Course
          </Link>
        </div>

        <ProgressBar className="mb-6 h-3" value={completionPercentage} />

        <div className="space-y-4">
          {lessons.map((lesson) => {
            const completed = isLessonCompleted(courseSlug, lesson.slug);
            return (
              <Card key={lesson.id}>
                <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                  <div>
                    <h3 className="font-semibold text-slate-900">{lesson.title}</h3>
                    <p className="text-slate-600 mt-1">
                      {lesson.excerpt}{lesson.excerpt?.length >= 180 ? '...' : ''}
                    </p>
                  </div>
                  <div className="flex items-center gap-2">
                    {completed ? (
                      <span className="rounded-full bg-emerald-100 px-3 py-1 text-xs font-medium text-emerald-800">
                        Completed
                      </span>
                    ) : (
                      <Button
                        onClick={() => markLessonComplete(courseSlug, lesson.slug, true)}
                        className="px-3 py-1.5"
                      >
                        Mark Complete
                      </Button>
                    )}
                    <Link
                      to={`/lesson/${lesson.slug}?course=${courseSlug}`}
                      className="rounded-lg border border-slate-300 px-3 py-1.5 text-sm font-medium text-slate-700"
                    >
                      Open
                    </Link>
                  </div>
                </div>
              </Card>
            );
          })}

          {!lessons.length && (
            <div className="rounded-2xl border border-dashed border-slate-300 bg-white p-8 text-center text-slate-500">
              No lessons available yet.
            </div>
          )}
        </div>
      </div>
    </div>
  );
}