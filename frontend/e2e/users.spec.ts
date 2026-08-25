import { test, expect } from '@playwright/test';
import { api, uniqueName } from './helpers';

test.describe('Admin / Users', () => {
  test('create, edit, and delete a user', async ({ page }) => {
    const name = uniqueName('E2E User');
    const editedName = `${name} Edited`;

    await page.goto('/admin/users');
    await expect(page.getByRole('heading', { name: 'Users' })).toBeVisible();

    await page.getByRole('button', { name: 'Add user' }).click();
    await page.locator('#user-name').fill(name);
    await page.locator('#user-role').selectOption('Lead');
    await page.getByRole('button', { name: 'Save' }).click();

    const row = page.locator('tr', { hasText: name });
    await expect(row).toBeVisible();
    await expect(row.getByText('Lead')).toBeVisible();

    await row.getByRole('button', { name: 'Edit' }).click();
    await page.locator('#user-name').fill(editedName);
    await page.locator('#user-role').selectOption('Developer');
    await page.getByRole('button', { name: 'Save' }).click();

    const editedRow = page.locator('tr', { hasText: editedName });
    await expect(editedRow).toBeVisible();
    await expect(editedRow.getByText('Developer')).toBeVisible();

    await editedRow.getByRole('button', { name: 'Delete' }).click();
    const confirmDialog = page.getByRole('alertdialog');
    await expect(confirmDialog.getByText(`Delete ${editedName}?`)).toBeVisible();
    await confirmDialog.getByRole('button', { name: 'Delete' }).click();

    await expect(page.locator('tr', { hasText: editedName })).toHaveCount(0);
  });

  test('shows a validation error for an empty name', async ({ page }) => {
    await page.goto('/admin/users');
    await page.getByRole('button', { name: 'Add user' }).click();
    // Leave #user-name empty; the browser's native `required` should block
    // submission, but if it doesn't, the backend rejects it and the error
    // surfaces inline rather than the dialog silently closing.
    await page.locator('#user-name').fill('');
    await page.getByRole('button', { name: 'Save' }).click();
    await expect(page.getByRole('heading', { name: 'Add user' })).toBeVisible();
  });
});
