import type { ReactNode } from 'react';

/** String (optionally a key, when followed by a colon) | literal | number. */
const TOKEN = /("(?:\\.|[^"\\])*")(\s*:)?|\b(?:true|false|null)\b|-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?/g;

/**
 * Monochrome JSON highlighting by weight and tone only: keys `fg`, strings
 * `muted`, numbers and literals `fg` medium, punctuation `subtle`.
 */
function highlight(json: string): ReactNode[] {
  const out: ReactNode[] = [];
  let last = 0;
  for (const match of json.matchAll(TOKEN)) {
    const start = match.index;
    if (start > last) out.push(<span key={`p${last}`} className="text-subtle">{json.slice(last, start)}</span>);
    const [token, str, colon] = match;
    if (str && colon) {
      out.push(
        <span key={start} className="text-fg">
          {str}
        </span>,
        <span key={`c${start}`} className="text-subtle">
          {colon}
        </span>,
      );
    } else {
      out.push(
        <span key={start} className={str ? 'text-muted' : 'font-medium text-fg'}>
          {token}
        </span>,
      );
    }
    last = start + token.length;
  }
  if (last < json.length) out.push(<span key={`p${last}`} className="text-subtle">{json.slice(last)}</span>);
  return out;
}

interface CodeBlockProps {
  /** Pretty-printed JSON, or plain text when `plain` is set. */
  code: string;
  plain?: boolean;
  label: string;
}

/** Scrollable, bounded code view. Wraps long lines so nothing hides sideways. */
export function CodeBlock({ code, plain, label }: CodeBlockProps) {
  return (
    <pre
      aria-label={label}
      tabIndex={0}
      className="max-h-96 overflow-auto whitespace-pre-wrap break-words p-4 font-mono text-caption text-muted"
    >
      <code>{plain ? code : highlight(code)}</code>
    </pre>
  );
}
