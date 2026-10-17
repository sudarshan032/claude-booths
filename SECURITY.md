# Security policy

## Supported versions

Only the latest release gets security fixes. Upgrade by running the install script again.

## Reporting a vulnerability

Please don't open a public issue. Report it privately through GitHub:
**[Report a vulnerability](https://github.com/sudarshan032/claude-booths/security/advisories/new)**
(Security tab → "Report a vulnerability").

Include the `claude-booths version` output, your OS and tmux version, and steps to reproduce.
You'll get a reply within 7 days. Once a fix is released, the advisory is published with credit
to you unless you'd rather stay anonymous.

## Scope

The `claude-booths` binary, `install.sh`, and the release artifacts on GitHub. Bugs in the
agents it runs (Claude Code, Codex, Gemini CLI, ...) or in tmux belong to those projects.

## Verifying a download

Every release has a `checksums.txt` and a GitHub build provenance attestation:

```bash
shasum -a 256 -c checksums.txt --ignore-missing
gh attestation verify claude-booths_<version>_<os>_<arch>.tar.gz -R sudarshan032/claude-booths
```
