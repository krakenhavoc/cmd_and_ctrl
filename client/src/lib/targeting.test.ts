import { describe, it, expect } from "vitest";
import { get } from "svelte/store";

import {
  advance,
  allPicks,
  begin,
  beginChoice,
  beginForAbility,
  beginForModes,
  cancel,
  canConfirm,
  confirm,
  hasXCost,
  castLocksXAtZero,
  isModal,
  isMultiPick,
  isPicked,
  modeOptionCastable,
  setConfirmHandler,
  togglePick,
  isLegalCardTarget,
  isLegalPlayerTarget,
  legalTargetCount,
  targeting,
} from "./targeting";
import type { ActivatedAbilityView, CardView, PendingChoiceView } from "./protocol";

function card(extras: Partial<CardView> = {}): CardView {
  return { instance_id: "spell", name: "Spell", owner: "p0", controller: "p0", ...extras };
}

describe("targeting store — S20 legal sets", () => {
  it("uses the server legal set when the card carries one", () => {
    begin(card({ legal_targets: { players: ["p1"], cards: ["bear"] } }), "any");
    const t = get(targeting)!;
    expect(isLegalCardTarget(t, "bear")).toBe(true);
    expect(isLegalCardTarget(t, "rock")).toBe(false);
    expect(isLegalPlayerTarget(t, "p1")).toBe(true);
    expect(isLegalPlayerTarget(t, "p0")).toBe(false);
    expect(legalTargetCount(t)).toBe(2);
    cancel();
    expect(get(targeting)).toBeNull();
  });

  it("falls back to mode heuristics for free-form cards", () => {
    begin(card(), "creature");
    const t = get(targeting)!;
    expect(t.legal).toBeUndefined();
    expect(isLegalCardTarget(t, "anything")).toBe(true);
    expect(isLegalPlayerTarget(t, "p1")).toBe(false);
    expect(legalTargetCount(t)).toBe(-1);
    begin(card(), "player");
    expect(isLegalPlayerTarget(get(targeting)!, "p1")).toBe(true);
    cancel();
  });
});

describe("hasXCost + xValue on the prompt — S20 sub-PR 3", () => {
  it("detects {X} in the printed cost", () => {
    expect(hasXCost(card({ mana_cost: "{X}{R}" }))).toBe(true);
    expect(hasXCost(card({ mana_cost: "{2}{U}" }))).toBe(false);
    expect(hasXCost(card({}))).toBe(false);
  });

  it("carries the announced X through the targeting prompt", () => {
    begin(card({ mana_cost: "{X}{R}", legal_targets: { players: ["p1"] } }), "any", { xValue: 4 });
    expect(get(targeting)?.choices?.xValue).toBe(4);
    cancel();
  });

  // S23: Toxic Deluge prints a flat {2}{B} and still has an X to
  // announce, because its "pay X life" additional cost carries one.
  // A card whose mana cost alone decided this would never open the
  // prompt and would always cast for X = 0.
  it("detects an X on a pay-X-life additional cost", () => {
    expect(
      hasXCost(
        card({ mana_cost: "{2}{B}", additional_cost: { demands_x: true, label: "Pay X life" } }),
      ),
    ).toBe(true);
    expect(hasXCost(card({ mana_cost: "{2}{B}", additional_cost: { discard_cards: 1 } }))).toBe(
      false,
    );
  });
});

