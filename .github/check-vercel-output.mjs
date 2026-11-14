import assert from 'node:assert/strict';
import { accessSync, readFileSync, readdirSync } from 'node:fs';

const output = '.vercel/output';
const config = JSON.parse(readFileSync(`${output}/config.json`, 'utf8'));
const functions = readdirSync(`${output}/functions`, { recursive: true })
  .filter(path => path.endsWith('.func'));

assert.equal(config.version, 3, 'Expected Build Output API v3');
assert.deepEqual(functions, ['api/index.func'], 'Expected only the api/index function');
accessSync(`${output}/functions/api/index.func/.vc-config.json`);

const rewrite = config.routes.find(route => route.dest === '/api/index.go');
assert.ok(rewrite, 'Expected the /api/index.go rewrite destination');
assert.equal(new RegExp(rewrite.src).source, new RegExp('^/api(.*)$').source,
  'Expected the /api(.*) rewrite source');

console.log('Verified api/index.func and /api(.*) -> /api/index.go routing.');
