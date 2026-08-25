import { test, expect } from '@playwright/test';
import { api, deleteTicketsReferencing, uniqueName, type ApiProject, type ApiUser } from './helpers';

test.describe('Admin / Tickets', () => {
  let project: ApiProject;
  let user: ApiUser;

  test.beforeAll(async () => {
    project = await api.createProject(uniqueName('E2E Ticket Project'));
    user = await api.createUser(uniqueName('E2E Ticket User'), 'Developer');
  });

  test.afterAll(async () => {
    // Covers the case where the test failed before its own UI-driven ticket
    // delete ran — otherwise deleting the project would 409 on the FK.
    await deleteTicketsReferencing([project.id]);
    await api.deleteProject(project.id);
    await api.deleteUser(user.id);
  });

  test('create a ticket, assign it, then unassign and delete it', async ({ page }) => {
    const title = uniqueName('E2E Ticket');

    await page.goto('/admin/tickets');
    await expect(page.getByRole('heading', { name: 'Tickets' })).toBeVisible();

    await page.getByRole('button', { name: 'Add ticket' }).click();
    await page.locator('#ticket-project').selectOption({ label: project.name });
    await page.locator('#ticket-title').fill(title);
    await page.locator('#ticket-points').fill('8');
    await page.locator('#ticket-assignee').selectOption({ label: user.name });
    await page.getByRole('button', { name: 'Save' }).click();

    const row = page.locator('tr', { hasText: title });
    await expect(row).toBeVisible();
    await expect(row.getByText(project.name)).toBeVisible();
    await expect(row.getByText(user.name)).toBeVisible();
    await expect(row.getByRole('cell', { name: '8', exact: true })).toBeVisible();

    await row.getByRole('button', { name: 'Edit' }).click();
    await page.locator('#ticket-assignee').selectOption('unassigned');
    await page.getByRole('button', { name: 'Save' }).click();

    await expect(row.getByText('Unassigned')).toBeVisible();

    await row.getByRole('button', { name: 'Delete' }).click();
    await page.getByRole('alertdialog').getByRole('button', { name: 'Delete' }).click();

    await expect(page.locator('tr', { hasText: title })).toHaveCount(0);
  });
});
