type ReadyResponse = {
  status: string;
  db?: string;
  error?: string;
};

async function getBackendStatus(): Promise<{ ok: boolean; body: ReadyResponse | { error: string } }> {
  const backendUrl = process.env.BACKEND_INTERNAL_URL ?? 'http://localhost:8080';

  try {
    const res = await fetch(`${backendUrl}/readyz`, { cache: 'no-store' });
    const body = (await res.json()) as ReadyResponse;
    return { ok: res.ok, body };
  } catch (err) {
    return { ok: false, body: { error: err instanceof Error ? err.message : String(err) } };
  }
}

export default async function Home() {
  const { ok, body } = await getBackendStatus();

  return (
    <main style={{ fontFamily: 'sans-serif', padding: '2rem' }}>
      <h1>Velocity Tracker</h1>
      <p>Scaffold status check — this page confirms frontend → backend → database connectivity.</p>
      <p>
        Backend: <strong>{ok ? 'ok' : 'unavailable'}</strong>
      </p>
      <pre style={{ background: '#f4f4f4', padding: '1rem' }}>{JSON.stringify(body, null, 2)}</pre>
    </main>
  );
}
