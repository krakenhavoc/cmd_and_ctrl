// @vitest-environment jsdom
//
// sideboarding.render.test.ts — the screen between two games of a match
// (ADR 0222 §3), mounted on its own. A move is kept on the screen until
// the deck is one the format's check takes, then sent whole; Ready waits
// for a complete, saved deck; the loser alone gets the play-or-draw
// choice; and conceding the match asks first.

import { afterEach, describe, expect, it, vi } from "vitest";

import { L } from "./labels";
import type { CardView, MatchPoolView, MatchView, PlayerView, SideboardingView } from "./protocol";
import { cleanup, click, flushSync, render } from "./test/render.svelte";
import SideboardingSheet from "./components/board/SideboardingSheet.svelte";

function card(id: string, name: string): CardView {
  return { instance_id: id, name, owner: "bob", controller: "bob" } as CardView;
}

// Sixty in the deck (fifty-six Mountains and four Shocks) and two
// Abrades in the sideboard.
function pool(): MatchPoolView {
  const deck: CardView[] = [];
  for (let i = 0; i < 56; i++) deck.push(card(`land-${i}`, "Mountain"));
  for (let i = 0; i < 4; i++) deck.push(card(`shock-${i}`, "Shock"));
  return {
    seat: 1,
    player_id: "bob",
    deck,
    sideboard: [card("abrade-0", "Abrade"), card("abrade-1", "Abrade")],
    registered: deck.map((c) => c.instance_id),
  };
}

const seats = [
  { id: "alice", seat: 0, name: "Alice" },
  { id: "bob", seat: 1, name: "Bob" },
] as PlayerView[];

function mount(over: Partial<SideboardingView> = {}, p: MatchPoolView = pool()) {
  const sideboarding: SideboardingView = {
    deadline: "2026-11-11T00:03:00Z",
    seconds_left: 125,
    ready: [],
    chooser_seat: 1,
    play_first: true,
    deck_min: 60,
    sideboard_max: 15,
    pools: [p],
    ...over,
  };
  const match: MatchView = {
    best_of: 3,
    game: 1,
    wins: [1, 0],
    games: [{ winner_seat: 0, starting_seat: 0, turns: 6 }],
    sideboarding,
  };
  const calls = {
    setDeck: vi.fn<(main: string[]) => void>(),
    choose: vi.fn<(playFirst: boolean) => void>(),
    ready: vi.fn<(ready: boolean) => void>(),
    concede: vi.fn(),
    hide: vi.fn(),
  };
  const r = render(
    SideboardingSheet as never,
    {
      match,
      sideboarding,
      pool: p,
      seats,
      viewerSeat: 1,
      onSetDeck: calls.setDeck,
      onChoosePlayFirst: calls.choose,
      onReady: calls.ready,
      onConcedeMatch: calls.concede,
      onHide: calls.hide,
    } as never,
  );
  return { ...r, calls };
}

function button(container: HTMLElement, text: string): HTMLButtonElement {
  const found = [...container.querySelectorAll("button")].find(
    (b) => b.textContent?.trim() === text,
  );
  if (!found) throw new Error(`no button "${text}"`);
  return found as HTMLButtonElement;
}

function byLabel(container: HTMLElement, label: string): HTMLButtonElement {
  const found = container.querySelector(`[aria-label="${label}"]`);
  if (!found) throw new Error(`nothing labelled "${label}"`);
  return found as HTMLButtonElement;
}

afterEach(cleanup);

