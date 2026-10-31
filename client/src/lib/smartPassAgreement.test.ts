// The TypeScript half of the smart auto-pass agreement contract (ADR 0009's
// amendment of 2026-10-30, #4734), as timingAgreement.test.ts is for cast
// timing.
//
// server/internal/legal/testdata/smart_pass_agreement.json holds real game
// frames, each as the wire carries it to one seat (`legal_moves` with the
// server's `answers_now` on them) and a verdict declared BY HAND, with the
// reason: "stop" (the table keeps the cursor) or "pass" (smart auto-pass
// passes for the player). The verdicts assume the default settings.
// server/internal/legal/smart_pass_agreement_test.go checks the Go side
// writes the frames it declares; this file asserts autopassDecision reaches
// every frame's verdict from the frame alone. The MCP seat asserts the same
// frames from smart_test.go (slice 4). Either reader drifting fails a test,
// and neither is the other's oracle.
//
// Regenerate the fixture after a change to the enumerator, the window match
// or the view:
//
//   scripts/go-docker.sh go test ./internal/legal/ -run TestSmartPassAgreement -update
//
// A frame the browser disagrees with is not settled by this file. Work out
// which side is wrong first: the fixture's declaration is a claim about
// ADR 0009, and it can be the wrong one.

import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

import { describe, it, expect } from "vitest";

import { autopassDecision, stepStopFor, type AutopassGates } from "./autopassDecision";
import { engineMayMissMana } from "./engineMayMissMana";
import { ownsEveryStackItem } from "./holdPriority";
import type { GameView, LegalMoveView } from "./protocol";
import {
  hasMatchedResponse,
  hasPlay,
  hasResponse,
  hasUnmatchedAnswer,
  keyWindow,
  type ResponseCategories,
} from "./responseWindow";
import { defaultSettings } from "./settings";
import { hasPriority, isActivePlayer, stackEmpty } from "./timing";
import {
  owesAttackRequirement,
  owesBlockDecision,
  owesChosenDeclaration,
  owesFreeCast,
} from "./priority";

interface FixtureMove {
  source: string;
  name: string;
  kind: string;
  // true, false, or null for a move that carries no verdict.
  answers_now: boolean | null;
  why: string;
}

interface Frame {
  name: string;
  note: string;
  viewer: string;
  view: GameView;
  moves: FixtureMove[];
  verdict: "stop" | "pass";
  why: string;
}

const fixtureURL = new URL(
  "../../../server/internal/legal/testdata/smart_pass_agreement.json",
  import.meta.url,
);

const frames: Frame[] = JSON.parse(readFileSync(fileURLToPath(fixtureURL), "utf8"));

// gatesFor builds the decision's inputs for a frame at the default
// settings, as Game.svelte's autopassGates does for the viewer's own seat:
// no pin, no Skip to my turn, no hold, no bluff armed, nothing owed that
// the frame does not show. When Game.svelte reads a new input, read it here.
function gatesFor(view: GameView, me: string, strip = false): AutopassGates {
  const gp = defaultSettings().gameplay;
  const cats: ResponseCategories = {
    counter: gp.respondCounterspells,
    instant: gp.respondInstants,
    valueInstant: gp.respondValueInstants,
    ability: gp.respondAbilities,
    untargeted: gp.respondUntargetedAbilities,
    special: gp.respondSpecialActions,
  };
  const v = strip ? withoutVerdicts(view) : view;
  const step = v.turn?.step;
  const viewerIsActive = isActivePlayer(v, me);
  const kw = keyWindow(v, me);
  const respond = hasResponse(v, me, cats);
  return {
    viewerHasPriority: hasPriority(v, me),
    tableBusy: v.mulligans_open || v.state !== "active",
    hasPendingChoice: !!v.pending_choices?.some((c) => c.chooser === me),
    owesBlockDecision: owesBlockDecision(v, me),
    owesAttackRequirement: owesAttackRequirement(v, me) || owesChosenDeclaration(v, me),
    owesFreeCast: owesFreeCast(v, me),
    loopSuspended: false,
    step,
    autopassToggle: false,
    viewerIsActive,
    manualStop: false,
    passMode: gp.passMode,
    stackEmpty: stackEmpty(v),
    holdPriority: false,
    autoPassOwnStack: gp.autoPassOwnStack,
    ownsEveryStackItem: ownsEveryStackItem(v, me),
    stepStop: stepStopFor(gp, step, viewerIsActive),
    stepStopsOnlyWhenCanAct: gp.stepStopsOnlyWhenCanAct,
    hasResponse: respond,
    hasPlay: hasPlay(v, me, cats),
    matchedResponse: hasMatchedResponse(v, me, cats),
    heldUnmatchedAnswer: !respond && hasUnmatchedAnswer(v, me, cats),
    engineMayMissMana: engineMayMissMana(v, me),
    combatWindow: kw.combat,
    oppEndWindow: kw.oppEnd,
    bluffCounter: false,
    bluffInstant: false,
    bluffManual: false,
  };
}