// CR 107.3b (#831): a free cast of an {X} spell has exactly one legal
// X and it is 0, so the picker must not open. The client never
// re-derives the rule from the cost strings — the server ships the
// answer on the offer being taken.
describe("castLocksXAtZero — CR 107.3b", () => {
  const stroke = (extras: Partial<CardView> = {}) => card({ mana_cost: "{X}{U}", ...extras });

  it("is false for an ordinary cast, which still announces X", () => {
    expect(castLocksXAtZero(stroke(), undefined)).toBe(false);
    expect(hasXCost(stroke())).toBe(true);
  });

  it("locks X under a free-cast exile grant (cascade, a Siege)", () => {
    const hit = stroke({
      exile_play: { player: "p0", cast_only: true, cost_override: "{0}", x_locked_at_zero: true },
    });
    expect(castLocksXAtZero(hit, undefined)).toBe(true);
  });

  it("leaves an unpriced impulse grant alone — it pays the printed cost", () => {
    expect(castLocksXAtZero(stroke({ exile_play: { player: "p0" } }), undefined)).toBe(false);
  });

  it("locks X for an alternative cost without X, and only when claimed", () => {
    const c = stroke({
      alternative_costs: [
        { key: "free", label: "Cast without paying its mana cost", x_locked_at_zero: true },
        { key: "kicked", label: "Pay {X}{R}", mana_cost: "{X}{R}" },
      ],
    });
    expect(castLocksXAtZero(c, "free")).toBe(true);
    expect(castLocksXAtZero(c, "kicked")).toBe(false);
    // Nothing claimed: the printed cost is being paid.
    expect(castLocksXAtZero(c, undefined)).toBe(false);
    // A key the card doesn't offer is not an offer.
    expect(castLocksXAtZero(c, "overload")).toBe(false);
  });
});

// --- S20 sub-PR 4: modal spells ----------------------------------

describe("modal targeting", () => {
  const charm = {
    instance_id: "c-charm",
    name: "Rakdos Charm",
    modes: {
      prompt: "Choose one",
      min: 1,
      max: 1,
      options: [
        {
          label: "Exile target player's graveyard.",
          target_mode: "player",
          legal_targets: { players: ["p1"] },
        },
        {
          label: "Destroy target artifact.",
          target_mode: "permanent",
          legal_targets: { cards: [] },
        },
        { label: "Each creature deals 1 damage to its controller." },
      ],
    },
  } as unknown as CardView;

  it("isModal / modeOptionCastable", () => {
    expect(isModal(charm)).toBe(true);
    expect(isModal({ instance_id: "x", name: "Bolt" } as unknown as CardView)).toBe(false);
    expect(modeOptionCastable(charm.modes!.options[0])).toBe(true);
    expect(modeOptionCastable(charm.modes!.options[1])).toBe(false);
    expect(modeOptionCastable(charm.modes!.options[2])).toBe(true);
  });

  it("beginForModes takes the option's legal set and label, and carries the modes", () => {
    expect(beginForModes(charm, [0], { xValue: 3 })).toBe(true);
    const t = get(targeting)!;
    expect(t.mode).toBe("player");
    expect(t.label).toBe("Exile target player's graveyard.");
    expect(t.modes).toEqual([0]);
    expect(t.choices?.xValue).toBe(3);
    expect(isLegalPlayerTarget(t, "p1")).toBe(true);
    expect(isLegalPlayerTarget(t, "p0")).toBe(false);
    expect(isLegalCardTarget(t, "c-anything")).toBe(false);
    cancel();
  });

  it("an all-untargeted selection asks nothing", () => {
    expect(beginForModes(charm, [2])).toBe(false);
    expect(get(targeting)).toBe(null);
  });

  // #1172: `modes` is PUBLIC on a public pile — it is the card's
  // printed text — and since #1172 the legal sets inside it are not.
  // A bystander's frame therefore carries the bullets with no
  // `legal_targets` and no `clauses` at all, and this pins that the
  // picker reads the frame it was handed rather than assuming a
  // targeted bullet always arrives with a set: it opens nothing,
  // exactly as it does for a genuinely untargeted bullet, instead of
  // falling back to a free-form prompt over the whole board.
  //
  // Nothing about the reader had to change — frames are per viewer,
  // and the only frame that offers this cast is the one the server
  // promoted the sets into. The test is the statement that this stays
  // true.
  it("a mode block stripped of its nested legal sets opens no picker (#1172)", () => {
    const bystander = {
      instance_id: "c-charm",
      name: "Fixture Charm",
      modes: {
        prompt: "Choose one",
        min: 1,
        max: 1,
        options: [
          // The same two bullets the owner's frame carries, with the
          // per-viewer half gone: label and target_mode are printed
          // text and stay.
          { label: "Exile target player's graveyard.", target_mode: "player" },
          { label: "Destroy target artifact.", target_mode: "permanent" },
        ],
      },
    } as unknown as CardView;
    expect(isModal(bystander)).toBe(true);
    expect(beginForModes(bystander, [0])).toBe(false);
    expect(get(targeting)).toBe(null);
  });
});

