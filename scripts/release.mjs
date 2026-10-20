import { readFile, readdir, lstat } from 'node:fs/promises';
import { createHash } from 'node:crypto';
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
    requireMetadata(target.package === `${main.name}-${key}` && typeof target.goos === 'string' && /^[a-z0-9]+$/.test(target.goos) && typeof target.goarch === 'string' && /^[a-z0-9]+$/.test(target.goarch) && target.executable === (target.platform === 'win32' ? 'learning-wiki.exe' : 'learning-wiki'), `Invalid target metadata: ${key}`);
    const pkg = platforms[key];
    requireMetadata(pkg?.name === target.package && pkg.version === version && Array.isArray(pkg.os) && pkg.os.length === 1 && pkg.os[0] === target.platform && Array.isArray(pkg.cpu) && pkg.cpu.length === 1 && pkg.cpu[0] === target.arch, `Platform package differs from target: ${key}`);
    names.push(target.package);
  }
  const dependencies = main.optionalDependencies;
  requireMetadata(dependencies && Object.keys(dependencies).length === names.length && names.every(name => dependencies[name] === version), 'Optional dependencies differ from platform targets');
  return version;
}

function requireArtifact(condition, message) {
  if (!condition) throw Object.assign(new Error(message), { code: 'ERR_RELEASE_ARTIFACT' });
}

export async function verifyArtifacts(dir, version) {
  try {
    requireArtifact(typeof version === 'string' && /^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$/.exec(version)?.[0] === version, 'Invalid artifact release version');
    const root = dirname(dirname(fileURLToPath(import.meta.url)));
    const targets = JSON.parse(await readFile(join(root, 'npm/learning-wiki/bin/targets.json'), 'utf8'));
    const main = JSON.parse(await readFile(join(root, 'npm/learning-wiki/package.json'), 'utf8'));
    const names = [...Object.values(targets).map(target => target.package), main.name];
    const artifacts = JSON.parse(await readFile(join(dir, 'artifacts.json'), 'utf8'));
    requireArtifact(Array.isArray(artifacts) && artifacts.length === names.length, 'Artifact package set differs from targets');
    const files = new Set();
    for (const [index, artifact] of artifacts.entries()) {
      requireArtifact(artifact && typeof artifact === 'object' && !Array.isArray(artifact) && Object.keys(artifact).length === 4 && ['name', 'version', 'file', 'integrity'].every(key => Object.hasOwn(artifact, key)), 'Invalid artifact manifest schema');
      requireArtifact(artifact.name === names[index] && artifact.version === version, 'Artifact identity or order differs from release');
      requireArtifact(typeof artifact.file === 'string' && /^[a-zA-Z0-9][a-zA-Z0-9._-]*\.tgz$/.exec(artifact.file)?.[0] === artifact.file && !files.has(artifact.file), 'Unsafe or duplicate artifact filename');
      files.add(artifact.file);
      requireArtifact(typeof artifact.integrity === 'string' && /^sha512-[A-Za-z0-9+/]{86}==$/.exec(artifact.integrity)?.[0] === artifact.integrity, 'Invalid artifact integrity');
    }
    const actual = (await readdir(dir)).filter(file => file.endsWith('.tgz'));
    requireArtifact(actual.length === files.size && actual.every(file => files.has(file)), 'Tarball files differ from manifest');
    for (const artifact of artifacts) {
      const path = resolve(dir, artifact.file);
      requireArtifact((await lstat(path)).isFile(), 'Artifact must be a regular file');
      const integrity = 'sha512-' + createHash('sha512').update(await readFile(path)).digest('base64');
      requireArtifact(integrity === artifact.integrity, 'Artifact bytes differ from integrity');
      const entries = execFileSync('tar', ['-tzf', path], { encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] }).split('\n');
      requireArtifact(entries.filter(entry => entry === 'package/package.json').length === 1, 'Archive must contain one package manifest');
      const pkg = JSON.parse(execFileSync('tar', ['-xOzf', path, 'package/package.json'], { encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] }));
      requireArtifact(pkg?.name === artifact.name && pkg.version === version, 'Archived package differs from release');
    }
    return artifacts;
  } catch (cause) {
    if (cause.code === 'ERR_RELEASE_ARTIFACT') throw cause;
    throw Object.assign(new Error('Unable to verify release artifacts', { cause }), { code: 'ERR_RELEASE_ARTIFACT' });
  }
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
    if (command === 'preflight' && tag && extra.length === 0) {
      await preflight(tag);
    } else if (command === 'verify' && typeof tag === 'string' && /^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$/.exec(tag)?.[0] === tag && extra.length === 1) {
      await verifyArtifacts(extra[0], tag.slice(1));
      console.log(`Release artifacts verified for ${tag.slice(1)}`);
    } else {
      throw new Error('Usage: node scripts/release.mjs preflight vX.Y.Z | verify vX.Y.Z dist');
    }
  } catch (error) {
    console.error(error.message);
    process.exitCode = 1;
  }
}
