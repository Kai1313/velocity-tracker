import { test, expect } from '@playwright/test';
import { api, deleteEntriesReferencing, uniqueName, type ApiProject, type ApiTicket, type ApiSprint } from './helpers';

test.describe('Dashboard / Retrospective', () => {
  let project: ApiProject;
  let committedTicket: ApiTicket;
  let lateAddTicket: ApiTicket;
  let sprint: ApiSprint;

  test.beforeAll(async () => {
    project = await api.createProject(uniqueName('E2E Retro Project'));
    committedTicket = await api.createTicket(project.id, uniqueName('E2E Committed Ticket'), 6);
    lateAddTicket = await api.createTicket(project.id, uniqueName('E2E Late-Add Ticket'), 4);

    const open = await api.createSprint(uniqueName('E2E Retro Sprint'), '2026-07-01T00:00:00Z', '2026-07-14T00:00:00Z');
    await api.createSprintEntry(committedTicket.id, open.id, 6, { status: 'Done' });
    await api.createSprintEntry(lateAddTicket.id, open.id, 4, { addedAfterSprintStart: true });
    sprint = await api.closeSprint(open);
  });

  test.afterAll(async () => {
    await deleteEntriesReferencing([committedTicket.id, lateAddTicket.id], [sprint.id]);
    await api.deleteTicket(committedTicket.id);
    await api.deleteTicket(lateAddTicket.id);
    await api.deleteSprint(sprint.id);
    await api.deleteProject(project.id);
  });

  test('shows Planning Accuracy and Late-Add Rate for the most recently closed sprint', async ({ page }) => {
    await page.goto('/dashboard/retrospective?limit=1');
    await expect(page.getByRole('heading', { name: 'Sprint Retrospective' })).toBeVisible();

    const row = page.locator('tr', { hasText: sprint.name });
    await expect(row).toBeVisible();
    // Committed 6 (Done) + late-add 4: Planning Accuracy = 6/6*100 = 100%,
    // Late-Add Rate = 4/6*100 ≈ 67%.
    await expect(row.getByText('100%', { exact: true })).toBeVisible();
    await expect(row.getByText('67%', { exact: true })).toBeVisible();
  });

  test('the window switcher links update the limit param', async ({ page }) => {
    await page.goto('/dashboard/retrospective');
    await expect(page).toHaveURL(/\/dashboard\/retrospective$/);

    await page.getByRole('link', { name: '6 sprints' }).click();
    await expect(page).toHaveURL(/limit=6/);
    await expect(page.getByText('last 6 closed sprints')).toBeVisible();
  });
});
