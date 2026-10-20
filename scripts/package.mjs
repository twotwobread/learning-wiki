import { readFile, writeFile, mkdir, copyFile, chmod, readdir, rm } from 'node:fs/promises';
import { execFileSync } from 'node:child_process';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = dirname(dirname(fileURLToPath(import.meta.url)));
const dist = join(root, 'dist');
const main = JSON.parse(await readFile(join(root, 'npm/learning-wiki/package.json'), 'utf8'));
const targets = JSON.parse(await readFile(join(root, 'npm/learning-wiki/bin/targets.json'), 'utf8'));
if (!/^\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?$/.test(main.version)) throw new Error('Invalid release version');
const targetEntries = Object.entries(targets);
if (!targetEntries.length) throw new Error('No platform targets configured');
const packageNames = targetEntries.map(([key, target]) => {
  if (!/^[a-z0-9]+-[a-z0-9]+$/.test(key) || key !== `${target.platform}-${target.arch}`) throw new Error(`Invalid target key: ${key}`);
  const name = `learning-wiki-${key}`;
  const expectedPackage = `@suyoung9137/${name}`;
  if (target.package !== expectedPackage || !/^[a-z0-9]+$/.test(target.goos) || !/^[a-z0-9]+$/.test(target.goarch) || !/^learning-wiki(?:\.exe)?$/.test(target.executable)) throw new Error(`Invalid target metadata: ${key}`);
  return name;
});
if (new Set(packageNames).size !== packageNames.length) throw new Error('Duplicate platform package');
if (Object.keys(main.optionalDependencies).length !== packageNames.length) throw new Error('Optional dependency list differs from platform targets');
for (const [key, target] of targetEntries) {
  const pkg = JSON.parse(await readFile(join(root, `npm/learning-wiki-${key}/package.json`), 'utf8'));
  if (pkg.name !== target.package || pkg.version !== main.version || main.optionalDependencies[pkg.name] !== main.version || JSON.stringify(pkg.os) !== JSON.stringify([target.platform]) || JSON.stringify(pkg.cpu) !== JSON.stringify([target.arch])) {
    throw new Error(`Package metadata differs from platform target: ${key}`);
  }
}

// Only generated staging directories are reset, never source or user configuration.
await rm(join(dist, 'npm'), { recursive: true, force: true });
await mkdir(dist, { recursive: true });
for (const name of ['learning-wiki', ...packageNames]) {
  const target = join(dist, 'npm', name);
  await mkdir(join(target, 'bin'), { recursive: true });
  await copyFile(join(root, 'npm', name, 'package.json'), join(target, 'package.json'));
  await copyFile(join(root, 'docs/npm.md'), join(target, 'README.md'));
  await copyFile(join(root, 'LICENSE'), join(target, 'LICENSE'));
}
const launcher = join(dist, 'npm/learning-wiki/bin/learning-wiki.cjs');
await copyFile(join(root, 'npm/learning-wiki/bin/learning-wiki.cjs'), launcher);
await chmod(launcher, 0o755);
await copyFile(join(root, 'npm/learning-wiki/bin/targets.json'), join(dist, 'npm/learning-wiki/bin/targets.json'));

// Preserve dependency license notices alongside the distributed Go binary.
const goroot = execFileSync('go', ['env', 'GOROOT'], { cwd: root, encoding: 'utf8' }).trim();
const standardLicense = await readFile(join(goroot, 'LICENSE'), 'utf8');
for (const [key, target] of targetEntries) {
  const targetDir = join(dist, 'npm', `learning-wiki-${key}`);
  const binary = join(targetDir, 'bin', target.executable);
  const env = { ...process.env, GOOS: target.goos, GOARCH: target.goarch, CGO_ENABLED: '0' };
  execFileSync('go', ['build', '-trimpath', '-ldflags', `-s -w -X main.version=${main.version}`, '-o', binary, './cmd/learning-wiki'], {
    cwd: root, stdio: 'inherit', env
  });
  await chmod(binary, 0o755);
  let notices = '# Third-party notices\n\n## Go standard library/runtime\n\n' + standardLicense;
  const modules = execFileSync('go', ['list', '-deps', '-f', '{{if .Module}}{{if not .Module.Main}}{{.Module.Path}}|{{.Module.Version}}|{{.Module.Dir}}{{end}}{{end}}', './cmd/learning-wiki'], {
    cwd: root, encoding: 'utf8', env
  });
  for (const line of new Set(modules.trim().split('\n').filter(Boolean))) {
    const [name, version, dir] = line.split('|');
    if (!dir) throw new Error(`Module not downloaded: ${name}. Run go mod download first.`);
    const files = (await readdir(dir, { withFileTypes: true })).filter(f => f.isFile() && /^(license|copying|notice)([._-].*)?$/i.test(f.name));
    if (!files.length) throw new Error(`No root license notice found for ${name}; review before packaging`);
    notices += `\n\n## ${name} ${version}\n`;
    for (const file of files) notices += `\n### ${file.name}\n\n${await readFile(join(dir, file.name), 'utf8')}\n`;
  }
  await writeFile(join(targetDir, 'THIRD-PARTY-NOTICES.md'), notices);
}
const artifacts = [];
for (const name of [...packageNames, 'learning-wiki']) {
  const output = execFileSync('npm', ['pack', '--json', '--ignore-scripts', '--pack-destination', dist], { cwd: join(dist, 'npm', name), encoding: 'utf8' });
  const [packed] = JSON.parse(output);
  artifacts.push({ name: packed.name, version: packed.version, file: packed.filename, integrity: packed.integrity });
}
await writeFile(join(dist, 'artifacts.json'), JSON.stringify(artifacts, null, 2) + '\n');
console.log(JSON.stringify(artifacts, null, 2));
