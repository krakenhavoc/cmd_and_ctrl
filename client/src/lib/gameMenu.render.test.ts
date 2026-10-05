// @vitest-environment jsdom
//
// gameMenu.render.test.ts — ADR 0111 Delivery PR 7 (S56, #1958), owner
// decision 3: the ⋯ menu moved from the command bar into the action
// dock, as the last chip on its toggles row, opening upward. It holds
// the sandbox tools, life history, the table, spawn, the vote launcher
// (out of the board's top-left corner) and Concede with its confirm. A
// viewer with no dock (a spectator) keeps a smaller one on the command
// bar.
//
// The first half renders GameMenu on its own: its items, their names
// and gates, Concede's confirm, the vote launcher, and the keys. The
// second half mounts the real Game route: where the menu is for a
// seated player and for a spectator, and what its entries send.

import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { readFileSync } from "node:fs";
import { join } from "node:path";
import { tick } from "svelte";

import GameMenu from "./components/board/GameMenu.svelte";
import { clampMulligan, parseVote, type GameMenuOptions } from "./gameMenu";
import { PROTOCOL_VERSION, type GameView, type PlayerView } from "./protocol";
import { session, type Session } from "./session";
import { _resetForTests as resetDock } from "./dock";
import { _resetForTests as resetModals, modalOpen } from "./modalLayers";
import { shortcutsHelpOpen } from "./shortcutRuntime";
import { render, click, cleanup, flushSync } from "./test/render.svelte";
import { get } from "svelte/store";

vi.mock("./components/board/Board.svelte", async () => ({
  default: (await import("./test/BoardAttentionStub.svelte")).default,
}));

const Game = await import("../routes/Game.svelte").then((m) => m.default);

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

class FakeSocket {
  static readonly OPEN = 1;
  static last: FakeSocket | null = null;
  readyState = FakeSocket.OPEN;
  sent: Array<{ kind: string; payload?: { type?: string; params?: unknown } }> = [];
  private listeners: Record<string, ((ev: unknown) => void)[]> = {};
  constructor(readonly url: string) {
    FakeSocket.last = this;
  }
  addEventListener(type: string, fn: (ev: unknown) => void): void {
    (this.listeners[type] ??= []).push(fn);
  }
  removeEventListener(): void {}
  send(data: string): void {
    this.sent.push(JSON.parse(data));
  }
  close(): void {}
  emit(type: string, ev: unknown): void {
    for (const fn of this.listeners[type] ?? []) fn(ev);
  }
  actions(): string[] {
    return this.sent.filter((f) => f.kind === "action").map((f) => f.payload?.type ?? "");
  }
  last(type: string) {
    return this.sent.filter((f) => f.kind === "action" && f.payload?.type === type).at(-1);
  }
}

let realWebSocket: unknown;
let realFetch: unknown;

beforeEach(() => {
  const g = globalThis as Record<string, unknown>;
  realWebSocket = g.WebSocket;
  realFetch = g.fetch;
  g.WebSocket = FakeSocket;
  g.ResizeObserver ??= FakeObserver;
  g.IntersectionObserver ??= FakeObserver;
  g.fetch = () => Promise.resolve(new Response("{}", { status: 404 }));
  FakeSocket.last = null;
  resetDock();
  resetModals();
  vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue(undefined);
});

afterEach(() => {
  cleanup();
  const g = globalThis as Record<string, unknown>;
  g.WebSocket = realWebSocket;
  g.fetch = realFetch;
  session.set(null);
  vi.restoreAllMocks();
});

function accessibleName(el: Element): string {
  const label = el.getAttribute("aria-label");
  if (label) return label;
  const copy = el.cloneNode(true) as Element;
  copy.querySelectorAll('[aria-hidden="true"]').forEach((n) => n.remove());
  return (copy.textContent ?? "").replace(/\s+/g, " ").trim();
}
const menuItems = (root: ParentNode): HTMLElement[] => [
  ...root.querySelectorAll<HTMLElement>('[role="menu"] [role="menuitem"]'),
];
const itemNames = (root: ParentNode): string[] => menuItems(root).map(accessibleName);
const item = (root: ParentNode, name: string | RegExp): HTMLElement | undefined =>
  menuItems(root).find((el) =>
    typeof name === "string" ? accessibleName(el) === name : name.test(accessibleName(el)),
  );
