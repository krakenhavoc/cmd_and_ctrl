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

import { describe, it, expect, afterEach } from "vitest";

import ChoicePromptModal from "./components/board/ChoicePromptModal.svelte";
import type { ActionType, GameView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const snapWithOptionPick = (): GameView =>
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
        pick_options: [
          { label: "Lose 3 life", life_cost: 3 },
          { label: "Sacrifice a nonland permanent" },
          { label: "Discard a card" },
        ],
      },
    ],
  }) as unknown as GameView;

interface Sent {
  type: ActionType;
  params?: unknown;
}

function mountOptionPick(): { container: HTMLElement; sent: Sent[] } {
  const sent: Sent[] = [];
  const view = render(
    ChoicePromptModal as never,
    {
      snap: snapWithOptionPick(),
      viewerID: "me",
      sendAction: (type: ActionType, params?: unknown) => sent.push({ type, params }),
      lastError: null,
    } as never,
  );
  return { container: view.container, sent };
}

const optionButtons = (container: HTMLElement): HTMLButtonElement[] => [
  ...container.querySelectorAll<HTMLButtonElement>("button.pick-option"),
];

describe("ChoicePromptModal — option_pick (#568)", () => {
  it("renders one button per branch, in printed order", () => {
    const { container } = mountOptionPick();
    const labels = optionButtons(container).map((b) => b.textContent?.trim());
    expect(labels).toEqual(["Lose 3 life", "Sacrifice a nonland permanent", "Discard a card"]);
    // The card grid is the fallback arm — an option pick must not
    // land in it, or the answer goes out as card_ids.
    expect(container.querySelector(".card-grid")).toBeNull();
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
});
