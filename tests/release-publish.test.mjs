import test from 'node:test';
import assert from 'node:assert/strict';
import * as release from '../scripts/release.mjs';

const entries = [
  { name: '@scope/tool-linux-x64', version: '1.0.0', file: 'linux.tgz', integrity: 'sha512-platform' },
  { name: '@scope/tool', version: '1.0.0', file: 'main.tgz', integrity: 'sha512-main' },
];
const metadata = entry => ({ dist: { integrity: entry.integrity } });

test('publishes absent platform artifacts before main and confirms integrity', async () => {
  const registry = new Map();
  const published = [];
  await release.publishArtifacts(entries, '/tmp/dist', {
    lookup: async entry => registry.get(entry.name) ?? null,
    publish: async (path, entry) => { published.push(path); registry.set(entry.name, metadata(entry)); },
    wait: async () => {},
  });
  assert.deepEqual(published, ['/tmp/dist/linux.tgz', '/tmp/dist/main.tgz']);
});

test('skips existing artifacts only when integrity matches', async () => {
  await release.publishArtifacts(entries, '/tmp/dist', {
    lookup: async entry => metadata(entry),
    publish: async () => assert.fail('existing artifact published'),
    wait: async () => assert.fail('existing artifact waited'),
  });
});

test('rejects existing integrity mismatch before publishing', async () => {
  await assert.rejects(release.publishArtifacts(entries, '/tmp/dist', {
    lookup: async () => ({ dist: { integrity: 'different' } }),
    publish: async () => assert.fail('mismatched artifact published'),
    wait: async () => {},
  }), { code: 'ERR_RELEASE_INTEGRITY' });
});

test('waits for published artifact visibility before publishing main', async () => {
  let published = false;
  let visible = false;
  await release.publishArtifacts(entries, '/tmp/dist', {
    lookup: async entry => entry === entries[0] && !visible ? null : metadata(entry),
    publish: async () => { published = true; },
    wait: async () => { assert.equal(published, true); visible = true; },
  });
  assert.equal(visible, true);
});

test('retry after partial publication skips completed platform', async () => {
  const registry = new Map();
  const published = [];
  const deps = {
    lookup: async entry => registry.get(entry.name) ?? null,
    publish: async (path, entry) => {
      published.push(entry.name);
      if (entry === entries[1] && published.length === 2) throw Object.assign(new Error('network'), { code: 'ENETDOWN' });
      registry.set(entry.name, metadata(entry));
    },
    wait: async () => {},
  };
  await assert.rejects(release.publishArtifacts(entries, '/tmp/dist', deps), { code: 'ENETDOWN' });
  await release.publishArtifacts(entries, '/tmp/dist', deps);
  assert.deepEqual(published, [entries[0].name, entries[1].name, entries[1].name]);
});

test('rejects changed registry integrity after publication', async () => {
  let published = false;
  await assert.rejects(release.publishArtifacts(entries, '/tmp/dist', {
    lookup: async () => published ? { dist: { integrity: 'different' } } : null,
    publish: async () => { published = true; },
    wait: async () => {},
  }), { code: 'ERR_RELEASE_INTEGRITY' });
});

test('stops waiting after five minutes without publishing main', async () => {
  let elapsed = 0;
  const published = [];
  await assert.rejects(release.publishArtifacts(entries, '/tmp/dist', {
    lookup: async () => null,
    publish: async (_, entry) => published.push(entry.name),
    wait: async ms => { elapsed += ms; },
  }), { code: 'ERR_RELEASE_VISIBILITY' });
  assert.equal(elapsed, 300000);
  assert.deepEqual(published, [entries[0].name]);
});
