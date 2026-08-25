import Link from 'next/link';

import { ThemeToggle } from '@/components/theme-toggle';

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  return (
    <>
      <div className="flex items-center justify-between border-b border-border px-8 py-3">
        <Link href="/admin/users" className="text-sm text-muted-foreground hover:underline">
          Manage data &rarr;
        </Link>
        <ThemeToggle />
      </div>
      {children}
    </>
  );
}
