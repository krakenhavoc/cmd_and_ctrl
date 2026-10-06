// labels.test.ts — the guard on the contract labels (ADR 0125 §2.3).
//
// It reads source files and renders nothing, so it runs in well under a
// second on every PR that touches the client:
//
//   1. every entry is rendered by its owners: each owner file exists and
//      references `L.<key>`;
//   2. no other file under client/src carries a literal copy of a
//      registered static `aria` name, so a component cannot keep
//      rendering a contract name the registry does not see;
//   3. every tutorial anchor, detours included, and every first-use
//      hint's anchor (the shipped hints and the test fixtures) is a
//      registered `aria` label;
//   4. docs/labels.md is what the registry generates. To rewrite it:
//      `UPDATE_LABELS_DOC=1 npm test -- labels`.
//
// What it cannot see: that the element is on screen when an anchor needs
// it. A label behind an {#if} that never comes true passes here; the
// nightly tutorial walk is for that.

import { describe, it, expect } from "vitest";
import { existsSync, readFileSync, readdirSync, writeFileSync } from "node:fs";
import { join, relative, sep } from "node:path";
import { fileURLToPath } from "node:url";

import {
  L,
  LABEL_SPECS,
  registeredAria,
  stemMatches,
  type LabelKey,
  type LabelRef,
  type LabelSpec,
} from "./labels";
import { anchorsOf, type Anchor, type StepContext } from "./tutorial";
import { TUTORIAL_STEPS } from "./tutorialSteps";
import { auto, board, ctx, elves, forest, type Roll } from "./test/tutorialBoards";
import { HINTS } from "./hints";
import type { Hint } from "./hints/hint";
import { FIXTURE_HINTS, hintContexts } from "./test/hintContexts";

const SRC = fileURLToPath(new URL("../", import.meta.url));
const DOC = fileURLToPath(new URL("../../../docs/labels.md", import.meta.url));
const SPECS = Object.entries(LABEL_SPECS) as [LabelKey, LabelSpec][];

const read = (rel: string): string => readFileSync(join(SRC, rel), "utf8");
const escapeRe = (s: string): string => s.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
const show = (spec: LabelSpec): string => (spec.type === "static" ? spec.name : spec.shape);

// ---- Rule 3's walk: every tutorial anchor over the step tests' boards ----

const TURN_STEPS = [
  "untap",
  "upkeep",
  "draw",
  "precombat_main",
  "begin_combat",
  "declare_attackers",
  "postcombat_main",
  "end",
];

/** Boards enough to reach every step's anchor and every detour. */
function contexts(): StepContext[] {
  const out: StepContext[] = [ctx(null)];
  const permanents = [
    [],
    [forest()],
    [forest(), elves()],
    [forest({ tapped: true }), elves({ summoning_sick: false })],
  ];
  for (const step of TURN_STEPS) {
    for (const active of [0, 1]) {
      for (const mine of permanents) {
        for (const stack of [[], [elves()]]) {
          const v = board({ step, active, mine, stack });
          out.push(ctx(v), auto(v));
        }
      }
    }
  }
  // The opening roll (step 2): owed, rolled and waiting, won, lost.
  const rolls: Roll[] = [
    {},
    { rolls: [[0, 12]] },
    {
      rolls: [
        [0, 19],
        [1, 4],
      ],
      chooser: 0,
    },
    {
      rolls: [
        [0, 3],
        [1, 20],
      ],
      chooser: 1,
    },
  ];
  for (const roll of rolls) out.push(ctx(board({ hand: [], step: "untap", roll })));
  // The opening hand still to keep, and kept (steps 4 and 7).
  for (const kept of [false, true]) out.push(ctx(board({ step: "upkeep", mulligan: { kept } })));
  return out;
}

interface Found {
  /** "step <id>" or "step <id>, detour <id>". */
  where: string;
  anchor: Anchor;
}

function tutorialAnchors(): { found: Found[]; detoursBySteps: Map<string, Set<string>> } {
  const found: Found[] = [];
  const detoursBySteps = new Map<string, Set<string>>();
  for (const c of contexts()) {
    for (const step of TUTORIAL_STEPS) {
      for (const anchor of anchorsOf(step, c)) found.push({ where: `step ${step.id}`, anchor });
      const d = step.first?.(c);
      if (!d) continue;
      if (!detoursBySteps.has(step.id)) detoursBySteps.set(step.id, new Set());
      detoursBySteps.get(step.id)!.add(d.id);
      const as = d.anchor === undefined ? [] : Array.isArray(d.anchor) ? d.anchor : [d.anchor];
      for (const anchor of as) found.push({ where: `step ${step.id}, detour ${d.id}`, anchor });
    }
  }
  return { found, detoursBySteps };
}

// ---- Rule 3's walk, for hints: every hint's anchor over its contexts ----

/** Every hint the guard walks: the shipped ones and the test fixtures. */
const ALL_HINTS: readonly Hint[] = [...HINTS, ...FIXTURE_HINTS];

