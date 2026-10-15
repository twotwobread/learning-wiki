import { readFile, realpath, lstat, open, unlink } from 'node:fs/promises';
import { dirname, join, isAbsolute } from 'node:path';
import { fileURLToPath } from 'node:url';
import { randomUUID } from 'node:crypto';

async function main() {
  const args = process.argv.slice(2);
  if (args.length && !(args.length === 1 && args[0] === '--template')) {
    throw new Error('Usage: capture.mjs [--template]; Markdown is read from stdin');
  }
  const skillDir = dirname(dirname(fileURLToPath(import.meta.url)));
  const binding = JSON.parse(await readFile(join(skillDir, 'binding.json'), 'utf8'));
  if (typeof binding.config_path !== 'string' || !isAbsolute(binding.config_path)) throw new Error('Invalid configuration binding');
  const config = JSON.parse(await readFile(binding.config_path, 'utf8'));
  if (config.version !== 1 || typeof config.vault_path !== 'string' || !isAbsolute(config.vault_path)) throw new Error('Invalid local vault configuration; run init');
  const root = await realpath(config.vault_path);
  if (args[0] === '--template') {
    const template = join(root, 'templates', 'learning-note.md');
    if (await realpath(template) !== template) throw new Error('Template symlinks are not supported');
    process.stdout.write(await readFile(template, 'utf8'));
    return;
  }
  const sessions = join(root, 'sessions');
  if (!(await lstat(sessions)).isDirectory() || await realpath(sessions) !== sessions) throw new Error('sessions must be a real directory in the vault');
  const chunks = []; let size = 0;
  for await (const chunk of process.stdin) {
    size += chunk.length;
    if (size > 2 * 1024 * 1024) throw new Error('Note exceeds 2 MiB limit');
    chunks.push(chunk);
  }
  const content = new TextDecoder('utf-8', { fatal: true }).decode(Buffer.concat(chunks));
  if (!content.trim()) throw new Error('Empty note refused');
  const name = `${new Date().toISOString().replaceAll(':', '-')}-${randomUUID()}.md`;
  const path = join(sessions, name);
  const file = await open(path, 'wx', 0o600);
  try {
    await file.writeFile(content, 'utf8');
    await file.close();
  } catch (err) {
    await file.close().catch(() => {});
    await unlink(path).catch(() => {});
    throw err;
  }
  process.stdout.write(JSON.stringify({ path }) + '\n');
}
main().catch(err => { console.error(err.message); process.exitCode = 1; });
