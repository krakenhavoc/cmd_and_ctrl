// #3898: declaring blockers on the board — click your creature, then
// the attacker. These pin the state machine: select, pair, clear, change,
// multi-block, illegal pairs, the menace group, and the pile helpers.

import { describe, expect, it } from "vitest";

import {
  NO_BLOCK_PICK,
  blockHighlight,
  blockTargetFor,
  blockTargets,
  canSelectBlocker,
  fanPile,
  fansOnClick,
  inBlockGroup,
  liveBlockPick,
  reduceBlockPick,
  releaseSends,
  type BlockBoard,
  type BlockContext,
  type BlockPickState,
} from "./blockPick";
import { legalActionsOf } from "./legalActions";
import type { CardView, GameView } from "./protocol";

const ME = "me";
const OPP = "opp";

function creature(id: string, over: Partial<CardView> = {}): CardView {
  return {
    instance_id: id,
    name: id,
    owner: ME,
    controller: ME,
    type_line: "Creature — Bear",
    ...over,
  } as CardView;
}

function attacker(id: string, over: Partial<CardView> = {}): CardView {
  return creature(id, {
    owner: OPP,
    controller: OPP,
    attacking_target: ME,
    attacking_target_kind: "player",
    defending_player: ME,
    tapped: true,
    ...over,
  });
}

// ctx builds a frame whose digest lets each creature in `blocks` block
// the attackers listed for it.
function ctx(cards: CardView[], blocks: Record<string, string[]> = {}): BlockContext {
  const sources: Record<string, { kinds: string[]; moves: number; blocks: string[] }> = {};
  for (const [id, atks] of Object.entries(blocks)) {
    sources[id] = { kinds: ["block"], moves: atks.length, blocks: atks };
  }
  const view = {
    battlefield: { kind: "battlefield", count: cards.length, cards },
    legal_actions: { sources },
    seats: [],
  } as unknown as GameView;
  return { view, viewerID: ME, legal: legalActionsOf(view) };
}

const step = (
  s: BlockPickState,
  c: BlockContext,
  ...evs: Parameters<typeof reduceBlockPick>[1][]
) => {
  let state = s;
  const sent: unknown[] = [];
  let refused: string | undefined;
  for (const ev of evs) {
    const r = reduceBlockPick(state, ev, c);
    state = r.state;
    sent.push(...r.send);
    refused = r.refused ?? refused;
  }
  return { state, sent, refused };
};

describe("blockPick: select, then pair", () => {
  const c = ctx([creature("bear"), creature("wolf"), attacker("ogre"), attacker("imp")], {
    bear: ["ogre", "imp"],
    wolf: ["ogre"],
  });

  it("selects your creature, and a second click lets it go", () => {
    const one = step(NO_BLOCK_PICK, c, { kind: "own", id: "bear" });
    expect(one.state.selected).toBe("bear");
    expect(one.sent).toEqual([]);
    expect(step(one.state, c, { kind: "own", id: "bear" }).state.selected).toBeNull();
  });

  it("a click on another creature moves the selection", () => {
    const r = step(NO_BLOCK_PICK, c, { kind: "own", id: "bear" }, { kind: "own", id: "wolf" });
    expect(r.state.selected).toBe("wolf");
  });

  it("then the attacker declares the block and lets go", () => {
    const r = step(NO_BLOCK_PICK, c, { kind: "own", id: "bear" }, { kind: "attacker", id: "ogre" });
    expect(r.sent).toEqual([
      { type: "declare_blocker", params: { blocker: "bear", attacker: "ogre" } },
    ]);
    expect(r.state).toEqual(NO_BLOCK_PICK);
  });

  it("an attacker click with nothing selected does nothing", () => {
    const r = step(NO_BLOCK_PICK, c, { kind: "attacker", id: "ogre" });
    expect(r.sent).toEqual([]);
    expect(r.state).toEqual(NO_BLOCK_PICK);
  });

  it("a creature that takes no part can't be selected", () => {
    const tapped = ctx([creature("bear", { tapped: true }), attacker("ogre")]);
    expect(canSelectBlocker(tapped, "bear")).toBe(false);
    expect(step(NO_BLOCK_PICK, tapped, { kind: "own", id: "bear" }).state.selected).toBeNull();
    // Nor an opponent's creature, nor a non-creature.
    expect(canSelectBlocker(c, "ogre")).toBe(false);
  });

  it("cancel drops the selection", () => {
    const r = step(NO_BLOCK_PICK, c, { kind: "own", id: "bear" }, { kind: "cancel" });
    expect(r.state).toEqual(NO_BLOCK_PICK);
  });
});

