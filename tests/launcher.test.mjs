import { test } from 'node:test';
import assert from 'node:assert/strict';
import { createRequire } from 'node:module';
import { mkdtemp, mkdir, writeFile, copyFile, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { spawn, spawnSync } from 'node:child_process';
import { once } from 'node:events';

const require = createRequire(import.meta.url);
const { resolveBinary } = require('../npm/learning-wiki/bin/learning-wiki.cjs');

test('unsupported platforms fail with a semantic error before resolution', () => {
  assert.throws(() => resolveBinary('win32', 'x64'), { code: 'UNSUPPORTED_PLATFORM' });
  assert.throws(() => resolveBinary('linux', 'arm64'), { code: 'UNSUPPORTED_PLATFORM' });
});

test('selects the matching binary package for Linux x64 and macOS arm64', () => {
  for (const [platform, arch, expected] of [
    ['linux', 'x64', '@suyoung9137/learning-wiki-linux-x64/bin/learning-wiki'],
    ['darwin', 'arm64', '@suyoung9137/learning-wiki-darwin-arm64/bin/learning-wiki'],
  ]) {
    let requested;
    const result = resolveBinary(platform, arch, specifier => { requested = specifier; return '/fixture/learning-wiki'; });
    assert.equal(requested, expected);
    assert.equal(result, '/fixture/learning-wiki');
  }
});

test('missing platform package has an actionable error classification', () => {
  assert.throws(() => resolveBinary('linux', 'x64', () => { throw Object.assign(new Error('missing'), { code: 'MODULE_NOT_FOUND' }); }), { code: 'MISSING_BINARY_PACKAGE' });
});

async function fixture(t, body = 'process.stdout.write(JSON.stringify(process.argv.slice(2))); process.stderr.write("diagnostic"); process.exitCode = 7;') {
  const dir = await mkdtemp(join(tmpdir(), 'learning-wiki-launcher-'));
  t.after(() => rm(dir, { recursive: true, force: true }));
  const launcher = join(dir, 'bin/learning-wiki.cjs');
  await mkdir(join(dir, 'bin'));
  await copyFile(new URL('../npm/learning-wiki/bin/learning-wiki.cjs', import.meta.url), launcher);
  await copyFile(new URL('../npm/learning-wiki/bin/targets.json', import.meta.url), join(dir, 'bin/targets.json'));
  const packageName = process.platform === 'darwin' && process.arch === 'arm64'
    ? '@suyoung9137/learning-wiki-darwin-arm64'
    : '@suyoung9137/learning-wiki-linux-x64';
  const binaryDir = join(dir, 'node_modules', packageName);
  await mkdir(join(binaryDir, 'bin'), { recursive: true });
  await writeFile(join(binaryDir, 'package.json'), JSON.stringify({ name: packageName, version: '0.2.0' }));
  await writeFile(join(binaryDir, 'bin/learning-wiki'), `#!/usr/bin/env node\n${body}\n`, { mode: 0o755 });
  return launcher;
}

test('launcher preserves argument values, output streams and exit code', async t => {
  const launcher = await fixture(t);
  const args = ['--vault', 'C:\\내 문서\\Learning Wiki', ';echo injected', '$(touch unwanted)'];
  const result = spawnSync(process.execPath, [launcher, ...args], { encoding: 'utf8', timeout: 5000 });
  assert.equal(result.status, 7);
  assert.deepEqual(JSON.parse(result.stdout), args);
  assert.equal(result.stderr, 'diagnostic');
});

test('termination signal reaches the child and produces a signal exit status', { timeout: 10000 }, async t => {
  const launcher = await fixture(t, 'process.stdout.write("ready"); setTimeout(() => process.exit(99), 5000);');
  const child = spawn(process.execPath, [launcher], { stdio: ['ignore', 'pipe', 'pipe'] });
  t.after(() => { if (child.exitCode === null) child.kill('SIGKILL'); });
  const closed = once(child, 'close');
  await once(child.stdout, 'data');
  child.kill('SIGTERM');
  const [code, signal] = await closed;
  assert.equal(code, 143);
  assert.equal(signal, null);
});
