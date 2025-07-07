import { test, expect } from '@playwright/test';
import { clearCookies, login, navigateToPayloadImporterExporter, navigateToPayloadMaker, navigateToPayloads, navigateToSettings, triggerXSS, triggerXSSWithCustomHTML, generateHTML } from '../helper';

const crypto = require('crypto');

test('Logging in Successfully', async ({ page, context }) => {
  await page.goto('http://localhost:1449/');
  await clearCookies(context);

  await login(page);
  await navigateToSettings(page);
  await navigateToPayloads(page);
  await navigateToPayloadMaker(page);
  await navigateToPayloadImporterExporter(page);
});

test('Correlation Trigger XSS', async ({ page, context }) => {
  await page.goto('http://localhost:1449/');
  await clearCookies(context);

  await login(page);
  await navigateToSettings(page);

  const randomInjectionKey = crypto.randomBytes(20).toString('hex');
  const randomRequest = crypto.randomBytes(20).toString('hex');

  await expect(page.locator('#CORRELATION_API_KEY')).toBeVisible();

  const CORRELATION_API_KEY = await page.locator('#CORRELATION_API_KEY').inputValue();

  const form_data = new FormData();
  form_data.append('owner_correlation_key', CORRELATION_API_KEY);
  form_data.append('injection_key', randomInjectionKey);
  form_data.append('request', randomRequest);

  const resp = await fetch(`http://localhost:1449/api/v1/record_injection`,
    {
      method: 'POST',
      body: form_data,
    },
  );

  await expect(resp.status).toBe(200);
  
  const injection_requests_id = await resp.text().then((text) => text.replace(/\r?\n|\r/g, ''));

  await triggerXSS(page, context, randomInjectionKey);
  await page.goto('http://localhost:1449/admin');

  // Added First because payload for some reason doubles in the pipeline and not sure why
  await page.locator(`button[id="injection-request-id-${injection_requests_id}"]`).first().click();
});

test('Basic Trigger XSS', async ({ page, context }) => {
  await page.goto('http://localhost:1449/');
  await clearCookies(context);

  const randomInjectionKey = crypto.randomBytes(20).toString('hex');
  await triggerXSS(page, context, randomInjectionKey);
  
  await page.goto('http://localhost:1449/admin');
  await clearCookies(context);
  await login(page);
  
  await page.locator('.action_button').filter({ hasText: 'Expand' }).first().click();
  
  await expect(page.locator('.modal_div')).toContainText(randomInjectionKey);
});

test('Update Settings', async ({ page, context }) => {
  await page.goto('http://localhost:1449/admin');
  await clearCookies(context);

  await login(page);
  await navigateToSettings(page);

  const randomString = crypto.randomBytes(20).toString('hex');

  // await page.locator('#CORRELATION_API_KEY').fill(randomString);
  await page.fill('#CORRELATION_API_KEY', randomString);
  await page.getByRole('button', { name: 'Save' }).click();
  await page.goto('http://localhost:1449/admin');
  await navigateToSettings(page);

  await expect(page.locator('#CORRELATION_API_KEY')).toHaveValue(randomString);
});

test('Create Payload', async ({ page, context }) => {
  await page.goto('http://localhost:1449/admin');
  await clearCookies(context);

  await login(page);
  await navigateToPayloadMaker(page);

  const randomPayload = crypto.randomBytes(20).toString('hex');
  const randomTitle = crypto.randomBytes(20).toString('hex');
  const randomDesc = crypto.randomBytes(20).toString('hex');
  const randomAuthor = crypto.randomBytes(20).toString('hex');

  await page.locator('#payload_input').fill(randomPayload + ` /script_hostname/`);
  await page.getByRole('textbox').nth(1).fill(randomTitle);
  await page.getByRole('textbox').nth(2).fill(randomDesc);
  await page.getByRole('textbox').nth(3).fill(randomAuthor);
  await page.getByRole('button', { name: 'Create Payload' }).click();
  await expect(page.locator('#payload_maker')).toContainText('Payload Added');
  await page.getByRole('button', { name: 'Payloads' }).click();
  await expect(page.getByText(randomPayload + ` localhost:1449`)).toBeVisible();
  await expect(page.getByText(randomTitle)).toBeVisible();
});

test('Basic Trigger XSS with a lot of HTML', async ({ page, context }) => {
  await page.goto('http://localhost:1449/');
  await clearCookies(context);

  const randomInjectionKey = crypto.randomBytes(20).toString('hex');
  let skeletonHTML = "<div id='addtional-text'>";

  let longPregeneratedHTML = generateHTML(200000, 250);

  skeletonHTML += longPregeneratedHTML + "</div>";

  await triggerXSS(page, context, randomInjectionKey, skeletonHTML);

  await page.waitForSelector('#addtional-text');
  let substringToCheck = longPregeneratedHTML.slice(-100);
  await expect(page.locator('#addtional-text')).toContainText(substringToCheck);

  await page.goto('http://localhost:1449/admin');
  await clearCookies(context);
  await login(page);
  
  await page.locator('.action_button').filter({ hasText: 'Expand' }).first().click();
  
  await expect(page.locator('.modal_div')).toContainText(randomInjectionKey);
});

