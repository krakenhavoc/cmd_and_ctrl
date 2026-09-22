// @vitest-environment jsdom
//
// #1214's three resolution-time picks in ChoicePromptModal.
//
// All three carry the same {choice_id, card_ids} payload and the same
// choose_min / choose_max bounds as choose_cards, so they render
// through the shared card grid — and the bug a render test catches is
// exactly that they might NOT: a kind that misses `isCardSetPick`
// falls into the last `{:else}` arm, which reads `count` instead of
// the bounds, so a "choose any number" prompt renders as
// "pick 0 cards" with a Confirm button that can never enable.
//
// Three properties per kind: the grid renders the candidates, the
// bounds come from choose_min / choose_max rather than from `count`,
// and a submitted pick goes out as card_ids.

import { describe, it, expect, afterEach } from "vitest";

import ChoicePromptModal from "./components/board/ChoicePromptModal.svelte";
import type { ActionType, GameView, PendingChoiceView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const card = (id: string, name: string) => ({
  instance_id: id,
  name,
  type_line: "Creature — Bear",
  known_by_you: true,
});

const snapWith = (choice: Partial<PendingChoiceView>): GameView =>
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
        id: "choice-1",
        chooser: "me",
        from_player: "them",
        options: [card("c1", "Alpha"), card("c2", "Beta"), card("c3", "Gamma")],
        ...choice,
      },
    ],
  }) as unknown as GameView;

interface Sent {
  type: ActionType;
  params?: unknown;
}

function mount(choice: Partial<PendingChoiceView>): { container: HTMLElement; sent: Sent[] } {
  const sent: Sent[] = [];
  const view = render(
    ChoicePromptModal as never,
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

const submitButton = (container: HTMLElement): HTMLButtonElement =>
  container.querySelector<HTMLButtonElement>("button.primary")!;

describe("ChoicePromptModal — the three resolution-time picks (#1214)", () => {
  for (const kind of ["reveal_pick", "their_permanents", "own_permanents"] as const) {
    it(`${kind} renders the candidate grid and submits card_ids`, () => {
      const { container, sent } = mount({
        kind,
        reason: "test — choose one",
        choose_min: 1,
        choose_max: 1,
      });
      expect(cardButtons(container)).toHaveLength(3);
      // Nothing picked yet, so the primary button is inert: the floor
      // is one.
      expect(submitButton(container).disabled).toBe(true);

      click(cardButtons(container)[1]);
      expect(submitButton(container).disabled).toBe(false);
      click(submitButton(container));

      expect(sent).toHaveLength(1);
      expect(sent[0].type).toBe("resolve_choice");
      expect(sent[0].params).toMatchObject({ choice_id: "choice-1", card_ids: ["c2"] });
    });

    it(`${kind} reads its ceiling from choose_max, not from count`, () => {
      // `count` is absent on these prompts. A kind that fell through
      // to the fallback arm would read it as 0 and disable every card
      // in the grid.
      const { container } = mount({
        kind,
        reason: "test — choose two",
        choose_min: 0,
        choose_max: 2,
      });
      const cards = cardButtons(container);
      expect(cards.some((b) => b.disabled)).toBe(false);
      click(cards[0]);
      click(cards[1]);
      // At the ceiling, the unpicked card is locked out.
      expect(cards[2].disabled).toBe(true);
      // A floor of zero is a real answer: the empty submit is live.
      expect(submitButton(container).disabled).toBe(false);
    });
  }

  it("a zero-floor pick can be cleared back to nothing", () => {
    // "Sacrifice ANY NUMBER of lands" — Clear is offered, and the
    // empty answer stays submittable.
    const { container, sent } = mount({
      kind: "own_permanents",
      reason: "test — sacrifice any number",
      choose_min: 0,
      choose_max: 3,
    });
    click(cardButtons(container)[0]);
    const clear = [...container.querySelectorAll<HTMLButtonElement>(".prompt-foot button")].find(
      (b) => b.textContent?.trim() === "Clear",
    );
    expect(clear).toBeDefined();
    click(clear!);
    click(submitButton(container));
    expect(sent[0].params).toMatchObject({ choice_id: "choice-1", card_ids: [] });
  });

  it("a their_permanents prompt names whose board is on offer", () => {
    const { container } = mount({
      kind: "their_permanents",
      reason: "test — choose one of theirs",
      choose_min: 1,
      choose_max: 1,
    });
    expect(container.querySelector(".prompt-hint")?.textContent).toContain("Them");
  });
});
