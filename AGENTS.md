# AGENTS.md — working on cmd_and_ctrl

This file is for AI coding agents (Claude Code, Codex, Cursor, etc.) operating in
this repository. Humans should read [PLAN.md](PLAN.md) first for the project
vision and phased roadmap; this file is about *how* to work, not *what* to build.

---

## 1. What this project is

A private, personal client for 4-player Magic: The Gathering Commander, layered
on top of the XMage rules engine. Personal-use only — not a product. See
[PLAN.md](PLAN.md) for scope, stack, and the Option A vs B decision.

Three hard problems, ranked: **rules engine** (delegated to XMage), **multiplayer
state sync** (inherited from XMage's protocol), **UX polish** (this is where all
original work goes).

---

## 2. Ground rules

1. **Read [PLAN.md](PLAN.md) before making architectural suggestions.** The
   central decision (reuse XMage vs build from scratch) is already made. Do not
   relitigate it without new information.
2. **This is a hobby project at ~10 hours/week.** Optimise for momentum and
   clarity, not enterprise rigour. No microservices, no k8s, no premature
   abstractions. One VPS. One database. One deployable per service.
3. **Personal use only changes the calculus.** Legal risk is low (Cockatrice and
   XMage exist). Scale is ≤8 users. Don't build for hypothetical public launch.
4. **UX polish is the entire point.** If a task is "make rules work" the answer
   is almost always "delegate to XMage". If a task is "make it feel good"
   that's core product work and deserves real care.
5. **Don't touch XMage internals unless explicitly in scope.** Fork discipline:
   the protocol bridge is the seam. Keep Java ugliness on one side of it.

---

## 3. Repo layout (evolving)

```
cmd_and_ctrl/
├── PLAN.md              # vision, roadmap, open decisions
├── AGENTS.md            # this file
├── .devcontainer/       # Java 21 + Go + Node dev environment
├── bridge/              # protocol bridge (TS or Go) — not yet created
├── client/              # web client (React/Svelte + PixiJS) — not yet created
├── xmage/               # XMage submodule or vendored fork — not yet created
├── scripts/             # one-off tools, Scryfall pipeline, etc.
└── docs/                # research notes, protocol capture, decision records
```

When you create a new top-level directory, add it here.

---

## 4. Sprint and tracking discipline

Work is organised into 2-week sprints tracked in:

- **Project board:** https://github.com/users/krakenhavoc/projects/5
- **Milestones:** https://github.com/krakenhavoc/cmd_and_ctrl/milestones
- **Sprint plan with task checklists:** [docs/sprints.md](docs/sprints.md)

Each sprint has one tracking issue (`#1` through `#12`) whose body is the
checklist of sub-tasks. Commits reference the sprint issue number.

**Every commit and pull request must reference the sprint and issue it relates
to.** This is non-negotiable — it's how we keep a part-time, multi-month project
coherent.

### Commit message format

```
<type>(<scope>): <subject>

<body>

Sprint: S<NN> — <sprint name>
Issue: #<issue_number>
```

Example:

```
feat(bridge): add websocket handshake scaffolding

Thin TS service accepting WS connections on :8080 and echoing frames.
No XMage protocol yet — this is just the seam.

Sprint: S02 — Protocol bridge foundations
Issue: #14
```

`<type>` is one of: `feat`, `fix`, `chore`, `docs`, `refactor`, `test`, `spike`.
`<scope>` is the top-level dir being touched (`bridge`, `client`, `xmage`, etc.).

### Pull request format

PR descriptions must include:

```
## Summary
<1-3 bullets>

## Sprint
S<NN> — <sprint name>

## Issues
Closes #<n>, relates to #<n>

## Test plan
<bulleted checklist>
```

If a PR does not belong to the active sprint, say so explicitly and justify it.

---

## 5. Commands you'll actually run

*(Populate as the project takes shape. Empty sections are fine — don't invent.)*

### Dev environment

The devcontainer already installs Java 21, Maven, Go, Node, and the GitHub CLI.
Ports 3000, 5173, 8080, and 17171 are forwarded.

### XMage
- Build: *(TBD — captured during Phase 0 spike)*
- Run server: *(TBD)*
- Run legacy client: *(TBD)*

### Bridge
- *(TBD — created in Sprint 2)*

### Client
- *(TBD — created in Sprint 4)*

---

## 6. When you're unsure

1. Re-read [PLAN.md](PLAN.md) section 2 (architectural decision) and section 6
   (roadmap).
2. Check the current sprint in [docs/sprints.md](docs/sprints.md) — the scope
   there is authoritative for "what should I be working on right now".
3. Look for an open decision in [PLAN.md](PLAN.md) section 7. If the question is
   listed there, surface it to the user rather than guessing.
4. Prefer a spike (time-boxed, throwaway) over speculative architecture.

---

## 7. Things to explicitly *not* do

- Do not add CI/CD beyond basic lint + test until there's code to protect.
- Do not introduce a database, auth provider, or payment anything without
  discussion. Shared password is fine for now (see [PLAN.md](PLAN.md#4-tech-stack-assuming-option-a)).
- Do not build a full rules engine. That is Option C in PLAN.md, and the plan
  explicitly rejects it.
- Do not generate card art, card text, or anything else that would pull this
  project out of "private, personal use" territory.
- Do not create commits or PRs that do not reference a sprint (see section 4).
