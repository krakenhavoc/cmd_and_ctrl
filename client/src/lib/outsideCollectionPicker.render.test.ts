// @vitest-environment jsdom
//
// ADR 0198 — the outside_collection picker: "a card you own from outside
// the game" for a seat whose deck has no sideboard. The suggestions show
// first, clicking one picks it, the dock's take button sends its name,
// and "Take nothing" declines when the floor is zero.

import { describe, it, expect, afterEach } from "vitest";

import ChoiceDockHarness from "./test/ChoiceDockHarness.svelte";
import type { ActionType, GameView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const snapWithCollectionPrompt = (min = 0): GameView =>
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
        id: "choice-wish",
        kind: "outside_collection",
        chooser: "me",
        from_player: "me",
        reason: "Burning Wish — reveal a sorcery card you own from outside the game, or none",
        choose_min: min,
        choose_max: 1,
        choose_destination: "outside",
        collection_options: [
          { instance_id: "", name: "Wrath of God", type_line: "Sorcery", mana_cost: "{2}{W}{W}" },
          { instance_id: "", name: "Cultivate", type_line: "Sorcery", mana_cost: "{2}{G}" },
        ],
        options: [],
      },
    ],
  }) as unknown as GameView;

interface Sent {
  type: ActionType;
  params?: unknown;
}

function mount(min = 0): { container: HTMLElement; sent: Sent[] } {
  const sent: Sent[] = [];
  const view = render(
    ChoiceDockHarness as never,
    {
      snap: snapWithCollectionPrompt(min),
      viewerID: "me",
      sendAction: (type: ActionType, params?: unknown) => sent.push({ type, params }),
      lastError: null,
    } as never,
  );
  return { container: view.container, sent };
}

const rows = (container: HTMLElement): HTMLElement[] =>
  Array.from(container.querySelectorAll("button.collection-row"));

const buttonNamed = (container: HTMLElement, text: string): HTMLButtonElement | undefined =>
  Array.from(container.querySelectorAll("button")).find((b) =>
    b.textContent?.trim().startsWith(text),
  ) as HTMLButtonElement | undefined;

describe("ChoicePromptModal — a card from outside the game (ADR 0198)", () => {
  it("shows the suggestions as cards", () => {
    const { container } = mount();
    expect(rows(container).map((r) => r.querySelector(".collection-name")?.textContent)).toEqual([
      "Wrath of God",
      "Cultivate",
    ]);
  });

  it("sends the picked card's name", () => {
    const { container, sent } = mount();
    const take = buttonNamed(container, "Take…");
    expect(take?.disabled).toBe(true);
    click(rows(container)[1]);
    const ready = buttonNamed(container, "Take Cultivate");
    expect(ready?.disabled).toBe(false);
    click(ready as HTMLElement);
    expect(sent).toEqual([
      { type: "resolve_choice", params: { choice_id: "choice-wish", card_names: ["Cultivate"] } },
    ]);
  });

  it("declines with an empty list when the pick is a may", () => {
    const { container, sent } = mount();
    click(buttonNamed(container, "Take nothing") as HTMLElement);
    expect(sent).toEqual([
      { type: "resolve_choice", params: { choice_id: "choice-wish", card_names: [] } },
    ]);
  });

  it("offers no decline when a card must be taken", () => {
    const { container } = mount(1);
    expect(buttonNamed(container, "Take nothing")).toBeUndefined();
  });
});
