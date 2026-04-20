import { Link, useParams } from 'react-router-dom';
import { useCourseContent } from '../context/CourseContentContext';
import Card from '../components/Card';
import Button from '../components/Button';
import ProgressBar from '../components/ProgressBar';
import CourseDetailsPanel from '../components/CourseDetailsPanel';

export default function CourseView() {
  const { id: courseSlug } = useParams();
  const { getCourseBySlug, getCourseProgress, isLessonCompleted } = useCourseContent();

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
            <h1 className="text-3xl font-bold text-slate-900">{course.title}</h1>
            <p className="mt-2 text-slate-600">{course.description}</p>
          </div>
          <div className="flex gap-2">
            <Link
              to="/dashboard"
              className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700"
            >
              Dashboard
            </Link>
            <Link to="/courses" className="rounded-lg bg-sky-600 px-3 py-2 text-sm font-medium text-white">
              All Courses
            </Link>
          </div>
        </div>

        <Card className="mb-4">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <p className="text-sm font-medium text-slate-700">Progress</p>
              <p className="text-sm text-slate-600 mt-1">
                {completedLessons}/{totalLessons} lessons completed ({completionPercentage}%)
              </p>
            </div>
            <Link
              to={`/course/${course.slug}/progress`}
              className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700"
            >
              Open Progress Details
            </Link>
          </div>
          <ProgressBar className="mt-3" value={completionPercentage} />
        </Card>

        <div className="mb-4">
          <CourseDetailsPanel course={course} />
        </div>

        <div className="space-y-4">
          {lessons.map((lesson) => {
            const isDone = isLessonCompleted(course.slug, lesson.slug);
            return (
              <Card key={lesson.id}>
                <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                  <div>
                    <div className="flex items-center gap-2">
                      <h3 className="font-semibold text-slate-900">{lesson.title}</h3>
                      {isDone && (
                        <span className="rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-medium text-emerald-800">
                          Completed
                        </span>
                      )}
                    </div>
                    <p className="mt-2 text-slate-600">
                      {lesson.excerpt}{lesson.excerpt?.length >= 180 ? '...' : ''}
                    </p>
                  </div>
                  <Link
                    to={`/lesson/${lesson.slug}?course=${course.slug}`}
                    className="inline-flex"
                  >
                    <Button variant="warning">Open Lesson</Button>
                  </Link>
                </div>
              </Card>
            );
          })}

          {!lessons.length && (
            <div className="rounded-2xl border border-dashed border-slate-300 bg-white p-8 text-center text-slate-500">
              This course has no lessons yet.
            </div>
          )}
        </div>
      </div>
    </div>
  );
}