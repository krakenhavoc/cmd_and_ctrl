# ADR 0224 — Splitting the card catalog package: a core, card shards and test packages

**Status:** Accepted · 2026-11-19 · S82 — Card coverage: sets outside Standard, newest first (#5906 is an unscheduled P1 CI bug; it rides the active sprint)
**Issues:** [#5906](https://github.com/krakenhavoc/cmd_and_ctrl/issues/5906)
**Owner decisions:** option A of the #5906 spike (`spike/5906-options`, `OPTIONS.md`), done by a generated `go/types` rewrite; card files reach the core through a **qualified** import (`effects.Register(effects.Spec{…})`), never a dot import; card work never stops (no freeze), and a roll-up merge waits at most an hour or two while the generated PR runs CI; a file added in the old place must fail CI with one exact command that fixes it.
**Numbering:** after `git fetch --all --prune`, `git ls-tree --name-only <ref> docs/decisions/` over all 240 remote branches. The highest number anywhere was **0223** (develop's mutate ADR), so this is **0224**.
**Builds on:** [ADR 0148](0148-capacity-limits-for-a-public-launch.md) (one VM, a hobby budget: no paid runners), #2766 (the sharded `server-test` job and `scripts/ci-test-shard.sh`), #522 and ADR 0041 P9 (snapshot fixtures and effect keys are on-disk identities).

---

## Context

`server/internal/cards/effects` held the whole catalog: 18,262 source files (17,193 of them one card each) and 2,940 test files, one Go package. Go compiles a package as one unit, and its test variant compiles the sources and the tests together. In November 2026 that unit needed more memory than the machines have:

| measured on develop (4 cores, 13.35 GiB cgroup) | largest process | result |
|---|---|---|
| `go build ./internal/cards/effects/` | 7.05 GiB | ok |
| `go test -race -cover -run XXX ./internal/cards/effects/` | 13.84 GB at the kill | OOM-killed |
| `go test -race -run XXX` (what CI ran, `-cover` already dropped) | 13.24 GiB | OOM-killed |
| `go test -run XXX` (no race, no cover) | 13.13 GiB | OOM-killed |

CI's 16 GB hosted runner still fit the no-cover build by about 2 GB, and every card batch ate into that. Every other package whose tests link the catalog compiled it with `-race` as one 6 GiB unit too, so the `rest` shard peaked at 12.8 GiB. A bigger runner costs money (ADR 0148 and PLAN.md §3 say no), and dropping `-race` would have bought little: plain `go test` no longer fit either.

The spike measured three options. B (move only the tests out) leaves the part that grows with every batch in one unit. C (GOGC, flags) is a stopgap. A splits the package itself.

What makes any split hard is the helpers: 13,984 unexported top-level names, 4,400 used in another file, and 11,301 of the 17,193 card files using at least one of them. No hand migration moves that; a type-checked rewriter does it in about a minute.

## Decision

### 1. The layout

```
internal/cards/effects/               package effects      the core: Spec, Register, Lookup, All, every seam and shared
                                                           helper, and the card files listed in core_cards.txt
internal/cards/effects/cards/         package cards        the aggregator: blank-imports every shard (generated)
internal/cards/effects/cards/<id>/    package cards<id>    card files only: one init() each, nothing another file uses
internal/cards/effects/effectstest/   package effectstest  test helpers: declarations lifted out of _test.go files, exported
internal/cards/effects/tests/<id>/    package tests<id>    _test.go files only, plus a generated main_test.go
```

Dependencies point one way: `tests/* → effectstest → cards/* → effects → game, legal, …`, and every test package also imports the aggregator, because card tests look cards up by name at run time.

- **A card file is placed by its name** through the frozen range table `internal/cards/effects/shards.txt`: `card mi mi` means a card file whose name sorts from `mi` up to the next row goes in `cards/mi/`. Tests are placed the same way (`test m m`). The table was sized on develop at 2af192538 for at most 600 card files a shard (37 shards) and 150 test files a package (23 packages). It changes only in a PR whose purpose is to change it, when a shard passes about 1,000 card files or a test package about 300.
- **A shard file declares nothing another source file uses.** A card file whose declarations another card or core file reads stays in the core (about 1,740 core files: the framework plus such card files), and is listed in `core_cards.txt`. New shared helpers go in a core file with no `init()`, exported. Moving the listed card files' helpers out, so they can move to their shards, is follow-up hygiene, not part of the split.
- **Qualified, never dot-imported.** A shard file says `effects.Register(effects.Spec{…})` and `effects.OnBattlefield(…)`; a test says `effectstest.NewCatalogGame(t)`. Collisions are impossible across packages, and grep finds every use of a helper by its qualified name.
- **Exported names.** Every unexported name that ends up used across a package boundary is exported by upper-casing its first letter (`onBattlefield` → `OnBattlefield`), about 13,000 in all. Where that name is already taken, `scripts/split-effects/overrides.txt` holds a reviewed name (20 at the split, for example `hasSubtype` → `CardHasSubtype` beside the existing `HasSubtype` predicate); a method is never exported to a name the standard library gives meaning to (`String`, `Error`, …).

### 2. The catalog completeness guard

Importing `internal/cards/effects` used to register every card. After the split only the aggregator links the shards, and a binary that forgot it would read a partial catalog in silence: `Lookup` misses, and a restore that names a card's effect key cannot find it, against "a restore never drops a stack item". So the guard is loud:

- the aggregator's `init()` calls `effects.MarkCatalogComplete()` once every shard has registered;
- `effects.Lookup`, `All` and `Has` panic, naming the import to add, if they run before it;
- `const catalogSharded` in `catalog_complete.go` turns the guard on. It is `false` until the split sets it, so the guard ships first as a no-op.

The split links the aggregator wherever the catalog is reachable: a generated `catalog_link.go` in each main package that depends on the core, and a generated external test file, `effects_catalog_link_test.go`, in each other package whose tests reach it.

### 3. The tool, and how the split lands

`scripts/split-effects/` (a standard-library Go program, run through `scripts/split-effects.sh`, with or without a local Go) does the move in one type-checked pass, and keeps the tree split afterwards:

1. **Placement** is a fixpoint over the use → declaration edges of sources and tests (rules in `split.go`'s header). Test files are split declaration by declaration: a declaration another package needs becomes an `effectstest` helper. Every test `init()`, `TestMain` and package-level var that registers something goes to `effectstest`, so each test binary sees what the one binary saw.
2. **Renames** go through the overrides table; struct literals that now cross a package boundary get their keys written in; an unnamed struct type in a crossing signature has its fields exported on both sides.
3. **Moves** keep file names and contents apart from the package clause, the imports and the qualified names, so git follows each file (ort's base-name rename detection pairs them).
4. **Test code** moved two directories down: relative paths gain `../../`. A test that scans the package's own sources reads them through `effectstest`'s flat view (`FlatGlob`, `FlatReadFile`, `FlatParseFile`, …), which shows the tree as the one directory it was, with the effects qualifiers removed, so its AST matching keeps working; `file.go:N` allowlist keys follow the lines the imports added.
5. **Generated:** the aggregator, a `main_test.go` per test package (calling `effectstest.RunTestMain`, the old `TestMain`), `core_cards.txt`, and `scripts/split-effects/table.txt`, every name the split moved or renamed.
6. **Outside the tree:** the CI test shard pins `internal/cards/effects/...`, `-cover` comes back for it, and the token-dump jobs point at the test package that holds the token test.

It lands in two PRs. **PR 1** (#5906, this ADR) adds the tool, the guard (a no-op), the range table, the fix-it check, the CI script's `/...` support and the docs written for after the split; it moves nothing. **PR 2** is the tool's output on the newest develop, regenerated right after a roll-up merges (`scripts/split-effects/REGENERATE.md`), and is reviewed by re-running the tool and diffing, not by reading 38,000 changed files.

### 4. Card work does not stop: the fix-it check

A branch cut before PR 2 adds its card and test files in the old directory. After rebasing they do not compile there. `TestEffectsTreeLayout` (`server/internal/docsguard`) runs `scripts/split-effects --check`, which lists every misplaced file (a card file or test in the core, a test in a shard, a file in the wrong shard) and prints one command:

```
scripts/split-effects.sh --place <files>
```

run from the repository root. It moves each file to its shard or test package (git mv) and rewrites the names it uses: through `table.txt` first, then to the core's or effectstest's exported name, exporting an unexported core helper in place if the file needs one. A card that edits an existing card file rebases onto the moved file; its conflicts are only where it touched a renamed name.

## Consequences

- **Memory.** In the spike's prototype of this layout, the largest compile in the effects tree fell from more than 13.3 GiB (OOM) to 1.47 GiB (the core), with 2.9 GiB summed over four parallel jobs, under `-race -cover`; repo-wide, the tree peak halved. PR 2's body records the numbers measured on the real split.
- **More test binaries.** 23 test binaries each link the whole catalog (about 1 GiB and a few seconds each, plus a few seconds of catalog `init()` under race). That is CPU the one binary did not spend, and far less than a build that cannot finish.
- **`unused` sees less.** About 13,000 helpers are exported, and the linter assumes an exported name is used. An occasional `deadcode` run is the mitigation.
- **The core still grows** with every helper, and about 600 card files sit in it because siblings reuse their helpers. Moving those helpers into named core files lets the cards move to their shards; that is follow-up work, one file at a time.
- **No on-disk identity changes.** Oracle IDs, effect keys, `ScopedEffect` mod kinds and stack-item catalog refs are string literals, which the tool never edits (the only literals it touches are relative paths in moved tests and identifier names in source-scanning tests). Snapshot fixtures, `SnapshotSchemaVersion` and the closure ratchet are untouched.
- **Docs.** `AGENTS.md` §7, `docs/adding-cards.md` ("Where a card goes"), `docs/repo-layout.md` and `docs/commands.md` describe the split layout, marked as taking effect when PR 2 lands. `docs/sprints.md` keeps its history.
