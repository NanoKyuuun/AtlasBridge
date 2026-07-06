# Changelog

All notable changes to this release will be documented in this file.

## [0.1.0] - Initial Release

### Added
- OpenAI-compatible API endpoint at `/v1/chat/completions`
- Smart routing with task classification (coding, debugging, documentation, architecture)
- Route profiles (balanced, speed, cost, quality)
- Web UI dashboard at `/admin`
- System tray icon with context menu
- Runtime modes: Always On, Manual, Disabled
- Auto-start on Windows login
- Single-instance lock
- 8 smart aliases: smart-auto, smart-debug, smart-cheap, smart-docs, smart-architect, smart-code, smart-fast, smart-long-context

### Requirements
- 9Router running on port 20128
- Windows x64

### Known Issues
- Port conflict notification in tray needs manual resolution
- No installer (portable only)

### Quick Start
1. Extract portable ZIP
2. Run `smart-ai-proxy.exe`
3. Open http://127.0.0.1:20127/admin
4. Configure routing preferences