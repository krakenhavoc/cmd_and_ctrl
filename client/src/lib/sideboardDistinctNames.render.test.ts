// @vitest-environment jsdom
//
// #4962 — with a sideboard seated, an outside-the-game pick "with
// different names" (Turtles Forever's outside half) is a choose_cards over
// the sideboard that carries `distinct_names`. A sideboard may hold two
// cards with one name, so once one of them is picked the other is
// disabled, rather than letting it be clicked and the server refuse the
// answer. A pick without the flag (Research) takes both copies.

import { describe, it, expect, afterEach } from "vitest";

import ChoiceDockHarness from "./test/ChoiceDockHarness.svelte";
import { nameAlreadyPicked } from "./outsideCollection";
import type { ActionType, CardView, GameView } from "./protocol";
import { barPrimary } from "./test/dockView";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const card = (id: string, name: string): CardView =>
  ({
    instance_id: id,
    name,
    type_line: "Legendary Creature — Mutant Ninja Turtle",
    known_by_you: true,
  }) as CardView;

const options = [
  card("leo-1", "Leonardo, Leader in Blue"),
  card("leo-2", "Leonardo, Leader in Blue"),
  card("raph", "Raphael, Fiery Fighter"),
];

const snap = (distinct: boolean): GameView =>
  ({
    id: "g",
    state: "active",
    seats: [
      { id: "me", name: "Me" },
      { id: "them", name: "Them" },
    ],
    battlefield: { kind: "battlefield", count: 0, cards: [] },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: 0, cards: [] },
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    mulligans_open: false,
    pending_choices: [
      {
        id: "choice-side",
        kind: "choose_cards",
        chooser: "me",
        from_player: "me",
        reason: distinct
          ? "Turtles Forever — choose legendary creature cards you own from outside the game with different names"
          : "Research — choose up to four cards you own from outside the game",
        options,
        choose_min: 0,
        choose_max: 4,
        choose_destination: "outside",
        ...(distinct ? { distinct_names: true } : {}),
      },
    ],
  }) as unknown as GameView;

interface Sent {
  type: ActionType;
  params?: unknown;
}

function mount(distinct: boolean): { container: HTMLElement; sent: Sent[] } {
  const sent: Sent[] = [];
  const view = render(
    ChoiceDockHarness as never,
    {
      snap: snap(distinct),
      viewerID: "me",
      sendAction: (type: ActionType, params?: unknown) => sent.push({ type, params }),
      lastError: null,
    } as never,
  );
  return { container: view.container, sent };
}

const cardButtons = (c: HTMLElement): HTMLButtonElement[] => [
  ...c.querySelectorAll<HTMLButtonElement>("button.card-pick"),
];

describe("a sideboard pick with different names", () => {
  it("blocks a card whose name another picked card has", () => {
    const picked = new Set(["leo-1"]);
    expect(nameAlreadyPicked(options, picked, "leo-2")).toBe(true);
    expect(nameAlreadyPicked(options, picked, "leo-1")).toBe(false);
    expect(nameAlreadyPicked(options, picked, "raph")).toBe(false);
    expect(nameAlreadyPicked(options, new Set(), "leo-2")).toBe(false);
    // Compared as the server compares them.
    const spaced = [card("a", "Leonardo"), card("b", " leonardo ")];
    expect(nameAlreadyPicked(spaced, new Set(["a"]), "b")).toBe(true);
  });

  it("disables the second copy once the first is picked, and frees it when taken back", () => {
    const { container } = mount(true);
    const [leo1, leo2, raph] = cardButtons(container);
    expect(leo2.disabled).toBe(false);

    click(leo1);
    expect(leo1.getAttribute("aria-pressed")).toBe("true");
    expect(cardButtons(container)[1].disabled).toBe(true);
    expect(raph.disabled).toBe(false);

    click(cardButtons(container)[0]);
    expect(cardButtons(container)[1].disabled).toBe(false);
  });

  it("sends each name once", () => {
    const { container, sent } = mount(true);
    click(cardButtons(container)[0]);
    click(cardButtons(container)[1]);
    click(cardButtons(container)[2]);
    expect(cardButtons(container)[1].getAttribute("aria-pressed")).toBe("false");
    click(barPrimary(container)!);
    expect(sent).toHaveLength(1);
    expect(sent[0].params).toMatchObject({ choice_id: "choice-side", card_ids: ["leo-1", "raph"] });
  });

  it("leaves a pick whose names may repeat as it was", () => {
    const { container, sent } = mount(false);
    click(cardButtons(container)[0]);
    expect(cardButtons(container)[1].disabled).toBe(false);
    click(cardButtons(container)[1]);
    click(barPrimary(container)!);
    expect(sent[0].params).toMatchObject({
      choice_id: "choice-side",
      card_ids: ["leo-1", "leo-2"],
    });
  });
});
