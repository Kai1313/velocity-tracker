import { defineConfig } from '@playwright/test';

// These are browser-driven end-to-end tests against a real backend + Postgres
// (not mocked), the same way the backend's own integration tests need a real
// Postgres via TEST_DATABASE_URL. Run `docker compose up` first — see README.
//
// fullyParallel/workers are 1 because tests create and delete real rows
// through the shared API; concurrent runs could collide on unique-name
// checks or FK ordering during cleanup.
export default defineConfig({
  testDir: './e2e',
  timeout: 30_000,
  fullyParallel: false,
  workers: 1,
  retries: 0,
  reporter: 'list',
  use: {
    baseURL: process.env.E2E_BASE_URL ?? 'http://localhost:3000',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
  },
});
