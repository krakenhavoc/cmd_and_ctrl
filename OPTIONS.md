# #5906: options for splitting `internal/cards/effects`

Design spike, S-unassigned, issue #5906. No code on `develop` changes. This
branch (`spike/5906-options`, an orphan) holds only this file and the
throwaway `prototype.sh` that produced the measured prototype. The owner
picks an option; this file recommends one.

Tree measured: `origin/develop` at `39ccb36ef` (2026-11-19), go1.27.0.

---

## 0. Summary

- **The problem is worse than the issue says.** On this machine (4 cores, cgroup
  limit 13.35 GiB), compiling the effects *test* variant **OOMs even without
  `-cover`, and even without `-race`**. The compiler process alone reached
  13.1 to 13.3 GiB in every mode before it was killed. CI's current workaround
  (drop `-cover` on the 16 GB runner) works only because the runner has about
  2 GB more than this machine, and each card batch eats into that margin. This
  is a P1 with a short fuse, not a tidy-up.
- **Splitting works, and by a large margin.** In the prototype of option A
  (39 card shards, 35 test packages, a core, a test-helper package), the
  whole effects tree's `go test -race -cover -run XXX` peaks at **1.47 GiB for
  the largest process and 2.9 GiB summed over 4 parallel jobs**, and finishes
  in 4m46s. The baseline OOMs above 13.3 GiB. Plain `go build` falls from
  7.05 GiB to 1.33 GiB.
- **Repo-wide, it helps the other packages too.** Prototype A's
  `go test -race -cover ./...` (every package, effects included) took 28:35
  and peaked at 6.47 GiB in total. Develop *without* the effects package took
  21:57 and peaked at **12.82 GiB**, because every package whose tests link the
  catalog compiles `effects` with `-race` as one 6 GiB unit.
- **Helpers are what make a split hard.** Package `effects` has **13,984
  unexported top-level names**. **4,400** of them are used outside the file
  that declares them. **11,301 of 17,193 card files** use at least one
  unexported name from another file. Any split that moves cards out of the
  package has to **export about 8–13k names** (12,852 in the prototype,
  including test helpers), touching about 15k files. Done by hand, that is
  impossible. A type-checked rewriter does the analysis and the rewrite in seconds.
- **Recommendation: option A**, done as one generated PR from a
  `go/types` script, behind a short card-merge freeze. Section 6 has the
  details.

---

## 1. Measurements

All numbers were taken on this cloud machine: 4 vCPU, a 15 GiB host, and a
**memory cgroup limit of 13.35 GiB** (14,338,756,608 bytes). The GitHub
`ubuntu-latest` runner has 4 vCPU and 16 GB with no cgroup cap, so a build that
needs more than 13.35 GiB is killed here but might still pass there.

How I measured:
- **Go and Docker.** Docker has no daemon here, so `scripts/go-docker.sh` could
  not run. I used the host's Go with `GOTOOLCHAIN=auto`, which downloaded
  go1.27.0, the version `go.mod` declares.
- **Peak memory.** `/usr/bin/time -v` gives the maximum RSS of the single
  largest process. A sampler (every 0.2 s) gives the summed RSS of the whole
  process tree, which is what the OOM killer sees when `go` runs compiles in
  parallel (default `-p 4`).
- **Per-package numbers** come from a `-toolexec` wrapper that runs every
  `compile` and `link` under `/usr/bin/time`.
- **Caches.** Every run starts from a copy of one warm `GOCACHE` that already
  holds every dependency of effects (plain and `-race`). Only the effects
  packages are compiled cold.

### 1.1 Baseline (develop, one package)

| command | wall | largest process | tree sum | result |
|---|---|---|---|---|
| `go build ./internal/cards/effects/` | 0:53 | **7.05 GiB** | 7.10 GiB | ok |
| `go test -race -cover -run XXX ./internal/cards/effects/` | 2:33 to the kill | **13.20 GiB sampled; 13.84 GB anon-rss at the kill (kernel log)** | 13.29 GiB | **OOM-killed** (`compile: signal: killed`) |
| `go test -race -run XXX ./internal/cards/effects/` (what CI runs now) | 6:00 to the kill | 13.24 GiB | 13.30 GiB | **OOM-killed** here |
| `go test -run XXX ./internal/cards/effects/` (no race, no cover) | 3:37 to the kill | 13.13 GiB | 13.21 GiB | **OOM-killed** here |

