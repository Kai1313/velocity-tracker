import { test, expect } from '@playwright/test';
import { api, deleteEntriesReferencing, uniqueName, type ApiProject, type ApiTicket, type ApiSprint } from './helpers';

test.describe('Dashboard / Sprint detail', () => {
  let project: ApiProject;
  let freshTicket: ApiTicket;
  let carriedTicket: ApiTicket;
  let priorSprint: ApiSprint;
  let currentSprint: ApiSprint;

  test.beforeAll(async () => {
    project = await api.createProject(uniqueName('E2E Dashboard Project'));
    freshTicket = await api.createTicket(project.id, uniqueName('E2E Fresh Ticket'), 5);
    carriedTicket = await api.createTicket(project.id, uniqueName('E2E Carried Ticket'), 3);

    const priorSprintOpen = await api.createSprint(
      uniqueName('E2E Prior Sprint'),
      '2026-11-01T00:00:00Z',
      '2026-11-14T00:00:00Z',
    );
    const originEntry = await api.createSprintEntry(carriedTicket.id, priorSprintOpen.id, 3, { status: 'NotDone' });
    priorSprint = await api.closeSprint(priorSprintOpen);

    currentSprint = await api.createSprint(
      uniqueName('E2E Current Sprint'),
      '2026-11-15T00:00:00Z',
      '2026-11-28T00:00:00Z',
    );
    await api.createSprintEntry(freshTicket.id, currentSprint.id, 5, { status: 'NotDone' });
    await api.createSprintEntry(carriedTicket.id, currentSprint.id, 3, {
      status: 'NotDone',
      carriedFrom: originEntry.id,
    });
  });

  test.afterAll(async () => {
    await deleteEntriesReferencing([freshTicket.id, carriedTicket.id], [priorSprint.id, currentSprint.id]);
    await api.deleteTicket(freshTicket.id);
    await api.deleteTicket(carriedTicket.id);
    await api.deleteSprint(priorSprint.id);
    await api.deleteSprint(currentSprint.id);
    await api.deleteProject(project.id);
  });

  test('clicking sprint workload navigates to the current-vs-carried-over breakdown', async ({ page }) => {
    await page.goto(`/dashboard/${currentSprint.id}`);
    await expect(page.getByRole('heading', { name: currentSprint.name })).toBeVisible();

    await page.getByText('Sprint workload (pts)').click();
    await expect(page).toHaveURL(`/dashboard/${currentSprint.id}/entries`);

    const currentCard = page.locator('div', { has: page.getByRole('heading', { name: 'Current sprint workload' }) }).first();
    await expect(currentCard.getByText('5', { exact: true })).toBeVisible();
    await expect(currentCard.getByText('1 ticket', { exact: true })).toBeVisible();

    const carriedCard = page
      .locator('div', { has: page.getByRole('heading', { name: 'Carry-over sprint workload' }) })
      .first();
    await expect(carriedCard.getByText('3', { exact: true })).toBeVisible();
    await expect(carriedCard.getByText('1 ticket', { exact: true })).toBeVisible();
  });
});
