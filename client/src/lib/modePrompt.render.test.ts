// @vitest-environment jsdom
//
// #764: the mode_pick arm of ChoicePromptModal — a modal TRIGGERED
// ability asking its controller which bullet to take as the ability
// is put on the stack (CR 603.3c).
//
// Two things live only in the markup and so need a DOM. One: the
// bullets on the prompt are the ONLY ones offered — the server drops
// a bullet whose clause has no legal target, and mode_indexes is what
// says which ModeSpec index each surviving label belongs to, so a
// picker that answered with the render index would answer the wrong
// question. Two: the repeatable case (CR 700.2d) is a count per
// option, not a toggle, and the answer is a multiset in click order.

import { describe, it, expect, afterEach } from "vitest";

import ChoicePromptModal from "./components/board/ChoicePromptModal.svelte";
import type { ActionType, GameView, PendingChoiceView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

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
        kind: "mode_pick",
        chooser: "me",
        from_player: "me",
        reason: "Choose one",
        options: [],
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

const optionButtons = (c: HTMLElement): HTMLElement[] =>
  Array.from(c.querySelectorAll(".prompt-options .prompt-opt")) as HTMLElement[];

const primary = (c: HTMLElement): HTMLButtonElement =>
  c.querySelector(".prompt-foot .primary") as HTMLButtonElement;

describe("ChoicePromptModal — mode_pick (#764, CR 603.3c)", () => {
  it("offers only the bullets the server sent, and answers with their ModeSpec index", () => {
    // Glissa Sunslayer with no enchantment and no counter on the
    // board: bullets 0 and 2 survive, bullet 1 does not.
    const { container, sent } = mount({
      mode_options: ["You draw a card and lose 1 life.", "Remove up to three counters."],
      mode_indexes: [0, 2],
      mode_min: 1,
      mode_max: 1,
    });
    const opts = optionButtons(container);
    expect(opts.length).toBe(2);
    expect(opts[1].textContent).toContain("Remove up to three counters.");

    // The submit is dead until a bullet is chosen (min 1).
    expect(primary(container).disabled).toBe(true);
    click(opts[1]);
    expect(primary(container).disabled).toBe(false);
    click(primary(container));

    expect(sent.length).toBe(1);
    expect(sent[0].type).toBe("resolve_choice");
    // 2, not 1: the answer names the ModeSpec index, not the row.
    expect(sent[0].params).toMatchObject({ choice_id: "choice-1", modes: [2] });
  });

  it("a repeatable prompt counts occurrences and keeps click order (CR 700.2d)", () => {
    const { container, sent } = mount({
      reason: "Choose three",
      mode_options: ["Bounce a creature.", "Draw a card."],
      mode_indexes: [0, 1],
      mode_min: 3,
      mode_max: 3,
      mode_repeatable: true,
    });
    const opts = optionButtons(container);
    click(opts[1]);
    click(opts[0]);
    click(opts[1]);
    expect(primary(container).disabled).toBe(false);
    // A fourth click is refused — max 3.
    click(opts[1]);
    click(primary(container));
    expect(sent[0].params).toMatchObject({ modes: [1, 0, 1] });
  });

  it("says the ability is not on the stack until the mode is chosen", () => {
    const { container } = mount({
      mode_options: ["Draw a card."],
      mode_indexes: [0],
      mode_min: 1,
      mode_max: 1,
    });
    const hint = container.querySelector(".prompt-hint")?.textContent ?? "";
    expect(hint).toContain("not on the stack");
  });
});
