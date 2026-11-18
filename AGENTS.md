## Agent skills

### Issue tracker

Issues and specs live in this repository's GitHub Issues. See `docs/agents/issue-tracker.md`.

### Triage labels

Use `needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, and `wontfix`. See `docs/agents/triage-labels.md`.

### Domain docs

This is a single-context repository. See `docs/agents/domain.md`.
## Version

- `VERSION` holds this repository's own semantic version; the plugin reports it as registration metadata.
- Every requirement that changes this repository ships a version change: run `D:\WILL\AGENT\CPA\ops\update-versions.ps1 -BumpChanged` before committing, or `-BumpEx minor|major` / `-ExVersion x.y.z` when a patch is too small.
- `D:\WILL\AGENT\CPA\ops\update-versions.ps1 -Check` must pass before the task is reported finished.
- Never edit the `version_generated.go` file by hand; a test guards the shape.
