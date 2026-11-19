// banding.test.ts — ADR 0224 (#5723): the pure half of forming a band.

import { describe, expect, it } from "vitest";

import { attackAllParams, planAttackAll } from "./attackAll";
import { bandProblem, bandTitle, hasBanding, withBand } from "./banding";
import { legalActionsOf } from "./legalActions";
import type { CardView, GameView, PlayerView, ZoneView } from "./protocol";

const card = (name: string, abilities: string[] = []) => ({ name, abilities });

describe("banding", () => {
  it("reads banding off the effective ability list", () => {
    expect(hasBanding(card("Benalish Hero", ["banding"]))).toBe(true);
    expect(hasBanding(card("Mesa Pegasus", ["flying", "banding"]))).toBe(true);
    expect(hasBanding(card("Grizzly Bears"))).toBe(false);
    expect(hasBanding(null)).toBe(false);
  });

  it("allows any number with banding and at most one without (CR 702.22c)", () => {
    const hero = card("Benalish Hero", ["banding"]);
    const wolves = card("Timber Wolves", ["banding"]);
    expect(bandProblem([hero])).toBeNull();
    expect(bandProblem([hero, card("Bear")])).toBeNull();
    expect(bandProblem([hero, wolves, card("Bear")])).toBeNull();
    expect(bandProblem([hero, card("Bear"), card("Giant")])).toBe(
      "A band can have only one creature without banding, and Bear and Giant both lack it.",
    );
    expect(bandProblem([card("Bear"), card("Giant")])).not.toBeNull();
  });

  it("bands two or more named entries and leaves one alone", () => {
    const entries = [
      { attacker: "a", target: "bob" },
      { attacker: "b", target: "bob" },
      { attacker: "c", target: "bob" },
    ];
    expect(withBand(entries, ["a", "c"])).toEqual([
      { attacker: "a", target: "bob", band: 1 },
      { attacker: "b", target: "bob" },
      { attacker: "c", target: "bob", band: 1 },
    ]);
    expect(withBand(entries, ["a"])).toEqual(entries);
    expect(bandTitle(2)).toContain("band 2");
  });
});

function zone(kind: string, owner: string | undefined, cards: CardView[]): ZoneView {
  return { kind, owner, count: cards.length, cards };
}

function seat(id: string, name: string, n: number): PlayerView {
  return {
    id,
    name,
    seat: n,
    life: 40,
    library: zone("library", id, []),
    hand: zone("hand", id, []),
    graveyard: zone("graveyard", id, []),
    command: zone("command", id, []),
    commander_damage: {},
    life_history: [],
  };
}

describe("attackAllParams with a band", () => {
  it("numbers the picker's band on the wire", () => {
    const cards: CardView[] = ["hero", "bear"].map((id) => ({
      instance_id: id,
      name: id,
      owner: "me",
      controller: "me",
      type_line: "Creature — Test",
      abilities: id === "hero" ? ["banding"] : [],
    }));
    const v = {
      id: "g1",
      state: "active",
      seats: [seat("me", "Me", 0), seat("bob", "Bob", 1)],
      battlefield: zone("battlefield", undefined, cards),
      stack: zone("stack", undefined, []),
      exile: zone("exile", undefined, []),
      turn: {
        seq: 1,
        number: 3,
        active_seat: 0,
        priority_holder: 0,
        phase: "combat",
        step: "declare_attackers",
      },
      mulligans_open: false,
      legal_actions: {
        pass: true,
        sources: {
          hero: { kinds: ["attack"], moves: 1, attack_targets: ["bob"] },
          bear: { kinds: ["attack"], moves: 1, attack_targets: ["bob"] },
        },
      },
    } as GameView;
    const plan = planAttackAll(v, "me", legalActionsOf(v));
    expect(attackAllParams(plan, "bob", { band: ["hero", "bear"] })?.attackers).toEqual([
      { attacker: "hero", target: "bob", band: 1 },
      { attacker: "bear", target: "bob", band: 1 },
    ]);
    expect(attackAllParams(plan, "bob")?.attackers).toEqual([
      { attacker: "hero", target: "bob" },
      { attacker: "bear", target: "bob" },
    ]);
  });
});