// --- #764: per-mode targets, repeated modes, multi-clause walks ----

describe("the two-step picker (#764)", () => {
  // Kolaghan's Command's shape: two bullets that each target.
  const command = {
    instance_id: "c-cmd",
    name: "Kolaghan's Command",
    modes: {
      prompt: "Choose two",
      min: 2,
      max: 2,
      options: [
        {
          label: "Destroy target artifact.",
          target_mode: "permanent",
          legal_targets: { cards: ["rock"], min: 1, max: 1 },
        },
        {
          label: "Deals 2 damage to any target.",
          target_mode: "any",
          legal_targets: { cards: ["bear"], players: ["p1"], min: 1, max: 1 },
        },
      ],
    },
  } as unknown as CardView;

  it("walks one clause per chosen mode, stamping each pick with its occurrence", () => {
    expect(beginForModes(command, [0, 1])).toBe(true);
    let t = get(targeting)!;
    expect(t.steps.length).toBe(2);
    expect(t.step).toBe(0);
    expect(t.label).toBe("Destroy target artifact.");
    expect(isLegalCardTarget(t, "rock")).toBe(true);
    expect(isLegalCardTarget(t, "bear")).toBe(false);

    // Answer the first clause and step on.
    t = { ...t, picked: [{ kind: "card", id: "rock" }] };
    const second = advance(t)!;
    expect(second.step).toBe(1);
    expect(second.label).toBe("Deals 2 damage to any target.");
    expect(second.done).toEqual([{ kind: "card", id: "rock", mode: 0, slot: 0 }]);
    expect(isLegalCardTarget(second, "bear")).toBe(true);

    // Answer the second and finish.
    const last = { ...second, picked: [{ kind: "card" as const, id: "bear" }] };
    expect(advance(last)).toBe(null);
    expect(allPicks(last)).toEqual([
      { kind: "card", id: "rock", mode: 0, slot: 0 },
      { kind: "card", id: "bear", mode: 1, slot: 0 },
    ]);
    cancel();
  });

  it("a repeated mode (CR 700.2d) gets one step per occurrence", () => {
    const confluence = {
      instance_id: "c-conf",
      name: "Mystic Confluence",
      modes: {
        prompt: "Choose three",
        min: 3,
        max: 3,
        repeatable: true,
        options: [
          {
            label: "Return target creature to its owner's hand.",
            target_mode: "creature",
            legal_targets: { cards: ["a", "b"], min: 1, max: 1 },
          },
          { label: "Draw a card." },
        ],
      },
    } as unknown as CardView;
    expect(beginForModes(confluence, [0, 0, 1])).toBe(true);
    const t = get(targeting)!;
    expect(t.steps.length).toBe(2);
    expect(t.steps.map((s) => s.modeIndex)).toEqual([0, 1]);
    cancel();
  });

  it("a two-clause card walks its clauses in printed order", () => {
    const bite = {
      instance_id: "c-bite",
      name: "Bite Down",
      target_mode: "permanent",
      legal_targets: { cards: ["mine"], min: 1, max: 1 },
      clauses: [
        { cards: ["mine"], min: 1, max: 1, label: "target creature you control" },
        {
          cards: ["theirs"],
          min: 1,
          max: 1,
          label: "target creature or planeswalker you don't control",
          distinct: true,
        },
      ],
    } as unknown as CardView;
    begin(bite, "permanent");
    const t = get(targeting)!;
    expect(t.steps.length).toBe(2);
    expect(t.label).toBe("target creature you control");
    expect(isLegalCardTarget(t, "mine")).toBe(true);
    expect(isLegalCardTarget(t, "theirs")).toBe(false);
    const second = advance({ ...t, picked: [{ kind: "card", id: "mine" }] })!;
    expect(second.label).toBe("target creature or planeswalker you don't control");
    expect(isLegalCardTarget(second, "theirs")).toBe(true);
    cancel();
  });

  it("a distinct clause drops what an earlier clause already took", () => {
    const move = {
      instance_id: "c-move",
      name: "Resourceful Defense",
      target_mode: "permanent",
      legal_targets: { cards: ["a", "b"], min: 1, max: 1 },
      clauses: [
        { cards: ["a", "b"], min: 1, max: 1, label: "target permanent you control" },
        {
          cards: ["a", "b"],
          min: 1,
          max: 1,
          label: "a second target permanent you control",
          distinct: true,
        },
      ],
    } as unknown as CardView;
    begin(move, "permanent");
    const t = get(targeting)!;
    const second = advance({ ...t, picked: [{ kind: "card", id: "a" }] })!;
    expect(isLegalCardTarget(second, "a")).toBe(false);
    expect(isLegalCardTarget(second, "b")).toBe(true);
    cancel();
  });
});

