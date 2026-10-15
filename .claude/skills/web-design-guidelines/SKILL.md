---
name: web-design-guidelines
description: Review UI code for Web Interface Guidelines compliance. Use when asked to "review my UI", "check accessibility", "audit design", "review UX", or "check my site against best practices".
metadata:
  argument-hint: <file-or-pattern>
---

# Web Interface Guidelines

Review files for compliance with Web Interface Guidelines.

## How it works

1. Read [rules.md](rules.md), the vendored guidelines. Do not fetch them from the
   network: the copy here is pinned so a review is reproducible.
2. Read the specified files (or ask the user for files or a pattern).
3. Check them against every rule in `rules.md`.
4. Output findings in the terse `file:line` format that `rules.md` specifies.

Read [C2.md](C2.md) first: it lists where this project deliberately departs from the
rules, so those are not reported as defects.

If no files are specified, ask the user which files to review.
