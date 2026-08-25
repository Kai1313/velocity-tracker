import { ThemeToggle } from '@/components/theme-toggle';

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  return (
    <>
      <div className="flex justify-end border-b border-border px-8 py-3">
        <ThemeToggle />
      </div>
      {children}
    </>
  );
}
