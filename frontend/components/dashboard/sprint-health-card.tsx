'use client';

import { useEffect, useState } from 'react';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { SprintHealthTable } from '@/components/dashboard/sprint-health-table';
import type { SprintHealth } from '@/lib/api';

const STORAGE_KEY = 'dashboard.sprintHealthCard.collapsed';

function ChevronIcon({ open }: { open: boolean }) {
  return (
    <svg
      viewBox="0 0 24 24"
      width="16"
      height="16"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      className={`shrink-0 text-muted-foreground transition-transform ${open ? '' : '-rotate-90'}`}
    >
      <path d="M6 9l6 6 6-6" />
    </svg>
  );
}

// Defaults to expanded on every render (including the server-rendered first
// paint) so there's no hydration mismatch; a stored "collapsed" preference is
// applied right after mount instead.
export function SprintHealthCard({ rows }: { rows: SprintHealth[] }) {
  const [open, setOpen] = useState(true);

  useEffect(() => {
    try {
      if (localStorage.getItem(STORAGE_KEY) === 'true') {
        setOpen(false);
      }
    } catch {
      // localStorage unavailable (private browsing, etc.) — stay expanded.
    }
  }, []);

  function toggle() {
    setOpen((prev) => {
      const next = !prev;
      try {
        localStorage.setItem(STORAGE_KEY, next ? 'false' : 'true');
      } catch {
        // ignore — collapse state just won't persist this session
      }
      return next;
    });
  }

  return (
    <Card>
      <CardHeader
        role="button"
        tabIndex={0}
        aria-expanded={open}
        onClick={toggle}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            toggle();
          }
        }}
        className="flex-row items-center justify-between gap-4 space-y-0 cursor-pointer select-none"
      >
        <div className="space-y-1.5">
          <CardTitle>Sprint Health</CardTitle>
          <p className="text-sm font-normal text-muted-foreground">
            Comparing &quot;Required Velocity&quot; vs &quot;Achieved Velocity&quot; as an early warning signal for
            sprint risk. One row per open sprint, aggregated across every active project&apos;s work in it — the team&apos;s
            developers share capacity across projects, so splitting this out per project would distort the pace
            comparison.
          </p>
        </div>
        <ChevronIcon open={open} />
      </CardHeader>
      {open && (
        <CardContent className="overflow-x-auto p-0 pb-6">
          <SprintHealthTable rows={rows} />
        </CardContent>
      )}
    </Card>
  );
}
