import './globals.css';

export const metadata = {
  title: 'Velocity Tracker',
  description: 'Sprint velocity, planning accuracy, and late-add rate for a small software team.',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
