// @vitest-environment jsdom
import { beforeEach, describe, expect, it } from "vitest";
import type { GameView } from "./protocol";
import { SIGNAL_PREFIX, shouldSignal, showTurnSignal, waitingOnViewer } from "./turnSignal";

const ME = "p-me";
const OTHER = "p-other";

function view(over: Record<string, unknown> = {}): GameView {
  return {
    state: "active",
    mulligans_open: false,
    seats: [{ id: ME }, { id: OTHER }],
    turn: { priority_holder: 1, active_seat: 1 },
    ...over,
  } as unknown as GameView;
}

describe("waitingOnViewer", () => {
  it("is true when the viewer holds priority", () => {
    expect(waitingOnViewer(view({ turn: { priority_holder: 0 } }), ME)).toBe(true);
  });
  it("is false when someone else holds priority", () => {
    expect(waitingOnViewer(view(), ME)).toBe(false);
  });
  it("is true for an owed prompt without priority", () => {
    expect(waitingOnViewer(view({ pending_choices: [{ chooser: ME }] }), ME)).toBe(true);
    expect(waitingOnViewer(view({ pending_choices: [{ chooser: OTHER }] }), ME)).toBe(false);
    expect(waitingOnViewer(view({ discard_pending: { [ME]: 1 } }), ME)).toBe(true);
  });
  it("is true for an owed block declaration", () => {
    expect(
      waitingOnViewer(view({ turn: { priority_holder: -1, block_decision_seats: [0] } }), ME),
    ).toBe(true);
  });
  it("follows the mulligan turn while the window is open", () => {
    const seats = (t: boolean) => [{ id: ME, mulligan_turn: t }, { id: OTHER }];
    const open = { mulligans_open: true, turn: { priority_holder: 0 } };
    expect(waitingOnViewer(view({ ...open, seats: seats(true) }), ME)).toBe(true);
    expect(waitingOnViewer(view({ ...open, seats: seats(false) }), ME)).toBe(false);
  });
  it("never signals a spectator, a finished game or an eliminated seat", () => {
    const mine = { turn: { priority_holder: 0 } };
    expect(waitingOnViewer(view(mine), null)).toBe(false);
    expect(waitingOnViewer(view({ ...mine, state: "ended" }), ME)).toBe(false);
    expect(waitingOnViewer(view({ ...mine, seats: [{ id: ME, eliminated: true }] }), ME)).toBe(
      false,
    );
    expect(waitingOnViewer(null, ME)).toBe(false);
  });
});

describe("shouldSignal", () => {
  const mine = view({ turn: { priority_holder: 0 } });
  it("needs a hidden tab", () => {
    expect(shouldSignal(mine, ME, true)).toBe(true);
    expect(shouldSignal(mine, ME, false)).toBe(false);
  });
});

describe("showTurnSignal", () => {
  beforeEach(() => {
    document.head.innerHTML = '<link rel="icon" href="/icons/favicon.svg">';
    document.title = "CMD & CTRL";
    document.documentElement.style.setProperty("--bg", "#0b0a09");
    document.documentElement.style.setProperty("--accent", "#d9b45c");
    document.documentElement.style.setProperty("--fg", "#f1ece2");
  });
  it("dots the title and icon, then restores both exactly", () => {
    const clear = showTurnSignal(document);
    expect(document.title).toBe(SIGNAL_PREFIX + "CMD & CTRL");
    const href = document.querySelector('link[rel="icon"]')?.getAttribute("href") ?? "";
    expect(decodeURIComponent(href)).toContain("<circle");
    clear();
    expect(document.title).toBe("CMD & CTRL");
    expect(document.querySelector('link[rel="icon"]')?.getAttribute("href")).toBe(
      "/icons/favicon.svg",
    );
  });
});
