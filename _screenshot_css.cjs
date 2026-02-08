const puppeteer = require('puppeteer');
(async () => {
  const browser = await puppeteer.launch({
    headless: 'new',
    args: ['--force-device-scale-factor=1', '--no-sandbox'],
  });
  const page = await browser.newPage();
  await page.setViewport({ width: 1070, height: 780, deviceScaleFactor: 1 });
  await page.goto('file:///E:/works/ebitenui-xml/output/reference.html', { waitUntil: 'networkidle0', timeout: 30000 });
  await new Promise(r => setTimeout(r, 2000));
  await page.screenshot({ path: 'E:\\works\\ebitenui-xml\\output\\browser.png', type: 'png', fullPage: true });
  await browser.close();
})();
