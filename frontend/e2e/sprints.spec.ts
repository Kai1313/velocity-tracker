import { test, expect } from '@playwright/test';
import { uniqueName } from './helpers';

test.describe('Admin / Sprints', () => {
  test('create a sprint, verify the date round-trip, then close and delete it', async ({ page }) => {
    const name = uniqueName('E2E Sprint');

    await page.goto('/admin/sprints');
    await expect(page.getByRole('heading', { name: 'Sprints' })).toBeVisible();

    await page.getByRole('button', { name: 'Add sprint' }).click();
    // The create form shouldn't expose Status — new sprints always start Open.
    await expect(page.locator('#sprint-status')).toHaveCount(0);
    await page.locator('#sprint-name').fill(name);
    await page.locator('#sprint-start').fill('2026-10-01');
    await page.locator('#sprint-end').fill('2026-10-14');
    await page.getByRole('button', { name: 'Save' }).click();

    const row = page.locator('tr', { hasText: name });
    await expect(row).toBeVisible();
    await expect(row.getByText('2026-10-01 – 2026-10-14')).toBeVisible();
    await expect(row.getByText('Open')).toBeVisible();

    // Re-opening edit must show the same dates back (RFC3339 -> YYYY-MM-DD
    // round-trip through the backend, not just what was typed).
    await row.getByRole('button', { name: 'Edit' }).click();
    await expect(page.locator('#sprint-start')).toHaveValue('2026-10-01');
    await expect(page.locator('#sprint-end')).toHaveValue('2026-10-14');

    await page.locator('#sprint-status').selectOption('Closed');
    await page.getByRole('button', { name: 'Save' }).click();

    await expect(row.getByText('Closed')).toBeVisible();

    await row.getByRole('button', { name: 'Delete' }).click();
    await page.getByRole('alertdialog').getByRole('button', { name: 'Delete' }).click();

    await expect(page.locator('tr', { hasText: name })).toHaveCount(0);
  });
});
