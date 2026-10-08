// @vitest-environment jsdom
//
// ADR 0071 decision 4, the client half: the designation badge.
//
// A Class's level (CR 716.2) and a Case's solved marker (CR 719.3)
// say WHICH of the lines printed on the card are live right now, and
// that is the one thing a player cannot read off the art. Everything
// else the gate does is invisible by design — an ability that is off
// is simply absent from the card's `activated_abilities`, and a
// static or a trigger has no per-ability representation on the wire —
// so these two pips are the whole of what the client renders for it.
//
// #1705 adds the two later ADR 0071 amendments to the same slot:
// harnessed (CR 701.64, its "∞ — [ability]" lines) and monstrous
// (CR 701.37b, its "as long as this creature is monstrous" lines).
// Both had wire fields (`CardView.harnessed` / `CardView.monstrous`)
// long before either had a badge — this file is what closes that gap.

import { describe, it, expect, afterEach } from "vitest";

import Card from "./components/board/Card.svelte";
import type { CardView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

afterEach(cleanup);

function mount(card: CardView) {
  return render(Card as never, { card } as Record<string, unknown>);
}

const permanent = (extra: Record<string, unknown>): CardView =>
  ({
    instance_id: "perm",
    name: "Wizard Class",
    owner: "me",
    controller: "me",
    known_by_you: true,
    type_line: "Enchantment — Class",
    ...extra,
  }) as unknown as CardView;

function badgeText(container: HTMLElement): string {
  return container.querySelector(".badge.designation")?.textContent?.trim() ?? "";
}

describe("the designation badge", () => {
  it("is absent on an ordinary permanent", () => {
    const { container } = mount(permanent({ name: "Sol Ring", type_line: "Artifact" }));
    expect(container.querySelector(".badge.designation")).toBeNull();
  });

  it("shows a Class's level", () => {
    const { container } = mount(permanent({ class_level: 1 }));
    expect(badgeText(container)).toBe("LVL 1");
  });

  it("follows the level up", () => {
    const { container } = mount(permanent({ class_level: 3 }));
    expect(badgeText(container)).toBe("LVL 3");
  });

  it("shows a solved Case", () => {
    const { container } = mount(
      permanent({
        name: "Case of the Shattered Pact",
        type_line: "Enchantment — Case",
        solved: true,
      }),
    );
    expect(badgeText(container)).toBe("SOLVED");
  });

  it("is absent on an unsolved Case", () => {
    const { container } = mount(
      permanent({ name: "Case of the Shattered Pact", type_line: "Enchantment — Case" }),
    );
    expect(container.querySelector(".badge.designation")).toBeNull();
  });

  // ADR 0090 (CR 722.3a): a prepared preparation creature.
  it("shows a prepared permanent", () => {
    const { container } = mount(
      permanent({
        name: "Skycoach Conductor",
        type_line: "Creature — Bird Pilot",
        prepared: true,
      }),
    );
    expect(badgeText(container)).toBe("PREPARED");
  });

  it("is absent on an unprepared preparation creature", () => {
    const { container } = mount(
      permanent({ name: "Skycoach Conductor", type_line: "Creature — Bird Pilot" }),
    );
    expect(container.querySelector(".badge.designation")).toBeNull();
  });

  // ADR 0071 amendment, #1321 (CR 701.64): a harnessed permanent.
  it("shows a harnessed permanent", () => {
    const { container } = mount(
      permanent({
        name: "The Mind Stone",
        type_line: "Legendary Artifact",
        harnessed: true,
      }),
    );
    expect(badgeText(container)).toBe("HARNESSED");
  });

  it("is absent on an unharnessed permanent", () => {
    const { container } = mount(
      permanent({ name: "The Mind Stone", type_line: "Legendary Artifact" }),
    );
    expect(container.querySelector(".badge.designation")).toBeNull();
  });

  // ADR 0071 amendment, #1700 (CR 701.37b): a monstrous creature.
  it("shows a monstrous creature", () => {
    const { container } = mount(
      permanent({
        name: "Stormbreath Dragon",
        type_line: "Creature — Dragon",
        monstrous: true,
      }),
    );
    expect(badgeText(container)).toBe("MONSTROUS");
  });

  // #2695 (CR 702.171): a saddled Mount says its "attacks while saddled"
  // abilities are live, which is nowhere else on the card.
  it("shows a saddled Mount", () => {
    const { container } = mount(
      permanent({ name: "Gilded Ghoda", type_line: "Creature — Horse Mount", saddled: true }),
    );
    expect(badgeText(container)).toBe("SADDLED");
  });

  it("is absent on a Mount that isn't saddled", () => {
    const { container } = mount(
      permanent({ name: "Gilded Ghoda", type_line: "Creature — Horse Mount" }),
    );
    expect(container.querySelector(".badge.designation")).toBeNull();
  });

  it("is absent on a creature that hasn't become monstrous", () => {
    const { container } = mount(
      permanent({ name: "Stormbreath Dragon", type_line: "Creature — Dragon" }),
    );
    expect(container.querySelector(".badge.designation")).toBeNull();
  });

  // The server clears both fields in the non-knower redaction — a
  // level says "Class" and a solved flag says "Case" as loudly as
  // loyalty says "planeswalker" — so a hidden card has nothing to
  // render. This pins the client half of that: it renders what it is
  // given and invents nothing for a card it cannot read.
  it("renders nothing for a card the viewer does not know", () => {
    const { container } = mount(
      permanent({ known_by_you: false, name: "", type_line: "", class_level: undefined }),
    );
    expect(container.querySelector(".badge.designation")).toBeNull();
  });
});
