import { readFile, writeFile, mkdtemp, mkdir, rm, stat } from 'node:fs/promises';
import { execFileSync, spawnSync } from 'node:child_process';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import assert from 'node:assert/strict';

const root = dirname(dirname(fileURLToPath(import.meta.url)));
const artifacts = JSON.parse(await readFile(join(root, 'dist/artifacts.json'), 'utf8'));
const temp = await mkdtemp(join(tmpdir(), 'learning-wiki-package-'));
try {
  const project = join(temp, 'unrelated-project');
  const home = join(temp, 'home');
  await mkdir(project); await mkdir(home);
  await writeFile(join(project, 'package.json'), JSON.stringify({ private: true }));
  const env = { ...process.env, HOME: home, XDG_CONFIG_HOME: join(home, 'config'), npm_config_cache: join(temp, 'npm-cache') };
  const packed = artifacts.map(a => join(root, 'dist', a.file));
  execFileSync('npm', ['install', '--offline', '--ignore-scripts', '--no-audit', '--no-fund', ...packed], { cwd: project, env, stdio: 'inherit' });
  const version = artifacts.find(a => a.name === '@twotwobread/learning-wiki').version;
  const output = execFileSync('npx', ['--offline', '--no-install', 'learning-wiki', '--version'], { cwd: project, env, encoding: 'utf8' });
  assert.ok(output.includes(version), `Expected version ${version}`);
  const launcher = join(project, 'node_modules/.bin/learning-wiki');
  for (const [args, code] of [[['--help'], 0], [['--invalid'], 2], [['init'], 2]]) {
    const result = spawnSync(launcher, args, { cwd: project, env, encoding: 'utf8', timeout: 10000 });
    assert.equal(result.status, code, result.stderr);
  }
  const vault = join(temp, '한글 Vault');
  const args = ['init', '--vault', vault, '--create'];
  for (let i = 0; i < 2; i++) {
    execFileSync('npx', ['--offline', '--no-install', 'learning-wiki', ...args], { cwd: project, env, stdio: 'inherit' });
  }
  assert.equal(JSON.parse(await readFile(join(home, 'config/learning-wiki/config.json'), 'utf8')).vault_path, vault);
  assert.ok((await stat(join(vault, 'templates/learning-note.md'))).isFile());
  console.log('Packed local install + npx + version + exit codes + repeated init: OK');
} finally {
  await rm(temp, { recursive: true, force: true });
}
