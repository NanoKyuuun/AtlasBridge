#!/usr/bin/env node

const { spawn, execSync } = require('child_process');
const path = require('path');
const os = require('os');
const fs = require('fs');

const APP_NAME = 'AtlasBridge';
const DEFAULT_PORT = '20127';
const API_BASE = `http://127.0.0.1:${DEFAULT_PORT}/v1`;
const ADMIN_BASE = `http://127.0.0.1:${DEFAULT_PORT}/admin`;

function getConfigDir() {
    if (process.platform === 'win32') {
        return path.join(process.env.APPDATA || '', APP_NAME);
    }
    return path.join(os.homedir(), '.config', APP_NAME);
}

function getBinaryPath() {
    const configDir = getConfigDir();
    const binaryName = process.platform === 'win32' ? 'atlasbridge.exe' : 'atlasbridge';
    return path.join(configDir, binaryName);
}

function findBinary() {
    const binaryPath = getBinaryPath();
    if (fs.existsSync(binaryPath)) {
        return binaryPath;
    }
    
    const localBinary = process.platform === 'win32' ? 'atlasbridge.exe' : './atlasbridge';
    if (fs.existsSync(localBinary)) {
        return path.resolve(localBinary);
    }
    
    return null;
}

function checkStatus() {
    try {
        const http = require('http');
        return new Promise((resolve) => {
            const req = http.get(API_BASE.replace('/v1', '/health'), (res) => {
                resolve(res.statusCode === 200 ? 'running' : 'unknown');
            });
            req.on('error', () => resolve('stopped'));
            req.setTimeout(2000, () => {
                req.destroy();
                resolve('stopped');
            });
        });
    } catch {
        return Promise.resolve('stopped');
    }
}

function openBrowser(url) {
    const start = process.platform === 'win32' ? 'start' : 
                 process.platform === 'darwin' ? 'open' : 'xdg-open';
    require('child_process').spawn(start, [url], { detached: true, stdio: 'ignore' });
}

async function cmdStart() {
    const binary = findBinary();
    if (!binary) {
        console.error('Error: Binary not found.');
        console.error('Please ensure atlasbridge is installed or exists in current directory.');
        process.exit(1);
    }
    
    console.log('Starting AtlasBridge...');
    const child = spawn(binary, [], { 
        detached: true, 
        stdio: 'ignore',
        cwd: path.dirname(binary)
    });
    child.unref();
    
    await new Promise(r => setTimeout(r, 1000));
    console.log('AtlasBridge started.');
    console.log(`API: ${API_BASE}`);
    console.log(`Admin: ${ADMIN_BASE}`);
}

async function cmdStatus() {
    const status = await checkStatus();
    console.log(`Status: ${status}`);
    if (status === 'running') {
        console.log(`API: ${API_BASE}`);
        console.log(`Admin: ${ADMIN_BASE}`);
    }
}

function cmdOpen() {
    openBrowser(ADMIN_BASE);
    console.log(`Opened ${ADMIN_BASE} in browser`);
}

function cmdVersion() {
    console.log('AtlasBridge v0.1.0');
}

function cmdTray() {
    console.log('Tray icon should be visible in system tray.');
}

function showHelp() {
    console.log(`
AtlasBridge CLI

Usage: atlasbridge <command>

Commands:
  start     Start the proxy (background)
  stop      Stop the proxy
  status    Show current status
  open      Open dashboard in browser
  tray      Show tray icon
  version   Show version
  help      Show this help message

Examples:
  atlasbridge start
  atlasbridge status
  atlasbridge open

For more info: https://github.com/atlasbridge/atlasbridge
`);
}

const command = process.argv[2] || 'help';

switch (command) {
    case 'start':
        cmdStart();
        break;
    case 'status':
        cmdStatus();
        break;
    case 'open':
        cmdOpen();
        break;
    case 'tray':
        cmdTray();
        break;
    case 'version':
        cmdVersion();
        break;
    case 'help':
    default:
        showHelp();
        break;
}
