// A loyalty ability that a card grants to EVERY planeswalker its controller
// controls ("Planeswalkers you control have '[−8]: …'", Kiora of Salt and
// Sand, #2797) reaches the client as an ordinary granted activated row on the
// walker: a `grant:` ref, the grantor's name and a loyalty cost. The menu
// already knows how to draw and gate such a row (the Talents, ADR 0109 §2);
// this pins that the class form needs nothing new on the client.

import { describe, expect, it } from "vitest";

import { abilityRowBlocked, abilityRowContext, buildMenuSections } from "./contextMenu.logic";
import type { CardView, GameView } from "./protocol";

const KIORA = { id: "kiora", name: "Kiora of Salt and Sand" };

const walker = (extra: Partial<CardView> = {}): CardView =>
  ({
    instance_id: "w1",
    name: "Teferi, Time Raveler",
    owner: "me",
    controller: "me",
    type_line: "Legendary Planeswalker — Teferi",
    counters: { loyalty: 8 },
    activated_abilities: [
      { index: 0, ref: "own:0", label: "+1: Fixture.", loyalty_cost: 1 },
      {
        index: 1,
        ref: "grant:kiora-of-salt-and-sand/leviathan:0:0",
        granted_by: KIORA,
        label: "−8: Create an 8/8 blue Leviathan creature token with hexproof.",
        loyalty_cost: -8,
      },
    ],
    ...extra,
  }) as CardView;

const view = (c: CardView): GameView =>
  ({
    seats: [{ id: "me", name: "Me" }],
    battlefield: { kind: "battlefield", count: 1, cards: [c] },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: 0, cards: [] },
    turn: { number: 1, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
  }) as unknown as GameView;

describe("a loyalty ability granted to every planeswalker you control", () => {
  it("names the grantor in the walker's menu and leaves the printed row alone", () => {
    const w = walker();
    const labels = buildMenuSections(view(w), w, "me", false)
      .flatMap((s) => s.items)
      .map((i) => i.label);
    expect(labels).toContain(
      "−8: Create an 8/8 blue Leviathan creature token with hexproof. (from Kiora of Salt and Sand)",
    );
    expect(labels.some((l) => l.startsWith("+1: Fixture.") && l.includes("(from"))).toBe(false);
  });

  it("gates the granted −8 on the WALKER's loyalty, like a printed minus", () => {
    const row = walker().activated_abilities![1];
    const enough = walker({ counters: { loyalty: 8 } });
    const short = walker({ counters: { loyalty: 7 } });
    expect(
      abilityRowBlocked(
        row,
        "activated",
        abilityRowContext(enough, { viewerID: "me", view: view(enough) }),
      ),
    ).toBe("");
    expect(
      abilityRowBlocked(
        row,
        "activated",
        abilityRowContext(short, { viewerID: "me", view: view(short) }),
      ),
    ).toMatch(/not enough loyalty/);
  });

  it("shares the walker's one activation a turn, whichever row used it", () => {
    const used = walker({ loyalty_activated: true });
    for (const row of used.activated_abilities!) {
      expect(
        abilityRowBlocked(
          row,
          "activated",
          abilityRowContext(used, { viewerID: "me", view: view(used) }),
        ),
      ).toBe("Already activated this turn");
    }
  });
});
