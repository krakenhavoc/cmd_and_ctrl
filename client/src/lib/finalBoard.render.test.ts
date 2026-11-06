// @vitest-environment jsdom
//
// #4912: an eliminated player's final board. The tombstone on their panel
// carries a Final board button that opens a read-only viewer of what they
// had when they left the game, from the seat's own `final_board`: the
// battlefield (attachments after their host, tapped cards noted), the
// hand, graveyard, exile and command zone, never the library. It works on
// a page opened after the player left, the same as on one that saw it.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushSync } from "svelte";

vi.mock("./sounds", () => ({ play: vi.fn() }));

import Board from "./components/board/Board.svelte";
import Tombstone from "./components/board/Tombstone.svelte";
import { L } from "./labels";
import {
  closeFinalBoard,
  finalBattlefieldOrder,
  finalBoardCount,
  finalCardNotes,
  firstFinalZone,
} from "./finalBoard";
import type { CardView, FinalBoardView, FinalCardView, GameView, PlayerView } from "./protocol";
import { resetSettings, updateSettings } from "./settings";
import { cleanup, render } from "./test/render.svelte";

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

const IDS = ["me", "bob", "cat", "dee"];
const NAMES = ["Me", "Bob", "Cat", "Dee"];

const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

const fc = (id: string, name: string, extra: Partial<FinalCardView> = {}): FinalCardView => ({
  instance_id: id,
  name,
  ...extra,
});

// What Cat had when she left.
const CAT_BOARD: FinalBoardView = {
  battlefield: [
    fc("cat-sword", "Sword of Feast and Famine", {
      type_line: "Artifact — Equipment",
      attached_to: "cat-bear",
      attached_to_name: "Grizzly Bears",
    }),
    fc("cat-bear", "Grizzly Bears", {
      type_line: "Creature — Bear",
      power: 4,
      toughness: 4,
      tapped: true,
      counters: { "+1/+1": 2 },
    }),
    fc("cat-curse", "Curse of Opulence", {
      type_line: "Enchantment — Aura Curse",
      attached_to: "bob",
      attached_to_name: "Bob",
    }),
  ],
  hand: [fc("cat-h1", "Counterspell"), fc("cat-h2", "Swords to Plowshares")],
  graveyard: [fc("cat-g1", "Sol Ring")],
  command: [fc("cat-cmdr", "Edea", { is_commander: true, type_line: "Legendary Creature" })],
};

function seatView(n: number, out: boolean, finalBoard?: FinalBoardView): PlayerView {
  const id = IDS[n];
  return {
    id,
    name: NAMES[n],
    seat: n,
    life: out ? 0 : 40,
    library: zone("library", id),
    hand: zone("hand", id),
    graveyard: zone("graveyard", id),
    command: zone("command", id),
    commander_damage: {},
    life_history: [],
    mana_pool: [],
    ...(out ? { eliminated: true, loss_cause: "life" } : {}),
    ...(out && finalBoard ? { final_board: finalBoard } : {}),
  } as unknown as PlayerView;
}

// A four-player game; Cat (seat 2) has left when `catOut`.
function gameView(catOut: boolean, withRecord = true): GameView {
  return {
    id: "g1",
    state: "active",
    seats: IDS.map((_, n) => seatView(n, catOut && n === 2, withRecord ? CAT_BOARD : undefined)),
    battlefield: zone("battlefield", undefined),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: { number: 4, active_seat: 0, priority_holder: 0, phase: "main1", step: "precombat_main" },
    mulligans_open: false,
    log: catOut ? [{ seq: 10, kind: "eliminated", seat: 2, cause: "life", turn: 4, text: "" }] : [],
  } as unknown as GameView;
}

beforeEach(() => {
  vi.useFakeTimers();
  vi.stubGlobal("ResizeObserver", FakeObserver);
  vi.stubGlobal("requestAnimationFrame", (cb: FrameRequestCallback) =>
    setTimeout(() => cb(performance.now()), 16),
  );
  vi.stubGlobal("cancelAnimationFrame", (id: number) => clearTimeout(id));
  resetSettings();
});

afterEach(() => {
  closeFinalBoard();
  cleanup();
  vi.useRealTimers();
  vi.unstubAllGlobals();
  resetSettings();
});

describe("the final board's order and notes", () => {
  it("puts each attachment straight after its host, and keeps one whose host is elsewhere", () => {
    const order = finalBattlefieldOrder(CAT_BOARD.battlefield!).map((c) => c.instance_id);
    expect(order).toEqual(["cat-bear", "cat-sword", "cat-curse"]);
  });

  it("says what a card was on and that it was tapped", () => {
    const [sword, bear, curse] = CAT_BOARD.battlefield!;
    expect(finalCardNotes(sword)).toEqual(["on Grizzly Bears"]);
    expect(finalCardNotes(bear)).toEqual(["tapped"]);
    expect(finalCardNotes(curse)).toEqual(["on Bob"]);
    expect(finalCardNotes(CAT_BOARD.hand![0])).toEqual([]);
  });

  it("counts every zone and opens on the first one with a card", () => {
    expect(finalBoardCount(CAT_BOARD)).toBe(7);
    expect(firstFinalZone(CAT_BOARD)).toBe("battlefield");
    expect(firstFinalZone({ graveyard: [fc("x", "X")] })).toBe("graveyard");
    expect(firstFinalZone(undefined)).toBe("battlefield");
  });
});

