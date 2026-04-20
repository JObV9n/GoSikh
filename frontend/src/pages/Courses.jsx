import { Link } from 'react-router-dom';
import { useCourseContent } from '../context/CourseContentContext';
import Card from '../components/Card';
import ProgressBar from '../components/ProgressBar';

export default function Courses() {
  const { courses, getCourseProgress } = useCourseContent();

  return (
    <div className="min-h-screen px-4 py-8">
      <div className="mx-auto max-w-6xl">
        <div className="mb-6 flex items-center justify-between gap-3">
          <h1 className="text-3xl font-bold text-slate-900">Courses</h1>
          <Link
            to="/dashboard"
            className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700"
          >
            Dashboard
          </Link>
        </div>

        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          {courses.map((course) => (
            <Card key={course.id}>
              <h2 className="text-xl font-semibold text-slate-900">{course.title}</h2>
              <p className="mt-2 text-slate-600">{course.description}</p>
              <ProgressBar className="mt-3" value={getCourseProgress(course.slug).completionPercentage} />
              <p className="mt-1 text-xs text-slate-500">{Math.round(getCourseProgress(course.slug).completionPercentage)}% complete</p>
              <Link
                to={`/course/${course.slug}`}
                className="mt-4 inline-flex rounded-lg bg-sky-600 px-4 py-2 text-sm font-medium text-white"
              >
                Open Course
              </Link>
            </Card>
          ))}

          {!courses.length && (
            <div className="rounded-2xl border border-dashed border-slate-300 bg-white p-8 text-center text-slate-500 md:col-span-2">
              No courses found yet.
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
