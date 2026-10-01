import { describe, expect, it } from "vitest";

import { cantAttackByCard, cantAttackChip } from "./cantAttack";
import type { CardView, PlayerView } from "./protocol";

// ADR 0106 §2 (#1794, owner decision 3): the CAN'T ATTACK chip is read
// off the card view's attack_target_restrictions; this only names the
// seat.

const seats = [
  { id: "a", name: "Alice" },
  { id: "b", name: "Bob" },
] as unknown as PlayerView[];

function card(id: string, rows: CardView["attack_target_restrictions"]): CardView {
  return {
    instance_id: id,
    name: id,
    owner: "a",
    controller: "b",
    attack_target_restrictions: rows,
  };
}

describe("cantAttackChip", () => {
  it("names the owner, with the planeswalker clause and the source in the tooltip", () => {
    expect(
      cantAttackChip(
        card("x", [{ player: "a", planeswalkers: true, source: "Xantcha, Sleeper Agent" }]),
        seats,
      ),
    ).toEqual({
      label: "Alice",
      title: "can't attack Alice or planeswalkers Alice controls (Xantcha, Sleeper Agent)",
    });
  });

  it("names the owner alone for the shorter restriction", () => {
    expect(cantAttackChip(card("x", [{ player: "a" }]), seats)?.title).toBe("can't attack Alice");
  });

  it("merges two restrictions naming the same player", () => {
    const chip = cantAttackChip(
      card("x", [
        { player: "a", source: "Alexios, Deimos of Kosmos" },
        { player: "a", planeswalkers: true, source: "Elrond of the White Council" },
      ]),
      seats,
    );
    expect(chip?.label).toBe("Alice");
    expect(chip?.title).toBe(
      "can't attack Alice or planeswalkers Alice controls (Alexios, Deimos of Kosmos, Elrond of the White Council)",
    );
  });

  // ADR 0107 §2 (#1879): "can't attack unless defending player controls
  // an Island" names each opponent with no Island, and says why.
  it("names every opponent with no Island, and why", () => {
    const chip = cantAttackChip(
      card("x", [
        { player: "a", planeswalkers: true, source: "Sea Serpent", unless: "Island" },
        { player: "b", planeswalkers: true, source: "Sea Serpent", unless: "Island" },
      ]),
      seats,
    );
    expect(chip).toEqual({
      label: "Alice, Bob",
      title:
        "can't attack Alice or planeswalkers Alice controls: Alice controls no Island (Sea Serpent); " +
        "can't attack Bob or planeswalkers Bob controls: Bob controls no Island (Sea Serpent)",
    });
  });

  it("falls back to a neutral name for a seat the frame does not list", () => {
    expect(cantAttackChip(card("x", [{ player: "gone" }]), seats)?.label).toBe("another player");
  });

  it("is null for a card with no restriction", () => {
    expect(cantAttackChip(card("x", undefined), seats)).toBeNull();
    expect(cantAttackChip(card("x", []), seats)).toBeNull();
  });
});

describe("cantAttackByCard", () => {
  it("keys only the cards that have a chip", () => {
    const out = cantAttackByCard([card("x", [{ player: "a" }]), card("y", undefined)], seats);
    expect(Object.keys(out)).toEqual(["x"]);
  });

  it("tolerates a frame with no battlefield", () => {
    expect(cantAttackByCard(undefined, undefined)).toEqual({});
  });
});