// --- S21 sub-PR 2: activated abilities ----------------------------

describe("activated-ability targeting", () => {
  const bombardment = {
    instance_id: "c-bomb",
    name: "Goblin Bombardment",
  } as unknown as CardView;
  const ability = {
    index: 0,
    label: "Sacrifice a creature: deal 1 damage to any target",
    sacrifice_label: "a creature",
    sacrifice_options: { cards: ["c-fodder"] },
    target_mode: "any",
    legal_targets: { players: ["p1"], cards: ["c-bear"] },
  } as unknown as ActivatedAbilityView;

  it("carries the ability index and the paid sacrifice through the prompt", () => {
    beginForAbility(bombardment, ability, ["c-fodder"]);
    const t = get(targeting)!;
    expect(t.mode).toBe("any");
    // crewIDs defaults to empty: this ability has no crew component,
    // and the server treats absent and empty the same.
    expect(t.ability).toEqual({ index: 0, sacrificeIDs: ["c-fodder"], crewIDs: [] });
    // The legal set comes from the ability, not the source card.
    expect(isLegalPlayerTarget(t, "p1")).toBe(true);
    expect(isLegalCardTarget(t, "c-bear")).toBe(true);
    expect(isLegalCardTarget(t, "c-fodder")).toBe(false);
    cancel();
    expect(get(targeting)).toBeNull();
  });

  it("carries a counter-cost payment through the prompt (#625)", () => {
    // Benevolent Hydra-shaped: the counter is chosen before the target,
    // and has to reach the one activate_ability the confirm sends.
    beginForAbility(bombardment, ability, [], [], undefined, {
      counter_source_ids: ["c-walker"],
      counter_kind: "stun",
    });
    expect(get(targeting)!.ability?.counter).toEqual({
      counter_source_ids: ["c-walker"],
      counter_kind: "stun",
    });
    cancel();
  });
});

// --- S20 sub-PR 5: multi-target ------------------------------------

