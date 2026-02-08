import puppeteer from 'puppeteer';
import { fileURLToPath } from 'url';
import path from 'path';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const htmlPath = path.join(__dirname, 'index.html');
const outPath = path.join(__dirname, 'browser_showcase.png');

const browser = await puppeteer.launch({
  headless: true,
  args: ['--no-sandbox', '--disable-setuid-sandbox'],
});
const page = await browser.newPage();
await page.setViewport({ width: 800, height: 600, deviceScaleFactor: 1 });
await page.goto(`file:///${htmlPath.replace(/\\/g, '/')}`, { waitUntil: 'networkidle0' });

// Screenshot just the #main div at exact 800x600
const el = await page.$('#main');
await el.screenshot({ path: outPath });

console.log(`Saved browser screenshot: ${outPath}`);
await browser.close();