The package is 18,262 non-test files and 2,940 `_test.go` files, 21,202 in
all. Measured by disk usage: about 70 MB of card files, 7.4 MB of non-card
("framework") files and 45 MB of tests. (`du -sh` of the directory reports
123M on disk.) The test variant compiles all of it as one unit. That is why
every test mode costs about twice what `go build` costs.

**Repo-wide `go test ./...` on develop cannot be measured here.** It contains
the effects package, which OOMs in every mode. Section 1.4 gives the
comparator I could run.

### 1.2 How tangled are the helpers? (go/types, exact)

A type-checked pass over the package (`go/types`, every `Uses` edge):

| | count |
|---|---|
| unexported top-level names in `effects` (sources only) | **13,984** |
| ... used in a file other than the one that declares them | **4,400** |
| card files (an `init()`) / non-card files | 17,193 / 1,069 |
| card files that use an unexported name, field or method from another file | **11,301** (66 %) |
| distinct framework-declared unexported names, fields and methods used by card files | 3,271 (2,806 funcs, 321 vars, 55 consts, 42 types, 25 methods, 22 fields) |
| names declared in a **card** file and used by **other card** files | 913, declared in 594 card files (e.g. `onBattlefield` in `monk_gyatso.go`: 481 users; `selfOnly` in `wire.go`: 145) |
| names declared in a card file and used by **framework** files | 169 |
| test-file declarations used by other test files (test helpers) | 4,961 (e.g. `newCatalogGame`: 2,713 users; `passPriorityAroundTable`: 2,540) |
| card-file names used by tests / framework names used by tests | 632 / 302 |

Most-used framework helpers: `b16FirstLegalTargetCard` (580 card files),
`enteredUnderYourControl` (199), `sourceIsStillThisPermanent` (184),
`thisGetsUntilEndOfTurn` (181), `attackDeclaredByYou` (170).

What this means:
1. Cards can leave the package only if their helpers become exported.
2. The 913 card-to-card names and the 169 card-to-framework names force about
   **640 card files to stay in core** (their declarations are other files'
   helpers). Without a hand refactor, the core ends up as 1,069 framework
   files plus about 640 card files.
3. Tests depend on test helpers far more than on card internals, so the
   helpers need a package of their own (`effectstest`).

### 1.3 Prototype of option A (scratch branch, never pushed)

The tool (`proto2`, about 600 lines of `go/types`; `prototype.sh` drives it)
made these numbers:

- 1,708 files stay in `effects` (8.2 MB).
- 16,554 card files go to 39 shards of at most 700 files each
  (`effects/cards/{a1,a2,b1,…,z}`).
- The test helpers (3,481 declarations lifted out of 1,331 test files, 1.6 MB)
  go to `effects/effectstest`.
- The test declarations go to 35 test-only packages of at most 150 files each
  (`effects/tests/{a1,…,z}`, 37 MB).
- `effects/cards` is the aggregator.
- 12,852 names are exported. 264 of them needed an `X` prefix because the
  capitalised name was already taken. The renames edit 15,371 files.
- One hand fix: an anonymous struct type with unexported fields
  (`HornTopFive`) crossing a package boundary.

Everything compiles: `go build`, `go vet` and the test binaries. `go vet`
reports **157 "composite literal uses unkeyed fields"** sites, where a struct
literal now crosses a package boundary. A real migration fixes those (or the
tool writes the keys).

| command (prototype) | wall | largest process | tree sum | result |
|---|---|---|---|---|
| `go build ./internal/cards/effects/...` | 1:05 | **1.33 GiB** | 2.52 GiB | ok |
| `go test -race -cover -run XXX ./internal/cards/effects/...` | 4:42 (4:46 under toolexec) | **1.47 GiB** | 2.87 GiB | exit 0: 77 packages (35 test packages ok, 42 with no tests) |
| `go test -count=1 ./internal/cards/effects/...` (runs all 34,323 tests, no race) | 1:40 | 0.85 GiB | 2.96 GiB | 34,191 pass, **132 fail** (see 1.5) |

