import Link from 'next/link';

import { ThemeToggle } from '@/components/theme-toggle';

const NAV_ITEMS = [
  { href: '/admin/users', label: 'Users' },
  { href: '/admin/projects', label: 'Projects' },
];

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen">
      <aside className="flex w-56 flex-col border-r border-border">
        <div className="p-4">
          <Link href="/dashboard" className="text-sm text-muted-foreground hover:underline">
            &larr; Dashboard
          </Link>
        </div>
        <nav className="flex flex-1 flex-col gap-1 px-2">
          {NAV_ITEMS.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className="rounded-md px-3 py-2 text-sm font-medium text-foreground hover:bg-muted"
            >
              {item.label}
            </Link>
          ))}
        </nav>
        <div className="flex justify-end border-t border-border p-4">
          <ThemeToggle />
        </div>
      </aside>
      <main className="flex-1 p-8">{children}</main>
    </div>
  );
}
