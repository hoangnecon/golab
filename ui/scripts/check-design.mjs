#!/usr/bin/env node
// Enforces ../DESIGN.md. Tailwind ignores unknown utilities silently, so this
// is the gate that rejects off-system styling. Exit code 1 on any violation.
import { readdirSync, readFileSync, statSync } from 'node:fs';
import { join, relative } from 'node:path';

const ROOT = new URL('../src/', import.meta.url).pathname;

const RULES = [
  {
    name: 'raw palette color (use tokens: app, canvas, surface, raised, hover, line, line-strong, fg, muted, subtle)',
    pattern:
      /\b(?:bg|text|border|ring|outline|divide|fill|stroke|from|to|via|shadow|placeholder|decoration|accent|caret)-(?:slate|gray|zinc|neutral|stone|red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose|black|white)\b/,
  },
  {
    name: 'removed unsupported hue token',
    pattern: /\b(?:bg|text|border|ring|outline|divide|fill|stroke)-(?:accent|on-accent|info|sidebar)\b/,
  },
  { name: 'hex or rgb literal in source', pattern: /#[0-9a-fA-F]{3,8}\b|\brgba?\(/ },
  { name: 'arbitrary Tailwind value', pattern: /\b[a-z-]+-\[[^\]]+\]/ },
  { name: 'font size outside caption|body|title|display', pattern: /\btext-(?:xs|sm|base|lg|xl|[2-9]xl)\b/ },
  { name: 'radius other than `rounded` / `rounded-lg` / `rounded-xl` / `rounded-full`', pattern: /\brounded-(?:sm|md|2xl|3xl)\b/ },
  { name: 'shadow other than `shadow-overlay`', pattern: /\bshadow-(?:sm|md|lg|xl|2xl|inner)\b/ },
  { name: 'centered page container (layout is fluid and left-aligned)', pattern: /\bmx-auto\b/ },
  { name: 'gradient / blur / glass', pattern: /\b(?:bg-gradient-|backdrop-blur|blur-)/ },
  { name: 'pulse / bounce / ping animation', pattern: /\banimate-(?:pulse|bounce|ping|spin)\b/ },
  { name: 'native alert/confirm/prompt', pattern: /\b(?:alert|confirm|prompt)\(/, skip: /useConfirm|confirm\(\{/ },
  { name: 'console logging', pattern: /\bconsole\.(?:log|debug|info)\b/ },
  { name: '`any` type', pattern: /:\s*any\b|\bas any\b|<any>/ },
];

function* files(dir) {
  for (const entry of readdirSync(dir)) {
    const path = join(dir, entry);
    if (statSync(path).isDirectory()) yield* files(path);
    else if (/\.(ts|tsx|css)$/.test(entry)) yield path;
  }
}

let violations = 0;
for (const file of files(ROOT)) {
  const isCss = file.endsWith('.css');
  readFileSync(file, 'utf8')
    .split('\n')
    .forEach((line, index) => {
      for (const rule of RULES) {
        // index.css is the one place hex-like/rgb values may live (token definitions).
        if (isCss && rule.name.startsWith('hex')) continue;
        if (rule.skip?.test(line)) continue;
        if (rule.pattern.test(line)) {
          violations += 1;
          console.error(`${relative(ROOT, file)}:${index + 1}  ${rule.name}\n    ${line.trim()}`);
        }
      }
    });
}

if (violations > 0) {
  console.error(`\n${violations} design violation(s). See DESIGN.md.`);
  process.exit(1);
}
console.log('check:design passed');
