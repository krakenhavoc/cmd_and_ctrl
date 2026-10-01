// @vitest-environment jsdom
//
// ADR 0106 §2 (#1794, owner decision 3): a creature that can't attack
// its owner wears a CAN'T ATTACK chip naming the owner, read off the
// card view's restriction.

import { describe, it, expect, afterEach } from "vitest";

import Card from "./components/board/Card.svelte";
import { cantAttackChip } from "./cantAttack";
import type { CardView, PlayerView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const xantcha = {
  instance_id: "xantcha",
  name: "Xantcha, Sleeper Agent",
  owner: "alice",
  controller: "bob",
  known_by_you: true,
  type_line: "Legendary Creature — Phyrexian Minion",
  attack_target_restrictions: [
    { player: "alice", planeswalkers: true, source: "Xantcha, Sleeper Agent" },
  ],
} as unknown as CardView;

const seats = [
  { id: "alice", name: "Alice" },
  { id: "bob", name: "Bob" },
] as unknown as PlayerView[];

describe("the CAN'T ATTACK chip", () => {
  it("names the owner, beside TAKEN FROM", () => {
    const { container } = render(
      Card as never,
      {
        card: xantcha,
        takenFrom: "Alice",
        cantAttack: cantAttackChip(xantcha, seats),
      } as Record<string, unknown>,
    );
    const chip = container.querySelector<HTMLElement>(".badge.cant-attack");
    expect(chip?.textContent?.trim()).toBe("CAN'T ATTACK Alice");
    expect(chip?.getAttribute("aria-label")).toBe("can't attack Alice");
    expect(chip?.getAttribute("title")).toBe(
      "can't attack Alice or planeswalkers Alice controls (Xantcha, Sleeper Agent)",
    );
    expect(container.querySelector(".badge.taken")).not.toBeNull();
  });

  it("is absent for a creature with no restriction", () => {
    const { container } = render(Card as never, { card: xantcha } as Record<string, unknown>);
    expect(container.querySelector(".badge.cant-attack")).toBeNull();
  });
});
