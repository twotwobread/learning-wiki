import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';

test('main package and platform packages agree on supported targets and versions', async () => {
  const main = JSON.parse(await readFile(new URL('../npm/learning-wiki/package.json', import.meta.url)));
  const targets = JSON.parse(await readFile(new URL('../npm/learning-wiki/bin/targets.json', import.meta.url)));
  const expected = {
    'linux-x64': { platform: 'linux', arch: 'x64', goos: 'linux', goarch: 'amd64', package: '@suyoung9137/learning-wiki-linux-x64', executable: 'learning-wiki' },
    'darwin-arm64': { platform: 'darwin', arch: 'arm64', goos: 'darwin', goarch: 'arm64', package: '@suyoung9137/learning-wiki-darwin-arm64', executable: 'learning-wiki' },
  };
  assert.deepEqual(targets, expected);
  for (const [name, target] of Object.entries(expected)) {
    const pkg = JSON.parse(await readFile(new URL(`../npm/learning-wiki-${name}/package.json`, import.meta.url)));
    assert.equal(pkg.name, target.package);
    assert.equal(pkg.version, main.version);
    assert.deepEqual(pkg.os, [target.platform]);
    assert.deepEqual(pkg.cpu, [target.arch]);
    assert.equal(main.optionalDependencies[pkg.name], main.version);
  }
  assert.deepEqual(Object.keys(main.optionalDependencies).sort(), Object.values(expected).map(target => target.package).sort());
});
