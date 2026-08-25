'use client';

import { useEffect, useState } from 'react';
import { useTheme } from 'next-themes';

const NEXT_THEME: Record<string, string> = {
  light: 'dark',
  dark: 'system',
  system: 'light',
};

const LABEL: Record<string, string> = {
  light: 'Light theme',
  dark: 'Dark theme',
  system: 'System theme',
};

function SunIcon() {
  return (
    <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2">
      <circle cx="12" cy="12" r="4" />
      <path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41" />
    </svg>
  );
}

function MoonIcon() {
  return (
    <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M20 14.5A8.5 8.5 0 1 1 9.5 4a6.5 6.5 0 0 0 10.5 10.5Z" />
    </svg>
  );
}

function SystemIcon() {
  return (
    <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2">
      <rect x="3" y="4" width="18" height="12" rx="1" />
      <path d="M8 20h8M12 16v4" />
    </svg>
  );
}

const ICON: Record<string, () => JSX.Element> = {
  light: SunIcon,
  dark: MoonIcon,
  system: SystemIcon,
};

export function ThemeToggle() {
  const { theme, setTheme } = useTheme();
  const [mounted, setMounted] = useState(false);

  useEffect(() => setMounted(true), []);

  const current = mounted ? (theme ?? 'system') : 'system';
  const Icon = ICON[current];

  return (
    <button
      type="button"
      onClick={() => setTheme(NEXT_THEME[current])}
      aria-label={`${LABEL[current]} — click to switch to ${LABEL[NEXT_THEME[current]].toLowerCase()}`}
      title={LABEL[current]}
      className="inline-flex h-8 w-8 items-center justify-center rounded-md border border-border text-foreground transition-colors hover:bg-muted"
    >
      {mounted ? <Icon /> : null}
    </button>
  );
}
