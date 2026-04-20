import { useEffect, useMemo, useState } from 'react';
import { Link, useLocation, useParams, useNavigate } from 'react-router-dom';
import { api } from '../services/api';
import MonacoEditor from '@monaco-editor/react';
import { useCourseContent } from '../context/CourseContentContext';
import { useAuth } from '../context/AuthContext';
import Card from '../components/Card';
import Button from '../components/Button';
import MarkdownLessonContent from '../components/MarkdownLessonContent';

function defaultStarterCode(language) {
  switch (language) {
    case 'javascript':
      return 'console.log("Hello World!")';
    case 'python':
      return 'print("Hello World!")';
    case 'go':
    default:
      return 'package main\n\nimport "fmt"\n\nfunc main() {\n  fmt.Println("Hello World!")\n}';
  }
}

function extractFenceBlock(markdown, language) {
  const languagePattern = new RegExp('```(?:' + language + ')\\s*\\n([\\s\\S]*?)```', 'i');
  const languageMatch = markdown.match(languagePattern);
  if (languageMatch?.[1]) {
    return languageMatch[1].trim();
  }

  const genericCodeMatch = markdown.match(/```\w*\s*\n([\s\S]*?)```/i);
  if (genericCodeMatch?.[1]) {
    return genericCodeMatch[1].trim();
  }

  return '';
}

function sanitizeGoStarterCode(code) {
  if (!code) {
    return '';
  }

  const lines = code.split('\n');
  const packageIndex = lines.findIndex((line) => line.trim().startsWith('package '));
  if (packageIndex > 0) {
    return lines.slice(packageIndex).join('\n').trim();
  }

  return code.trim();
}

function extractStarterCode(markdown, language) {
  const extracted = extractFenceBlock(markdown, language);
  if (!extracted) {
    return defaultStarterCode(language);
  }

  if (language === 'go') {
    const sanitized = sanitizeGoStarterCode(extracted);
    return sanitized || defaultStarterCode(language);
  }

  return extracted;
}