// withoutVerdicts is the frame as a server older than #4734 sent it: the
// same moves, none of them carrying answers_now.
function withoutVerdicts(view: GameView): GameView {
  return {
    ...view,
    legal_moves: (view.legal_moves ?? []).map((m): LegalMoveView => {
      const copy = { ...m };
      delete copy.answers_now;
      return copy;
    }),
  };
}

// browserVerdict maps the decision onto the fixture's two words. A bluff
// is a hold with a face on it, and nothing here arms one.
function browserVerdict(g: AutopassGates): "stop" | "pass" {
  return autopassDecision(g) === "pass" ? "pass" : "stop";
}

describe("autopassDecision agrees with smart_pass_agreement.json", () => {
  it("the fixture is present and non-trivial", () => {
    expect(frames.length).toBeGreaterThan(0);
    for (const f of frames) {
      expect(f.moves.length, `${f.name} declares no moves`).toBeGreaterThan(0);
      expect(["stop", "pass"], `${f.name} verdict`).toContain(f.verdict);
      expect(f.why, `${f.name} gives no reason`).not.toBe("");
    }
  });

  for (const f of frames) {
    it(`${f.name}: ${f.verdict}, ${f.why}`, () => {
      expect(browserVerdict(gatesFor(f.view, f.viewer)), `${f.note} | ${f.why}`).toBe(f.verdict);
    });
  }
});

// The fixture's `moves` list is its own statement of what the server sent.
// Check it against the wire frame the browser reads, so a declared move
// cannot drift from the legal_moves the verdicts above are computed from.
describe("each frame's declared moves are the ones on its wire frame", () => {
  for (const f of frames) {
    it(f.name, () => {
      const wire = f.view.legal_moves ?? [];
      for (const m of f.moves) {
        const found = wire.find((w) => w.source === m.source && w.kind === m.kind);
        expect(found, `${m.name} (${m.kind} ${m.source}) is not in legal_moves`).toBeDefined();
        expect(found?.answers_now ?? null, `${m.name}: ${m.why}`).toBe(m.answers_now);
      }
    });
  }
});

describe("no fixture frame leaks another seat's moves", () => {
  for (const f of frames) {
    it(f.name, () => {
      for (const m of f.view.legal_moves ?? []) {
        expect(m.player, `${f.name} leaked a move for ${m.player}`).toBe(f.viewer);
      }
    });
  }
});

// The receipt for #4734. Take the verdicts off the same frames, as an older
// server sent them, and the browser stops at every opponent window of the
// 3376a15c game: the 23 pointless stops the issue counted. The frames that
// carry the verdict and pass are the ones that changed.
describe("without answers_now the browser stops where #4734 reported", () => {
  const passes = frames.filter((f) => f.name.startsWith("3376a15c_") && f.verdict === "pass");

  it("the fixture holds the game's passing windows", () => {
    expect(passes.length).toBeGreaterThan(0);
  });

  for (const f of passes) {
    it(`${f.name} stopped before, passes now`, () => {
      expect(browserVerdict(gatesFor(f.view, f.viewer, true)), f.note).toBe("stop");
      expect(browserVerdict(gatesFor(f.view, f.viewer)), f.note).toBe("pass");
    });
  }

  it("a pass that a held answer caused says so on the chip's gate", () => {
    for (const f of passes) {
      expect(gatesFor(f.view, f.viewer).heldUnmatchedAnswer, f.name).toBe(true);
    }
  });
});

// The owner's decision of 2026-10-31 (#4734): animate and combat_grant (a
// manland, a Crew row, a keyword grant) no longer stop the table at the
// viewer's own beginning of combat. The step is not ticked by default, so
// the frame passes; a player who ticks it still stops there through the
// step grid, with "only when I can do something" on, because the server
// sends no verdict on those moves there and they keep the class they had.
describe("your own beginning of combat with a manland or a keyword grant", () => {
  const f = frames.find((x) => x.name === "own_beginning_of_combat_manland_and_keyword_grant");

  it("the fixture holds the frame", () => {
    expect(f).toBeDefined();
    expect(f?.view.turn.step).toBe("begin_combat");
  });

  it("passes at the default settings, where the step is not ticked", () => {
    if (!f) return;
    const g = gatesFor(f.view, f.viewer);
    expect(g.stepStop).toBe(false);
    expect(browserVerdict(g)).toBe("pass");
    expect(g.matchedResponse).toBe(false);
  });

  it("still stops once the player ticks the step", () => {
    if (!f) return;
    const g = gatesFor(f.view, f.viewer);
    expect(g.stepStopsOnlyWhenCanAct).toBe(true);
    expect(browserVerdict({ ...g, stepStop: true })).toBe("stop");
  });
});