function hintAnchors(): Found[] {
  const found: Found[] = [];
  for (const h of ALL_HINTS) {
    for (const c of hintContexts(h)) {
      const a = typeof h.anchor === "function" ? h.anchor(c) : h.anchor;
      if (a) found.push({ where: `hint ${h.id}`, anchor: a });
    }
  }
  return found;
}

const WALK = tutorialAnchors();
const HINT_WALK = hintAnchors();

/** Which entries each place anchors to, for rule 1's message. */
function anchoredBy(): Map<LabelKey, Set<string>> {
  const by = new Map<LabelKey, Set<string>>();
  const note = (ref: LabelRef, where: string) => {
    const key = registeredAria(ref);
    if (!key) return;
    if (!by.has(key)) by.set(key, new Set());
    by.get(key)!.add(where);
  };
  for (const { where, anchor } of [...WALK.found, ...HINT_WALK]) {
    if (!("label" in anchor)) continue;
    note(anchor.label, where);
    if (anchor.within !== undefined) note(anchor.within, `${where} (as its container)`);
  }
  return by;
}

const ANCHORED_BY = anchoredBy();

// ---- Rule 2's scan ----

function sourceFiles(dir: string): string[] {
  const out: string[] = [];
  for (const e of readdirSync(dir, { withFileTypes: true })) {
    const p = join(dir, e.name);
    if (e.isDirectory()) out.push(...sourceFiles(p));
    else if (/\.(ts|svelte)$/.test(e.name) && !e.name.endsWith(".test.ts")) out.push(p);
  }
  return out;
}

/** The literal forms rule 2 refuses: `aria-label="…"`, `aria-label={"…"}`, `label="…"`, `label: "…"`, … */
function literalCopy(name: string): RegExp {
  return new RegExp(
    `\\b(?:aria-label|label|ariaLabel|group)\\s*(?:=\\s*\\{?\\s*|:\\s*)(["'\`])${escapeRe(name)}\\1`,
  );
}

// ---- Rule 4's document ----

const mdCell = (s: string): string => s.replace(/\|/g, "\\|");

function nameCell(spec: LabelSpec): string {
  if (spec.type === "static") return `\`${spec.name}\``;
  const where = { prefix: "starts with", suffix: "ends with", contains: "contains" }[spec.match];
  return `\`${spec.shape}\` (${where} \`${spec.stem}\`)`;
}

/** renderLabelsDoc is docs/labels.md, generated from the registry. */
function renderLabelsDoc(): string {
  const rows = SPECS.map(([key, spec]) =>
    [
      `\`${key}\``,
      nameCell(spec),
      spec.role,
      spec.kind,
      spec.owners.map((o) => `\`${o}\``).join("<br>"),
      mdCell(spec.doc),
    ].join(" | "),
  );
  return [
    "# Contract labels",
    "",
    "<!-- Generated from client/src/lib/labels.ts by client/src/lib/labels.test.ts. Do not edit by hand:",
    "     change the registry, then run `UPDATE_LABELS_DOC=1 npm test -- labels` in client/. -->",
    "",
    "The accessible names the tutorial, the first-use hints and the e2e suite rely on",
    "([ADR 0125](decisions/0125-a-walkthrough-that-keeps-up.md) §2,",
    "[ADR 0076](decisions/0076-tutorial.md) §2.4, [ADR 0111](decisions/0111-action-dock.md) §10).",
    "Each lives once, in `client/src/lib/labels.ts`, and its owner files render it as `L.<key>`",
    "(or `L.<key>(…)`), never as a literal. `labels.test.ts` fails when an owner stops rendering",
    "its entry, when a registered `aria` name is copied as a literal anywhere else under",
    "`client/src`, when a tutorial step or a first-use hint anchors to a name that is not",
    "registered, or when this file is stale.",
    "",
    "- **Kind** `aria` is an `aria-label`, or a dialog's or group's name passed through a prop",
    "  that becomes one; only these can be anchors. `text` is a name a button, link or menu item",
    "  gets from its visible text.",
    "- A name with a variable part is matched by its fixed stem when used as an anchor",
    "  (`L.<key>.any`), with the CSS operators `^=`, `$=` or `*=`.",
    "- A trigger's dialog is named by the trigger's reason, which comes from the game, so it has",
    "  no entry here.",
    "- To rename a label, change its value in the registry: tutorial steps and hints follow on",
    "  their own. The older e2e specs select on literals on purpose, so the nightly E2E goes red",
    '  until they are updated; run it on the branch (`gh workflow run "cmd_and_ctrl E2E" --ref <branch>`).',
    "",
    "| Key | Name | Role | Kind | Owners | What it is |",
    "|---|---|---|---|---|---|",
    ...rows.map((r) => `| ${r} |`),
    "",
  ].join("\n");
}

// ---- The guard ----