function extractHints(markdown) {
  const text = markdown
    .replace(/```[\s\S]*?```/g, ' ')
    .replace(/!\[[^\]]*\]\([^\)]+\)/g, ' ')
    .replace(/\[[^\]]+\]\([^\)]+\)/g, '$1')
    .replace(/\*\*([^*]+)\*\*/g, '$1')
    .replace(/`([^`]+)`/g, '$1');

  const lines = text
    .split('\n')
    .map((line) => line.replace(/^#+\s*/, '').trim())
    .filter((line) => line.length > 20)
    .filter((line) => !line.startsWith('$ '))
    .filter((line) => !line.toLowerCase().includes('congratulations, you\'ve finished the course'));

  const unique = [];
  for (const line of lines) {
    if (unique.includes(line)) {
      continue;
    }
    unique.push(line);
    if (unique.length >= 6) {
      break;
    }
  }

  return unique;
}

function normalizeOutput(value) {
  return (value || '')
    .replace(/\r\n/g, '\n')
    .split('\n')
    .map((line) => line.trimEnd())
    .join('\n')
    .trim();
}

function extractExpectedOutput(markdown) {
  const bashBlocks = markdown.match(/```bash\s*\n[\s\S]*?```/gi) || [];

  for (const block of bashBlocks) {
    const withoutFence = block.replace(/^```bash\s*\n/i, '').replace(/```$/, '');
    const lines = withoutFence.split('\n');
    const commandIndex = lines.findIndex((line) => line.trim().startsWith('$ go run '));

    if (commandIndex === -1) {
      continue;
    }

    const outputLines = [];
    for (let i = commandIndex + 1; i < lines.length; i += 1) {
      const line = lines[i];
      if (line.trim().startsWith('$ ')) {
        break;
      }
      outputLines.push(line);
    }

    const expected = normalizeOutput(outputLines.join('\n'));
    if (expected) {
      return expected;
    }
  }

  return null;
}

function extractChallengeContext(markdown) {
  if (!markdown) {
    return '';
  }

  const sectionRegex = /^###\s+Challenge Context\s*\n([\s\S]*?)(?=^###\s+Show More Knowledge\s*$|^#\s+|\Z)/im;
  const match = markdown.match(sectionRegex);
  return match?.[1]?.trim() || '';
}

function extractShowMoreKnowledge(markdown) {
  if (!markdown) {
    return '';
  }

  const sectionRegex = /^###\s+Show More Knowledge\s*\n([\s\S]*?)(?=^#\s+|\Z)/im;
  const match = markdown.match(sectionRegex);
  return match?.[1]?.trim() || '';
}

function stripMarkdownForSummary(markdown) {
  return markdown
    .replace(/```[\s\S]*?```/g, ' ')
    .replace(/!\[[^\]]*\]\([^\)]+\)/g, ' ')
    .replace(/\[([^\]]+)\]\([^\)]+\)/g, '$1')
    .replace(/`([^`]+)`/g, '$1')
    .replace(/[*_>#-]/g, ' ')
    .replace(/\s+/g, ' ')
    .trim();
}

function summarizeLessonContext(markdown) {
  const plain = stripMarkdownForSummary(markdown);
  const sentences = plain
    .split(/(?<=[.!?])\s+/)
    .map((item) => item.trim())
    .filter((item) => item.length > 30);

  const summary = sentences.slice(0, 3);
  const headingTopics = markdown
    .split('\n')
    .map((line) => line.trim())
    .map((line) => line.match(/^#{2,3}\s+(.+)$/)?.[1]?.trim())
    .filter(Boolean)
    .filter((topic, index, arr) => arr.indexOf(topic) === index)
    .slice(0, 5);

  if (summary.length === 0 && plain) {
    return {
      summary: [plain.slice(0, 220)],
      topics: headingTopics,
    };
  }

  return {
    summary,
    topics: headingTopics,
  };
}

export default function LessonView() {
  const { id: lessonSlug } = useParams();
  const location = useLocation();
  const navigate = useNavigate();
  const { getCourseBySlug, getLessonBySlug, isLessonCompleted, markLessonComplete } = useCourseContent();
  const { isAuthenticated } = useAuth();

  const courseSlug = useMemo(() => {
    const search = new URLSearchParams(location.search);
    return search.get('course') || '';
  }, [location.search]);

  const course = getCourseBySlug(courseSlug);
  const lessons = course?.lessons || [];
  const lesson = getLessonBySlug(courseSlug, lessonSlug);
  const completed = lesson ? isLessonCompleted(courseSlug, lesson.slug) : false;
  const expectedOutput = lesson ? extractExpectedOutput(lesson.content || '') : null;
  const challengeContext = useMemo(() => extractChallengeContext(lesson?.content || ''), [lesson?.content]);
  const extendedKnowledge = useMemo(() => extractShowMoreKnowledge(lesson?.content || ''), [lesson?.content]);
  const knowledgeBody = extendedKnowledge || lesson?.content || '';
  const lessonHints = useMemo(() => extractHints(knowledgeBody), [knowledgeBody]);
  const lessonSummary = useMemo(() => summarizeLessonContext(lesson?.content || ''), [lesson?.content]);

  const [editorValue, setEditorValue] = useState('');
  const [output, setOutput] = useState(null);
  const [executing, setExecuting] = useState(false);
  const [error, setError] = useState('');
  const [runResult, setRunResult] = useState(null);
  const [showKnowledge, setShowKnowledge] = useState(false);

  useEffect(() => {
    if (!courseSlug) {
      navigate('/courses', { replace: true });
      return;
    }

    if (!course) {
      navigate('/courses', { replace: true });
      return;
    }

    if (!lesson) {
      navigate(`/course/${courseSlug}`, { replace: true });
      return;
    }

    setError('');
    setOutput(null);
    setRunResult(null);
    setShowKnowledge(false);
    setExecuting(false);
    setEditorValue(extractStarterCode(lesson.content || '', course.language || 'go'));
  }, [courseSlug, course, lesson, navigate]);

  if (!lesson || !course) {
    return (
      <div className="min-h-screen grid place-items-center">
        <p className="text-slate-600">Lesson not found.</p>
      </div>
    );
  }

  const handleRun = async () => {
    setExecuting(true);
    setError('');
    setOutput(null);
    setRunResult(null);
    try {
      const hasExpected = Boolean(expectedOutput);

      if (isAuthenticated) {
        const submission = await api.execute.submitLesson(
          courseSlug,
          lessonSlug,
          course.language || 'go',
          editorValue,
          expectedOutput || ''
        );

        setOutput(submission);
        setRunResult({
          passed: Boolean(submission.passed),
          hasExpected,
          expectedOutput: submission.expected_output || (hasExpected ? normalizeOutput(expectedOutput || '') : null),
          actualOutput: submission.actual_output || normalizeOutput(submission.stdout || ''),
        });

        if (submission.passed) {
          markLessonComplete(courseSlug, lessonSlug, true);
        }
      } else {
        const result = await api.execute.runCode(course.language || 'go', editorValue);
        setOutput(result);

        const cleanStdout = normalizeOutput(result.stdout);
        const expected = normalizeOutput(expectedOutput || '');
        const passedByExpected = hasExpected && cleanStdout === expected;
        const passedByExecution = !hasExpected && result.exit_code === 0 && !result.stderr?.trim();
        const passed = passedByExpected || passedByExecution;

        setRunResult({
          passed,
          hasExpected,
          expectedOutput: hasExpected ? expected : null,
          actualOutput: cleanStdout,
        });

        if (passed) {
          markLessonComplete(courseSlug, lessonSlug, true);
        }
      }
    } catch (err) {
      const message = err.message || 'Execution failed';
      const timeoutMessage = message.toLowerCase().includes('timed out')
        ? 'Execution timed out. Reduce loops/input size and try again.'
        : message;
      setOutput({ error: timeoutMessage });
      setRunResult({
        passed: false,
        hasExpected: Boolean(expectedOutput),
        expectedOutput: expectedOutput ? normalizeOutput(expectedOutput) : null,
        actualOutput: '',
      });
    } finally {
      setExecuting(false);
    }
  };

  const handleComplete = async () => {
    markLessonComplete(courseSlug, lessonSlug, true);
  };

  const currentLessonIndex = lessons.findIndex((item) => item.slug === lesson.slug);
  const prevLesson = currentLessonIndex > 0 ? lessons[currentLessonIndex - 1] : null;
  const nextLesson =
    currentLessonIndex >= 0 && currentLessonIndex < lessons.length - 1 ? lessons[currentLessonIndex + 1] : null;

  const goToLesson = (targetLesson) => {
    if (!targetLesson) return;
    navigate(`/lesson/${targetLesson.slug}?course=${courseSlug}`);
  };

  return (
    <div className="min-h-screen px-4 py-8">
      <div className="mx-auto max-w-6xl space-y-6">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h1 className="text-3xl font-bold text-slate-900">{lesson.title}</h1>
            <p className="text-slate-600 mt-2">
              {course.title} • Lesson {currentLessonIndex + 1} of {lessons.length}
            </p>
          </div>
          <div className="flex items-center gap-3">
            <span
              className={`rounded-full px-3 py-1 text-xs font-semibold ${
                completed ? 'bg-emerald-100 text-emerald-800' : 'bg-slate-200 text-slate-700'
              }`}
            >
              {completed ? 'Completed' : 'Not completed'}
            </span>
            {!completed && (
              <Button
                onClick={handleComplete}
                variant="success"
              >
                Mark Complete
              </Button>
            )}
            <Link
              to={`/course/${courseSlug}`}
              className="rounded-lg border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700"
            >
              Back to Course
            </Link>
          </div>
        </div>

        <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
          <Button variant="secondary" disabled={!prevLesson} onClick={() => goToLesson(prevLesson)}>
            Previous Lesson
          </Button>
          <Button variant="secondary" disabled={!nextLesson} onClick={() => goToLesson(nextLesson)}>
            Next Lesson
          </Button>
        </div>

        {error && <p className="text-red-600">{error}</p>}

        <Card>
          <h2 className="text-xl font-semibold text-slate-900 mb-3">Challenge Brief</h2>
          {challengeContext ? (
            <MarkdownLessonContent content={challengeContext} />
          ) : (
            <div className="space-y-3 text-slate-700">
              {lessonSummary.summary.map((point) => (
                <p key={point}>{point}</p>
              ))}
            </div>
          )}

          {lessonSummary.topics.length > 0 && (
            <div className="mt-4">
              <p className="text-sm font-semibold text-slate-800 mb-2">Key Topics</p>
              <div className="flex flex-wrap gap-2">
                {lessonSummary.topics.map((topic) => (
                  <span key={topic} className="rounded-full bg-slate-100 px-3 py-1 text-xs text-slate-700">
                    {topic}
                  </span>
                ))}
              </div>
            </div>
          )}

          <div className="mt-5 border-t border-slate-200 pt-4">
            <Button variant="secondary" onClick={() => setShowKnowledge((prev) => !prev)}>
              {showKnowledge ? 'Hide More Knowledge' : 'Show More Knowledge'}
            </Button>
          </div>

          {showKnowledge && (
            <div className="mt-5">
              <h3 className="text-lg font-semibold text-slate-900 mb-3">Full Lesson Knowledge</h3>
              <MarkdownLessonContent content={knowledgeBody} />
            </div>
          )}
        </Card>

        {lessonHints.length > 0 && (
          <Card>
            <h2 className="text-xl font-semibold text-slate-900 mb-3">Hints</h2>
            <ul className="list-disc pl-6 space-y-2 text-slate-700">
              {lessonHints.map((hint) => (
                <li key={hint}>{hint}</li>
              ))}
            </ul>
          </Card>
        )}

        <Card>
          <div className="mb-3 flex items-center justify-between">
            <h2 className="text-xl font-semibold text-slate-900">Code Editor</h2>
            <Button
              onClick={handleRun}
              disabled={executing}
            >
              {executing ? 'Running...' : 'Run Code'}
            </Button>
          </div>

          <div className="overflow-hidden rounded-xl border border-slate-200">
            <MonacoEditor
              height="400px"
              defaultLanguage={course.language || 'go'}
              value={editorValue}
              onChange={(value) => setEditorValue(value ?? '')}
              theme="vs-dark"
              options={{
                minimap: { enabled: false },
                fontSize: 14,
              }}
            />
          </div>
        </Card>

      {output && (
        <Card>
          <h2 className="text-xl font-semibold text-slate-900 mb-3">Execution Result</h2>
          {output.error && <p className="mb-3 text-red-600">{output.error}</p>}
          <div className="grid gap-3 text-sm">
            <div className="rounded-lg bg-slate-100 p-3">
              <p className="font-semibold text-slate-800 mb-1">Stdout</p>
              <pre className="overflow-auto whitespace-pre-wrap text-slate-700">{output.stdout || '(empty)'}</pre>
            </div>
            <div className="rounded-lg bg-slate-100 p-3">
              <p className="font-semibold text-slate-800 mb-1">Stderr</p>
              <pre className="overflow-auto whitespace-pre-wrap text-slate-700">{output.stderr || '(empty)'}</pre>
            </div>
            <p className="text-slate-600">Exit code: {output.exit_code} | Timed out: {String(output.timed_out)}</p>
            <p className="text-slate-600">
              Runner: {output.runner || '-'} | Sandboxed: {String(output.sandboxed ?? false)}
              {output.runner_image ? ` | Image: ${output.runner_image}` : ''}
            </p>
            {typeof output.progress_marked === 'boolean' && (
              <p className="text-slate-600">Progress updated: {output.progress_marked ? 'yes' : 'no'}</p>
            )}
          </div>
        </Card>
      )}

      {runResult && runResult.passed && (
        <Card className="border-emerald-200 bg-emerald-50">
          <h2 className="text-xl font-semibold text-emerald-900">Congratulations!</h2>
          <p className="mt-2 text-emerald-800">
            You successfully completed this section.
          </p>
          <div className="mt-4 flex flex-wrap gap-2">
            <Button variant="success" disabled={!nextLesson} onClick={() => goToLesson(nextLesson)}>
              Go to Next Section
            </Button>
            <Button variant="secondary" onClick={() => navigate(`/course/${courseSlug}/progress`)}>
              View Progress
            </Button>
          </div>
        </Card>
      )}

      {runResult && !runResult.passed && runResult.hasExpected && (
        <Card className="border-amber-200 bg-amber-50">
          <h2 className="text-xl font-semibold text-amber-900">Not quite yet</h2>
          <p className="mt-2 text-amber-800">
            Your program ran, but the output does not match the expected lesson output.
          </p>
          <div className="mt-4 grid gap-3 text-sm">
            <div className="rounded-lg bg-white p-3 border border-amber-200">
              <p className="font-semibold text-amber-900 mb-1">Expected Output</p>
              <pre className="overflow-auto whitespace-pre-wrap text-amber-900">{runResult.expectedOutput || '(empty)'}</pre>
            </div>
            <div className="rounded-lg bg-white p-3 border border-amber-200">
              <p className="font-semibold text-amber-900 mb-1">Your Output</p>
              <pre className="overflow-auto whitespace-pre-wrap text-amber-900">{runResult.actualOutput || '(empty)'}</pre>
            </div>
          </div>
        </Card>
      )}
      </div>
    </div>
  );
}