import { test } from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, mkdir, writeFile, copyFile, readFile, readdir, rm, symlink } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { spawnSync } from 'node:child_process';

async function fixture(t) {
  const root = await mkdtemp(join(tmpdir(), 'learning-capture-'));
  t.after(() => rm(root, { recursive: true, force: true }));
  const vault = join(root, '한글 Vault');
  const skill = join(root, 'skill');
  await mkdir(join(vault, 'sessions'), { recursive: true });
  await mkdir(join(vault, 'templates'));
  await mkdir(join(skill, 'scripts'), { recursive: true });
  const script = join(skill, 'scripts/capture.mjs');
  await copyFile(new URL('../internal/resources/assets/capture.mjs', import.meta.url), script);
  const config = join(root, 'config.json');
  await writeFile(config, JSON.stringify({ version: 1, vault_path: vault }));
  await writeFile(join(skill, 'binding.json'), JSON.stringify({ config_path: config }));
  return { root, vault, config, run: (input, args = []) => spawnSync(process.execPath, [script, ...args], { input, encoding: 'utf8', cwd: tmpdir(), timeout: 5000 }) };
}

test('preserves content, returns actual path, never overwrites existing notes', async t => {
  const f = await fixture(t);
  const content = '# 모듈 학습\n\n정확한 내용 $HOME `command`\n';
  const first = f.run(content); assert.equal(first.status, 0, first.stderr);
  const path = JSON.parse(first.stdout).path;
  assert.equal(await readFile(path, 'utf8'), content);
  const second = f.run(content); assert.equal(second.status, 0, second.stderr);
  assert.notEqual(JSON.parse(second.stdout).path, path);
  assert.equal(await readFile(path, 'utf8'), content);
  assert.equal((await readdir(join(f.vault, 'sessions'))).length, 2);
});

test('reads the vault template without capturing a note', async t => {
  const f = await fixture(t); const template = '# Custom template\n';
  await writeFile(join(f.vault, 'templates/learning-note.md'), template);
  const result = f.run('', ['--template']); assert.equal(result.status, 0, result.stderr);
  assert.equal(result.stdout, template);
  assert.deepEqual(await readdir(join(f.vault, 'sessions')), []);
});

test('rejects empty, oversized and invalid UTF-8 input without writes', async t => {
  const f = await fixture(t);
  for (const input of ['  \n', 'x'.repeat(2 * 1024 * 1024 + 1), Buffer.from([0xff])]) {
    const result = f.run(input); assert.notEqual(result.status, 0); assert.equal(result.stdout, '');
    assert.deepEqual(await readdir(join(f.vault, 'sessions')), []);
  }
});

test('rejects symlinked sessions and missing configuration', async t => {
  const f = await fixture(t); const outside = join(f.root, 'outside');
  await mkdir(outside); await rm(join(f.vault, 'sessions'), { recursive: true });
  await symlink(outside, join(f.vault, 'sessions'));
  assert.notEqual(f.run('# Note').status, 0); assert.deepEqual(await readdir(outside), []);
  await rm(f.config); assert.notEqual(f.run('# Note').status, 0);
});
