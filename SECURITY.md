# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability in `ccs`, please report it by emailing **tony@virtuallytd.com**.

**Please do NOT open a public GitHub issue for security vulnerabilities.**

### What to Include

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if you have one)

### What to Expect

- **Acknowledgment:** Within 48 hours
- **Initial assessment:** Within 1 week
- **Fix timeline:** Depends on severity, but critical issues will be prioritized

## Security Considerations

### Credential Storage

`ccs` stores Claude Code session tokens in your system keyring:
- **macOS:** Keychain Access (encrypted at rest)
- **Linux:** Secret Service (gnome-keyring, KWallet)

Tokens are stored under the service name `"Claude Code"` and `"Claude Code - <profile-name>"`.

### Environment Variables

Vertex AI profiles source environment variables from `~/.claude/profiles/<name>/env.zsh`. These files should have restricted permissions:

```bash
chmod 600 ~/.claude/profiles/*/env.zsh
```

### State Files

The following files contain non-sensitive data but should still be protected:
- `~/.claude/.current-profile` — Plain text, just profile name
- `~/.claude/profiles/*/settings.json` — May contain MCP server URLs

### MCP Server Configurations

Be careful not to commit sensitive MCP server configurations (OAuth tokens, API keys) to public repositories. Keep profile directories in `~/.claude/profiles/` local only.

## Security Best Practices

1. **Don't commit profile directories** — They may contain sensitive configurations
2. **Use restrictive file permissions** — `chmod 600` for env.zsh files
3. **Audit MCP server configs** — Review settings.json before sharing
4. **Keep ccs updated** — Install security updates promptly

## Known Limitations

- **Windows not supported** — Keyring backend needed
- **Plain text env vars** — Vertex AI credentials in env.zsh are plain text (rely on file permissions and `gcloud auth`)
