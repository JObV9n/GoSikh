import beginnerGolangMarkdown from './beginner-golang-course.md?raw';

export const courseCatalog = [
  {
    id: 'beginner-golang-course',
    slug: 'beginner-golang',
    title: 'Beginner Golang Course',
    shortTitle: 'Go for Beginners',
    description:
      'Learn Go from first principles with hands-on examples covering syntax, control flow, functions, data structures, modules, testing, concurrency, and context.',
    level: 'Beginner',
    language: 'go',
    estimatedHours: 20,
    tags: ['Go', 'Backend', 'Concurrency', 'Testing'],
    prerequisites: ['Basic programming familiarity'],
    outcomes: [
      'Write and run idiomatic Go programs',
      'Understand core language features and standard tooling',
      'Build confidence with testing, modules, and package design',
      'Use goroutines, channels, and context for concurrent programs',
    ],
    markdown: beginnerGolangMarkdown,
  },
];
