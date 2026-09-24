// #1508: a dragged cast always pays strictly and lets the engine tap
// its lands (owner decision 3), whatever the viewer's strictMana
// setting says; a clicked cast is byte-identical to what it was.
//
// The drag flag rides CastChoices through the whole prompt chain, and
// applyCastChoices — the one function every cast_spell send site in
// Board.svelte writes its payload with — turns it into the two wire
// flags. So these tests pin the stamp at that function, show the flag
// survives the targeting walk (the longest hop in the chain, and the
// one that stores choices in a module store rather than a local), and
// show Game.svelte's strictMana stamp leaves it alone.

import { describe, it, expect, afterEach } from "vitest";
import { get } from "svelte/store";

import {
  allPicks,
  applyCastChoices,
  begin,
  beginForModes,
  cancel,
  castChoicesBase,
  targeting,
  togglePick,
  type CastChoices,
} from "./targeting";
import { stampManaEnforcement } from "./manaEnforcement";
import type { CardView } from "./protocol";

afterEach(() => cancel());

function payload(choices: CastChoices | undefined): Record<string, unknown> {
  const params: Record<string, unknown> = { instance_id: "spell" };
  applyCastChoices(params, choices);
  return params;
}

function card(extras: Partial<CardView> = {}): CardView {
  return { instance_id: "spell", name: "Spell", owner: "me", controller: "me", ...extras };
}

describe("castChoicesBase", () => {
  it("a hand click starts with nothing", () => {
    expect(castChoicesBase()).toEqual({});
  });

  it("a drag starts with the flag, and keeps the zone", () => {
    expect(castChoicesBase(undefined, true)).toEqual({ viaDrag: true });
    expect(castChoicesBase("graveyard", true)).toEqual({ fromZone: "graveyard", viaDrag: true });
    expect(castChoicesBase("exile")).toEqual({ fromZone: "exile" });
  });
});

describe("applyCastChoices — the drag stamp", () => {
  it("a dragged cast carries strict and auto_tap", () => {
    expect(payload(castChoicesBase(undefined, true))).toEqual({
      instance_id: "spell",
      strict: true,
      auto_tap: true,
    });
  });

  it("a clicked cast is unchanged", () => {
    expect(payload(castChoicesBase())).toEqual({ instance_id: "spell" });
    expect(payload(undefined)).toEqual({ instance_id: "spell" });
    expect(payload({ xValue: 3 })).toEqual({ instance_id: "spell", x_value: 3 });
  });

  it("the stamp sits beside every other announce-time choice", () => {
    expect(
      payload({ ...castChoicesBase(undefined, true), xValue: 2, altCost: "evoke", face: 1 }),
    ).toEqual({
      instance_id: "spell",
      x_value: 2,
      alternative_cost: "evoke",
      face: 1,
      strict: true,
      auto_tap: true,
    });
  });
});

describe("the flag survives the prompt chain", () => {
  it("through a targeted spell's picker", () => {
    const bolt = card({ legal_targets: { players: ["opp"], cards: ["bear"] } });
    begin(bolt, "any", { ...castChoicesBase(undefined, true), xValue: 0 });
    let t = get(targeting)!;
    t = togglePick(t, { kind: "player", id: "opp" });
    // This is Board.svelte's fireTargets, less the send.
    const params: Record<string, unknown> = {
      instance_id: t.card.instance_id,
      targets: allPicks(t),
    };
    applyCastChoices(params, t.choices);
    expect(params).toMatchObject({
      instance_id: "spell",
      targets: [{ kind: "player", id: "opp" }],
      x_value: 0,
      strict: true,
      auto_tap: true,
    });
  });

  it("through a modal spell's targeted mode", () => {
    const command = card({
      modes: {
        min: 1,
        max: 1,
        options: [
          {
            label: "Destroy target artifact",
            target_mode: "permanent",
            legal_targets: { players: [], cards: ["rock"] },
          },
        ],
      },
    } as unknown as Partial<CardView>);
    beginForModes(command, [0], castChoicesBase(undefined, true));
    const t = get(targeting);
    expect(t?.choices?.viaDrag).toBe(true);
  });

  it("a clicked targeted cast carries no stamp", () => {
    begin(card({ legal_targets: { players: ["opp"], cards: [] } }), "player", castChoicesBase());
    const t = get(targeting)!;
    const params: Record<string, unknown> = { instance_id: "spell" };
    applyCastChoices(params, t.choices);
    expect(params).toEqual({ instance_id: "spell" });
  });
});

describe("Game.svelte's strictMana stamp", () => {
  it("leaves a dragged cast strict even with strictMana off", () => {
    const dragged = payload(castChoicesBase(undefined, true));
    expect(stampManaEnforcement("cast_spell", dragged, false)).toEqual({
      instance_id: "spell",
      strict: true,
      auto_tap: true,
    });
  });

  it("still stamps a clicked cast from the setting", () => {
    const clicked = payload(castChoicesBase());
    expect(stampManaEnforcement("cast_spell", clicked, false)).toEqual({
      instance_id: "spell",
      strict: false,
    });
    expect(stampManaEnforcement("cast_spell", clicked, true)).toEqual({
      instance_id: "spell",
      strict: true,
    });
  });
});
