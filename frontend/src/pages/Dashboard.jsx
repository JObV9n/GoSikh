import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { useCourseContent } from '../context/CourseContentContext';
import Card from '../components/Card';
import Button from '../components/Button';
import ProgressBar from '../components/ProgressBar';

export default function Dashboard() {
  const { logout } = useAuth();
  const { courses, getCourseProgress } = useCourseContent();
  const navigate = useNavigate();

  const enrolledCourses = courses.filter((course) => getCourseProgress(course.slug).completionPercentage > 0);
  const totalCompletedCourses = courses.filter((course) => getCourseProgress(course.slug).completionPercentage >= 100).length;
  const totalProgress = courses.reduce((sum, course) => sum + getCourseProgress(course.slug).completionPercentage, 0);
  const averageProgress = courses.length ? Math.round(totalProgress / courses.length) : 0;

  const nextCourse =
    courses.find((course) => {
      const completion = getCourseProgress(course.slug).completionPercentage;
      return completion > 0 && completion < 100;
    }) ||
    courses.find((course) => getCourseProgress(course.slug).completionPercentage === 0) ||
    null;

  return (
    <div className="min-h-screen px-4 py-8">
      <div className="mx-auto max-w-6xl">
        <header className="mb-8 rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
          <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.2em] text-sky-700">Dashboard</p>
              <h1 className="text-3xl font-bold text-slate-900 mt-2">Welcome back, Learner</h1>
              <p className="text-slate-600 mt-2">Pick a course and continue your hands-on path.</p>
            </div>
            <Button
              onClick={logout}
              variant="secondary"
            >
              Logout
            </Button>
          </div>
        </header>

        <section className="mb-6 grid grid-cols-1 gap-4 md:grid-cols-3">
          <Card>
            <p className="text-xs font-semibold uppercase tracking-[0.16em] text-slate-500">Progress Summary</p>
            <p className="mt-3 text-3xl font-bold text-slate-900">{averageProgress}%</p>
            <p className="mt-1 text-sm text-slate-600">Average completion across your courses</p>
          </Card>

          <Card>
            <p className="text-xs font-semibold uppercase tracking-[0.16em] text-slate-500">Enrolled Courses</p>
            <p className="mt-3 text-3xl font-bold text-slate-900">{enrolledCourses.length}</p>
            <p className="mt-1 text-sm text-slate-600">Courses where progress has started</p>
          </Card>

          <Card>
            <p className="text-xs font-semibold uppercase tracking-[0.16em] text-slate-500">Completed Courses</p>
            <p className="mt-3 text-3xl font-bold text-slate-900">{totalCompletedCourses}</p>
            <p className="mt-1 text-sm text-slate-600">Courses finished at 100%</p>
          </Card>
        </section>

        <Card className="mb-8">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.16em] text-slate-500">Continue Learning</p>
              {nextCourse ? (
                <>
                  <h2 className="mt-2 text-xl font-semibold text-slate-900">{nextCourse.title}</h2>
                  <p className="mt-1 text-sm text-slate-600">
                    Current progress: {Math.round(getCourseProgress(nextCourse.slug).completionPercentage)}%
                  </p>
                </>
              ) : (
                <p className="mt-2 text-sm text-slate-600">No course activity yet. Pick one to start learning.</p>
              )}
            </div>

            {nextCourse && (
              <Button
                onClick={() => navigate(`/course/${nextCourse.slug}`)}
                variant="warning"
              >
                Continue
              </Button>
            )}
          </div>
        </Card>

        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          {courses.map((course) => (
            <Card key={course.id}>
              <h2 className="text-xl font-semibold text-slate-900">{course.title}</h2>
              <p className="mt-2 text-slate-600">{course.description}</p>
              <ProgressBar className="mt-3" value={getCourseProgress(course.slug).completionPercentage} />
              <p className="mt-1 text-xs text-slate-500">{Math.round(getCourseProgress(course.slug).completionPercentage)}% complete</p>
              <div className="mt-4 flex gap-3">
                <Button
                  onClick={() => navigate(`/course/${course.slug}`)}
                >
                  Open Course
                </Button>
                <Link
                  to="/courses"
                  className="rounded-lg border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100"
                >
                  Browse Courses
                </Link>
              </div>
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