Per compile unit (`-race -cover` test run, from toolexec):

| unit | files | peak RSS | wall |
|---|---|---|---|
| `effects` (core) | 1,708 | **1.47 GiB** | 24.4 s (the critical path: everything waits on it) |
| `effects/effectstest` | 1,331 (helpers only) | 0.47 GiB | 7.6 s |
| a card shard (39 of them) | ≤ 700 | 0.53 GiB on average | 5.7 s on average |
| a test package (35 of them) | ≤ 150 | 0.38 GiB on average, 0.80 GiB max (`tests/b`) | 7.0 s on average, 16.7 s max |
| a test binary link (36) | — | ≤ 1.13 GiB | 2–3 s each, 89 s in all |

Sum of compile wall over the 77 units: 496 s, about 2 min of wall on 4
cores. The rest of the 4m42s is linking 35 race binaries, each holding the
whole catalog, and starting each one: about 4 s of catalog `init()` per
binary under race+cover, even with `-run XXX`.

**A first, naive version** split by whole files (every test file that held a
helper went whole into `effectstest`; tests lived in their card's shard; 27
letter shards). Its peaks were **6.68 GiB** (`effectstest`, 1,488 files) and
**5.73 GiB** (shard `s`, 2,426 files), and the tests failed to run (section
1.5). Two lessons, both kept in option A:
- **Shard size is the lever on memory.** About 0.5 GiB per 700 card files, and
  roughly linear.
- **Tests must not live in card shards** (section 1.5).

### 1.4 Repo-wide `go test -race -cover ./...`

A repo-wide baseline on develop is impossible here, because the effects test
binary OOMs (section 1.1). I ran the comparator that *can* be run instead:
develop with every package **except** `internal/cards/effects` (73
packages). Both runs started from the same warm cache, ran one after the
other on an otherwise idle machine, and used `-timeout 60m`.

| run | wall | largest process | tree sum | packages |
|---|---|---|---|---|
| develop, `./...` minus `internal/cards/effects` | **21:57** | 6.07 GiB | **12.82 GiB** | 73 ok |
| prototype A, all of `./...` (includes the whole effects tree: 77 effects packages, 34,323 effects tests) | **28:35** | 3.89 GiB | 6.47 GiB | 86 ok, 22 FAIL |

Reading it:
- **The whole effects tree costs about 6.5 min of wall in prototype A**:
  compile, link, and run all 34k tests under race+cover, on 4 cores. The issue
  measured the same tests at about 15 min in one race binary, on top of a
  compile that no longer fits.
- **Peak memory halves even for the *rest* of the repo.** On develop, the other
  packages' tests compile `effects` as a dependency, with `-race`, as one
  6 GiB unit. That is the 6.07 GiB process, and the 12.82 GiB tree peak, in a
  run that does not even test effects. CI's `rest` shard, and every shard that
  links the catalog, sits uncomfortably close to the runner too. In A, the
  largest process anywhere in `./...` is 3.89 GiB (a non-effects package), and
  the tree peaks at 6.47 GiB.
- The 22 failing packages are the 21 effects test packages of section 1.5 plus
  `internal/cards/coverage`. Coverage fails one test,
  `TestNoNewExactClonesInTheCatalog` (the clone gate). It reports a "new"
  duplicate between `angraths_marauders.go` and `batch07_helpers.go`, a pair
  the gate's baseline lists by its old location. The gate needs to read the
  new directories.
- Slowest packages are unchanged between the runs: `internal/game` 925 s and
  `lobby` 782–804 s.

### 1.5 What breaks when it runs (prototype)

