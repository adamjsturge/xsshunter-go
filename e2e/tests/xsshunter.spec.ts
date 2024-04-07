import { test, expect } from '@playwright/test';

require('dotenv').config();

const password = process.env.TEMP_E2E_PLAYWRIGHT_PASSWORD ?? '';

test('Logging in Successfully', async ({ page, context }) => {
  await page.goto('http://localhost:1449/');
  await context.clearCookies({ domain: 'localhost' });

  await page.goto('http://localhost:1449/admin');


  await page.getByLabel('Password:').fill(password);
  await page.getByRole('button', { name: 'Submit' }).click();

  await page.getByRole('button', { name: 'Settings' }).click();
  await expect(page.getByRole('heading')).toContainText('Settings');
  await expect(page.getByRole('rowgroup')).toContainText('PAGES_TO_COLLECT');
  await page.getByRole('button', { name: 'Payloads' }).click();
  await expect(page.locator('#payloads')).toContainText('Payloads');
  await page.getByRole('button', { name: 'Payload Maker' }).click();
  await expect(page.locator('#payloadsTable')).toContainText('Basic Payload');
  await expect(page.locator('#payload_maker')).toContainText('Payload Maker');
  await page.getByRole('button', { name: 'Payload Importer/Exporter' }).click();
  await expect(page.getByRole('heading')).toContainText('Payload Importer/Exporter');
});

test('Trigger XSS', async ({ page, context }) => {
  await page.goto('about:blank');

  await page.route('http://localhost:1449/', async (route) => {
    const response = await route.fetch();

    expect(response.status()).toBe(200);

    route.continue();
  });

  await context.route('**/js_callback', async (route) => {
    const response = await route.fetch();

    expect(response.status()).toBe(200);

    await route.continue();
  });

  const customHTML = `
    <html>
      <body>
        <h1>Test</h1>
        <script src='http://localhost:1449'></script>
      </body>
    </html>
  `;
  
  await page.setContent(customHTML);

  await expect(page.getByText('Test')).toBeVisible();  

  await page.goto('http://localhost:1449/admin');
  await context.clearCookies();

});