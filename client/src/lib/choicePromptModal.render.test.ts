// @vitest-environment jsdom
//
// The first component render test on this client (#689), and the one
// #624 wanted and could not have: the refusal line inside
// ChoicePromptModal.
//
// `choiceRejection.ts` already pins WHICH error belongs to the prompt,
// as a pure function, and that test stays the cheaper one. What it
// cannot say is that the decision reaches the player: that the message
// is rendered at all, that it is announced (`role="alert"` — a refusal
// that appears silently under a full-screen backdrop is a refusal the
// player does not know about), and that it appears only once they have
// answered. All three live in the markup, so they need a DOM.

import { describe, it, expect, afterEach } from "vitest";

import ChoicePromptModal from "./components/board/ChoicePromptModal.svelte";
import type { ActionType, GameView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

// A snapshot with one pay_unless prompt addressed to "me". pay_unless
// is the cheapest kind to stand up — two buttons, no card grid — and
// the rejection line is shared by every kind.
const snapWithPrompt = (): GameView =>
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
        kind: "pay_unless",
        chooser: "me",
        from_player: "me",
        pay_cost: "{2}",
        reason: "Pay {2} or Propaganda stops the attack",
        options: [],
      },
    ],
  }) as unknown as GameView;

interface Sent {
  type: ActionType;
  params?: unknown;
}

function mountPrompt(): {
  container: HTMLElement;
  sent: Sent[];
  setProps: (next: Record<string, unknown>) => void;
} {
  const sent: Sent[] = [];
  const view = render(
    ChoicePromptModal as never,
    {
      snap: snapWithPrompt(),
      viewerID: "me",
      sendAction: (type: ActionType, params?: unknown) => sent.push({ type, params }),
      lastError: null,
    } as never,
  );
  return { container: view.container, sent, setProps: view.setProps as never };
}

const alertText = (container: HTMLElement): string | null =>
  container.querySelector('[role="alert"]')?.textContent?.replace(/\s+/g, " ").trim() ?? null;

describe("ChoicePromptModal — the server's refusal of the open prompt", () => {
  it("renders the prompt for the choice addressed to the viewer", () => {
    const { container } = mountPrompt();

    const title = container.querySelector("#choice-title");
    expect(title?.textContent).toContain("Pay {2} or Propaganda stops the attack");
    // Nothing has been refused yet, so nothing is announced.
    expect(alertText(container)).toBeNull();
  });

  it("stays silent about an error that arrived before the player answered", () => {
    const { container, setProps } = mountPrompt();

    // A refusal of some other action entirely — the board's toast owns
    // this one. The modal must not claim it.
    setProps({
      lastError: { code: "illegal_action", message: "not your priority", at: new Date() },
    });

    expect(alertText(container)).toBeNull();
  });

  it("announces the refusal in a role=alert line once the answer is refused", () => {
    const { container, sent, setProps } = mountPrompt();

    const pay = [...container.querySelectorAll("button")].find((b) =>
      /Pay \{2\}/.test(b.textContent ?? ""),
    );
    expect(pay, "the pay button should be rendered").toBeTruthy();
    click(pay!);

    // The answer went out, and it named this prompt.
    expect(sent).toHaveLength(1);
    expect(sent[0].type).toBe("resolve_choice");
    expect(sent[0].params).toMatchObject({ choice_id: "choice-1", apply: true });

    // The server refuses it. The prompt is still open, so the reason
    // has to be readable here — under the modal's own backdrop the
    // board's toast is not visible.
    setProps({
      lastError: { code: "cannot_pay", message: "not enough mana in pool", at: new Date() },
    });

    const shown = alertText(container);
    expect(shown).toContain("Not accepted");
    expect(shown).toContain("not enough mana in pool");
  });

  it("keeps the prompt open and answerable after a refusal", () => {
    const { container, sent, setProps } = mountPrompt();

    click(
      [...container.querySelectorAll("button")].find((b) => /Pay \{2\}/.test(b.textContent ?? ""))!,
    );
    setProps({
      lastError: { code: "cannot_pay", message: "not enough mana in pool", at: new Date() },
    });

    // The whole point of showing the refusal in place: the player gets
    // another go without the modal closing out from under them.
    const decline = [...container.querySelectorAll("button")].find((b) =>
      /Don't pay/.test(b.textContent ?? ""),
    );
    expect(decline, "the decline button should still be rendered").toBeTruthy();
    click(decline!);

    expect(sent).toHaveLength(2);
    expect(sent[1].params).toMatchObject({ choice_id: "choice-1", apply: false });
  });
});
