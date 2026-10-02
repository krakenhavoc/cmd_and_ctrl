// @vitest-environment jsdom
//
// #568's option_pick arm in ChoicePromptModal.
//
// The kind is rendered by the LAST `{:else}` branch's neighbours, and
// that fallback is the shared card grid — so an arm that is missing or
// placed after it does not fail loudly, it renders a discard picker
// over an option pick's (usually empty) options and submits
// `card_ids` to a resolver that wants `option_index`. A render test is
// the only thing that catches that, which is the README's own bar for
// writing one.
//
// Two properties: one button per branch, in the card's printed order,
// and a click that submits the INDEX.
//
// ADR 0111 PR 5: a short option pick (no cards, six options at most,
// short labels) is answered inline in the action dock; one whose
// options embed cards is still the modal (a sheet in PR 6).

import { describe, it, expect, afterEach, beforeEach } from "vitest";

import ChoiceDockHarness from "./test/ChoiceDockHarness.svelte";
import { _resetForTests as resetDock } from "./dock";
import { _resetForTests as resetModals } from "./modalLayers";
import type { ActionType, CardView, GameView, PickOptionView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}
beforeEach(() => {
  const g = globalThis as Record<string, unknown>;
  g.ResizeObserver ??= FakeObserver;
  g.IntersectionObserver ??= FakeObserver;
  resetDock();
  resetModals();
});
afterEach(cleanup);

const HAILFIRE: PickOptionView[] = [
  { label: "Lose 3 life", life_cost: 3 },
  { label: "Sacrifice a nonland permanent" },
  { label: "Discard a card" },
];

const snapWithOptionPick = (pickOptions: PickOptionView[] = HAILFIRE): GameView =>
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
        kind: "option_pick",
        chooser: "me",
        from_player: "them",
        reason: "Torment of Hailfire — lose 3 life, or...",
        pick_options: pickOptions,
      },
    ],
  }) as unknown as GameView;

interface Sent {
  type: ActionType;
  params?: unknown;
}

function mountOptionPick(pickOptions?: PickOptionView[]): {
  container: HTMLElement;
  sent: Sent[];
} {
  const sent: Sent[] = [];
  const view = render(
    ChoiceDockHarness as never,
    {
      snap: snapWithOptionPick(pickOptions),
      viewerID: "me",
      sendAction: (type: ActionType, params?: unknown) => sent.push({ type, params }),
      lastError: null,
    } as never,
  );
  return { container: view.container, sent };
}

// The dock's dialog, named as the modal was: by the reason.
const dockDialog = (container: HTMLElement): HTMLElement | null =>
  container.querySelector<HTMLElement>(
    'section[aria-label="actions"] [role="dialog"][aria-label="Torment of Hailfire — lose 3 life, or..."]',
  );
const optionButtons = (container: HTMLElement): HTMLButtonElement[] => [
  ...(dockDialog(container)?.querySelectorAll<HTMLButtonElement>(".dock-row button") ?? []),
];

describe("ChoicePromptModal — option_pick (#568)", () => {
  it("renders one button per branch, in printed order", () => {
    const { container } = mountOptionPick();
    const labels = optionButtons(container).map((b) => b.textContent?.trim());
    expect(labels).toEqual(["Lose 3 life", "Sacrifice a nonland permanent", "Discard a card"]);
    // The card grid is the fallback arm — an option pick must not
    // land in it, or the answer goes out as card_ids. And it is not a
    // modal any more.
    expect(container.querySelector(".card-grid")).toBeNull();
    expect(container.querySelector(".prompt-backdrop")).toBeNull();
  });

  it("submits the chosen branch's index", () => {
    const { container, sent } = mountOptionPick();
    click(optionButtons(container)[2]);
    expect(sent).toHaveLength(1);
    expect(sent[0].type).toBe("resolve_choice");
    expect(sent[0].params).toMatchObject({ choice_id: "choice-1", option_index: 2 });
  });

  it("submits index 0 for the first branch, which is a real answer", () => {
    // Zero is the commonest answer and the field's own zero value,
    // which is why the server routes this kind by KIND rather than by
    // the key's presence. The client still has to send it.
    const { container, sent } = mountOptionPick();
    click(optionButtons(container)[0]);
    expect(sent[0].params).toMatchObject({ option_index: 0 });
  });

  it("renders an option whose cards the server redacted away", () => {
    // A pile the viewer may not see arrives as a label with no cards.
    // That is the redaction pass working, not a missing render, so the
    // button stays live.
    const { container } = mountOptionPick();
    const buttons = optionButtons(container);
    expect(buttons[1].querySelector(".pick-cards")).toBeNull();
    expect(buttons[1].disabled).toBe(false);
  });

  it("keeps an option pick whose options embed cards in the modal (a sheet, PR 6)", () => {
    // Fact or Fiction's piles: the cards are the point, and they need
    // room. Not inline.
    const card = { instance_id: "c1", name: "Island" } as CardView;
    const { container, sent } = mountOptionPick([
      { label: "Pile 1", cards: [card] },
      { label: "Pile 2", cards: [] },
    ]);
    expect(dockDialog(container)).toBeNull();
    const modalButtons = [...container.querySelectorAll<HTMLButtonElement>("button.pick-option")];
    expect(modalButtons).toHaveLength(2);
    expect(modalButtons[0].querySelector(".pick-cards")).not.toBeNull();
    click(modalButtons[1]);
    expect(sent[0].params).toMatchObject({ choice_id: "choice-1", option_index: 1 });
  });

  it("keeps an option pick with more than six options, or a long label, in the modal", () => {
    const seven = Array.from({ length: 7 }, (_, i) => ({ label: `Option ${i + 1}` }));
    const many = mountOptionPick(seven);
    expect(dockDialog(many.container)).toBeNull();
    expect(many.container.querySelectorAll("button.pick-option")).toHaveLength(7);
    cleanup();
    resetDock();

    const long = mountOptionPick([
      { label: "Fame — its caster gains control of a creature you control until end of turn" },
      { label: "Fortune — its caster draws a card and creates a Treasure token" },
    ]);
    expect(dockDialog(long.container)).toBeNull();
    expect(long.container.querySelectorAll("button.pick-option")).toHaveLength(2);
  });
});
