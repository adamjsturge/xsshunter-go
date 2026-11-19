import { test, expect } from '@playwright/test';
import { clearCookies } from '../helper';

test('Custom Frontend Directory', async ({ page, context }) => {
  // Test direct access to custom login page
  await page.goto('http://localhost:1449/admin');
  await expect(page.locator('#custom-login-marker')).toBeVisible();
  
  // Test static file access
  await page.goto('http://localhost:1449/static/test.html');
  await expect(page.locator('#custom-static-marker')).toBeVisible();
});
