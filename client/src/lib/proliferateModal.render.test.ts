// @vitest-environment jsdom
//
// #2525 (CR 701.34a): the "proliferate" prompt in ChoicePromptModal —
// "choose any number of permanents and/or players with counters on
// them". The permanents ride the card grid, the seats on offer are chips
// beside it, both are picked into one set and sent as card_ids, and the
// engine's suggested pick is already selected when the prompt opens.

import { describe, it, expect, afterEach } from "vitest";

import ChoiceDockHarness from "./test/ChoiceDockHarness.svelte";
import type { ActionType, CardView, GameView, PendingChoiceView } from "./protocol";
import { barButton, barPrimary, nameOf } from "./test/dockView";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const creature = (id: string, name: string): CardView =>
  ({
    instance_id: id,
    name,
    controller: "me",
    owner: "me",
    type_line: "Creature — Bear",
    known_by_you: true,
  }) as CardView;

const mine = creature("c1", "Mine");
const theirs = creature("c2", "Theirs");

const snapWith = (choice: Partial<PendingChoiceView>): GameView =>
  ({
    id: "g",
    state: "active",
    seats: [
      { id: "me", name: "Me", graveyard: { kind: "graveyard", count: 0, cards: [] } },
      { id: "opp", name: "Opponent", graveyard: { kind: "graveyard", count: 0, cards: [] } },
    ],
    battlefield: { kind: "battlefield", count: 2, cards: [mine, theirs] },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: 0, cards: [] },
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    mulligans_open: false,
    pending_choices: [
      {
        id: "choice-1",
        kind: "proliferate",
        chooser: "me",
        from_player: "me",
        count: 3,
        reason: "Proliferate: choose any number of permanents and players with counters",
        choose_min: 0,
        choose_max: 3,
        options: [mine, theirs],
        choose_players: ["opp"],
        choose_suggested: ["c1", "opp"],
        ...choice,
      },
    ],
  }) as unknown as GameView;

interface Sent {
  type: ActionType;
  params?: unknown;
}

function mount(choice: Partial<PendingChoiceView> = {}): { container: HTMLElement; sent: Sent[] } {
  const sent: Sent[] = [];
  const view = render(
    ChoiceDockHarness as never,
    {
      snap: snapWith(choice),
      viewerID: "me",
      sendAction: (type: ActionType, params?: unknown) => sent.push({ type, params }),
      lastError: null,
    } as never,
  );
  return { container: view.container, sent };
}

const cardButtons = (container: HTMLElement): HTMLButtonElement[] => [
  ...container.querySelectorAll<HTMLButtonElement>("button.card-pick"),
];
const seatButtons = (container: HTMLElement): HTMLButtonElement[] => [
  ...container.querySelectorAll<HTMLButtonElement>("button.seat-pick"),
];

describe("ChoicePromptModal — proliferate (#2525)", () => {
  it("opens on the engine's suggested pick, permanents and seats alike", () => {
    const { container } = mount();
    expect(cardButtons(container)).toHaveLength(2);
    expect(seatButtons(container)).toHaveLength(1);
    expect(seatButtons(container)[0].textContent).toContain("Opponent");
    expect(cardButtons(container)[0].getAttribute("aria-pressed")).toBe("true");
    expect(cardButtons(container)[1].getAttribute("aria-pressed")).toBe("false");
    expect(seatButtons(container)[0].getAttribute("aria-pressed")).toBe("true");
    expect(container.textContent).toContain("CR 701.34");
    expect(nameOf(barPrimary(container)!)).toBe("Proliferate");
  });

  it("submits the suggested pick untouched, seats in the same card_ids list", () => {
    const { container, sent } = mount();
    click(barPrimary(container)!);
    expect(sent).toHaveLength(1);
    expect(sent[0].type).toBe("resolve_choice");
    const ids = (sent[0].params as { card_ids: string[] }).card_ids;
    expect([...ids].sort()).toEqual(["c1", "opp"]);
  });

  it("lets the player deviate: add a permanent, drop a seat", () => {
    const { container, sent } = mount();
    click(cardButtons(container)[1]);
    click(seatButtons(container)[0]);
    click(barPrimary(container)!);
    const ids = (sent[0].params as { card_ids: string[] }).card_ids;
    expect([...ids].sort()).toEqual(["c1", "c2"]);
  });

  it("lets the player choose nothing", () => {
    const { container, sent } = mount();
    click(barButton("Clear", container)!);
    const submit = barPrimary(container)!;
    expect(submit.disabled).toBe(false);
    expect(nameOf(submit)).toBe("Proliferate nothing");
    click(submit);
    expect(sent[0].params).toMatchObject({ choice_id: "choice-1", card_ids: [] });
  });

  it("shows no seat chips when no player has a counter", () => {
    const { container } = mount({
      choose_players: undefined,
      choose_suggested: ["c1"],
      choose_max: 2,
    });
    expect(seatButtons(container)).toHaveLength(0);
  });
});
