import { readFile } from 'node:fs/promises';
import { execFileSync } from 'node:child_process';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

function requireMetadata(condition, message) {
  if (!condition) throw Object.assign(new Error(message), { code: 'ERR_RELEASE_METADATA' });
}

export function validateRelease(tag, main, targets, platforms) {
  const match = typeof tag === 'string' && /^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$/.exec(tag);
  requireMetadata(match && match[0] === tag, 'Expected a stable release tag vX.Y.Z');
  const version = tag.slice(1);
  requireMetadata(main?.name === '@suyoung9137/learning-wiki' && main.version === version, 'Main package differs from release');
  const entries = Object.entries(targets ?? {});
  requireMetadata(entries.length > 0, 'No platform targets configured');
  const keys = entries.map(([key]) => key);
  requireMetadata(platforms && Object.keys(platforms).length === keys.length && Object.keys(platforms).every(key => keys.includes(key)), 'Platform manifest set differs from targets');
  const names = [];
  for (const [key, target] of entries) {
    requireMetadata(target && /^[a-z0-9]+-[a-z0-9]+$/.test(key) && key === `${target.platform}-${target.arch}`, `Invalid target key: ${key}`);
    requireMetadata(target.package === `${main.name}-${key}` && /^[a-z0-9]+$/.test(target.goos) && /^[a-z0-9]+$/.test(target.goarch) && target.executable === (target.platform === 'win32' ? 'learning-wiki.exe' : 'learning-wiki'), `Invalid target metadata: ${key}`);
    const pkg = platforms[key];
    requireMetadata(pkg?.name === target.package && pkg.version === version && Array.isArray(pkg.os) && pkg.os.length === 1 && pkg.os[0] === target.platform && Array.isArray(pkg.cpu) && pkg.cpu.length === 1 && pkg.cpu[0] === target.arch, `Platform package differs from target: ${key}`);
    names.push(target.package);
  }
  const dependencies = main.optionalDependencies;
  requireMetadata(dependencies && Object.keys(dependencies).length === names.length && names.every(name => dependencies[name] === version), 'Optional dependencies differ from platform targets');
  return version;
}

async function preflight(tag) {
  const root = dirname(dirname(fileURLToPath(import.meta.url)));
  const readJSON = async path => JSON.parse(await readFile(join(root, path), 'utf8'));
  const main = await readJSON('npm/learning-wiki/package.json');
  const targets = await readJSON('npm/learning-wiki/bin/targets.json');
  // Validate catalog keys before using them as directory names.
  for (const key of Object.keys(targets)) requireMetadata(/^[a-z0-9]+-[a-z0-9]+$/.test(key), `Invalid target key: ${key}`);
  const platforms = Object.fromEntries(await Promise.all(Object.keys(targets).map(async key => [key, await readJSON(`npm/learning-wiki-${key}/package.json`)])));
  const version = validateRelease(tag, main, targets, platforms);
  const git = args => execFileSync('git', args, { cwd: root, encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] });
  git(['diff', '--quiet']);
  git(['diff', '--cached', '--quiet']);
  if (git(['status', '--porcelain', '--untracked-files=normal']).trim()) throw Object.assign(new Error('Release checkout must be clean'), { code: 'ERR_RELEASE_CHECKOUT' });
  git(['merge-base', '--is-ancestor', 'HEAD', 'origin/main']);
  console.log(`Release preflight passed for ${version}`);
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  try {
    const [command, tag, ...extra] = process.argv.slice(2);
    if (command !== 'preflight' || !tag || extra.length) throw new Error('Usage: node scripts/release.mjs preflight vX.Y.Z');
    await preflight(tag);
  } catch (error) {
    console.error(error.message);
    process.exitCode = 1;
  }
}
