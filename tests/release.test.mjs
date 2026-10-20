import { test } from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, mkdir, writeFile, copyFile, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { execFileSync, spawnSync } from 'node:child_process';
import * as release from '../scripts/release.mjs';
import { createHash } from 'node:crypto';
import { readFile, readdir } from 'node:fs/promises';
import { validateRelease } from '../scripts/release.mjs';

function fixture() {
  const targets = {};
  const platforms = {};
  const main = { name: '@suyoung9137/learning-wiki', version: '1.2.3', optionalDependencies: {} };
  for (const [platform, arch, goarch] of [['linux', 'x64', 'amd64'], ['darwin', 'arm64', 'arm64'], ['win32', 'arm64', 'arm64']]) {
    const key = `${platform}-${arch}`;
    const name = `${main.name}-${key}`;
    targets[key] = { platform, arch, goos: platform === 'win32' ? 'windows' : platform, goarch, package: name, executable: platform === 'win32' ? 'learning-wiki.exe' : 'learning-wiki' };
    platforms[key] = { name, version: main.version, os: [platform], cpu: [arch] };
    main.optionalDependencies[name] = main.version;
  }
  return { main, targets, platforms };
}
const invalid = error => error instanceof Error && error.code === 'ERR_RELEASE_METADATA';
test('accepts stable release with every catalog target, including an additional OS', () => {
  const { main, targets, platforms } = fixture();
  assert.equal(validateRelease('v1.2.3', main, targets, platforms), '1.2.3');
});
for (const tag of ['1.2.3', 'v01.2.3', 'v1.02.3', 'v1.2.03', 'v1.2', 'v1.2.3-beta.1', 'v1.2.3+build', 'v1.2.3\n']) {
  test(`rejects invalid stable tag ${JSON.stringify(tag)}`, () => {
    const { main, targets, platforms } = fixture();
    assert.throws(() => validateRelease(tag, main, targets, platforms), invalid);
  });
}
for (const [name, change] of Object.entries({
  'main version': f => { f.main.version = '1.2.4'; },
  'main name': f => { f.main.name = '@other/tool'; },
  'target key': f => { f.targets['linux-x64'].arch = 'arm64'; },
  'target package': f => { f.targets['linux-x64'].package = '@other/tool'; },
  'target go metadata': f => { f.targets['linux-x64'].goos = '../linux'; },
  'missing target goos': f => { delete f.targets['linux-x64'].goos; },
  'null target goos': f => { f.targets['linux-x64'].goos = null; },
  'missing target goarch': f => { delete f.targets['linux-x64'].goarch; },
  'null target goarch': f => { f.targets['linux-x64'].goarch = null; },
  'target executable': f => { f.targets['linux-x64'].executable = '../tool'; },
  'empty catalog': f => { f.targets = {}; },
  'platform name': f => { f.platforms['linux-x64'].name = '@other/tool'; },
  'platform version': f => { f.platforms['linux-x64'].version = '1.2.4'; },
  'platform os': f => { f.platforms['linux-x64'].os = ['linux', 'darwin']; },
  'platform cpu': f => { f.platforms['linux-x64'].cpu = ['arm64']; },
  'missing platform': f => { delete f.platforms['linux-x64']; },
  'extra platform': f => { f.platforms.extra = {}; },
  'missing dependency': f => { delete f.main.optionalDependencies[f.targets['linux-x64'].package]; },
  'extra dependency': f => { f.main.optionalDependencies.extra = '1.2.3'; },
  'dependency version range': f => { f.main.optionalDependencies[f.targets['linux-x64'].package] = '^1.2.3'; },
})) {
  test(`rejects ${name} mismatch`, () => {
    const f = fixture();
    change(f);
    assert.throws(() => validateRelease('v1.2.3', f.main, f.targets, f.platforms), invalid);
  });
}

test('CLI preflight verifies isolated clean checkout and main ancestry without changing files', async t => {
  const root = await mkdtemp(join(tmpdir(), 'release ; test-'));
  t.after(() => rm(root, { recursive: true, force: true }));
  const env = { ...process.env, HOME: join(root, 'home'), GIT_CONFIG_NOSYSTEM: '1' };
  await mkdir(env.HOME);
  const git = (...args) => execFileSync('git', args, { cwd: root, env, stdio: 'pipe' });
  const { main, targets, platforms } = fixture();
  await mkdir(join(root, 'scripts'));
  await copyFile(new URL('../scripts/release.mjs', import.meta.url), join(root, 'scripts/release.mjs'));
  await mkdir(join(root, 'npm/learning-wiki/bin'), { recursive: true });
  await writeFile(join(root, 'npm/learning-wiki/package.json'), JSON.stringify(main));
  await writeFile(join(root, 'npm/learning-wiki/bin/targets.json'), JSON.stringify(targets));
  for (const [key, pkg] of Object.entries(platforms)) {
    await mkdir(join(root, `npm/learning-wiki-${key}`));
    await writeFile(join(root, `npm/learning-wiki-${key}/package.json`), JSON.stringify(pkg));
  }
  git('init', '-b', 'main');
  git('add', '.');
  git('-c', 'user.name=Test', '-c', 'user.email=test@example.invalid', 'commit', '-m', 'fixture');
  git('update-ref', 'refs/remotes/origin/main', 'HEAD');
  const run = (...args) => spawnSync(process.execPath, ['scripts/release.mjs', ...args], { cwd: root, env, encoding: 'utf8' });
  assert.equal(run('preflight', 'v1.2.3').status, 0);
  assert.notEqual(run('preflight', 'v1.2.3; touch injected').status, 0);
  assert.notEqual(run('unknown', 'v1.2.3').status, 0);
  await writeFile(join(root, 'untracked'), 'preserve');
  assert.notEqual(run('preflight', 'v1.2.3').status, 0);
  await rm(join(root, 'untracked'));
  await writeFile(join(root, 'npm/learning-wiki/package.json'), JSON.stringify({ ...main, description: 'changed' }));
  assert.notEqual(run('preflight', 'v1.2.3').status, 0);
  git('add', '.');
  assert.notEqual(run('preflight', 'v1.2.3').status, 0);
  git('-c', 'user.name=Test', '-c', 'user.email=test@example.invalid', 'commit', '-m', 'ahead');
  assert.notEqual(run('preflight', 'v1.2.3').status, 0);
  assert.equal(git('status', '--porcelain').toString(), '');
  git('update-ref', 'refs/remotes/origin/main', 'HEAD');
  assert.equal(run('preflight', 'v1.2.3').status, 0);
});


