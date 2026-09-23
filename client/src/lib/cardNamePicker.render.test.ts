// @vitest-environment jsdom
//
// #1210 — the "as this permanent enters, choose a card name" picker
// (CR 614.12): Pithing Needle, Phyrexian Revoker, Sorcerous Spyglass.
//
// It looks like the creature-type picker and behaves differently in
// the one way that matters, which is why it earns a test of its own:
// there is NO legal set. CR 201.2 lets a player name any card name,
// so `name_options` is a suggestion list of what is publicly visible
// and the text box is the real answer. The assertions below are the
// three ways that difference shows up in the DOM — a suggestion
// submits, free text submits, and free text that matches no
// suggestion still submits.

import { describe, it, expect, afterEach } from "vitest";

import ChoicePromptModal from "./components/board/ChoicePromptModal.svelte";
import type { ActionType, GameView } from "./protocol";
import { render, click, cleanup, flushSync } from "./test/render.svelte";

afterEach(cleanup);

const snapWithNamePrompt = (): GameView =>
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
        id: "choice-name",
        kind: "choose_card_name",
        chooser: "me",
        from_player: "me",
        reason: "Pithing Needle — choose a card name",
        name_options: ["Sol Ring", "Sensei's Divining Top", "Gaea's Cradle"],
        options: [],
      },
    ],
  }) as unknown as GameView;

interface Sent {
  type: ActionType;
  params?: unknown;
}

function mountNamePrompt(): { container: HTMLElement; sent: Sent[] } {
  const sent: Sent[] = [];
  const view = render(
    ChoicePromptModal as never,
    {
      snap: snapWithNamePrompt(),
      viewerID: "me",
      sendAction: (type: ActionType, params?: unknown) => sent.push({ type, params }),
      lastError: null,
    } as never,
  );
  return { container: view.container, sent };
}

const nameBox = (container: HTMLElement): HTMLInputElement =>
  container.querySelector("input.type-filter") as HTMLInputElement;

function type(container: HTMLElement, text: string): void {
  const box = nameBox(container);
  box.value = text;
  box.dispatchEvent(new Event("input", { bubbles: true }));
  flushSync();
}

function pressEnter(container: HTMLElement): void {
  nameBox(container).dispatchEvent(
    new KeyboardEvent("keydown", { key: "Enter", bubbles: true, cancelable: true }),
  );
  flushSync();
}

describe("ChoicePromptModal — choose a card name (#1210)", () => {
  it("offers the publicly visible names as suggestions", () => {
    const { container } = mountNamePrompt();
    const labels = Array.from(container.querySelectorAll("button.type-pick")).map((b) =>
      b.textContent?.trim(),
    );
    expect(labels).toEqual(["Sol Ring", "Sensei's Divining Top", "Gaea's Cradle"]);
  });

  it("sends card_name when a suggestion is clicked", () => {
    const { container, sent } = mountNamePrompt();
    const top = Array.from(container.querySelectorAll("button.type-pick")).find(
      (b) => b.textContent?.trim() === "Sol Ring",
    ) as HTMLElement;
    click(top);
    expect(sent).toEqual([
      { type: "resolve_choice", params: { choice_id: "choice-name", card_name: "Sol Ring" } },
    ]);
  });

  it("filters the suggestions, prefix matches first", () => {
    const { container } = mountNamePrompt();
    type(container, "se");
    const labels = Array.from(container.querySelectorAll("button.type-pick")).map((b) =>
      b.textContent?.trim(),
    );
    expect(labels).toEqual(["Sensei's Divining Top"]);
  });

  // The assertion this file exists for. CR 201.2 admits a name nobody
  // at the table can see — a card in a library, in a hand, or in no
  // deck here at all — and the picker must submit exactly what was
  // typed rather than the nearest suggestion.
  it("sends free text that matches no suggestion", () => {
    const { container, sent } = mountNamePrompt();
    type(container, "  Rhystic Study  ");
    pressEnter(container);
    expect(sent).toEqual([
      { type: "resolve_choice", params: { choice_id: "choice-name", card_name: "Rhystic Study" } },
    ]);
  });

  it("refuses to submit an empty name", () => {
    const { container, sent } = mountNamePrompt();
    type(container, "   ");
    pressEnter(container);
    expect(sent).toEqual([]);
  });
});