describe("the tombstone's Final board button", () => {
  const mount = (view: GameView, n: number, compact = false) =>
    render(Tombstone as never, { seat: view.seats[n], view, compact } as never);
  const button = () =>
    document.querySelector<HTMLButtonElement>(`button[aria-label="${L.finalBoardOpen("Cat")}"]`);

  it("is on the stone of a player who left with a final board", () => {
    mount(gameView(true), 2);
    const b = button();
    expect(b).not.toBeNull();
    expect(b!.textContent?.trim()).toBe(L.finalBoard);
    expect(b!.closest("[data-tombstone]")).not.toBeNull();
  });

  it("is on the small stone of a collapsed seat too", () => {
    mount(gameView(true), 2, true);
    expect(button()).not.toBeNull();
  });

  it("is not there for a seat restored without the record", () => {
    mount(gameView(true, false), 2);
    expect(document.querySelector("[data-tombstone]")).not.toBeNull();
    expect(button()).toBeNull();
  });
});

describe("the final board viewer on the board", () => {
  function mountBoard(view: GameView) {
    return render(
      Board as never,
      {
        view,
        viewerID: "me",
        isAdmin: false,
        sendAction: () => {},
        combatMode: "idle",
        selectedCombatCardID: null,
        onSelectCombatCard: () => {},
        onDeclareAttack: () => {},
        onDeclareBlock: () => {},
      } as never,
    );
  }
  const dialog = () =>
    document.querySelector<HTMLElement>(`[role="dialog"][aria-label="${L.finalBoardOf("Cat")}"]`);
  const shownCards = () =>
    [...dialog()!.querySelectorAll<HTMLElement>("[data-final-card]")].map(
      (e) => e.dataset.finalCard,
    );
  const tab = (kind: string) =>
    dialog()!.querySelector<HTMLButtonElement>(`[data-final-zone="${kind}"]`)!;
  const openButton = () =>
    document.querySelector<HTMLButtonElement>(`button[aria-label="${L.finalBoardOpen("Cat")}"]`);

  function open(): void {
    openButton()!.click();
    flushSync();
  }

  it("opens from the tombstone of a player who left on a page opened afterwards (a reload)", () => {
    updateSettings("display", "expandStyle", "overlay");
    mountBoard(gameView(true));
    expect(dialog()).toBeNull();
    open();
    const d = dialog();
    expect(d).not.toBeNull();
    expect(d!.textContent).toContain("Cat");
    expect(d!.textContent).toContain(L.finalBoard);
    expect(d!.textContent).toContain("left the game");
    // The battlefield, in reading order, with the notes.
    expect(shownCards()).toEqual(["cat-bear", "cat-sword", "cat-curse"]);
    expect(d!.textContent).toContain("tapped");
    expect(d!.textContent).toContain("on Grizzly Bears");
    expect(d!.textContent).toContain("on Bob");
    // The tabs carry their counts; there is no library tab.
    expect(tab("battlefield").textContent).toContain("3");
    expect(tab("hand").textContent).toContain("2");
    expect(tab("graveyard").textContent).toContain("1");
    expect(tab("exile").textContent).toContain("0");
    expect(tab("command").textContent).toContain("1");
    expect(dialog()!.querySelector('[data-final-zone="library"]')).toBeNull();
  });

  it("shows the recorded hand, graveyard and command zone to an opponent", () => {
    updateSettings("display", "expandStyle", "overlay");
    mountBoard(gameView(true));
    open();
    tab("hand").click();
    flushSync();
    expect(shownCards()).toEqual(["cat-h1", "cat-h2"]);
    tab("graveyard").click();
    flushSync();
    expect(shownCards()).toEqual(["cat-g1"]);
    tab("command").click();
    flushSync();
    expect(shownCards()).toEqual(["cat-cmdr"]);
    tab("exile").click();
    flushSync();
    expect(shownCards()).toEqual([]);
    expect(dialog()!.textContent).toContain("Nothing here");
  });

  it("opens when the player leaves in front of the viewer, and closes on Escape", () => {
    updateSettings("display", "expandStyle", "overlay");
    const r = mountBoard(gameView(false));
    expect(openButton()).toBeNull();
    r.setProps({ view: gameView(true) } as never);
    flushSync();
    open();
    expect(dialog()).not.toBeNull();
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    flushSync();
    expect(dialog()).toBeNull();
  });

  it("closes when the elimination is undone", () => {
    updateSettings("display", "expandStyle", "overlay");
    const r = mountBoard(gameView(true));
    open();
    expect(dialog()).not.toBeNull();
    r.setProps({ view: gameView(false) } as never);
    flushSync();
    expect(dialog()).toBeNull();
    // And it does not come back by itself when she leaves again.
    r.setProps({ view: gameView(true) } as never);
    flushSync();
    expect(dialog()).toBeNull();
  });
});