async function artifactFixture(t) {
  const dir = await mkdtemp(join(tmpdir(), 'release artifacts ; '));
  t.after(() => rm(dir, { recursive: true, force: true }));
  const targets = JSON.parse(await readFile(new URL('../npm/learning-wiki/bin/targets.json', import.meta.url)));
  const names = [...Object.values(targets).map(target => target.package), '@suyoung9137/learning-wiki'];
  const artifacts = [];
  async function pack(index, metadata) {
    const staging = join(dir, 'staging');
    await mkdir(join(staging, 'package'), { recursive: true });
    await writeFile(join(staging, 'package/package.json'), JSON.stringify(metadata));
    const file = `artifact-${index}.tgz`;
    execFileSync('tar', ['-czf', join(dir, file), '-C', staging, 'package']);
    await rm(staging, { recursive: true });
    return { name: names[index], version: '1.2.3', file, integrity: 'sha512-' + createHash('sha512').update(await readFile(join(dir, file))).digest('base64') };
  }
  for (const [index, name] of names.entries()) artifacts.push(await pack(index, { name, version: '1.2.3' }));
  const save = () => writeFile(join(dir, 'artifacts.json'), JSON.stringify(artifacts));
  await save();
  return { dir, artifacts, save, pack };
}
const artifactError = error => error instanceof Error && error.code === 'ERR_RELEASE_ARTIFACT';
test('verifies exact artifacts in platform-first order without changing files', async t => {
  const f = await artifactFixture(t);
  const before = await readdir(f.dir);
  assert.equal(typeof release.verifyArtifacts, 'function');
  assert.deepEqual(await release.verifyArtifacts(f.dir, '1.2.3'), f.artifacts);
  assert.deepEqual(await readdir(f.dir), before);
  const result = spawnSync(process.execPath, ['scripts/release.mjs', 'verify', 'v1.2.3', f.dir], { encoding: 'utf8' });
  assert.equal(result.status, 0, result.stderr);
  assert.notEqual(spawnSync(process.execPath, ['scripts/release.mjs', 'verify', '1.2.3', f.dir]).status, 0);
});
for (const [name, change] of Object.entries({
  'missing entry': f => { f.artifacts.pop(); },
  'duplicate package': f => { f.artifacts[1].name = f.artifacts[0].name; },
  'wrong order': f => { f.artifacts.reverse(); },
  'wrong version': f => { f.artifacts[0].version = '1.2.4'; },
  'unknown field': f => { f.artifacts[0].extra = true; },
  'missing field': f => { delete f.artifacts[0].integrity; },
  'unsafe filename': f => { f.artifacts[0].file = '../outside.tgz'; },
  'option filename': f => { f.artifacts[0].file = '--help.tgz'; },
  'wrong extension': f => { f.artifacts[0].file = 'file.tar'; },
  'duplicate filename': f => { f.artifacts[1].file = f.artifacts[0].file; },
  'wrong integrity': f => { f.artifacts[0].integrity = 'sha512-' + Buffer.alloc(64).toString('base64'); },
  'noncanonical integrity': f => { f.artifacts[0].integrity += '\n'; },
  'missing tarball': f => rm(join(f.dir, f.artifacts[0].file)),
  'extra tarball': f => writeFile(join(f.dir, 'extra.tgz'), 'extra'),
  'changed bytes': f => writeFile(join(f.dir, f.artifacts[0].file), 'corrupt'),
  'wrong archived name': async f => { f.artifacts[0] = await f.pack(0, { name: '@other/tool', version: '1.2.3' }); },
  'wrong archived version': async f => { f.artifacts[0] = await f.pack(0, { name: f.artifacts[0].name, version: '1.2.4' }); },
})) {
  test(`rejects artifact ${name}`, async t => {
    const f = await artifactFixture(t);
    await change(f);
    await f.save();
    assert.equal(typeof release.verifyArtifacts, 'function');
    await assert.rejects(() => release.verifyArtifacts(f.dir, '1.2.3'), artifactError);
  });
}
for (const manifest of [null, {}, [], [null]]) {
  test(`rejects malformed artifact manifest ${JSON.stringify(manifest)}`, async t => {
    const f = await artifactFixture(t);
    await writeFile(join(f.dir, 'artifacts.json'), JSON.stringify(manifest));
    assert.equal(typeof release.verifyArtifacts, 'function');
    await assert.rejects(() => release.verifyArtifacts(f.dir, '1.2.3'), artifactError);
  });
}
