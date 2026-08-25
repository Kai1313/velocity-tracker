import { test, expect } from '@playwright/test';
import { uniqueName } from './helpers';

test.describe('Admin / Projects', () => {
  test('create, archive, and delete a project', async ({ page }) => {
    const name = uniqueName('E2E Project');

    await page.goto('/admin/projects');
    await expect(page.getByRole('heading', { name: 'Projects' })).toBeVisible();

    await page.getByRole('button', { name: 'Add project' }).click();
    // Create only takes a name — the status field shouldn't appear until edit.
    await expect(page.locator('#project-status')).toHaveCount(0);
    await page.locator('#project-name').fill(name);
    await page.getByRole('button', { name: 'Save' }).click();

    const row = page.locator('tr', { hasText: name });
    await expect(row).toBeVisible();
    await expect(row.getByText('Active')).toBeVisible();

    await row.getByRole('button', { name: 'Edit' }).click();
    await expect(page.locator('#project-status')).toBeVisible();
    await page.locator('#project-status').selectOption('Archived');
    await page.getByRole('button', { name: 'Save' }).click();

    await expect(row.getByText('Archived')).toBeVisible();

    await row.getByRole('button', { name: 'Delete' }).click();
    const confirmDialog = page.getByRole('alertdialog');
    await confirmDialog.getByRole('button', { name: 'Delete' }).click();

    await expect(page.locator('tr', { hasText: name })).toHaveCount(0);
  });
});