describe("the label registry", () => {
  it("gives each static entry its own name, and L the same value", () => {
    const seen = new Map<string, LabelKey>();
    for (const [key, spec] of SPECS) {
      if (spec.type !== "static") continue;
      const other = seen.get(spec.name);
      expect(other, `${key} and ${other} are both registered as "${spec.name}"`).toBeUndefined();
      seen.set(spec.name, key);
      expect(L[key] as unknown, key).toBe(spec.name);
    }
  });

  it("makes each dynamic name with its stem where the entry says", () => {
    for (const [key, spec] of SPECS) {
      if (spec.type !== "dynamic") continue;
      const made = (L[key] as unknown as (...a: unknown[]) => string)(...spec.example);
      expect(stemMatches(made, spec.stem, spec.match), `${key}: "${made}"`).toBe(true);
      if (spec.kind === "aria") {
        expect(registeredAria(made), `${key}: "${made}"`).not.toBeNull();
        const any = (L[key] as unknown as { any: LabelRef }).any;
        expect(registeredAria(any), key).toBe(key);
      }
    }
  });

  it("rule 1: every entry is rendered by each of its owners as L.<key>", () => {
    const problems: string[] = [];
    for (const [key, spec] of SPECS) {
      expect(spec.owners.length, `${key} names no owner file`).toBeGreaterThan(0);
      for (const owner of spec.owners) {
        const ref = new RegExp(`\\bL\\.${key}\\b`);
        const why = !existsSync(join(SRC, owner))
          ? "which does not exist"
          : ref.test(read(owner))
            ? null
            : `which no longer references L.${key}`;
        if (!why) continue;
        const users = [...(ANCHORED_BY.get(key) ?? [])];
        problems.push(
          `labels.ts: ${key} ("${show(spec)}") names ${owner} as an owner, ${why}. ` +
            (users.length > 0
              ? `Anchored to by: ${users.join("; ")}.`
              : "No tutorial step or hint anchors to it.") +
            " Render it there as L." +
            key +
            ", or fix the entry's owners.",
        );
      }
    }
    expect(problems, problems.join("\n")).toEqual([]);
  });

  it("rule 2: no file under client/src copies a registered aria name as a literal", () => {
    const names = SPECS.flatMap(([key, spec]) =>
      spec.type === "static" && spec.kind === "aria" ? [{ key, name: spec.name }] : [],
    );
    const problems: string[] = [];
    for (const file of sourceFiles(SRC)) {
      const rel = relative(SRC, file).split(sep).join("/");
      if (rel === "lib/labels.ts") continue;
      const lines = readFileSync(file, "utf8").split("\n");
      for (const { key, name } of names) {
        const re = literalCopy(name);
        lines.forEach((line, i) => {
          if (re.test(line)) {
            problems.push(
              `${rel}:${i + 1} renders "${name}" as a literal; it is the contract label L.${key} (labels.ts). Use L.${key} instead.`,
            );
          }
        });
      }
    }
    expect(problems, problems.join("\n")).toEqual([]);
  });

  it("rule 3: every tutorial anchor and detour anchor is a registered aria label", () => {
    expect(WALK.found.length).toBeGreaterThan(0);
    // The walk has to reach the detours to check them.
    for (const step of TUTORIAL_STEPS) {
      if (step.first) {
        expect(WALK.detoursBySteps.has(step.id), `no board reached step ${step.id}'s detour`).toBe(
          true,
        );
      }
    }
    const problems: string[] = [];
    for (const { where, anchor } of WALK.found) {
      if (!("label" in anchor)) continue;
      for (const [part, ref] of [
        ["label", anchor.label],
        ["within", anchor.within],
      ] as const) {
        if (ref === undefined) continue;
        if (registeredAria(ref) === null) {
          const s = typeof ref === "string" ? `"${ref}"` : `stem "${ref.stem}"`;
          problems.push(
            `tutorial ${where}: its ${part} ${s} is not a registered aria label in labels.ts`,
          );
        }
      }
    }
    expect([...new Set(problems)], [...new Set(problems)].join("\n")).toEqual([]);
  });

  it("rule 3: every hint's anchor is a registered aria label", () => {
    // The fixtures are walked too, so the walk is proven to reach a hint.
    expect(HINT_WALK.length).toBeGreaterThan(0);
    const problems: string[] = [];
    for (const { where, anchor } of HINT_WALK) {
      if (!("label" in anchor)) continue;
      for (const [part, ref] of [
        ["label", anchor.label],
        ["within", anchor.within],
      ] as const) {
        if (ref === undefined) continue;
        if (registeredAria(ref) === null) {
          const s = typeof ref === "string" ? `"${ref}"` : `stem "${ref.stem}"`;
          problems.push(`${where}: its ${part} ${s} is not a registered aria label in labels.ts`);
        }
      }
    }
    const unique = [...new Set(problems)];
    expect(unique, unique.join("\n")).toEqual([]);
  });

  it("rule 4: docs/labels.md is generated from the registry", () => {
    const want = renderLabelsDoc();
    if (process.env.UPDATE_LABELS_DOC) {
      writeFileSync(DOC, want);
      return;
    }
    const have = existsSync(DOC) ? readFileSync(DOC, "utf8") : "";
    expect(
      have === want,
      "docs/labels.md is stale: run `UPDATE_LABELS_DOC=1 npm test -- labels` in client/",
    ).toBe(true);
  });
});
