import { expect } from '@playwright/test';
const path = require('path');

require('dotenv').config({ path: path.resolve(__dirname, '.env') });

const password = process.env.TEMP_E2E_PLAYWRIGHT_PASSWORD ?? '';

export async function login(page) {
  await page.goto('http://localhost:1449/admin');
  await page.getByLabel('Password:').fill(password);
  await page.getByRole('button', { name: 'Submit' }).click();
}

export async function navigateToSettings(page) {
  await page.getByRole('button', { name: 'Settings' }).click();
  await expect(page.getByRole('heading', { level: 1})).toContainText('Settings');
  await expect(page.getByRole('rowgroup')).toContainText('PAGES_TO_COLLECT');
}

export async function navigateToPayloads(page) {
  await page.getByRole('button', { name: 'Payloads' }).click();
  await expect(page.locator('#payloadsTable')).toContainText('Basic Payload');
  await expect(page.locator('#payloads')).toContainText('Payloads');
}

export async function navigateToPayloadMaker(page) {
  await page.getByRole('button', { name: 'Payload Maker' }).click();
  await expect(page.locator('#payload_maker')).toContainText('Payload Maker');
}

export async function navigateToPayloadImporterExporter(page) {
  await page.getByRole('button', { name: 'Payload Importer/Exporter' }).click();
  await expect(page.getByRole('heading')).toContainText('Payload Importer/Exporter');
}

export async function clearCookies(context) {
  await context.clearCookies({ domain: 'localhost' });
}

export async function triggerXSS(page, context, randomInjectionKey = "", longPregeneratedHTML = "") {
  await page.goto('about:blank');

  // page.on('request', request => console.log('>>', request.method(), request.url()));
  // page.on('response', response => console.log('<<', response.status(), response.url())); 

  const customHTML = `
  <html>
    <body>
      <h1>Test XSS Payload</h1>
      <script src='http://localhost:1449/${randomInjectionKey}'></script>
      ${longPregeneratedHTML}
    </body>
  </html>
`;

  const responsePromise = page.waitForResponse('**/js_callback');

  await page.setContent(customHTML);

  await responsePromise;

  await expect(page.getByText('Test XSS Payload')).toBeVisible();
}

export function generateHTML(length, lineBreakLength) {
  let charOptions = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ";
  let longPregeneratedHTML = "";

  for (let i = 0; i < length; i++) {
    longPregeneratedHTML += charOptions.charAt(Math.floor(Math.random() * charOptions.length));
    if (i % lineBreakLength == 0) {
      longPregeneratedHTML += "\n<br>\n";
    }
  }

  return longPregeneratedHTML;
}