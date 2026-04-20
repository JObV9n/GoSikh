import Card from './Card';

export default function CourseDetailsPanel({ course }) {
  if (!course) {
    return null;
  }

  return (
    <Card>
      <h2 className="text-xl font-semibold text-slate-900">About This Course</h2>
      <div className="mt-4 grid gap-3 text-sm text-slate-700 sm:grid-cols-2">
        <p>
          <span className="font-semibold text-slate-900">Level:</span> {course.level}
        </p>
        <p>
          <span className="font-semibold text-slate-900">Language:</span> {course.language}
        </p>
        <p>
          <span className="font-semibold text-slate-900">Estimated Time:</span> {course.estimatedHours} hours
        </p>
        <p>
          <span className="font-semibold text-slate-900">Lessons:</span> {course.totalLessons}
        </p>
      </div>

      <div className="mt-4">
        <p className="text-sm font-semibold text-slate-900">Prerequisites</p>
        <ul className="mt-2 list-disc space-y-1 pl-5 text-sm text-slate-700">
          {(course.prerequisites || []).map((item) => (
            <li key={item}>{item}</li>
          ))}
        </ul>
      </div>

      <div className="mt-4">
        <p className="text-sm font-semibold text-slate-900">Learning Outcomes</p>
        <ul className="mt-2 list-disc space-y-1 pl-5 text-sm text-slate-700">
          {(course.outcomes || []).map((item) => (
            <li key={item}>{item}</li>
          ))}
        </ul>
      </div>

      <div className="mt-4 flex flex-wrap gap-2">
        {(course.tags || []).map((tag) => (
          <span
            key={tag}
            className="rounded-full border border-slate-300 bg-slate-100 px-3 py-1 text-xs font-medium text-slate-700"
          >
            {tag}
          </span>
        ))}
      </div>
    </Card>
  );
}
