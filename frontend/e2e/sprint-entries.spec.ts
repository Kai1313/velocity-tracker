import { test, expect } from '@playwright/test';
import { api, deleteEntriesReferencing, uniqueName, type ApiProject, type ApiTicket, type ApiSprint } from './helpers';

test.describe('Admin / Sprint Entries', () => {
  let project: ApiProject;
  let ticket: ApiTicket;
  let openSprint: ApiSprint;
  let closedSprint: ApiSprint;

  test.beforeAll(async () => {
    project = await api.createProject(uniqueName('E2E Entry Project'));
    ticket = await api.createTicket(project.id, uniqueName('E2E Entry Ticket'), 5);
    openSprint = await api.createSprint(uniqueName('E2E Open Sprint'), '2026-11-01T00:00:00Z', '2026-11-14T00:00:00Z');
    const closedSprintOpen = await api.createSprint(
      uniqueName('E2E Closed Sprint'),
      '2026-10-01T00:00:00Z',
      '2026-10-14T00:00:00Z',
    );
    closedSprint = await api.closeSprint(closedSprintOpen);
  });

  test.afterAll(async () => {
    await deleteEntriesReferencing([ticket.id], [openSprint.id, closedSprint.id]);
    await api.deleteTicket(ticket.id);
    await api.deleteSprint(openSprint.id);
    await api.deleteSprint(closedSprint.id);
    await api.deleteProject(project.id);
  });

  test('creating an entry auto-fills points from the ticket, and it can be edited', async ({ page }) => {
    await page.goto('/admin/sprint-entries');
    await expect(page.getByRole('heading', { name: 'Sprint Entries' })).toBeVisible();

    await page.getByRole('button', { name: 'Add entry' }).click();
    await page.locator('#entry-ticket').selectOption({ label: ticket.title });
    await page.locator('#entry-sprint').selectOption({ label: `${openSprint.name} (Open)` });
    await expect(page.locator('#entry-points')).toHaveValue(String(ticket.storyPoints));

    await page.locator('input[type="checkbox"]').check();
    await page.getByRole('button', { name: 'Save' }).click();

    const row = page.locator('tr', { hasText: ticket.title }).filter({ hasText: openSprint.name });
    await expect(row).toBeVisible();
    await expect(row.getByText('Late add')).toBeVisible();

    // Ticket/sprint are fixed after creation — edit should show them as
    // read-only context, not editable selects.
    await row.getByRole('button', { name: 'Edit' }).click();
    await expect(page.locator('#entry-ticket')).toHaveCount(0);
    await expect(page.locator('#entry-sprint')).toHaveCount(0);
    await page.locator('#entry-status').selectOption('Done');
    await page.getByRole('button', { name: 'Save' }).click();

    await expect(row.getByText('Done')).toBeVisible();

    await row.getByRole('button', { name: 'Delete' }).click();
    await page.getByRole('alertdialog').getByRole('button', { name: 'Delete' }).click();
    await expect(row).toHaveCount(0);
  });

  test('an entry in a closed sprint is locked for edit but not for delete', async ({ page }) => {
    const entry = await api.createSprintEntry(ticket.id, closedSprint.id, ticket.storyPoints);

    await page.goto('/admin/sprint-entries');
    const row = page.locator('tr', { hasText: ticket.title }).filter({ hasText: closedSprint.name });
    await expect(row).toBeVisible();
    await expect(row.getByText('Locked')).toBeVisible();
    await expect(row.getByRole('button', { name: 'Edit' })).toBeDisabled();

    await row.getByRole('button', { name: 'Delete' }).click();
    await page.getByRole('alertdialog').getByRole('button', { name: 'Delete' }).click();
    await expect(row).toHaveCount(0);

    // Guard against double-cleanup if the UI delete above didn't go through.
    await api.deleteSprintEntry(entry.id).catch(() => {});
  });
});