describe("multi-pick targeting", () => {
  const pair = {
    instance_id: "c-ashes",
    name: "Ashes to Ashes",
    legal_targets: { cards: ["c-a", "c-b", "c-c"], min: 2, max: 2 },
  } as unknown as CardView;

  it("single-target prompts are not multi-pick", () => {
    begin(
      {
        instance_id: "c-bolt",
        name: "Bolt",
        legal_targets: { players: ["p1"] },
      } as unknown as CardView,
      "any",
    );
    expect(isMultiPick(get(targeting)!)).toBe(false);
    cancel();
  });

  it("toggles picks in click order, caps at max, confirms at min", () => {
    begin(pair, "creature");
    let t = get(targeting)!;
    expect(isMultiPick(t)).toBe(true);
    expect(canConfirm(t)).toBe(false);
    t = togglePick(t, { kind: "card", id: "c-b" });
    t = togglePick(t, { kind: "card", id: "c-a" });
    expect(t.picked.map((p) => p.id)).toEqual(["c-b", "c-a"]);
    expect(canConfirm(t)).toBe(true);
    // Third pick refused at max 2.
    t = togglePick(t, { kind: "card", id: "c-c" });
    expect(t.picked.length).toBe(2);
    // Toggling a picked one removes it.
    t = togglePick(t, { kind: "card", id: "c-b" });
    expect(t.picked.map((p) => p.id)).toEqual(["c-a"]);
    expect(isPicked(t, "c-a")).toBe(true);
    expect(isPicked(t, "c-b")).toBe(false);
    expect(canConfirm(t)).toBe(false);
    cancel();
  });

  it("up to N confirms with nothing picked; unbounded max never caps", () => {
    begin(
      { ...pair, legal_targets: { cards: ["c-a"], min: 0, max: 2 } } as unknown as CardView,
      "permanent",
    );
    expect(canConfirm(get(targeting)!)).toBe(true);
    cancel();
    begin(
      {
        ...pair,
        legal_targets: { cards: ["c-a", "c-b", "c-c"], min: 1, max: 0 },
      } as unknown as CardView,
      "creature",
    );
    let t = get(targeting)!;
    for (const id of ["c-a", "c-b", "c-c"]) t = togglePick(t, { kind: "card", id });
    expect(t.picked.length).toBe(3);
    cancel();
  });

  it("confirm() routes to the registered handler", () => {
    let fired = 0;
    setConfirmHandler(() => fired++);
    confirm();
    expect(fired).toBe(1);
    setConfirmHandler(null);
    confirm();
    expect(fired).toBe(1);
  });
});

// #1196, CR 115.7 — the retarget prompt. It reuses the pick_target
// payload and the board picker, so what needs pinning on the client
// is the two things that differ: the prompt knows which KIND it is
// (the banner says "click a new target", not "triggered"), and an
// OPTIONAL one (min 0) is confirmable with nothing picked, which is
// how "you may leave the target unchanged" is answered.
describe("retarget prompt — #1196", () => {
  it("carries the choice kind so the banner can say what is being asked", () => {
    beginChoice(
      {
        id: "choice-retarget",
        kind: "retarget",
        chooser: "p0",
        reason: "Deflecting Swat — choose a new target",
        pick_target: { players: ["p1", "p2"], cards: [], min: 0, max: 1 },
      } as unknown as PendingChoiceView,
      card({ instance_id: "swat", name: "Deflecting Swat" }),
    );
    const t = get(targeting)!;
    expect(t.choiceID).toBe("choice-retarget");
    expect(t.choiceKind).toBe("retarget");
    expect(isLegalPlayerTarget(t, "p2")).toBe(true);
    expect(isLegalPlayerTarget(t, "p0")).toBe(false);
    cancel();
    // A prompt is not cancellable — the server is waiting.
    expect(get(targeting)).not.toBeNull();
    targeting.set(null);
  });

  it("a you-may retarget is confirmable with nothing picked", () => {
    beginChoice(
      {
        id: "choice-may",
        kind: "retarget",
        chooser: "p0",
        reason: "Deflecting Swat — choose a new target",
        pick_target: { players: ["p1"], cards: [], min: 0, max: 1 },
      } as unknown as PendingChoiceView,
      card({ instance_id: "swat", name: "Deflecting Swat" }),
    );
    const t = get(targeting)!;
    expect(isMultiPick(t)).toBe(true);
    expect(canConfirm(t)).toBe(true);
    expect(allPicks(t)).toEqual([]);
    targeting.set(null);
  });

  it("a mandatory retarget needs a pick before it can be confirmed", () => {
    beginChoice(
      {
        id: "choice-must",
        kind: "retarget",
        chooser: "p0",
        reason: "Bolt Bend — change the target",
        pick_target: { players: ["p1"], cards: [], min: 1, max: 1 },
      } as unknown as PendingChoiceView,
      card({ instance_id: "bend", name: "Bolt Bend" }),
    );
    const t = get(targeting)!;
    expect(canConfirm(t)).toBe(false);
    expect(canConfirm(togglePick(t, { kind: "player", id: "p1" }))).toBe(true);
    targeting.set(null);
  });
});
