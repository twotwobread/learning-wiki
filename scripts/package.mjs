import { readFile, writeFile, mkdir, copyFile, chmod, readdir, rm } from 'node:fs/promises';
import { execFileSync } from 'node:child_process';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = dirname(dirname(fileURLToPath(import.meta.url)));
const dist = join(root, 'dist');
const main = JSON.parse(await readFile(join(root, 'npm/learning-wiki/package.json'), 'utf8'));
const platform = JSON.parse(await readFile(join(root, 'npm/learning-wiki-linux-x64/package.json'), 'utf8'));
if (main.version !== platform.version || main.optionalDependencies[platform.name] !== platform.version) throw new Error('Package versions must match exactly');
if (!/^\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?$/.test(main.version)) throw new Error('Invalid release version');

// Only generated staging directories are reset, never source or user configuration.
await rm(join(dist, 'npm'), { recursive: true, force: true });
await mkdir(dist, { recursive: true });
for (const name of ['learning-wiki', 'learning-wiki-linux-x64']) {
  const target = join(dist, 'npm', name);
  await mkdir(join(target, 'bin'), { recursive: true });
  await copyFile(join(root, 'npm', name, 'package.json'), join(target, 'package.json'));
  await copyFile(join(root, 'docs/npm.md'), join(target, 'README.md'));
  await copyFile(join(root, 'LICENSE'), join(target, 'LICENSE'));
}
const launcher = join(dist, 'npm/learning-wiki/bin/learning-wiki.cjs');
await copyFile(join(root, 'npm/learning-wiki/bin/learning-wiki.cjs'), launcher);
await chmod(launcher, 0o755);
const binary = join(dist, 'npm/learning-wiki-linux-x64/bin/learning-wiki');
execFileSync('go', ['build', '-trimpath', '-ldflags', `-s -w -X main.version=${main.version}`, '-o', binary, './cmd/learning-wiki'], {
  cwd: root, stdio: 'inherit', env: { ...process.env, GOOS: 'linux', GOARCH: 'amd64', CGO_ENABLED: '0' }
});
await chmod(binary, 0o755);

// Preserve dependency license notices alongside the distributed Go binary.
const goroot = execFileSync('go', ['env', 'GOROOT'], { cwd: root, encoding: 'utf8' }).trim();
let notices = '# Third-party notices\n\n## Go standard library/runtime\n\n' + await readFile(join(goroot, 'LICENSE'), 'utf8');
const modules = execFileSync('go', ['list', '-deps', '-f', '{{if .Module}}{{if not .Module.Main}}{{.Module.Path}}|{{.Module.Version}}|{{.Module.Dir}}{{end}}{{end}}', './cmd/learning-wiki'], {
  cwd: root, encoding: 'utf8', env: { ...process.env, GOOS: 'linux', GOARCH: 'amd64', CGO_ENABLED: '0' }
});
for (const line of new Set(modules.trim().split('\n').filter(Boolean))) {
  const [name, version, dir] = line.split('|');
  if (!dir) throw new Error(`Module not downloaded: ${name}. Run go mod download first.`);
  const files = (await readdir(dir, { withFileTypes: true })).filter(f => f.isFile() && /^(license|copying|notice)([._-].*)?$/i.test(f.name));
  if (!files.length) throw new Error(`No root license notice found for ${name}; review before packaging`);
  notices += `\n\n## ${name} ${version}\n`;
  for (const file of files) notices += `\n### ${file.name}\n\n${await readFile(join(dir, file.name), 'utf8')}\n`;
}
await writeFile(join(dist, 'npm/learning-wiki-linux-x64/THIRD-PARTY-NOTICES.md'), notices);
const artifacts = [];
for (const name of ['learning-wiki-linux-x64', 'learning-wiki']) {
  const output = execFileSync('npm', ['pack', '--json', '--ignore-scripts', '--pack-destination', dist], { cwd: join(dist, 'npm', name), encoding: 'utf8' });
  const [packed] = JSON.parse(output);
  artifacts.push({ name: packed.name, version: packed.version, file: packed.filename, integrity: packed.integrity });
}
await writeFile(join(dist, 'artifacts.json'), JSON.stringify(artifacts, null, 2) + '\n');
console.log(JSON.stringify(artifacts, null, 2));