describe("blockPick: illegal pairs", () => {
  const c = ctx(
    [
      creature("bear", {
        block_refusals: [
          {
            attacker: "drake",
            reason: "flying",
            message: "drake has flying, and bear has neither flying nor reach.",
          },
        ],
      }),
      attacker("ogre"),
      attacker("drake", { keywords: ["flying"] } as Partial<CardView>),
    ],
    { bear: ["ogre"] },
  );

  it("marks the refused attacker with the server's sentence", () => {
    const t = blockTargets(c, "bear");
    expect(t.get("ogre")).toEqual({ kind: "block" });
    expect(t.get("drake")).toEqual({
      kind: "refused",
      reason: "drake has flying, and bear has neither flying nor reach.",
    });
  });

  it("a click on it sends nothing, keeps the selection and says why", () => {
    const r = step(
      NO_BLOCK_PICK,
      c,
      { kind: "own", id: "bear" },
      { kind: "attacker", id: "drake" },
    );
    expect(r.sent).toEqual([]);
    expect(r.state.selected).toBe("bear");
    expect(r.refused).toMatch(/flying/);
  });

  it("a creature that can block nothing is still selectable, to show why", () => {
    const only = ctx([
      creature("bear", {
        block_refusals: [{ attacker: "drake", reason: "flying", message: "no" }],
      }),
      attacker("drake"),
    ]);
    expect(canSelectBlocker(only, "bear")).toBe(true);
  });
});

describe("blockPick: clear and change", () => {
  const c = ctx(
    [
      creature("bear", { blocking_target: "ogre" }),
      creature("wolf"),
      attacker("ogre"),
      attacker("imp"),
    ],
    { wolf: ["ogre", "imp"] },
  );

  it("a declared blocker can be selected", () => {
    expect(step(NO_BLOCK_PICK, c, { kind: "own", id: "bear" }).state.selected).toBe("bear");
  });

  it("its own attacker again takes the block back", () => {
    expect(blockTargets(c, "bear").get("ogre")).toEqual({ kind: "remove" });
    const r = step(NO_BLOCK_PICK, c, { kind: "own", id: "bear" }, { kind: "attacker", id: "ogre" });
    expect(r.sent).toEqual([{ type: "remove_from_combat", params: { instance_id: "bear" } }]);
    expect(r.state).toEqual(NO_BLOCK_PICK);
  });

  it("another attacker changes the block", () => {
    const r = step(NO_BLOCK_PICK, c, { kind: "own", id: "bear" }, { kind: "attacker", id: "imp" });
    expect(r.sent).toEqual([
      { type: "declare_blocker", params: { blocker: "bear", attacker: "imp" } },
    ]);
  });

  it("the dock's Remove block takes it back", () => {
    const r = step(NO_BLOCK_PICK, c, { kind: "own", id: "bear" }, { kind: "remove" });
    expect(r.sent).toEqual([{ type: "remove_from_combat", params: { instance_id: "bear" } }]);
  });
});

describe("blockPick: a creature that can block several", () => {
  const guard = (over: Partial<CardView> = {}) => creature("guard", { block_capacity: 2, ...over });

  it("stays selected while it has room, then lets go", () => {
    const c0 = ctx([guard(), attacker("ogre"), attacker("imp")], { guard: ["ogre", "imp"] });
    const one = step(
      NO_BLOCK_PICK,
      c0,
      { kind: "own", id: "guard" },
      { kind: "attacker", id: "ogre" },
    );
    expect(one.state.selected).toBe("guard");
    // The server's frame: it now blocks the ogre.
    const c1 = ctx([guard({ blocking_target: "ogre" }), attacker("ogre"), attacker("imp")], {
      guard: ["imp"],
    });
    const two = step(one.state, c1, { kind: "attacker", id: "imp" });
    expect(two.sent).toEqual([
      { type: "declare_blocker", params: { blocker: "guard", attacker: "imp" } },
    ]);
    expect(two.state.selected).toBeNull();
  });

  it("taking one block back keeps the other", () => {
    const c = ctx([
      guard({ blocking_target: "ogre", blocking_targets: ["ogre", "imp"] }),
      attacker("ogre"),
      attacker("imp"),
    ]);
    const r = step(NO_BLOCK_PICK, c, { kind: "own", id: "guard" }, { kind: "attacker", id: "imp" });
    expect(r.sent).toEqual([
      { type: "remove_from_combat", params: { instance_id: "guard" } },
      { type: "declare_blockers", params: { blocks: [{ blocker: "guard", attacker: "ogre" }] } },
    ]);
  });
});