Naive version, tests inside the card shards: 152 of the sampled shards' tests
failed. Card tests are named by batch, not by card (`yshtola_a_test.go` tests
Delney, G'raha Tia, Kambal, ...). They also find cards by name or oracle ID at
run time (`newCatalogGame("Card")`), which a static analysis cannot see. A test
binary that does not link the whole catalog fails. A card shard cannot import
the aggregator (that would be an import cycle), so **card tests belong in
packages that import the aggregator**.

Option A prototype: 132 of 34,323 top-level tests fail. Every cause seen is
mechanical:
- source-scanning lint tests (`filepath.Glob("*.go")`, `os.ReadDir(".")` in
  about 10 tests) now read the wrong directory. Example message: "scanned no
  catalog files — the glob is wrong".
- relative testdata paths (`../../game/testdata/...`) are one level off.
- the coverage oracle fixtures are found by relative path ("Zhalfirin Crusader
  has no oracle fixture").
- `TestMain` (card-index cross-check, plus `snapshotProductionDefKeys`) exists
  in only one of the 35 test packages ("TestMain did not capture the
  production catalog keys").
- tests that relied on a fixture another test file registered in `init()` in
  the same binary (the replacement-resume restore points).

### 1.6 What I could not measure, or what failed

- **Repo-wide `go test ./...` on develop**: impossible here, because the
  effects test build OOMs in every mode. Section 1.4 uses the "every package
  but effects" comparator instead.
- **`scripts/go-docker.sh`**: not used, because Docker has no daemon in this
  container. I used the host's Go with `GOTOOLCHAIN=auto` (go1.27.0).
- **Option B per-unit numbers**: the `-toolexec` wrapper failed on the
  18k-file `effects` compile ("argument list too long"). B was measured
  without the wrapper, as a whole run only.
- **The prototype's tests under `-race`**: compiled and linked with
  `-race -cover` (section 1.3), but run only without `-race` (1:40, 132
  failures) and in the repo-wide `-race -cover` run (1.4).
- **`golangci-lint` / `go vet` memory and time** on either layout: not
  measured. Lint type-checks the whole tree. Its 20 m timeout comment in
  `.golangci.yml` blames effects.
- **The real CI runner**: every number is from this 13.35 GiB machine, not
  from GitHub's 16 GB runner.

---

## 2. Facts every option has to respect

- **No identity is derived from Go symbols or package paths.** I grepped
  `internal/game` and the registry for `runtime.FuncForPC`, `runtime.Caller`,
  `%T`, `PkgPath()` and `reflect.Type.String()`. None of them feeds an
  identity.
  - The catalog key is the Scryfall `oracle_id` string (`Register` →
    `registry map[string]Spec`).
  - Effect keys and stack-item catalog refs (#1497, ADR 0041 P9) are string
    literals in the card files.
  - So **moving a file to another package changes no on-disk identity**:
    snapshots, effect keys and fixtures stay byte-identical, as long as the
    move does not edit those strings (the rewriter edits identifiers only, and
    never string literals).
- **Registration needs the code linked in.** Today, importing `effects`
  registers every card. After any split, a binary or test that reads the
  catalog (`effects.Lookup`, `All`, a snapshot restore that names a card's
  effect key) must link every shard. If one forgets, the result is silent: an
  empty `Lookup`, and a restore that cannot find an effect key. That
  contradicts "a restore never drops a stack item". Any option that moves
  cards therefore needs a guard (section 3, A.2).
- **The closure ratchet** (`internal/game/testdata/closure_fields.txt`,
  `TestClosureFieldsReachableFromGame`) classifies routes from `game.Game` to
  funcs and interfaces by **field path inside package game**. It names no
  effects package path (the only effects-like line is
  `scopedEffectAdapterMemo.effects`, a game field). A move changes no line,
  and the `census:` and `transient` counts do not move.
- **Init order.** No card's `init()` reads the registry (the only `Lookup(` or
  `registry[` uses in card files are in comments). Core's non-card files have
  no `init()`. Shards run after core, in import order, so registration order
  changes. That is harmless, because `Register` panics on a duplicate key
  regardless of order. One caveat: `snapshotProductionDefKeys()` in
  `TestMain` captures "the production catalog", so it must run in a binary
  that has linked every shard.
- **Imports into effects.** 17 packages import `effects` (125 files; protocol
  43, legal 26 as `legal_test`, aiseat 25, cards/coverage 17, roadmap 15,
  ...). They use its exported API (`Register` 229, `Spec` 234, `All` 46,
  `Lookup` 20, ...). That API stays in core in every option, so their
  qualified references do not change. Only the "link every card" side effect
  has to be added where the catalog is read. 73 files import `effects` for its
  side effect only today.

---

## 3. Option A — core + card shards + test packages (prototyped)

### A.1 Layout

```
internal/cards/effects/            package effects      core: Spec, Register, Lookup, All, the registry,
                                                        every seam and shared helper (~1,070 framework
                                                        files + ~640 card files whose declarations are
                                                        other files' helpers)
internal/cards/effects/cards/      package cards        the aggregator: blank-imports every shard, and
                                                        nothing else (generated, see A.4)
internal/cards/effects/cards/a1/   package cardsa1      card files only, one init() each, ≤ 700 per shard
  ... b1, b2, ..., s1..s4, ..., z
internal/cards/effects/effectstest/ package effectstest  test helpers (newCatalogGame → NewCatalogGame,
                                                        passPriorityAroundTable, castCatalogSpell, ...)
internal/cards/effects/tests/a1/   package testsa1      _test.go files only; each blank-imports
  ...                                                   effects/cards so the whole catalog is linked
```

Dependencies only point down:
`tests/* → effectstest → cards/* → effects → game, legal, ...`. Also
`tests/* → cards` (the aggregator) and `tests/* → cards/x` when a test calls
a card's own helper.

**Shard key.** The prototype used the first letter of the file name, splitting
a big letter into fixed alphabetical ranges (`s1` = `sa..sf`, ...). That is
predictable for agents: "a card file goes in `cards/<its letter range>/`".
**Freeze the ranges in a table** (`scripts/cards-shard.sh` or a Go
`shardFor(name)`) and let `docsguard` check that every file sits in its
shard. Rebalancing is then a deliberate PR that moves a range, never a side
effect of growth. Alternatives considered:
- **by set**: card files do not carry their set, and a reprint has none. No.
- **by hash of oracle ID**: balanced forever, but nobody can guess where a
  card lives. No.
- **first two letters** (about 400 packages of about 40 files each): too many
  packages. Every package re-reads core's export data, and every test binary
  links them all.

Recommended target: **≤ 700 card files per shard** (about 0.5 GiB to compile).
**≤ 150 test files per test package** (about 0.4 GiB). Re-split a range when
a shard passes about 1,000 files.

**Qualified or dot imports.** The prototype dot-imports core into card files
(`import . ".../effects"`). Card files then keep `Register(Spec{...})` and
every helper call unqualified, so a card diff and the
`docs/adding-cards.md` recipes keep their shape, and in-flight branches
rebase with fewer conflicts.
- Cost: staticcheck **ST1001** is on (`checks: all` in `.golangci.yml`), so
  the config needs `dot-import-whitelist` for the effects and effectstest
  paths.
- The alternative is qualified `effects.Register(effects.Spec{...})`,
  rewritten by the same tool. It is more idiomatic and easier to grep, but it
  touches every card line, and conflicts with every in-flight card.
- **Owner decision.** I lean to dot imports, confined by the whitelist to
  `cards/*`, `effectstest` and `tests/*`.

### A.2 Registration and the catalog

- `Register`, `Lookup`, `All` and the registry stay in core, unchanged.
- Each shard's `init()` functions register as today.
- Whatever needs the full catalog links `effects/cards`:
  - `cmd/server`, `cmd/boteval`, `mcpseat`;
  - the tests of `protocol`, `legal`, `aiseat`, `decks`, `lobby`, `ws`,
    `roadmap`, `coverage`, `catalog`, ...;
  - every `effects/tests/*` package.
  The prototype added the import next to every existing effects import
  (125 files). A real migration adds it only where the catalog is read.
- **Guard (required).** The aggregator's `init()` calls
  `effects.MarkCatalogComplete()`. `Lookup`, `All` and the snapshot-restore
  path panic if they run before it is called (or `docsguard` fails any package
  that calls them without importing `effects/cards` in its import graph).
  Without the guard, a forgotten import is a silent empty catalog, and in
  production a restore that loses an effect.
- **The aggregator is generated**, as `cards/cards.go`, by
  `go generate ./internal/cards/effects/cards`. A `docsguard` test fails if
  any shard directory is missing from it.
- Tests that inject fixtures (`registerForTest`) need an exported, test-only
  hook. The prototype exported `registry`. The real one should be
  `effects.RegisterForTest`, living in `effectstest` through an
  `export_test`-style seam.

### A.3 On-disk identities and ratchets

| thing | effect of A |
|---|---|
| oracle IDs, effect keys, `ScopedEffect` mod kinds, stack-item catalog refs | none: strings, not edited (the rewriter touches identifiers only) |
| snapshot fixtures (`internal/game/testdata/snapshots/v*`) | none; `SnapshotSchemaVersion` unchanged; no fixture edited |
| closure ratchet `closure_fields.txt` | none (game-internal field paths) |
| oracle fixtures (coverage) | file contents unchanged. The tests that read them by relative path need one path fix each (the 10 "no oracle fixture" failures). `-update-oracle -oracle-ids=...` keeps working |
| `testdata/legacy_trigger_builds.txt` (effects) | moves with the test that reads it, or is read through an `effectstest.TestdataDir()` helper |
| tests reaching unexported helpers | handled by export: 632 card-file names and 302 framework names used by tests become exported; 4,961 test-helper declarations become `effectstest` exports |
| `unused` linter | **loses sight of exported dead helpers** (exported = assumed used). Mitigation: an occasional `deadcode`/`staticcheck -checks U1000` run with `-unexported` semantics, or accept it |
| errcheck/unparam test exclusion (`path: _test\.go`) | `effectstest/*.go` are not `_test.go`. Add `path: internal/cards/effects/effectstest/` to that exclusion |
| CI shard script (`scripts/ci-test-shard.sh`, `TestCITestShardsCoverEveryPackageOnce`) | `effects` pins one package today. Pin `internal/cards/effects/...` (the script must expand a pattern). The `-cover` exception for the effects shard is deleted |
| `RealDumpEveryTokenTemplate` CI job (`go test ./internal/cards/effects/ -run ...`) | path changes to wherever the token test lands (`./internal/cards/effects/tests/...`) |

### A.4 How the move is done

1. **The script.** `scripts/split-effects/` is a small Go program, the
   productised `proto2`. In one type-checked pass it:
   1. assigns files to core, shards, `effectstest` and `tests/*` (the
      fixpoint rules in `prototype.sh`'s header);
   2. exports every name that crosses a package boundary, picking names
      deliberately instead of `X`-prefixing (a rename table, reviewed);
   3. writes the keys into the 157 unkeyed literals;
   4. splits test files declaration by declaration;
   5. `git mv`s whole files (blame survives, because a moved card file is
      unchanged apart from its package line and import);
   6. runs `goimports`;
   7. generates the aggregator.
   It is idempotent, so it can be re-run on top of a newer `develop`.
2. **PRs.**
   - **PR 1 (hand-written, small):** the guard (`MarkCatalogComplete`), the
     shard-range table and its `docsguard` test, the lint config
     (`dot-import-whitelist`, effectstest exclusion), the
     `ci-test-shard.sh` pattern support, and the docs. It merges first, with
     the guard a no-op until PR 2.
   - **PR 2 (generated, huge):** the script's output plus the hand fixes
     (testdata paths, the about 10 source-scanning lint tests pointed at the
     new dirs, a `TestMain` per test package via a generated `main_test.go`).
     It is reviewed by re-running the script and diffing, not by reading
     15k files.
   - **PR 3:** delete the CI `-cover` exception. Rebalance any test package
     that is slow.
3. **In-flight card branches.**
   - Announce a freeze of about one day on card PRs. Land PR 2 right after a
     roll-up.
   - A branch that adds a new card file `foo.go` rebases with **no conflict
     on the file itself**, but it lands in the wrong directory and fails to
     compile (unexported helpers). Fix:
     `scripts/split-effects --place <file>...` moves new card and test files
     into their shard and rewrites their helper references through the same
     rename table.
   - A branch that *edits* an existing card file conflicts only where the
     edit touches a renamed identifier (`git` follows the rename for the
     move).
   - Agents get one line in the brief: "after rebasing past #<PR2>, run
     `scripts/split-effects --place` on your new files".
4. **Docs and agent briefs that change:**
   - `AGENTS.md` §7 ("one file per card,
     `server/internal/cards/effects/<snake_name>.go`") becomes
     `effects/cards/<shard>/<snake_name>.go`, plus where tests and helpers go.
   - `docs/adding-cards.md`: 152 path mentions, the Recipe, and "where does a
     new helper go" (core, exported, if two shards will share it).
   - `docs/repo-layout.md`, `docs/snapshots.md` (3 mentions),
     `docs/commands.md` (test commands), `docs/sprints.md` (36 historical
     mentions: leave them).
   - `.claude/skills/triage/references/decision-rules.md` (1);
     `scripts/cost/cost_report.py` (2); `scripts/ci-test-shard.sh`;
     `.github/workflows/ci-cd.yml` (the effects `-cover` exception, the token
     job path) and `e2e-nightly.yml` (the token job path).
   - Any ADR that names the path: none found by grep outside `docs/sprints.md`
     and `docs/audit-2026-09-10.md`.
   - New ADR (take the next free number per `docs/decisions/README.md`;
     current range ends at 0222): "the catalog is sharded; the shard table;
     the completeness guard".

### A.5 Risks

- **The core keeps growing.** Core is 1,708 files and 8.2 MB, 1.47 GiB to
  compile, and it is the critical path (24 s, before any shard starts). About
  640 card files are stuck in core only because a sibling card reuses one of
  their helpers. Moving those helpers into named framework files
  (`b16_helpers.go` and the like, already the pattern) is follow-up hygiene,
  not part of the move. **New rule:** a card file declares nothing another
  file uses. A `docsguard` test can enforce it with `go/types` in about
  3 s.
- **Test binaries multiply.** 35 race binaries each link the whole catalog
  (about 1.1 GiB and 2–3 s per link), and each pays about 4 s of catalog
  `init()`. That is about 2–3 min of CPU that one binary did not cost. It is
  still far below the baseline, which cannot build at all. Fewer, larger test
  packages (≤ 300 files) halve it at about 0.8 GiB each.
- **Silent empty catalog** if a binary forgets the aggregator. Mitigated only
  by the guard in A.2. Do not ship without it.
- **Exported surface.** About 12.8k names become exported and `unused` stops
  checking them. Names chosen by a tool (the 264 `X…` collisions) need a
  human pass.
- **Rebase pain for about a day of in-flight work.** Mitigated by `--place`.
- **The prototype's 132 failing tests** are all of the kinds listed in 1.5.
  None was a behaviour change, but the real PR must take them to zero, not
  skip them (AGENTS.md: never skip a test to get green).

---

## 4. Option B — move only the tests out (cards stay in `effects`)

### B.1 Layout

`effects` keeps every card and framework file (18,262 files, as today) but no
`_test.go`. The test helpers go to `effects/effectstest` and the tests to
`effects/tests/*`, exactly as in A. `cards/` holds an empty aggregator, or
none (importing `effects` still registers everything).

### B.2 What it changes

- No registration change and no guard needed: `effects` still self-registers.
  No card file moves, so agent recipes, `AGENTS.md` §7 and in-flight card
  branches are untouched.
- Exports: tests use 632 card-file and 302 framework names, plus the 4,961
  test helpers. The prototype in B mode exported 9,148 names in 6,726 files.
- Identities and ratchets: as in A (nothing on disk changes).

### B.3 Measured

Prototype in B mode (`MAX_CARDS=0`): 18,262 files stay in `effects`, 1,331
helper files go to `effectstest`, 35 test packages. It exported 9,148 names in
6,726 files and compiles.

| command | wall | largest process | tree sum |
|---|---|---|---|
| `go test -race -cover -run XXX ./internal/cards/effects/...` (B) | 5:13 | **7.64 GiB** | 7.76 GiB |
| `go build -race -cover ./internal/cards/effects/` (develop: the same unit as B's core) | 2:58 | **8.00 GiB** | 8.07 GiB |

So B takes the effects test build from "does not fit in 13.35 GiB" to about
7.6–8 GiB, with about 5–6 GiB of headroom on a 16 GB runner. The package
itself (`go build`, 7.05 GiB plain today) keeps growing with every batch, and
so does the race copy every other package's tests compile (the 6.07 GiB
process in 1.4).

### B.4 Risks

- **It does not fix the growth.** The package itself still compiles at
  7.05 GiB plain today, and grows linearly with the catalog (about 0.4 MB of
  source per 100 cards). It buys months, not a lasting fix.
- It still pays the most painful part of A (the test-helper export and the
  test move) without the memory win where the catalog itself grows.

---

## 5. Option C — no split: buy headroom (stopgap only)

- **C1. `GOGC` for the compiler.** The issue reports `-race -cover` "just
  fits" at 13.4 GB with GOGC=30. That trades CPU for memory and buys a few
  hundred cards.
- **C2. A bigger runner** (GitHub's 8-core/32 GB larger runners are paid).
  PLAN.md and AGENTS.md §8 say "never charge", and a hobby project should not
  pay for CI. A self-hosted runner exists for lint (`.golangci.yml` mentions
  it), so the effects shard could run there, if it has more than 16 GB.
- **C3. Keep dropping flags.** `-race` is next. But measured here, even plain
  `go test` with no race and no cover needs more than 13.3 GiB. Nothing is
  left to drop after `-race`.

None of these stops growth. **C1 plus C2 is the bridge until A lands.** If the
effects shard starts dying before A is merged, set `GOGC=30` on that shard's
`go test` step only.

---

## 6. Recommendation

**Option A**, with dot imports confined to the new packages, ≤ 700 card files
per shard on a frozen alphabetical range table, card tests in `tests/*`
packages that link the aggregator, and the completeness guard. Order:

1. **Now** (if CI fails before A lands): `GOGC=30` on the effects shard
   (option C1). Section 1.1 shows how close to the cliff the no-cover build
   is.
2. **PR 1:** guard, shard table, lint and CI-script support, ADR, docs.
3. **PR 2:** run `scripts/split-effects` right after a roll-up, behind a card
   freeze of about one day. Fix the mechanical test failures. CI goes from
   OOM to a peak of about 1.5 GiB per process (measured) and about 3 GiB in
   total.
4. **PR 3:** restore `-cover` for the effects tree, and remove the
   exception and its comment in `ci-cd.yml`.
5. **Follow-up:** move helpers out of the about 640 card files that core still
   holds, and add the "a card file declares nothing another file uses" guard,
   so core stops growing with the catalog.

Why not B: it costs most of A's churn (the test-helper extraction is the
fiddly half) and leaves the part that grows with every batch in one compile
unit. Why not C: it is not a fix. Measured, even the cheapest test build
mode no longer fits on 13.35 GiB.

**Owner decisions needed:** (a) A or B; (b) dot vs qualified imports in card
files; (c) the shard size (700 cards / 150 tests as measured, or larger
shards for fewer binaries); (d) the length of the card freeze.

---

## Appendix: reproducing

- `prototype.sh` on this branch is self-contained. Its header documents the
  inputs and the placement rules. The `go/types` rewriter it runs (`proto2`,
  about 700 lines of Go) is embedded at the end of the script, and the script
  extracts and builds it when `PROTO` is unset.
  - Run it from the root of a scratch worktree of `develop` with
    `EXPORTS=<export map>`.
  - `MAX_CARDS=0` gives the option B layout.
  - It is throwaway code: a real `scripts/split-effects` should be written
    fresh and reviewed.
- Peak memory: `/usr/bin/time -v` plus a 0.2 s `/proc` sampler of the
  process-tree RSS sum. Per unit: `go test -toolexec=<wrapper>` timing each
  `compile` and `link`.
- Helper analysis: `go/types` `Info.Uses` over all 18,262 sources (and again
  with the 2,940 tests), counting edges whose object is unexported and
  declared in another file. A "card file" is one with an `init()`.
