import { test, expect } from '@playwright/test';
import { api, deleteEntriesReferencing, uniqueName, type ApiProject, type ApiTicket, type ApiSprint } from './helpers';

// The tests below share fixtures and run in order: the first one leaves the
// fresh ticket at 1 bounce, which the later per-developer assertion relies on.
test.describe.configure({ mode: 'serial' });

test.describe('Bounce tracking', () => {
  let project: ApiProject;
  let freshTicket: ApiTicket;
  let carriedTicket: ApiTicket;
  let retroDoneTicket: ApiTicket;
  let retroOpenTicket: ApiTicket;
  let priorSprint: ApiSprint;
  let currentSprint: ApiSprint;
  let retroSprint: ApiSprint;

  test.beforeAll(async () => {
    project = await api.createProject(uniqueName('E2E Bounce Project'));
    freshTicket = await api.createTicket(project.id, uniqueName('E2E Bounce Fresh'), 3);
    carriedTicket = await api.createTicket(project.id, uniqueName('E2E Bounce Carried'), 5);
    retroDoneTicket = await api.createTicket(project.id, uniqueName('E2E Bounce Retro Done'), 2);
    retroOpenTicket = await api.createTicket(project.id, uniqueName('E2E Bounce Retro Open'), 8);

    // The carried ticket bounced twice in the prior sprint, then carried over.
    const priorOpen = await api.createSprint(uniqueName('E2E Bounce Prior'), '2026-11-01T00:00:00Z', '2026-11-14T00:00:00Z');
    const origin = await api.createSprintEntry(carriedTicket.id, priorOpen.id, 5, { status: 'NotDone' });
    await api.setBounceCount(origin.id, 2);
    priorSprint = await api.closeSprint(priorOpen);

    currentSprint = await api.createSprint(uniqueName('E2E Bounce Current'), '2026-11-15T00:00:00Z', '2026-11-28T00:00:00Z');
    await api.createSprintEntry(freshTicket.id, currentSprint.id, 3, { status: 'NotDone' });
    await api.createSprintEntry(carriedTicket.id, currentSprint.id, 5, { status: 'NotDone', carriedFrom: origin.id });

    // Created last so it is the most recently closed sprint for ?limit=1.
    const retroOpen = await api.createSprint(uniqueName('E2E Bounce Retro'), '2026-12-01T00:00:00Z', '2026-12-14T00:00:00Z');
    const done = await api.createSprintEntry(retroDoneTicket.id, retroOpen.id, 2, { status: 'Done' });
    await api.createSprintEntry(retroOpenTicket.id, retroOpen.id, 8, { status: 'NotDone' });
    await api.setBounceCount(done.id, 1);
    retroSprint = await api.closeSprint(retroOpen);
  });

  test.afterAll(async () => {
    const tickets = [freshTicket, carriedTicket, retroDoneTicket, retroOpenTicket];
    const sprints = [priorSprint, currentSprint, retroSprint];
    await deleteEntriesReferencing(
      tickets.map((t) => t.id),
      sprints.map((s) => s.id),
    );
    for (const t of tickets) await api.deleteTicket(t.id);
    for (const s of sprints) await api.deleteSprint(s.id);
    await api.deleteProject(project.id);
  });

  test('the stepper changes an open-sprint bounce count, which persists, and shows earlier bounces', async ({ page }) => {
    await page.goto(`/admin/sprint-entries?sprintId=${currentSprint.id}`);
    const row = page.locator('tr', { hasText: freshTicket.title });
    const count = row.getByTestId('bounce-count');
    await expect(count).toHaveText('0');

    await row.getByRole('button', { name: 'Increase bounce count' }).click();
    await expect(count).toHaveText('1');
    await row.getByRole('button', { name: 'Increase bounce count' }).click();
    await expect(count).toHaveText('2');
    await row.getByRole('button', { name: 'Decrease bounce count' }).click();
    await expect(count).toHaveText('1');

    await page.reload();
    await expect(page.locator('tr', { hasText: freshTicket.title }).getByTestId('bounce-count')).toHaveText('1');

    // The carried-over entry restarts at 0 but still shows the ticket's history.
    const carriedRow = page.locator('tr', { hasText: carriedTicket.title });
    await expect(carriedRow.getByTestId('bounce-count')).toHaveText('0');
    await expect(carriedRow.getByText('(2 earlier)')).toBeVisible();
  });

  test('a closed sprint entry shows its count read-only', async ({ page }) => {
    await page.goto(`/admin/sprint-entries?sprintId=${priorSprint.id}`);
    const row = page.locator('tr', { hasText: carriedTicket.title });
    await expect(row.getByTestId('bounce-count')).toHaveText('2');
    await expect(row.getByRole('button', { name: 'Increase bounce count' })).toHaveCount(0);
  });

  test('the dashboard lists a carried-over ticket with earlier bounces as stuck, but not one below the threshold', async ({
    page,
  }) => {
    await page.goto('/dashboard');
    await expect(page.getByRole('heading', { name: 'Stuck tickets' })).toBeVisible();

    const stuckRow = page.locator('tr', { hasText: carriedTicket.title });
    await expect(stuckRow).toBeVisible();
    await expect(stuckRow.getByText(currentSprint.name)).toBeVisible();
    await expect(stuckRow.getByText('2 (0 this sprint, 2 earlier)')).toBeVisible();
    await expect(page.locator('tr', { hasText: freshTicket.title })).toHaveCount(0); // 1 bounce < threshold
  });

  test('the sprint detail per-developer table shows bounces with their raw counts', async ({ page }) => {
    await page.goto(`/dashboard/${currentSprint.id}`);
    // Unassigned: fresh (1 bounce) + carried (0) over 2 tracked tickets.
    await expect(page.getByText('1 bounce / 2 tickets = 0.5')).toBeVisible();
  });

  test('the retrospective shows bounce counts, and the by-size table lists every band', async ({ page }) => {
    await page.goto('/dashboard/retrospective?limit=1');
    const row = page.locator('tr', { hasText: retroSprint.name });
    await expect(row.getByText('1 bounce / 2 tickets = 0.5')).toBeVisible();
    await expect(row.getByText('1 of 2')).toBeVisible();

    await expect(page.getByRole('heading', { name: 'Bounces by ticket size' })).toBeVisible();
    for (const band of ['1-2 SP', '3-5 SP', '6+ SP']) {
      await expect(page.locator('tr', { hasText: band })).toBeVisible();
    }
  });
});
