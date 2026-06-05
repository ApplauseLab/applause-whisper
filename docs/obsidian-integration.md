# Obsidian Integration

## Goal

When the Obsidian extension is installed, Yap appends every successful transcription to an Obsidian daily note while preserving the normal Yap flow: history, audio archive, stats, clipboard, and auto-paste.

## Integration Strategy

Use direct Markdown file writes into the selected Obsidian vault. Obsidian vaults are local folders of Markdown files and Obsidian watches filesystem changes automatically. This avoids requiring Obsidian to be open and avoids mandatory community plugin dependencies.

Alternatives considered:

- Native `obsidian://` URIs can open vaults and notes, but native append support is limited.
- Advanced URI and Actions URI can append content, but require community plugins.
- Local REST API is powerful, but requires a community plugin, API key, local server, and certificate handling.

## Note Layout

Yap writes to a dedicated `Yap` folder in the configured vault. The note name is configurable, and defaults to a daily note:

```text
<Vault>/Yap/Transcriptions YYYY-MM-DD.md
```

The default note name template is `Transcriptions {{date}}`, where `{{date}}` expands to `YYYY-MM-DD`. A static name such as `Inbox` writes to `<Vault>/Yap/Inbox.md`.

Example:

```text
~/Documents/Obsidian/MainVault/Yap/Transcriptions 2026-06-04.md
```

New daily note content:

```md
---
source: yap
type: transcriptions
date: 2026-06-04
---

# Transcriptions 2026-06-04

## 14:32

Transcribed text goes here.
```

Subsequent captures append to the same daily note:

```md
## 16:08

Another capture goes here.
```

## Implemented Behavior

- The Obsidian extension is installed and uninstalled from the Extensions page.
- The extension stores an enabled flag, the selected Obsidian vault path, and the configured note name.
- The note name always resolves to a Markdown file inside the `Yap` folder.
- When installed, every successful transcription must be written to Obsidian.
- Normal Yap behavior remains active: history, audio saving, stats, clipboard, and auto-paste still run after a successful Obsidian write.
- If Obsidian writing fails after transcription, Yap surfaces the error and keeps the transcript in app state.
- Uninstalling the extension disables Obsidian writes and clears the vault path.

## First Version Scope

Ship plugin-free direct vault writing. REST API and URI-based integrations can be added later if users need remote vault operations, command execution, or richer Obsidian automation.
