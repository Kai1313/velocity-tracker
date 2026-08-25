import { test, expect } from '@playwright/test';

test.describe('Theme toggle and cross-section navigation', () => {
  test('cycles System -> Light -> Dark -> System on /admin', async ({ page }) => {
    await page.goto('/admin/users');
    const toggle = page.locator('aside button[title]');

    await expect(toggle).toHaveAttribute('title', 'System theme');
    await toggle.click();
    await expect(toggle).toHaveAttribute('title', 'Light theme');
    await toggle.click();
    await expect(toggle).toHaveAttribute('title', 'Dark theme');
    await expect(page.locator('html')).toHaveClass(/dark/);
    await toggle.click();
    await expect(toggle).toHaveAttribute('title', 'System theme');
  });

  test('"Manage data" link on /dashboard navigates to /admin/users', async ({ page }) => {
    await page.goto('/dashboard');
    await page.getByText('Manage data').click();
    await expect(page).toHaveURL(/\/admin\/users$/);
  });
});
