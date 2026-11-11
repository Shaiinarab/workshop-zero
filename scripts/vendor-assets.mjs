import fs from 'fs';
import path from 'path';
import {execSync} from 'child_process';

const root = process.cwd();
const v = '4.0.0';
const base = `https://cdn.jsdelivr.net/npm/htmx.org@${v}/dist`;
const files = [
  ['htmx.min.js', `${base}/htmx.min.js`],
  ['ext/hx-sse.js', `${base}/ext/hx-sse.js`],
  ['ext/hx-live.js', `${base}/ext/hx-live.js`],
  ['ext/hx-preload.js', `${base}/ext/hx-preload.js`],
];
for (const [out, url] of files) {
  const dest = path.join(root, 'web/static/vendor/runtime', out);
  fs.mkdirSync(path.dirname(dest), {recursive:true});
  execSync(`curl -fsSL ${JSON.stringify(url)} -o ${JSON.stringify(dest)}`, {stdio:'inherit'});
}
const fontDir = path.join(root,'web/static/fonts');
fs.mkdirSync(fontDir,{recursive:true});
const fontCandidates = [
  'https://cdn.jsdelivr.net/gh/rastikerdar/vazirmatn@v33.003/fonts/webfonts/vazirmatn-variable.woff2',
  'https://github.com/rastikerdar/vazirmatn/releases/download/v33.003/vazirmatn-variable.woff2',
  'https://cdn.jsdelivr.net/npm/@fontsource-variable/vazirmatn/files/vazirmatn-arabic-wght-normal.woff2',
];
let ok = false;
for (const u of fontCandidates) {
  try {
    execSync(`curl -fsSL ${JSON.stringify(u)} -o ${JSON.stringify(path.join(fontDir,'vazirmatn-variable.woff2'))}`, {stdio:'inherit'});
    ok = true; break;
  } catch {}
}
if (!ok) throw new Error('could not download Vazirmatn font');
console.log('vendored assets');
