// passReason.test.ts — #3343: why an automatic pass passed, and when the
// chip's Undo still works.

import { describe, it, expect, vi } from "vitest";

import { autopassDecision, type AutopassGates } from "./autopassDecision";
import {
  BLUFF_PASS_REASON,
  HOLD_PASS_REASON,
  autopassReason,
  passNoticeRequest,
  passNoticeText,
  passNoticeUndoable,
  timedPassReason,
  type PassNotice,
  type PassNoticeFacts,
} from "./passReason";
import { L } from "./labels";
import type { PlayerView } from "./protocol";

// The autopassDecision.test.ts baseline: an unstopped window that passes.
function gates(overrides: Partial<AutopassGates> = {}): AutopassGates {
  return {
    viewerHasPriority: true,
    tableBusy: false,
    hasPendingChoice: false,
    owesBlockDecision: false,
    owesAttackRequirement: false,
    loopSuspended: false,
    step: "upkeep",
    autopassToggle: false,
    viewerIsActive: false,
    manualStop: false,
    passMode: "smart",
    stackEmpty: true,
    holdPriority: false,
    autoPassOwnStack: true,
    ownsEveryStackItem: false,
    stepStop: undefined,
    stepStopsOnlyWhenCanAct: true,
    hasResponse: false,
    hasPlay: false,
    engineMayMissMana: false,
    combatWindow: false,
    oppEndWindow: false,
    bluffCounter: false,
    bluffInstant: false,
    bluffManual: false,
    ...overrides,
  };
}

describe("autopassReason — worded from the branch that passed", () => {
  it("names Skip to my turn", () => {
    expect(autopassReason(gates({ autopassToggle: true }))).toBe("skipping to your turn");
  });

  it("names an opponent's item with nothing to answer it", () => {
    expect(autopassReason(gates({ stackEmpty: false }))).toBe("nothing to respond with");
  });

  it("names a stack that is all the viewer's own", () => {
    expect(autopassReason(gates({ stackEmpty: false, ownsEveryStackItem: true }))).toBe(
      "everything on the stack is yours",
    );
  });

  it("names an unticked step", () => {
    expect(autopassReason(gates({ step: "upkeep" }))).toBe("no stop set for Upkeep");
  });

  it("names a ticked step with nothing to do in it", () => {
    expect(
      autopassReason(gates({ step: "declare_attackers", stepStop: true, hasPlay: false })),
    ).toBe("nothing to do in Declare Attackers");
  });

  it("names a combat or end-step window with nothing to respond with", () => {
    expect(autopassReason(gates({ step: "declare_blockers", combatWindow: true }))).toBe(
      "nothing to respond with in Declare Blockers",
    );
    expect(autopassReason(gates({ step: "end", oppEndWindow: true }))).toBe(
      "nothing to respond with in End",
    );
  });

  it("is a sentence, never empty, wherever the decision says pass", () => {
    // Walk the gates that can differ and check every pass has a reason.
    const bools = [false, true];
    let passes = 0;
    for (const autopassToggle of bools)
      for (const stackEmpty of bools)
        for (const ownsEveryStackItem of bools)
          for (const hasResponse of bools)
            for (const combatWindow of bools)
              for (const stepStop of [undefined, true, false])
                for (const passMode of ["smart", "careful", "manual"] as const) {
                  const g = gates({
                    autopassToggle,
                    stackEmpty,
                    ownsEveryStackItem,
                    hasResponse,
                    combatWindow,
                    stepStop,
                    passMode,
                  });
                  if (autopassDecision(g) !== "pass") continue;
                  passes++;
                  expect(autopassReason(g).trim(), JSON.stringify(g)).not.toBe("");
                }
    expect(passes).toBeGreaterThan(10);
  });
});