const moreButtons = (root: ParentNode) => [
  ...root.querySelectorAll<HTMLButtonElement>('button[aria-label="more actions"]'),
];
function key(target: EventTarget, k: string): KeyboardEvent {
  const e = new KeyboardEvent("keydown", { key: k, bubbles: true, cancelable: true });
  target.dispatchEvent(e);
  flushSync();
  return e;
}

// ---- GameMenu on its own --------------------------------------------------

interface Calls {
  [k: string]: unknown[];
}

function mountMenu(over: Partial<GameMenuOptions> & { placement?: "up" | "down" } = {}) {
  const calls: Calls = {};
  const rec =
    (name: string) =>
    (...args: unknown[]) => {
      (calls[name] ??= []).push(args);
    };
  const r = render(
    GameMenu as never,
    {
      seated: true,
      eliminated: false,
      gameEnded: false,
      drawKey: "D",
      drawTitle: "draw a card (D)",
      canManage: false,
      spawnAvailable: false,
      discordLink: null,
      myGames: false,
      voteOpen: false,
      onDraw: rec("draw"),
      onUntapAll: rec("untapAll"),
      onShuffle: rec("shuffle"),
      onMulligan: rec("mulligan"),
      onLifeHistory: rec("lifeHistory"),
      onTableSettings: rec("tableSettings"),
      onSpawn: rec("spawn"),
      onMyGames: rec("myGames"),
      onBack: rec("back"),
      onConcede: rec("concede"),
      onStartVote: rec("startVote"),
      ...over,
    } as never,
  );
  const trigger = r.container.querySelector<HTMLButtonElement>(
    'button[aria-label="more actions"]',
  )!;
  return { ...r, calls, trigger };
}

// The menu moves focus after its own `tick()`, which resolves a turn
// after the click's synchronous flush.
async function settle(): Promise<void> {
  await tick();
  await tick();
  flushSync();
}

async function open(trigger: HTMLButtonElement): Promise<void> {
  click(trigger);
  await settle();
  flushSync();
}

