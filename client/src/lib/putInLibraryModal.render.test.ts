// @vitest-environment jsdom
//
// #996 / ADR 0088: the put_in_library prompt reuses the scry dialog,
// with its lanes chosen by `placement`. What has to hold in the markup:
// a "bottom" placement shows only the bottom lane and lets the pile be
// reordered, a "top" placement is a pure reorder, and the answer goes
// out as {top_order, bottom} with both lists top-first.

import { describe, it, expect, afterEach } from "vitest";

import ChoicePromptModal from "./components/board/ChoicePromptModal.svelte";
import type { ActionType, GameView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const card = (id: string, name: string) => ({
  instance_id: id,
  name,
  type_line: "Sorcery",
  owner: "me",
  controller: "me",
});

type Extra = { top_count?: number; top_depth?: number; owner?: string; cards?: number };

const snapWith = (placement: "top" | "bottom" | "top_or_bottom", extra: Extra = {}): GameView =>
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
        kind: "put_in_library",
        chooser: "me",
        from_player: "me",
        count: extra.cards ?? 3,
        placement,
        top_count: extra.top_count,
        top_depth: extra.top_depth,
        reason: "Impulse — put the rest on the bottom of your library in any order",
        options: [card("a", "Alpha"), card("b", "Beta"), card("c", "Gamma")]
          .slice(0, extra.cards ?? 3)
          .map((c) => ({ ...c, owner: extra.owner ?? c.owner })),
      },
    ],
  }) as unknown as GameView;

interface Sent {
  type: ActionType;
  params?: unknown;
}

function mount(placement: "top" | "bottom" | "top_or_bottom", extra: Extra = {}) {
  const sent: Sent[] = [];
  const view = render(
    ChoicePromptModal as never,
    {
      snap: snapWith(placement, extra),
      viewerID: "me",
      sendAction: (type: ActionType, params?: unknown) => sent.push({ type, params }),
      lastError: null,
    } as never,
  );
  return { container: view.container, sent };
}

const button = (container: HTMLElement, pred: (b: HTMLButtonElement) => boolean) =>
  [...container.querySelectorAll("button")].find(pred);

const done = (container: HTMLElement) =>
  button(container, (b) => (b.textContent ?? "").trim() === "Done")!;

describe("ChoicePromptModal — put_in_library", () => {
  it("a bottom placement shows only the bottom lane and reorders the pile", () => {
    const { container, sent } = mount("bottom");
    const labels = [...container.querySelectorAll(".lane-label")].map((h) => h.textContent ?? "");
    expect(labels.some((l) => /On top/.test(l))).toBe(false);
    expect(labels.some((l) => /On the bottom \(3\)/.test(l))).toBe(true);
    expect(button(container, (b) => /Keep on top/.test(b.textContent ?? ""))).toBeUndefined();

    // Move Gamma up one: [Alpha, Gamma, Beta].
    click(
      button(
        container,
        (b) => b.getAttribute("aria-label") === "move Gamma up in the bottom pile",
      )!,
    );
    click(done(container));
    expect(sent).toHaveLength(1);
    expect(sent[0].params).toMatchObject({
      choice_id: "choice-1",
      bottom: ["a", "c", "b"],
      top_order: [],
    });
  });

  it("a top placement is a pure reorder on top", () => {
    const { container, sent } = mount("top");
    expect(button(container, (b) => /To bottom/.test(b.textContent ?? ""))).toBeUndefined();
    click(button(container, (b) => b.getAttribute("aria-label") === "move Beta up")!);
    click(done(container));
    expect(sent[0].params).toMatchObject({ top_order: ["b", "a", "c"], bottom: [] });
  });

  it("top_or_bottom offers both lanes", () => {
    const { container, sent } = mount("top_or_bottom");
    click(button(container, (b) => b.getAttribute("aria-label") === "put Alpha on the bottom")!);
    click(done(container));
    expect(sent[0].params).toMatchObject({ top_order: ["b", "c"], bottom: ["a"] });
  });

  // #1298: Cream of the Crop — exactly one card stays on top. The seed
  // is already a legal answer, and Done waits while the count is off.
  it("an exact top count seeds a legal answer and holds Done until it holds", () => {
    const { container, sent } = mount("top_or_bottom", { top_count: 1 });
    const labels = [...container.querySelectorAll(".lane-label")].map((h) => h.textContent ?? "");
    expect(labels.some((l) => /On top \(1 of 1\)/.test(l))).toBe(true);
    expect(done(container).disabled).toBe(false);

    click(button(container, (b) => b.getAttribute("aria-label") === "keep Beta on top")!);
    expect(done(container).disabled).toBe(true);
    click(done(container));
    expect(sent).toHaveLength(0);

    click(button(container, (b) => b.getAttribute("aria-label") === "put Alpha on the bottom")!);
    expect(done(container).disabled).toBe(false);
    click(done(container));
    expect(sent[0].params).toMatchObject({ top_order: ["b"], bottom: ["c", "a"] });
  });

  // #1298: Temporal Cleansing — the top lane is "second from the top".
  it("a top depth names the lane by its position", () => {
    const { container, sent } = mount("top_or_bottom", { top_depth: 2, cards: 1, owner: "them" });
    const labels = [...container.querySelectorAll(".lane-label")].map((h) => h.textContent ?? "");
    expect(labels.some((l) => /Second from the top \(1\)/.test(l))).toBe(true);
    expect(container.textContent ?? "").toMatch(
      /second from the top, or on the bottom of its owner's library/,
    );
    click(done(container));
    expect(sent[0].params).toMatchObject({ top_order: ["a"], bottom: [] });
  });

  // #1298: Jace's +2 and Portent order ANOTHER player's library — the
  // hint must not call its top card "your next draw".
  it("another player's library is not your next draw", () => {
    const { container } = mount("top", { owner: "them" });
    const text = container.textContent ?? "";
    expect(text).toMatch(/that player's next draw/);
    expect(text).not.toMatch(/your next draw/);
  });
});
