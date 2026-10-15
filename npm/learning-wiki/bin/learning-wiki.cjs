#!/usr/bin/env node
'use strict';

const { spawn } = require('node:child_process');
const { constants } = require('node:os');

function resolveBinary(platform = process.platform, arch = process.arch, resolve = require.resolve) {
  if (platform !== 'linux' || arch !== 'x64') {
    throw Object.assign(new Error(`Unsupported platform ${platform}/${arch}. Run inside WSL Linux x64, not Windows PowerShell.`), { code: 'UNSUPPORTED_PLATFORM' });
  }
  try {
    return resolve('@twotwobread/learning-wiki-linux-x64/bin/learning-wiki');
  } catch (cause) {
    throw Object.assign(new Error('Linux x64 binary package is missing. Reinstall with optional dependencies enabled (--include=optional).', { cause }), { code: 'MISSING_BINARY_PACKAGE' });
  }
}

function main() {
  let executable;
  try { executable = resolveBinary(); }
  catch (err) { console.error(err.message); process.exitCode = 1; return; }

  const child = spawn(executable, process.argv.slice(2), { stdio: 'inherit', shell: false });
  const handlers = new Map();
  for (const signal of ['SIGINT', 'SIGTERM', 'SIGHUP']) {
    const handler = () => { if (!child.killed) child.kill(signal); };
    handlers.set(signal, handler);
    process.on(signal, handler);
  }
  child.on('error', err => { console.error(`Unable to start learning-wiki: ${err.message}`); process.exitCode = 1; });
  child.on('close', (code, signal) => {
    for (const [name, handler] of handlers) process.removeListener(name, handler);
    process.exitCode = code ?? (signal ? 128 + (constants.signals[signal] ?? 1) : 1);
  });
}

module.exports = { resolveBinary };
if (require.main === module) main();
