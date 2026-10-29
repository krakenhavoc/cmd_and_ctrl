// attackStage.test.ts — #3266: declaring attackers by toggling them.

import { describe, it, expect } from "vitest";

import {
  isStaged,
  pruneStage,
  soleTarget,
  stageAt,
  stageMany,
  stagedParams,
  unstage,
  unstageMany,
  type StagedAttack,
} from "./attackStage";
import { primaryKind } from "./dockHint";
import type { GameView } from "./protocol";

const bear = { kind: "attacker" as const, cardID: "bear" };
const elf = { kind: "attacker" as const, cardID: "elf" };

describe("staging attackers", () => {
  it("stages a creature at its target, and a second toggle takes it out", () => {
    const one = stageAt([], bear, "bob")!;
    expect(one).toEqual([{ attacker: "bear", target: "bob" }]);
    expect(isStaged(one, "bear")).toBe(true);
    expect(unstage(one, "bear")).toEqual([]);
  });

  it("re-aims a staged creature rather than staging it twice", () => {
    const one = stageAt(stageAt([], bear, "bob")!, elf, "cy")!;
    expect(stageAt(one, bear, "cy")).toEqual([
      { attacker: "elf", target: "cy" },
      { attacker: "bear", target: "cy" },
    ]);
  });

  it("stages nothing while the exert question is open, and carries a yes", () => {
    expect(stageAt([], { ...bear, exert: null }, "bob")).toBeNull();
    expect(stageAt([], { ...bear, exert: true }, "bob")).toEqual([
      { attacker: "bear", target: "bob", exert: true },
    ]);
    expect(stageAt([], { ...bear, exert: false }, "bob")).toEqual([
      { attacker: "bear", target: "bob" },
    ]);
  });

  it("stages on the first click only with one legal attack target", () => {
    expect(soleTarget(["bob"])).toBe("bob");
    expect(soleTarget(["bob", "walker"])).toBeNull();
    expect(soleTarget([])).toBeNull();
  });

  it("drops creatures that can no longer attack, and keeps the array when none did", () => {
    const stage: StagedAttack[] = [
      { attacker: "bear", target: "bob" },
      { attacker: "elf", target: "bob" },
    ];
    expect(pruneStage(stage, (id) => id !== "elf")).toEqual([{ attacker: "bear", target: "bob" }]);
    expect(pruneStage(stage, () => true)).toBe(stage);
  });

  it("sends the whole stage as one declare_attackers with auto_tap", () => {
    expect(stagedParams([])).toBeNull();
    expect(
      stagedParams([
        { attacker: "bear", target: "bob" },
        { attacker: "elf", target: "cy", exert: true },
      ]),
    ).toEqual({
      attackers: [
        { attacker: "bear", target: "bob" },
        { attacker: "elf", target: "cy", exert: true },
      ],
      auto_tap: true,
    });
  });
});

describe("staging a token group's members (#4423)", () => {
  const stage: StagedAttack[] = [
    { attacker: "bear", target: "cy" },
    { attacker: "wolf1", target: "cy" },
  ];

  it("adds the group beside what is already staged, leaving it untouched", () => {
    const next = stageMany(stage, [
      { attacker: "wolf2", target: "bob" },
      { attacker: "wolf3", target: "bob" },
    ]);
    expect(next).toEqual([
      { attacker: "bear", target: "cy" },
      { attacker: "wolf1", target: "cy" },
      { attacker: "wolf2", target: "bob" },
      { attacker: "wolf3", target: "bob" },
    ]);
  });

  it("re-aims a member that is already staged instead of staging it twice", () => {
    expect(stageMany(stage, [{ attacker: "wolf1", target: "bob" }])).toEqual([
      { attacker: "bear", target: "cy" },
      { attacker: "wolf1", target: "bob" },
    ]);
  });

  it("takes some members back out, which is how the count changes", () => {
    expect(unstageMany(stage, ["wolf1", "nobody"])).toEqual([{ attacker: "bear", target: "cy" }]);
    expect(unstageMany(stage, [])).toEqual(stage);
  });
});

describe("the primary with attackers staged", () => {
  const view = (step: string) =>
    ({
      battlefield: { kind: "battlefield", count: 0, cards: [] },
      stack: { kind: "stack", count: 0, cards: [] },
      stack_items: [],
      turn: { step },
    }) as unknown as GameView;

  it("reads Attack with N in declare attackers, No attacks with none staged", () => {
    expect(primaryKind(view("declare_attackers"), 3, 2)).toBe("attack");
    expect(primaryKind(view("declare_attackers"), 3, 0)).toBe("noAttacks");
    expect(primaryKind(view("declare_blockers"), 0, 2)).toBe("next");
  });
});
