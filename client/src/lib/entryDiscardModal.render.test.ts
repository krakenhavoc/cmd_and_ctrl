// @vitest-environment jsdom
//
// ADR 0098 — Mox Diamond's "if this artifact would enter, you may
// discard a land card instead. If you do, put it onto the battlefield.
// If you don't, put it into its owner's graveyard."
//
// The decline destroys the card, so its button says so and names the
// card, which is on the stack (not the battlefield) while the prompt is
// open. Both answers leave through the ordinary {choice_id, card_ids}
// payload.

import { describe, it, expect, afterEach } from "vitest";

import ChoiceDockHarness from "./test/ChoiceDockHarness.svelte";
import type { ActionType, CardView, GameView } from "./protocol";
import { nameOf } from "./test/dockView";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const card = (id: string, name: string, typeLine: string): CardView =>
  ({ instance_id: id, name, type_line: typeLine }) as unknown as CardView;

const snapWith = (kind: string, min: number, max: number, options: CardView[]): GameView =>
  ({
    id: "g",
    state: "active",
    seats: [
      { id: "me", name: "Me" },
      { id: "them", name: "Them" },
    ],
    battlefield: { kind: "battlefield", count: 0, cards: [] },
    stack: { kind: "stack", count: 1, cards: [card("mox", "Mox Diamond", "Artifact")] },
    exile: { kind: "exile", count: 0, cards: [] },
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "precombat_main", step: "main1" },
    mulligans_open: false,
    pending_choices: [
      {
        id: "choice-1",
        kind,
        chooser: "me",
        from_player: "me",
        source: "mox",
        reason: "Mox Diamond — discard a land card so it enters?",
        choose_min: min,
        choose_max: max,
        options,
      },
    ],
  }) as unknown as GameView;

function mount(kind: string, min: number, max: number, options: CardView[]) {
  const sent: { type: ActionType; params?: unknown }[] = [];
  const view = render(
    ChoiceDockHarness as never,
    {
      snap: snapWith(kind, min, max, options),
      viewerID: "me",
      sendAction: (type: ActionType, params?: unknown) => sent.push({ type, params }),
      lastError: null,
    } as never,
  );
  return { container: view.container, sent };
}

const buttonNamed = (container: HTMLElement, text: string): HTMLButtonElement | undefined =>
  [...container.querySelectorAll("button")].find((b) => nameOf(b) === text);

describe("ChoicePromptModal — entry_discard_from_hand (ADR 0098)", () => {
  const forest = card("f", "Forest", "Basic Land — Forest");

  it("names the consequence on the decline", () => {
    const { container, sent } = mount("entry_discard_from_hand", 0, 1, [forest]);

    const decline = buttonNamed(
      container,
      "Don't discard — put Mox Diamond into its owner's graveyard",
    );
    expect(decline).toBeDefined();
    click(decline!);
    expect(sent).toHaveLength(1);
    expect(sent[0].params).toMatchObject({ choice_id: "choice-1", card_ids: [] });
  });

  it("sends the discarded land", () => {
    const { container, sent } = mount("entry_discard_from_hand", 0, 1, [forest]);

    click(container.querySelectorAll<HTMLButtonElement>(".card-pick")[0]);
    const discard = buttonNamed(container, "Discard");
    expect(discard).toBeDefined();
    click(discard!);
    expect(sent[0].params).toMatchObject({ choice_id: "choice-1", card_ids: ["f"] });
  });
});

describe("ChoicePromptModal — entry_sacrifice (ADR 0098)", () => {
  it("cannot be declined", () => {
    const { container } = mount("entry_sacrifice", 1, 1, [
      card("f", "Forest", "Basic Land — Forest"),
    ]);

    const sacrifice = buttonNamed(container, "Sacrifice");
    expect(sacrifice).toBeDefined();
    expect(sacrifice!.disabled).toBe(true);
  });
});