describe("timedPassReason", () => {
  it("says the bluff timer ended for a bluff verdict, whichever timer was armed", () => {
    const g = gates({ stackEmpty: false });
    expect(timedPassReason("bluff", { kind: "bluff", manual: false }, g)).toBe(BLUFF_PASS_REASON);
    expect(timedPassReason("hold", { kind: "bluff", manual: false }, g)).toBe(BLUFF_PASS_REASON);
  });

  it("says the hold ran out for a plain pass after a stack hold", () => {
    expect(timedPassReason("hold", "pass", gates({ stackEmpty: false }))).toBe(HOLD_PASS_REASON);
  });

  it("falls back to the ordinary reason when settings moved during a bluff wait", () => {
    expect(timedPassReason("bluff", "pass", gates({ stackEmpty: false }))).toBe(
      "nothing to respond with",
    );
  });
});

const seat = (over: Partial<PlayerView> = {}): PlayerView =>
  ({ id: "me", seat: 0, undos_remaining: 3, ...over }) as unknown as PlayerView;

const notice: PassNotice = { frame: "f-1", reason: "nothing to respond with", actionsSent: 4 };

function facts(over: Partial<PassNoticeFacts> = {}): PassNoticeFacts {
  return {
    ack: { id: "f-1", seq: 10 },
    lastSeq: 10,
    actionsSent: 4,
    seat: seat(),
    undoBlocked: null,
    ...over,
  };
}

describe("passNoticeUndoable — when the existing undo would take this pass back", () => {
  it("is true once the server acked the pass and nothing has happened since", () => {
    expect(passNoticeUndoable(notice, facts())).toBe(true);
    // A later snapshot is fine as long as the undo gates still say yes.
    expect(passNoticeUndoable(notice, facts({ lastSeq: 12 }))).toBe(true);
  });

  it("is false before the pass is acknowledged", () => {
    expect(passNoticeUndoable(notice, facts({ ack: null }))).toBe(false);
    // An ack for some other frame is not ours.
    expect(passNoticeUndoable(notice, facts({ ack: { id: "f-0", seq: 9 } }))).toBe(false);
    // The ack's snapshot has not been rendered yet.
    expect(passNoticeUndoable(notice, facts({ lastSeq: 9 }))).toBe(false);
  });

  it("is false once the viewer sent anything else", () => {
    expect(passNoticeUndoable(notice, facts({ actionsSent: 5 }))).toBe(false);
  });

  it("is false when the existing undo is blocked (another seat acted, no budget, game over)", () => {
    expect(
      passNoticeUndoable(
        notice,
        facts({ undoBlocked: "the last action was not yours — you can only undo your own" }),
      ),
    ).toBe(false);
    expect(passNoticeUndoable(notice, facts({ undoBlocked: "no undos remaining this turn" }))).toBe(
      false,
    );
  });

  it("is false when an automatic answer sits on top of the pass", () => {
    expect(passNoticeUndoable(notice, facts({ seat: seat({ undo_auto_answer: 12 }) }))).toBe(false);
  });
});

describe("the chip", () => {
  it("says what happened in the issue's words", () => {
    expect(passNoticeText(notice)).toBe("Passed for you: nothing to respond with");
  });

  it("is a step-rank status named for the e2e suite, with the reason as its question", () => {
    const r = passNoticeRequest(notice, false, () => {});
    expect(r.rank).toBe("step");
    expect(r.label).toBe(L.autoPassNotice);
    expect(r.question).toBe("Passed for you: nothing to respond with");
  });

  it("carries an Undo button only while undoable, and it calls back", () => {
    expect(passNoticeRequest(notice, false, () => {}).row).toBeUndefined();
    const onUndo = vi.fn();
    const row = passNoticeRequest(notice, true, onUndo).row!;
    expect(row).toHaveLength(1);
    expect(row[0].label).toBe("Undo");
    // Named apart from the dock's "Undo (N left)", which /^Undo/ counts.
    expect(row[0].ariaLabel).toBe(L.undoAutoPass);
    expect(row[0].ariaLabel).not.toMatch(/^Undo/);
    row[0].onPress?.();
    expect(onUndo).toHaveBeenCalledOnce();
  });
});