describe("blockPick: an attacker that takes two blockers (menace)", () => {
  const refusal = (atk: string) => ({
    attacker: atk,
    reason: "too_few_blockers",
    message: `${atk} can't be blocked by fewer than two creatures.`,
  });
  const c = ctx(
    [
      creature("bear", { block_refusals: [refusal("brute")] }),
      creature("wolf", { block_refusals: [refusal("brute")] }),
      creature("cat", {
        block_refusals: [refusal("brute"), { attacker: "drake", reason: "flying", message: "x" }],
      }),
      attacker("brute", { min_blockers: 2 }),
      attacker("drake"),
    ],
    { bear: ["brute"], wolf: ["brute"], cat: ["brute"] },
  );

  it("the click starts a group, and a second creature completes it", () => {
    const one = step(
      NO_BLOCK_PICK,
      c,
      { kind: "own", id: "bear" },
      { kind: "attacker", id: "brute" },
    );
    expect(one.sent).toEqual([]);
    expect(one.state.group).toEqual({ attacker: "brute", blockers: ["bear"], need: 2 });
    expect(one.state.selected).toBeNull();
    const two = step(one.state, c, { kind: "own", id: "wolf" });
    expect(two.sent).toEqual([
      {
        type: "declare_blockers",
        params: {
          blocks: [
            { blocker: "bear", attacker: "brute" },
            { blocker: "wolf", attacker: "brute" },
          ],
        },
      },
    ]);
    expect(two.state).toEqual(NO_BLOCK_PICK);
  });

  it("a creature in the group clicked again leaves it", () => {
    const r = step(
      NO_BLOCK_PICK,
      c,
      { kind: "own", id: "bear" },
      { kind: "attacker", id: "brute" },
      { kind: "own", id: "bear" },
    );
    expect(r.state).toEqual(NO_BLOCK_PICK);
    expect(r.sent).toEqual([]);
  });

  it("taking one of the pair out takes the other too", () => {
    const declared = ctx([
      creature("bear", { blocking_target: "brute" }),
      creature("wolf", { blocking_target: "brute" }),
      attacker("brute", { min_blockers: 2 }),
    ]);
    expect(releaseSends(declared, [{ blocker: "bear", attacker: "brute" }])).toEqual([
      { type: "remove_from_combat", params: { instance_id: "bear" } },
      { type: "remove_from_combat", params: { instance_id: "wolf" } },
    ]);
  });

  it("but not from a block with blockers to spare", () => {
    const three = ctx([
      creature("bear", { blocking_target: "brute" }),
      creature("wolf", { blocking_target: "brute" }),
      creature("cat", { blocking_target: "brute" }),
      attacker("brute", { min_blockers: 2 }),
    ]);
    expect(releaseSends(three, [{ blocker: "bear", attacker: "brute" }])).toEqual([
      { type: "remove_from_combat", params: { instance_id: "bear" } },
    ]);
  });
});

describe("blockPick: a new frame", () => {
  it("drops a selected creature that left", () => {
    const c = ctx([attacker("ogre")]);
    expect(liveBlockPick({ selected: "bear", group: null }, c)).toEqual(NO_BLOCK_PICK);
  });
  it("keeps the same object when nothing changed", () => {
    const c = ctx([creature("bear"), attacker("ogre")], { bear: ["ogre"] });
    const s: BlockPickState = { selected: "bear", group: null };
    expect(liveBlockPick(s, c)).toBe(s);
  });
  it("drops a group whose attacker is gone", () => {
    const c = ctx([creature("bear"), creature("wolf")], { bear: [], wolf: [] });
    expect(
      liveBlockPick(
        { selected: null, group: { attacker: "brute", blockers: ["bear"], need: 2 } },
        c,
      ).group,
    ).toBeNull();
  });
});

describe("blockPick: piles", () => {
  const board: BlockBoard = {
    viewerID: ME,
    selected: "bear",
    targets: new Map([
      ["tok-1", { kind: "block" }],
      ["tok-2", { kind: "refused", reason: "no" }],
    ]),
    group: { attacker: "brute", blockers: ["wolf"], need: 2 },
  };

  it("a pile of your creatures or of attackers fans out on click; others don't", () => {
    expect(fansOnClick(board, [creature("bear")])).toBe(true);
    expect(fansOnClick(board, [attacker("tok-1")])).toBe(true);
    expect(
      fansOnClick(board, [creature("treasure", { type_line: "Token Artifact — Treasure" })]),
    ).toBe(false);
    expect(fansOnClick(null, [creature("bear")])).toBe(false);
  });

  it("fanPile adds a key once", () => {
    const a = fanPile(new Set(), "k");
    expect([...a]).toEqual(["k"]);
    expect(fanPile(a, "k")).toBe(a);
  });

  it("a pile's card reads its members' click, and splits by it", () => {
    expect(blockTargetFor(board, ["x", "tok-1"])).toEqual({ kind: "block" });
    expect(blockTargetFor(null, ["tok-1"])).toBeUndefined();
    expect(blockHighlight(board, "tok-1")).not.toBe(blockHighlight(board, "tok-2"));
    expect(inBlockGroup(board, ["wolf"])).toBe(true);
    expect(blockHighlight(board, "wolf")).toBe("g");
  });
});
