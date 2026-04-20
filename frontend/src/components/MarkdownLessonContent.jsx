import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';

export default function MarkdownLessonContent({ content }) {
  return (
    <div className="max-w-none text-slate-700 leading-relaxed [&_h1]:mt-6 [&_h1]:text-2xl [&_h1]:font-bold [&_h2]:mt-6 [&_h2]:text-xl [&_h2]:font-semibold [&_h3]:mt-5 [&_h3]:text-lg [&_h3]:font-semibold [&_p]:mt-3 [&_pre]:mt-4 [&_pre]:overflow-auto [&_pre]:rounded-xl [&_pre]:bg-slate-900 [&_pre]:p-4 [&_pre]:text-slate-100 [&_code]:font-mono [&_ul]:mt-3 [&_ul]:list-disc [&_ul]:pl-6 [&_ol]:mt-3 [&_ol]:list-decimal [&_ol]:pl-6 [&_table]:mt-4 [&_table]:w-full [&_table]:overflow-hidden [&_table]:rounded-lg [&_table]:border [&_table]:border-slate-200 [&_th]:bg-slate-100 [&_th]:px-3 [&_th]:py-2 [&_td]:border-t [&_td]:border-slate-200 [&_td]:px-3 [&_td]:py-2">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          a: ({ node, ...props }) => (
            <a
              {...props}
              target="_blank"
              rel="noreferrer noopener"
              className="text-sky-700 underline underline-offset-2"
            />
          ),
          img: ({ node, ...props }) => (
            <img
              {...props}
              className="my-4 w-full rounded-xl border border-slate-200"
              loading="lazy"
            />
          ),
          code: ({ inline, className, children, ...props }) => {
            if (inline) {
              return (
                <code
                  {...props}
                  className="rounded bg-slate-100 px-1.5 py-0.5 text-sm text-slate-800"
                >
                  {children}
                </code>
              );
            }

            return (
              <code {...props} className={className}>
                {children}
              </code>
            );
          },
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
}
