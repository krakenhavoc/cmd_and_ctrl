import { describe, it, expect } from "vitest";

import {
  assignedTotal,
  autoAssignAnswer,
  autoAssignHolds,
  autoAssignText,
  canDealDamage,
  initialShares,
  lethalOf,
  stepBlocker,
  stepTrample,
  toAnswer,
  trampleShort,
} from "./damageAssignment";
import type { DamageAssignmentView, PendingChoiceView } from "./protocol";

// damageAssignment.test.ts — #2956, ADR 0147: the prompt opens on the
// server's canonical split, the steppers move damage one point at a
// time, and a split that covers lethal for every blocker is sent for
// the player while "Auto-assign combat damage" is on.

// The report: Vivi Ornitier (44, trample) blocked by Solphim (lethal
// 4), Professional Face-Breaker (3) and Corsair Captain (2).
const vivi: DamageAssignmentView = {
  attacker_card_id: "vivi",
  blocker_card_ids: ["solphim", "breaker", "captain"],
  attacker_power: 44,
  allow_trample: true,
  lethal: [4, 3, 2],
  covers_lethal: true,
  suggested: {
    assignments: [
      { blocker_id: "solphim", amount: 4 },
      { blocker_id: "breaker", amount: 3 },
      { blocker_id: "captain", amount: 2 },
    ],
    trample_to_player: 35,
  },
};

const choice = (frame: DamageAssignmentView, chooser = "me"): PendingChoiceView =>
  ({
    id: "c1",
    kind: "damage_assignment",
    chooser,
    from_player: chooser,
    count: frame.blocker_card_ids.length,
    damage_assignment: frame,
  }) as PendingChoiceView;

describe("the pre-filled split", () => {
  it("opens on the suggested split, trample leftover on the player", () => {
    const s = initialShares(vivi);
    expect(s.amounts).toEqual({ solphim: 4, breaker: 3, captain: 2 });
    expect(s.trample).toBe(35);
    expect(assignedTotal(s)).toBe(44);
    expect(canDealDamage(vivi, s)).toBe(true);
  });

  it("opens at 0 when the server sent no suggestion", () => {
    const frame = { ...vivi, suggested: undefined };
    const s = initialShares(frame);
    expect(s).toEqual({ amounts: { solphim: 0, breaker: 0, captain: 0 }, trample: 0 });
    expect(canDealDamage(frame, s)).toBe(false);
  });

  it("never puts damage over without trample", () => {
    const s = initialShares({ ...vivi, allow_trample: false });
    expect(s.trample).toBe(0);
  });

  it("reads each blocker's lethal from the server", () => {
    expect(lethalOf(vivi, "breaker")).toBe(3);
    expect(lethalOf({ ...vivi, lethal: undefined }, "breaker")).toBeNull();
  });
});

describe("ticking damage up and down", () => {
  it("+ on a full split moves a point off the trample, − frees one", () => {
    let s = initialShares(vivi);
    s = stepBlocker(s, "captain", 1, 44);
    expect(s.amounts.captain).toBe(3);
    expect(s.trample).toBe(34);
    expect(assignedTotal(s)).toBe(44);

    s = stepBlocker(s, "captain", -1, 44);
    expect(s.amounts.captain).toBe(2);
    expect(assignedTotal(s)).toBe(43);
    // The freed point goes back over with the trample stepper.
    s = stepTrample(s, 1, 44);
    expect(s.trample).toBe(35);
    expect(stepTrample(s, 1, 44)).toBe(s);
  });

  it("stops at 0 and at the attacker's power", () => {
    const zero = initialShares({ ...vivi, suggested: undefined });
    expect(stepBlocker(zero, "solphim", -1, 44)).toBe(zero);
    const full = { amounts: { a: 2 }, trample: 0 };
    expect(stepBlocker(full, "a", 1, 2)).toBe(full);
    expect(stepBlocker(full, "nobody", 1, 2)).toBe(full);
  });

  it("refuses to trample over a blocker short of lethal (CR 702.19b)", () => {
    let s = initialShares(vivi);
    s = stepBlocker(s, "solphim", -1, 44);
    s = stepTrample(s, 1, 44);
    expect(assignedTotal(s)).toBe(44);
    expect(trampleShort(vivi, s)).toEqual(["solphim"]);
    expect(canDealDamage(vivi, s)).toBe(false);
  });

  it("answers in the frame's blocker order", () => {
    expect(toAnswer(vivi, initialShares(vivi))).toEqual({
      assignments: [
        { blocker_id: "solphim", amount: 4 },
        { blocker_id: "breaker", amount: 3 },
        { blocker_id: "captain", amount: 2 },
      ],
      trample_to_player: 35,
    });
  });
});