describe("the sideboarding screen", () => {
  it("is a named dialog that says where the match stands", () => {
    const { container } = mount();
    const dialog = container.querySelector('[role="dialog"]');
    expect(dialog?.getAttribute("aria-label")).toBe(L.sideboarding);
    const text = container.textContent ?? "";
    expect(text).toContain("Game 2 of 3 is next");
    expect(text).toContain("Alice won game 1.");
    expect(text).toContain("Alice 1–0 Bob");
    expect(container.querySelector('[role="timer"]')?.textContent).toBe("2:05");
    expect(text).toContain("This is your registered deck.");
    // By name, with the copies.
    const lines = [...container.querySelectorAll("section.list li")].map((li) =>
      li.textContent?.replace(/\s+/g, " ").trim(),
    );
    expect(lines).toEqual(["56× Mountain Take out", "4× Shock Take out", "2× Abrade Bring in"]);
  });

  it("keeps a move here until the deck is complete, then sends it whole", () => {
    const { container, calls } = mount();
    click(byLabel(container, "move one Shock to your sideboard"));
    flushSync();
    // Fifty-nine cards: nothing is sent, and Ready waits.
    expect(calls.setDeck).not.toHaveBeenCalled();
    expect(container.textContent).toContain("Your deck has 59 cards. Bring in 1 more to reach 60.");
    expect(button(container, L.sideboardReady).disabled).toBe(true);

    click(byLabel(container, "move one Abrade to your deck"));
    flushSync();
    expect(calls.setDeck).toHaveBeenCalledTimes(1);
    const sent = calls.setDeck.mock.calls[0][0];
    expect(sent).toHaveLength(60);
    expect(sent).toContain("abrade-1");
    expect(sent).not.toContain("shock-3");
    // What moved is marked on each side.
    const text = container.textContent?.replace(/\s+/g, " ") ?? "";
    expect(text).toContain("1× Abrade 1 in");
    expect(text).toContain("1× Shock 1 out");
    // Until the server's deck catches up, Ready still waits.
    expect(text).toContain("Saving your deck…");
    expect(button(container, L.sideboardReady).disabled).toBe(true);
  });

  it("readies up with a complete deck, and locks the deck while ready", () => {
    const { container, calls } = mount();
    const ready = button(container, L.sideboardReady);
    expect(ready.disabled).toBe(false);
    click(ready);
    expect(calls.ready).toHaveBeenCalledWith(true);
    cleanup();

    const again = mount({ ready: [1] });
    expect(byLabel(again.container, "move one Shock to your sideboard").disabled).toBe(true);
    click(button(again.container, "Not ready"));
    expect(again.calls.ready).toHaveBeenCalledWith(false);
    expect(again.container.textContent?.replace(/\s+/g, " ")).toContain("Bob ready");
  });

  it("gives the play-or-draw choice to the loser alone", () => {
    const { container, calls } = mount();
    expect(container.textContent).toContain("You lost the last game, so you choose.");
    const radios = [...container.querySelectorAll<HTMLInputElement>('input[name="play-first"]')];
    expect(radios.map((r) => r.checked)).toEqual([true, false]);
    radios[1].checked = true;
    radios[1].dispatchEvent(new Event("change", { bubbles: true }));
    expect(calls.choose).toHaveBeenCalledWith(false);
    cleanup();

    const other = mount({ chooser_seat: 0, play_first: undefined });
    expect(other.container.querySelectorAll('input[name="play-first"]')).toHaveLength(0);
    expect(other.container.textContent).toContain("Alice chooses");
  });

  it("asks before conceding the match, and sets the screen aside to show the board", () => {
    const { container, calls } = mount();
    click(button(container, `${L.concedeMatch}…`));
    flushSync();
    expect(calls.concede).not.toHaveBeenCalled();
    expect(container.textContent).toContain("Concede the whole match?");
    click(button(container, "Keep playing"));
    flushSync();
    click(button(container, `${L.concedeMatch}…`));
    flushSync();
    click(button(container, L.concedeMatch));
    expect(calls.concede).toHaveBeenCalledTimes(1);
    cleanup();

    const again = mount();
    click(button(again.container, "Look at the board"));
    expect(again.calls.hide).toHaveBeenCalledTimes(1);
  });

  it("offers the way back to the registered deck once it has changed", () => {
    const p = pool();
    // The server has a one-for-one swap already.
    p.deck = [...p.deck.filter((c) => c.instance_id !== "shock-3"), p.sideboard[1]];
    p.sideboard = [p.sideboard[0], card("shock-3", "Shock")];
    const { container, calls } = mount({}, p);
    expect(container.textContent).toContain("Your deck is saved for the next game.");
    click(button(container, "Back to registered deck"));
    flushSync();
    expect(calls.setDeck).toHaveBeenCalledTimes(1);
    expect([...calls.setDeck.mock.calls[0][0]].sort()).toEqual([...p.registered].sort());
  });
});