test('Basic Trigger XSS with hidden HTML', async ({ page, context }) => {
  await page.goto('http://localhost:1449/');
  await clearCookies(context);

  const randomInjectionKey = crypto.randomBytes(20).toString('hex');
  let longPregeneratedHTML = `<div id='addtional-text' style='display: none;'>${generateHTML(500000, 1000)}</div>`;

  await triggerXSS(page, context, randomInjectionKey, longPregeneratedHTML);

  await page.goto('http://localhost:1449/admin');
  await clearCookies(context);
  await login(page);
  
  await page.locator('.action_button').filter({ hasText: 'Expand' }).first().click();
  
  await expect(page.locator('.modal_div')).toContainText(randomInjectionKey);
});

test('Trigger XSS in HTML without body tag', async ({ page, context }) => {
  await page.goto('http://localhost:1449/');
  await clearCookies(context);

  const randomInjectionKey = crypto.randomBytes(20).toString('hex');
  const htmlWithoutBody = `
    <!DOCTYPE html>
    <html>
      <head>
        <title>No Body Test</title>
        <script src='http://localhost:1449/${randomInjectionKey}'></script>
      </head>
    </html>
  `;

  await triggerXSSWithCustomHTML(page, context, htmlWithoutBody, randomInjectionKey);
  
  await page.goto('http://localhost:1449/admin');
  await clearCookies(context);
  await login(page);
  
  await page.locator('.action_button').filter({ hasText: 'Expand' }).first().click();
  
  await expect(page.locator('.modal_div')).toContainText(randomInjectionKey);
});

test('Failed Login Attempt', async ({ page, context }) => {
  await page.goto('http://localhost:1449/admin');
  await clearCookies(context);

  await page.getByLabel('Password:').fill('wrong_password');
  await page.getByRole('button', { name: 'Login' }).click();
  
  await expect(page.locator('#error')).toBeVisible();
  await expect(page.locator('#error')).toContainText('Invalid password');
  
  await expect(page.getByLabel('Password:')).toBeVisible();
});


test('Probe ID Source Tracking', async ({ page, context }) => {
  await page.goto('http://localhost:1449/');
  await clearCookies(context);

  const probeId = 'username' + crypto.randomBytes(10).toString('hex');
  await triggerXSS(page, context, probeId);
  
  await page.goto('http://localhost:1449/admin');
  await clearCookies(context);
  await login(page);
  
  // Wait for the payload fires table row to appear (excluding the header row)
  await page.waitForSelector('#payloadFiresTable tr.payload_fires_row', { timeout: 10000 });
  
  // Check if probe_id appears anywhere in the table
  await expect(page.locator('#payloadFiresTable')).toContainText(probeId);
  
  // More specifically, find the table row containing our probe_id and check it's in the third column
  const row = page.locator('#payloadFiresTable tr.payload_fires_row').filter({ hasText: probeId });
  await expect(row.locator('td:nth-child(3)')).toContainText(probeId);
  
  // Expand the payload fire to check the modal
  await page.locator('.action_button').filter({ hasText: 'Expand' }).first().click();
  
  // Verify probe_id appears in the modal
  await expect(page.locator('.modal_div')).toContainText(`Source: ${probeId}`);
});

test('XSS with Special Characters', async ({ page, context }) => {
  await page.goto('http://localhost:1449/');
  await clearCookies(context);

  const randomKey = crypto.randomBytes(10).toString('hex');
  const specialCharsHTML = `
    <!DOCTYPE html>
    <html>
      <head>
        <title>Special Characters Test</title>
      </head>
      <body>
        <h1>Special Characters: &lt;&gt;&amp;"'</h1>
        <div data-test="${randomKey}">
          <p>HTML entities: &copy; &reg; &euro; &hearts;</p>
          <p>Emojis: 😀 🚀 💻 🔥</p>
        </div>
        <script src='http://localhost:1449/${randomKey}'></script>
      </body>
    </html>
  `;

  await triggerXSSWithCustomHTML(page, context, specialCharsHTML, randomKey);

  await page.goto('http://localhost:1449/admin');
  await clearCookies(context);
  await login(page);
  
  await page.locator('.action_button').filter({ hasText: 'Expand' }).first().click();
  
  // Check that special characters were captured correctly
  await expect(page.locator('.modal_div')).toContainText('Special Characters:');
  await expect(page.locator('.modal_div')).toContainText('HTML entities:');
  await expect(page.locator('.modal_div')).toContainText('Emojis:');
});
