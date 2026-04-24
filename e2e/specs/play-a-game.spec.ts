import { test, expect } from '@playwright/test';

test('user can search for a champion and submit a guess', async ({ page }) => {
  await page.goto('/');
  await expect(page.getByText('LOLIDLE')).toBeVisible();

  const searchBox = page.getByRole('textbox', { name: /buscar/i });
  await searchBox.fill('a');

  const firstOption = page.getByRole('option').first();
  await expect(firstOption).toBeVisible({ timeout: 10_000 });

  await searchBox.press('Enter');

  const firstRow = page.locator('table.guess-table tbody tr').first();
  await expect(firstRow).toBeVisible({ timeout: 10_000 });
});

test('champions list loads from backend', async ({ page }) => {
  await page.goto('/');

  const searchBox = page.getByRole('textbox', { name: /buscar/i });
  await searchBox.fill('y');

  await expect(page.getByRole('option').first()).toBeVisible({ timeout: 10_000 });
});
