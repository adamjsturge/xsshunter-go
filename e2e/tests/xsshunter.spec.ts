import { test, expect } from '@playwright/test';

const password = process.env.INITIAL_PASSWORD ?? '';

test('Logging in Successfully', async ({ page }) => {
  await page.goto('http://localhost:1449/');

  const text = await page.innerText('text=http://localhost:1449/js_callback');
  await page.goto('http://localhost:1449/admin');

  await page.getByLabel('Password:').fill(password);
  await page.getByRole('button', { name: 'Submit' }).click();
  await page.getByRole('heading', { name: 'Admin Page' }).click();

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

test('Trigger XSS', async ({ page }) => {
  await page.goto('about:blank');

  const customHTML = `
    <html>
      <body>
        <h1>Test</h1>
        <script src='http://localhost:1449'></script>
      </body>
    </html>
  `;

  await page.setContent(customHTML);

  await page.waitForTimeout(5000);

  await page.goto('http://localhost:1449/admin');

  await page.getByLabel('Password:').fill(password);
  await page.getByRole('button', { name: 'Submit' }).click();

  await expect(page.getByRole('rowgroup')).toContainText('about:blank');

  // await page.getByText('Expand').getByRole('button').click();
  // await page.getByText('URL: about:blank').click();
  // await expect(page.locator('body')).toContainText('about:blank');
  // await expect(page.getByRole('code')).toContainText('<html><head></head><body> <h1>Test</h1> <script src="http://localhost:1449"></script> </body></html>');
});