// CR 702.2c with CR 702.19b: a 5-power deathtouch trampler blocked by
// three 4/4s. 1 is lethal from deathtouch, so the server's split is
// 1/1/1 and 2 over, and it covers lethal for every blocker.
const deathtouch: DamageAssignmentView = {
  attacker_card_id: "dt",
  blocker_card_ids: ["a", "b", "c"],
  attacker_power: 5,
  allow_trample: true,
  has_deathtouch: true,
  lethal: [1, 1, 1],
  covers_lethal: true,
  suggested: {
    assignments: [
      { blocker_id: "a", amount: 1 },
      { blocker_id: "b", amount: 1 },
      { blocker_id: "c", amount: 1 },
    ],
    trample_to_player: 2,
  },
};

describe("trample with deathtouch", () => {
  it("pre-fills 1 on each blocker and the rest over", () => {
    const s = initialShares(deathtouch);
    expect(s.amounts).toEqual({ a: 1, b: 1, c: 1 });
    expect(s.trample).toBe(2);
    expect(canDealDamage(deathtouch, s)).toBe(true);
  });

  it("is auto-assigned: 1 / 1 / 1 and 2 to the player", () => {
    expect(autoAssignAnswer(choice(deathtouch), "me", true)).toEqual({
      assignments: [
        { blocker_id: "a", amount: 1 },
        { blocker_id: "b", amount: 1 },
        { blocker_id: "c", amount: 1 },
      ],
      trample_to_player: 2,
    });
  });

  it("counts a blocker at 0 as short of its 1 lethal", () => {
    let s = stepBlocker(initialShares(deathtouch), "b", -1, 5);
    s = stepTrample(s, 1, 5);
    expect(trampleShort(deathtouch, s)).toEqual(["b"]);
    expect(canDealDamage(deathtouch, s)).toBe(false);
  });
});

describe("damage already marked", () => {
  // CR 702.19c: a 2/4 with 3 damage marked needs 1 more. The server
  // sends lethal 1 for it, and the split follows.
  const marked: DamageAssignmentView = {
    attacker_card_id: "atk",
    blocker_card_ids: ["hurt", "bear"],
    attacker_power: 4,
    allow_trample: true,
    lethal: [1, 2],
    covers_lethal: true,
    suggested: {
      assignments: [
        { blocker_id: "hurt", amount: 1 },
        { blocker_id: "bear", amount: 2 },
      ],
      trample_to_player: 1,
    },
  };

  it("pre-fills lethal less what is marked, and tramples the rest", () => {
    const s = initialShares(marked);
    expect(s).toEqual({ amounts: { hurt: 1, bear: 2 }, trample: 1 });
    expect(trampleShort(marked, s)).toEqual([]);
    expect(autoAssignAnswer(choice(marked), "me", true)?.trample_to_player).toBe(1);
  });
});

describe("auto-assign", () => {
  it("sends the suggested split when it covers lethal and the setting is on", () => {
    expect(autoAssignAnswer(choice(vivi), "me", true)).toEqual(toAnswer(vivi, initialShares(vivi)));
  });

  it("asks when the setting is off, the damage falls short, or the prompt is not mine", () => {
    expect(autoAssignAnswer(choice(vivi), "me", false)).toBeNull();
    expect(autoAssignAnswer(choice({ ...vivi, covers_lethal: false }), "me", true)).toBeNull();
    expect(autoAssignAnswer(choice(vivi, "them"), "me", true)).toBeNull();
    expect(autoAssignAnswer(choice({ ...vivi, suggested: undefined }), "me", true)).toBeNull();
    expect(autoAssignAnswer(null, "me", true)).toBeNull();
  });

  it("holds the sheet until the record expires", () => {
    const c = choice(vivi);
    expect(autoAssignHolds(c, "me", true, {})).toBe(true);
    expect(autoAssignHolds(c, "me", true, { c1: { sentAt: 1, expired: false } })).toBe(true);
    expect(autoAssignHolds(c, "me", true, { c1: { sentAt: 1, expired: true } })).toBe(false);
    expect(autoAssignHolds(c, "me", false, {})).toBe(false);
  });

  it("says what went where", () => {
    const names: Record<string, string> = {
      solphim: "Solphim",
      breaker: "Professional Face-Breaker",
      captain: "Corsair Captain",
    };
    const text = autoAssignText(
      "Vivi Ornitier",
      toAnswer(vivi, initialShares(vivi)),
      (id) => names[id],
      "Bob",
    );
    expect(text).toBe(
      "Vivi Ornitier: 4 to Solphim, 3 to Professional Face-Breaker, 2 to Corsair Captain, 35 to Bob (trample)",
    );
  });
});