describe("GameMenu", () => {
  it("is a ⋯ button with a menu popup, closed until pressed", () => {
    const m = mountMenu();
    expect(m.trigger.getAttribute("aria-haspopup")).toBe("menu");
    expect(m.trigger.getAttribute("aria-expanded")).toBe("false");
    expect(m.container.querySelector('[role="menu"]')).toBeNull();
  });

  it("holds every entry a seated player had, in order, with the names they had", async () => {
    const m = mountMenu({
      spawnAvailable: true,
      myGames: true,
      discordLink: { href: "/auth/discord/link?game=g", label: "Link Discord" },
    });
    await open(m.trigger);
    expect(m.trigger.getAttribute("aria-expanded")).toBe("true");
    const menu = m.container.querySelector('[role="menu"]')!;
    expect(menu.getAttribute("aria-label")).toBe("game actions");
    expect(itemNames(m.container)).toEqual([
      "Draw a card D",
      "Untap all",
      "Shuffle library",
      "Life history",
      "Table settings… view",
      "Spawn a card or token…",
      "Call a vote…",
      "Link Discord",
      "My games",
      "Back to lobby",
      "Tips for the table",
      "Keyboard shortcuts",
      "Concede…",
    ]);
    // Mulligan to N keeps its field and its Go.
    expect(menu.querySelector('input[aria-label="mulligan hand size"]')).not.toBeNull();
    // Link Discord is a navigation, not a button.
    expect(item(m.container, "Link Discord")!.tagName).toBe("A");
    // No Undo: it left the menu in PR 3.
    expect(itemNames(m.container).some((n) => /undo/i.test(n))).toBe(false);
  });

  it("gates its entries: spawn, Link Discord and My games only when allowed, Table settings read-only for a guest", async () => {
    const m = mountMenu();
    await open(m.trigger);
    const names = itemNames(m.container);
    expect(names).not.toContain("Spawn a card or token…");
    expect(names).not.toContain("Link Discord");
    expect(names).not.toContain("My games");
    expect(item(m.container, /^Table settings/)!.title).toMatch(/only the host can change them/);

    m.setProps({ canManage: true } as never);
    expect(accessibleName(item(m.container, /^Table settings/)!)).toBe("Table settings…");
  });

  it("disables Concede, not hides it, once you are out or the game is over", async () => {
    const m = mountMenu({ eliminated: true });
    await open(m.trigger);
    const concede = item(m.container, "Concede…") as HTMLButtonElement;
    expect(concede.disabled).toBe(true);
    expect(concede.title).toBe("you are already eliminated");
    m.setProps({ eliminated: false, gameEnded: true } as never);
    expect((item(m.container, "Concede…") as HTMLButtonElement).disabled).toBe(true);
    expect(item(m.container, "Concede…")!.title).toBe("the game has ended");
  });

  it("closes, then acts, for each entry", async () => {
    const m = mountMenu();
    for (const [name, call] of [
      ["Draw a card D", "draw"],
      ["Untap all", "untapAll"],
      ["Shuffle library", "shuffle"],
      ["Life history", "lifeHistory"],
      ["Table settings… view", "tableSettings"],
      ["Back to lobby", "back"],
    ] as const) {
      await open(m.trigger);
      click(item(m.container, name)!);
      expect(m.container.querySelector('[role="menu"]'), name).toBeNull();
      expect(m.calls[call], name).toHaveLength(1);
    }
    await open(m.trigger);
    const field = m.container.querySelector<HTMLInputElement>(
      'input[aria-label="mulligan hand size"]',
    )!;
    field.value = "25";
    field.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();
    const go = [...m.container.querySelectorAll("button")].find((b) => b.textContent === "Go")!;
    click(go);
    expect(m.calls.mulligan).toEqual([[20]]);
  });

  it("asks before it concedes, where the menu was, and Keep playing backs out", async () => {
    const m = mountMenu();
    await open(m.trigger);
    click(item(m.container, "Concede…")!);
    await settle();
    flushSync();
    expect(m.container.querySelector('[role="menu"]')).toBeNull();
    const dlg = m.container.querySelector('[role="dialog"][aria-label="concede the game?"]')!;
    expect(dlg).not.toBeNull();
    expect(dlg.getAttribute("aria-modal")).toBe("true");
    // The safe answer has focus.
    expect(document.activeElement?.textContent).toBe("Keep playing");
    click([...dlg.querySelectorAll("button")].find((b) => b.textContent === "Keep playing")!);
    expect(m.container.querySelector('[role="dialog"]')).toBeNull();
    expect(m.calls.concede).toBeUndefined();

    await open(m.trigger);
    click(item(m.container, "Concede…")!);
    await settle();
    flushSync();
    const again = m.container.querySelector('[role="dialog"][aria-label="concede the game?"]')!;
    click([...again.querySelectorAll("button")].find((b) => accessibleName(b) === "Concede")!);
    expect(m.calls.concede).toHaveLength(1);
    expect(m.container.querySelector('[role="dialog"]')).toBeNull();
  });

  it("has the vote launcher: a topic, the options, and start", async () => {
    const m = mountMenu();
    await open(m.trigger);
    click(item(m.container, "Call a vote…")!);
    await settle();
    flushSync();
    const form = m.container.querySelector<HTMLFormElement>('form[aria-label="call a vote"]')!;
    expect(form).not.toBeNull();
    const topicField = form.querySelector<HTMLInputElement>('input[aria-label="vote topic"]')!;
    expect(document.activeElement).toBe(topicField);
    const start = [...form.querySelectorAll("button")].find((b) => b.textContent === "start")!;
    expect(start.disabled).toBe(true);
    topicField.value = "monarchy?";
    topicField.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();
    expect(start.disabled).toBe(false);
    form.dispatchEvent(new SubmitEvent("submit", { bubbles: true, cancelable: true }));
    flushSync();
    expect(m.calls.startVote).toEqual([["monarchy?", ["yes", "no"]]]);
    expect(m.container.querySelector("form")).toBeNull();
  });

  it("has no vote launcher while a vote is open: the vote is the dock's", async () => {
    const m = mountMenu({ voteOpen: true });
    await open(m.trigger);
    expect(item(m.container, "Call a vote…")).toBeUndefined();
  });

  it("opens upward from the dock and downward from the command bar", async () => {
    const up = mountMenu();
    expect(up.container.querySelector(".game-menu.up")).not.toBeNull();
    const down = mountMenu({ placement: "down", seated: false });
    expect(down.container.querySelector(".game-menu.down")).not.toBeNull();
    // The rules that place it (jsdom does no layout).
    const src = readFileSync(join(__dirname, "components/board/GameMenu.svelte"), "utf8");
    expect(src).toMatch(/\.up \.pop \{\s*bottom: calc\(100% \+ 8px\);/);
    expect(src).toMatch(/\.down \.pop \{\s*top: calc\(100% \+ 6px\);/);
  });

  it("focuses its first entry, moves with the arrow keys, and closes on Escape back to ⋯", async () => {
    const m = mountMenu();
    m.trigger.focus();
    await open(m.trigger);
    const menu = m.container.querySelector<HTMLElement>('[role="menu"]')!;
    expect(accessibleName(document.activeElement!)).toBe("Draw a card D");
    key(menu, "ArrowDown");
    expect(accessibleName(document.activeElement!)).toBe("Untap all");
    key(menu, "ArrowUp");
    key(menu, "ArrowUp");
    expect(accessibleName(document.activeElement!)).toBe("Concede…");
    key(menu, "Home");
    expect(accessibleName(document.activeElement!)).toBe("Draw a card D");
    // While it is open the global shortcuts stand down.
    expect(get(modalOpen)).toBe(true);
    const esc = key(document.activeElement!, "Escape");
    expect(esc.defaultPrevented).toBe(true);
    expect(m.container.querySelector('[role="menu"]')).toBeNull();
    expect(document.activeElement).toBe(m.trigger);
    expect(get(modalOpen)).toBe(false);
  });

  it("closes on a press outside it, and not on one inside", async () => {
    const m = mountMenu();
    await open(m.trigger);
    const menu = m.container.querySelector<HTMLElement>('[role="menu"]')!;
    menu.dispatchEvent(new MouseEvent("pointerdown", { bubbles: true }));
    flushSync();
    expect(m.container.querySelector('[role="menu"]')).not.toBeNull();
    document.body.dispatchEvent(new MouseEvent("pointerdown", { bubbles: true }));
    flushSync();
    expect(m.container.querySelector('[role="menu"]')).toBeNull();
  });

  it("for a viewer with no seat: life history, the table and navigation, nothing that acts on a seat", async () => {
    const m = mountMenu({ seated: false, placement: "down", myGames: true });
    await open(m.trigger);
    expect(itemNames(m.container)).toEqual([
      "Life history",
      "Table settings… view",
      "My games",
      "Back to lobby",
      "Tips for the table",
      "Keyboard shortcuts",
    ]);
    expect(m.container.querySelector('input[aria-label="mulligan hand size"]')).toBeNull();
  });

  // ADR 0125 §6: the Help group, after Table, for everyone.
  it("has a Help group after Table: tips for the table and the keyboard shortcuts", async () => {
    const onTableTips = vi.fn();
    const onShortcuts = vi.fn();
    const m = mountMenu({ onTableTips, onShortcuts });
    await open(m.trigger);
    const names = itemNames(m.container);
    const at = names.indexOf("Tips for the table");
    expect(at).toBeGreaterThan(names.indexOf("Back to lobby"));
    expect(names.slice(at, at + 2)).toEqual(["Tips for the table", "Keyboard shortcuts"]);
    // Concede stays last.
    expect(names.at(-1)).toBe("Concede…");
    const headings = [...m.container.querySelectorAll(".menu-h")].map((h) => h.textContent);
    expect(headings).toContain("Help");
    // A real table offers no practice game and no tutorial.
    expect(names.some((n) => /practice|tutorial/i.test(n))).toBe(false);

    click(item(m.container, "Tips for the table")!);
    flushSync();
    expect(onTableTips).toHaveBeenCalledTimes(1);
    expect(m.container.querySelector('[role="menu"]')).toBeNull();

    await open(m.trigger);
    click(item(m.container, "Keyboard shortcuts")!);
    flushSync();
    expect(onShortcuts).toHaveBeenCalledTimes(1);
    expect(m.container.querySelector('[role="menu"]')).toBeNull();
  });

  it("offers Replay the tutorial on the practice table only", async () => {
    const onReplayTutorial = vi.fn();
    const m = mountMenu({ practice: true, onReplayTutorial });
    await open(m.trigger);
    const names = itemNames(m.container);
    expect(names.slice(names.indexOf("Tips for the table"), -1)).toEqual([
      "Tips for the table",
      "Keyboard shortcuts",
      "Replay the tutorial",
    ]);
    click(item(m.container, "Replay the tutorial")!);
    flushSync();
    expect(onReplayTutorial).toHaveBeenCalledTimes(1);
  });

  // ADR 0121 §5 and §8: the Dice section and its three names.
  it("has Roll a d6, Roll a d20 and Flip a coin for a seat in the game, after the sandbox", async () => {
    const onTableRoll = vi.fn();
    const m = mountMenu({ tableRoll: { ready: true }, onTableRoll });
    await open(m.trigger);
    const names = itemNames(m.container);
    const at = names.indexOf("Roll a d6");
    expect(names.slice(at, at + 3)).toEqual(["Roll a d6", "Roll a d20", "Flip a coin"]);
    expect(at).toBeGreaterThan(names.indexOf("Life history"));
    expect(at).toBeLessThan(names.findIndex((n) => /^Table settings/.test(n)));
    click(item(m.container, "Roll a d20")!);
    flushSync();
    expect(onTableRoll).toHaveBeenCalledWith("d20");
    expect(m.container.querySelector('[role="menu"]')).toBeNull();
  });

  it("disables the dice while the last roll lands, and has none without a table roll", async () => {
    const m = mountMenu({ tableRoll: { ready: false }, onTableRoll: vi.fn() });
    await open(m.trigger);
    for (const name of ["Roll a d6", "Roll a d20", "Flip a coin"]) {
      const el = item(m.container, name) as HTMLButtonElement;
      expect(el.disabled, name).toBe(true);
      expect(el.title).toBe("wait for your last roll to land");
    }
    cleanup();
    const none = mountMenu({ tableRoll: null });
    await open(none.trigger);
    expect(itemNames(none.container)).not.toContain("Roll a d20");
  });
});

describe("gameMenu rules", () => {
  it("clamps the mulligan size to 0–20 whole cards", () => {
    expect(clampMulligan(6.7)).toBe(6);
    expect(clampMulligan(-2)).toBe(0);
    expect(clampMulligan(99)).toBe(20);
    expect(clampMulligan(Number.NaN)).toBe(7);
  });
  it("needs a topic and two options for a vote", () => {
    expect(parseVote("  ", "a, b")).toBeNull();
    expect(parseVote("x", "a")).toBeNull();
    expect(parseVote(" x ", "a, , b ,")).toEqual({ topic: "x", options: ["a", "b"] });
  });
});

// ---- Through the Game route ----------------------------------------------

const ME = "p-1";
const OPP = "p-2";
const zone = (kind: string, owner?: string) => ({ kind, owner, count: 0, cards: [] });
const seat = (id: string, name: string, n: number): PlayerView =>
  ({
    id,
    name,
    seat: n,
    life: 40,
    library: zone("library", id),
    hand: zone("hand", id),
    graveyard: zone("graveyard", id),
    command: zone("command", id),
    commander_damage: {},
    life_history: [],
    mana_pool: [],
    hand_kept: true,
    undos_remaining: 1,
  }) as unknown as PlayerView;

function view(over: Partial<GameView> = {}): GameView {
  return {
    id: "g",
    state: "active",
    seats: [seat(ME, "Me", 0), seat(OPP, "Opp", 1)],
    battlefield: zone("battlefield"),
    stack: zone("stack"),
    exile: zone("exile"),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: {
      seq: 3,
      number: 3,
      active_seat: 0,
      priority_holder: 0,
      phase: "precombat_main",
      step: "precombat_main",
    },
    mulligans_open: false,
    ...over,
  } as unknown as GameView;
}

function sessionAs(role: "player" | "spectator"): Session {
  const expiresAt = new Date(Date.now() + 3_600_000).toISOString();
  const playerID = role === "player" ? ME : "";
  return {
    token: "tok",
    expiresAt,
    principal: {
      role,
      game_id: "game-1",
      player_id: playerID,
      issued_at: new Date().toISOString(),
      expires_at: expiresAt,
    },
    playerID,
    gameID: "game-1",
  } as Session;
}

let seq = 1;
async function mountGame(role: "player" | "spectator", game: GameView = view()) {
  session.set(sessionAs(role));
  const handle = render(Game as never, { gameID: "game-1" } as never);
  FakeSocket.last!.emit("open", {});
  FakeSocket.last!.emit("message", {
    data: JSON.stringify({
      v: PROTOCOL_VERSION,
      kind: "snapshot",
      id: `frame-${seq}`,
      payload: { seq: seq++, game },
    }),
  });
  flushSync();
  await new Promise((r) => setTimeout(r, 0));
  flushSync();
  return handle.container;
}

const dockOf = (c: ParentNode) =>
  c.querySelector<HTMLElement>('.play-area > section[aria-label="actions"]');

describe("the ⋯ menu at the table", () => {
  it("is the last chip on the dock's toggles row for a seated player, and gone from the command bar", async () => {
    const c = await mountGame("player");
    const bar = c.querySelector("header.bar")!;
    expect(moreButtons(bar)).toHaveLength(0);
    const all = moreButtons(c);
    expect(all).toHaveLength(1);
    const toggles = dockOf(c)!.querySelector('[role="group"][aria-label="priority controls"]')!;
    expect(toggles.contains(all[0]!)).toBe(true);
    const chips = [...toggles.querySelectorAll("button")].filter(
      (b) => b.getAttribute("aria-label") !== "bluff options",
    );
    expect(chips.map(accessibleName)).toEqual([
      "hold",
      "autopass",
      "bluff",
      "Undo (1 left)",
      "more actions",
    ]);
  });

  it("sends what its entries ask for, and Concede only once confirmed", async () => {
    const c = await mountGame("player");
    const more = moreButtons(c)[0]!;
    await open(more);
    click(item(c, "Draw a card D") ?? item(c, /^Draw a card/)!);
    expect(FakeSocket.last!.actions()).toContain("draw_card");

    await open(more);
    click(item(c, "Concede…")!);
    await settle();
    flushSync();
    expect(FakeSocket.last!.actions()).not.toContain("concede");
    const dlg = c.querySelector('[role="dialog"][aria-label="concede the game?"]')!;
    // It opens inside the dock, where the menu was.
    expect(dockOf(c)!.contains(dlg)).toBe(true);
    click([...dlg.querySelectorAll("button")].find((b) => accessibleName(b) === "Concede")!);
    expect(FakeSocket.last!.actions()).toContain("concede");
  });

  it("opens the keyboard shortcuts from Help, and offers no tutorial at a real table", async () => {
    const c = await mountGame("player");
    const more = moreButtons(c)[0]!;
    await open(more);
    expect(item(c, "Replay the tutorial")).toBeUndefined();
    click(item(c, "Keyboard shortcuts")!);
    flushSync();
    expect(get(shortcutsHelpOpen)).toBe(true);
    shortcutsHelpOpen.set(false);
  });

  it("rolls a d20 at the table, then waits 2 s before the next roll", async () => {
    const c = await mountGame("player");
    const more = moreButtons(c)[0]!;
    await open(more);
    vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout"] });
    try {
      click(item(c, "Roll a d20")!);
      flushSync();
      const sent = FakeSocket.last!.last("roll_table_die");
      expect(sent?.payload?.params).toEqual({ die: "d20" });
      await open(more);
      expect((item(c, "Flip a coin") as HTMLButtonElement).disabled).toBe(true);
      vi.advanceTimersByTime(2000);
      flushSync();
      expect((item(c, "Flip a coin") as HTMLButtonElement).disabled).toBe(false);
    } finally {
      vi.useRealTimers();
    }
  });

  it("starts a vote from the dock, and the board no longer has a launcher", async () => {
    const c = await mountGame("player");
    expect(c.querySelector('button[aria-label="call a vote"]')).toBeNull();
    await open(moreButtons(c)[0]!);
    click(item(c, "Call a vote…")!);
    await settle();
    flushSync();
    const form = c.querySelector<HTMLFormElement>('form[aria-label="call a vote"]')!;
    const topicField = form.querySelector<HTMLInputElement>('input[aria-label="vote topic"]')!;
    topicField.value = "monarchy?";
    topicField.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();
    form.dispatchEvent(new SubmitEvent("submit", { bubbles: true, cancelable: true }));
    flushSync();
    expect(FakeSocket.last!.last("start_vote")?.payload?.params).toEqual({
      topic: "monarchy?",
      options: ["yes", "no"],
    });
  });

  it("stays on the command bar for a spectator, who has no dock, with nothing that acts on a seat", async () => {
    const c = await mountGame("spectator");
    expect(dockOf(c)).toBeNull();
    const bar = c.querySelector("header.bar")!;
    const more = moreButtons(bar);
    expect(more).toHaveLength(1);
    // After settings, in the bar's icon row.
    expect(more[0]!.closest(".bar-icons")).not.toBeNull();
    await open(more[0]!);
    const names = itemNames(c);
    expect(names).toContain("Life history");
    expect(names).toContain("Back to lobby");
    expect(names.some((n) => /^Table settings/.test(n))).toBe(true);
    for (const seatOnly of [
      "Untap all",
      "Shuffle library",
      "Call a vote…",
      "Concede…",
      "Roll a d20",
    ]) {
      expect(names).not.toContain(seatOnly);
    }
    expect(names.some((n) => /^Draw a card/.test(n))).toBe(false);
  });
});
