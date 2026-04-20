import { useState } from 'react';
import MonacoEditor from '@monaco-editor/react';
import { api } from '../services/api';
import Card from '../components/Card';
import Button from '../components/Button';

const starterCodeByLanguage = {
  javascript: 'console.log("Hello from playground")',
  python: 'print("Hello from playground")',
  go: 'package main\n\nimport "fmt"\n\nfunc main() {\n  fmt.Println("Hello from playground")\n}',
};

export default function Playground() {
  const [language, setLanguage] = useState('javascript');
  const [code, setCode] = useState(starterCodeByLanguage.javascript);
  const [running, setRunning] = useState(false);
  const [result, setResult] = useState(null);

  const onLanguageChange = (event) => {
    const nextLanguage = event.target.value;
    setLanguage(nextLanguage);
    setCode(starterCodeByLanguage[nextLanguage] || '');
    setResult(null);
  };

  const runCode = async () => {
    setRunning(true);
    setResult(null);
    try {
      const data = await api.execute.runCode(language, code);
      setResult(data);
    } catch (err) {
      setResult({ error: err.message });
    } finally {
      setRunning(false);
    }
  };

  return (
    <div className="min-h-screen px-4 py-8">
      <div className="mx-auto max-w-6xl space-y-4">
        <div className="flex items-center justify-between gap-3">
          <h1 className="text-3xl font-bold text-slate-900">Playground</h1>
          <Button
            onClick={runCode}
            disabled={running}
          >
            {running ? 'Running...' : 'Run Code'}
          </Button>
        </div>

        <div className="flex items-center gap-3">
          <label htmlFor="language" className="text-sm font-medium text-slate-700">Language</label>
          <select
            id="language"
            value={language}
            onChange={onLanguageChange}
            className="rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm text-slate-800"
          >
            <option value="javascript">JavaScript</option>
            <option value="python">Python</option>
            <option value="go">Go</option>
          </select>
        </div>

        <div className="overflow-hidden rounded-xl border border-slate-200">
          <MonacoEditor
            height="420px"
            defaultLanguage={language}
            language={language}
            value={code}
            onChange={(value) => setCode(value || '')}
            theme="vs-dark"
            options={{ minimap: { enabled: false }, fontSize: 14 }}
          />
        </div>

        {result && (
          <Card>
            <h2 className="text-lg font-semibold text-slate-900 mb-3">Output</h2>
            {result.error && <p className="mb-3 text-red-600">{result.error}</p>}
            <pre className="rounded-lg bg-slate-100 p-3 whitespace-pre-wrap text-slate-700">{result.stdout || '(empty stdout)'}</pre>
            <pre className="mt-3 rounded-lg bg-slate-100 p-3 whitespace-pre-wrap text-slate-700">{result.stderr || '(empty stderr)'}</pre>
            <p className="mt-3 text-xs text-slate-500">
              Exit code: {result.exit_code ?? '-'} | Timed out: {String(result.timed_out ?? false)}
            </p>
            <p className="mt-1 text-xs text-slate-500">
              Runner: {result.runner || '-'} | Sandboxed: {String(result.sandboxed ?? false)}
              {result.runner_image ? ` | Image: ${result.runner_image}` : ''}
            </p>
          </Card>
        )}
      </div>
    </div>
  );
}
