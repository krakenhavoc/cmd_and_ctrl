// @vitest-environment jsdom
//
// openingRoll.render.test.ts — ADR 0121 §6 (PR 5), through the real
// Game route. The table during the opening roll: a seat that owes a die
// gets `roll for the first turn` (Roll, Enter); the host also gets `Roll
// for everyone left` (never Enter); the chooser gets the sheet `choose
// who takes the first turn`, where "I go first" sends at once and any
// other seat asks `Let <name> take the first turn?` first (owner
// decision 6: Confirm takes no Enter, Cancel and Escape go back). The
// strip shows the `opening roll` banner instead of `opening hand
// decisions`, and after the choice its pill reads the `starting_player`
// entry. The board is a stub that renders the strip Game.svelte hands it.

import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";

import {
  PROTOCOL_VERSION,
  type GameView,
  type LogEvent,
  type OpeningRollView,
  type PlayerView,
} from "./protocol";
import { session, type Session } from "./session";
import { _resetForTests as resetDock } from "./dock";
import { _resetForTests as resetModals } from "./modalLayers";
import { targeting, setConfirmHandler } from "./targeting";
import { defaultSettings, settings } from "./settings";
import { render, click, cleanup, flushSync } from "./test/render.svelte";
import { nameOf } from "./test/dockView";
import { DiceQueue } from "./diceQueue.svelte";
import { DICE_TUMBLE_MS, rollsFromLogs } from "./dice";
import OpeningRollBanner from "./components/board/OpeningRollBanner.svelte";

vi.mock("./components/board/Board.svelte", async () => ({
  default: (await import("./test/BoardAttentionStub.svelte")).default,
}));

const Game = await import("../routes/Game.svelte").then((m) => m.default);

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

interface SentFrame {
  kind: string;
  id: string;
  payload?: { type?: string; player?: string; params?: Record<string, unknown> };
}

class FakeSocket {
  static readonly OPEN = 1;
  static last: FakeSocket | null = null;
  readyState = FakeSocket.OPEN;
  sent: SentFrame[] = [];
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
  actions(type: string): SentFrame[] {
    return this.sent.filter((f) => f.kind === "action" && f.payload?.type === type);
  }
}

const ME = "p-1";
const OPP = "p-2";
const THIRD = "p-3";

const zone = (kind: string, owner?: string) => ({ kind, owner, count: 0, cards: [] });

const seat = (id: string, name: string, n: number, over: Partial<PlayerView> = {}): PlayerView =>
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
    hand_kept: false,
    undos_remaining: 1,
    ...over,
  }) as unknown as PlayerView;

// A three-seat table in the opening roll. The viewer is "Me", seat 0.
function rollTable(
  openingRoll: OpeningRollView | undefined,
  opts: { host?: boolean; log?: LogEvent[] } = {},
): GameView {
  return {
    id: "g",
    state: "active",
    seats: [
      seat(ME, "Me", 0, { is_host: opts.host ?? false } as never),
      seat(OPP, "Opp", 1),
      seat(THIRD, "Third", 2),
    ],
    battlefield: zone("battlefield"),
    stack: zone("stack"),
    exile: zone("exile"),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: {
      seq: 0,
      number: 0,
      active_seat: 0,
      priority_holder: -1,
      phase: "beginning",
      step: "untap",
    },
    mulligans_open: true,
    starting_seat: 0,
    log: opts.log ?? [],
    ...(openingRoll ? { opening_roll: openingRoll } : {}),
  } as unknown as GameView;
}

const ROUND_ONE: OpeningRollView = { rounds: [{ seats: [0, 1, 2], rolls: [] }] };

function playerSession(): Session {
  const expiresAt = new Date(Date.now() + 3_600_000).toISOString();
  return {
    token: "tok",
    expiresAt,
    principal: {
      role: "player",
      game_id: "game-1",
      player_id: ME,
      issued_at: new Date().toISOString(),
      expires_at: expiresAt,
    },
    playerID: ME,
    gameID: "game-1",
  } as Session;
}

function spectatorSession(): Session {
  const s = playerSession();
  return {
    ...s,
    principal: { ...s.principal, role: "spectator", player_id: "" },
    playerID: "",
  } as Session;
}

let realWebSocket: unknown;
let realFetch: unknown;
let seq = 1;

beforeEach(() => {
  const g = globalThis as Record<string, unknown>;
  realWebSocket = g.WebSocket;
  realFetch = g.fetch;
  g.WebSocket = FakeSocket;
  g.ResizeObserver ??= FakeObserver;
  g.IntersectionObserver ??= FakeObserver;
  g.fetch = () => Promise.resolve(new Response("{}", { status: 404 }));
  FakeSocket.last = null;
  settings.set(defaultSettings());
  resetDock();
  resetModals();
  targeting.set(null);
  setConfirmHandler(null);
  vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue(undefined);
  vi.spyOn(HTMLMediaElement.prototype, "load").mockImplementation(() => {});
  session.set(playerSession());
});

afterEach(() => {
  cleanup();
  resetModals();
  const g = globalThis as Record<string, unknown>;
  g.WebSocket = realWebSocket;
  g.fetch = realFetch;
  session.set(null);
  vi.restoreAllMocks();
});

function snapshot(game: GameView): void {
  FakeSocket.last!.emit("message", {
    data: JSON.stringify({
      v: PROTOCOL_VERSION,
      kind: "snapshot",
      id: `frame-${seq}`,
      payload: { seq: seq++, game },
    }),
  });
  flushSync();
}

async function mountGame(game: GameView): Promise<HTMLElement> {
  const handle = render(Game as never, { gameID: "game-1" } as never);
  FakeSocket.last!.emit("open", {});
  snapshot(game);
  await new Promise((r) => setTimeout(r, 0));
  flushSync();
  return handle.container;
}

const dockOf = (c: ParentNode) =>
  c.querySelector<HTMLElement>('.play-area > section[aria-label="actions"]');
const stripOf = (c: ParentNode) => c.querySelector<HTMLElement>('[data-testid="strip"]')!;
const dialogNamed = (root: ParentNode, name: string): HTMLElement | null =>
  [...root.querySelectorAll<HTMLElement>('[role="dialog"]')].find(
    (d) => d.getAttribute("aria-label") === name,
  ) ?? null;
const buttonNamed = (root: ParentNode, name: string): HTMLButtonElement | null =>
  [...root.querySelectorAll<HTMLButtonElement>("button")].find((b) => nameOf(b) === name) ?? null;
function keydown(key: string, target: EventTarget = document.body): void {
  target.dispatchEvent(new KeyboardEvent("keydown", { key, bubbles: true, cancelable: true }));
  flushSync();
}
function blur(): void {
  (document.activeElement as HTMLElement | null)?.blur?.();
}

describe("roll for the first turn", () => {
  it("asks a seat that owes a die, with Roll as the primary on Enter", async () => {
    const c = await mountGame(rollTable(ROUND_ONE));
    const dlg = dialogNamed(dockOf(c)!, "roll for the first turn");
    expect(dlg, "the roll request in the dock").not.toBeNull();
    expect(dlg!.getAttribute("aria-modal")).toBeNull();
    expect(dlg!.textContent).toContain("Roll a d20. The highest roll chooses who goes first.");
    const primary = dlg!.querySelector<HTMLButtonElement>(".dock-bar .request-primary")!;
    expect(nameOf(primary)).toBe("Roll");
    expect(primary.getAttribute("aria-keyshortcuts")).toBe("Enter");
    // Not the host: no "Roll for everyone left" anywhere.
    expect(buttonNamed(c, "Roll for everyone left")).toBeNull();
    // No hand yet, so no mulligan sheet.
    expect(dialogNamed(c, "keep or mulligan your hand")).toBeNull();

    click(primary);
    const sent = FakeSocket.last!.actions("roll_opening");
    expect(sent).toHaveLength(1);
    expect(sent[0].payload?.player).toBe(ME);
    // Pressed, it waits for the next frame rather than sending twice.
    expect(primary.disabled).toBe(true);
    click(primary);
    expect(FakeSocket.last!.actions("roll_opening")).toHaveLength(1);
  });

  it("rolls on Enter", async () => {
    await mountGame(rollTable(ROUND_ONE));
    blur();
    keydown("Enter");
    expect(FakeSocket.last!.actions("roll_opening")).toHaveLength(1);
  });

  it("reads 'You tied with N. Roll again.' in a reroll", async () => {
    const c = await mountGame(
      rollTable({
        rounds: [
          {
            seats: [0, 1, 2],
            rolls: [
              { seat: 0, result: 17 },
              { seat: 1, result: 17 },
              { seat: 2, result: 4 },
            ],
          },
          { seats: [0, 1], rolls: [] },
        ],
      }),
    );
    const dlg = dialogNamed(dockOf(c)!, "roll for the first turn")!;
    expect(dlg.textContent).toContain("You tied with 17. Roll again.");
  });

  it("gives the host 'Roll for everyone left' as a secondary that Enter never presses", async () => {
    const c = await mountGame(rollTable(ROUND_ONE, { host: true }));
    const dlg = dialogNamed(dockOf(c)!, "roll for the first turn")!;
    const secondaries = [...dlg.querySelectorAll<HTMLButtonElement>(".dock-bar .secondary")];
    expect(secondaries.map(nameOf)).toEqual(["Roll for everyone left"]);
    expect(secondaries[0].getAttribute("aria-keyshortcuts")).toBeNull();
    expect(nameOf(dlg.querySelector(".dock-bar .request-primary")!)).toBe("Roll");

    blur();
    keydown("Enter");
    expect(FakeSocket.last!.actions("host_roll_remaining")).toHaveLength(0);
    expect(FakeSocket.last!.actions("roll_opening")).toHaveLength(1);

    // A new frame (the host's own die landed) and the button sends.
    snapshot(
      rollTable(
        { rounds: [{ seats: [0, 1, 2], rolls: [{ seat: 0, result: 9 }] }] },
        { host: true },
      ),
    );
    const wait = dialogNamed(dockOf(c)!, "roll for the first turn")!;
    expect(wait.textContent).toContain("Waiting for Opp and Third to roll");
    expect(wait.querySelector(".dock-bar .request-primary")).toBeNull();
    click(buttonNamed(wait, "Roll for everyone left")!);
    const sent = FakeSocket.last!.actions("host_roll_remaining");
    expect(sent).toHaveLength(1);
    expect(sent[0].payload?.player).toBeUndefined();
  });

  it("leaves a seat that has rolled a status line, with next, Pass turn and the toggles disabled", async () => {
    const c = await mountGame(
      rollTable({ rounds: [{ seats: [0, 1, 2], rolls: [{ seat: 0, result: 12 }] }] }),
    );
    const dock = dockOf(c)!;
    expect(dock.querySelector('[role="dialog"]')).toBeNull();
    expect(dock.querySelector(".dock-status")!.textContent).toContain(
      "Waiting for Opp and Third to roll",
    );
    expect(dock.querySelector<HTMLButtonElement>(".dock-btn.next")!.disabled).toBe(true);
    expect(dock.querySelector<HTMLButtonElement>(".dock-btn.pass-turn")!.disabled).toBe(true);
    const toggles = dock.querySelector('[role="group"][aria-label="priority controls"]')!;
    for (const name of ["hold", "autopass", "Undo (1 left)", "bluff"]) {
      expect(buttonNamed(toggles, name)?.disabled, name).toBe(true);
    }
  });
});

describe("choose who takes the first turn", () => {
  const chosen: OpeningRollView = {
    rounds: [
      {
        seats: [0, 1, 2],
        rolls: [
          { seat: 1, result: 8 },
          { seat: 0, result: 20 },
          { seat: 2, result: 3 },
        ],
      },
    ],
    chooser: 0,
  };

  it("opens a sheet with a button per seat in turn order and no primary", async () => {
    const c = await mountGame(rollTable(chosen));
    const dlg = dialogNamed(dockOf(c)!, "choose who takes the first turn")!;
    expect(dlg).not.toBeNull();
    expect(dlg.querySelector(".dock-sheet")).not.toBeNull();
    expect(dlg.textContent).toContain("You won the roll with 20. Choose who takes the first turn.");
    const names = [...dlg.querySelectorAll<HTMLButtonElement>(".sheet-body button.choice")].map(
      nameOf,
    );
    expect(names).toEqual(["I go first", "Opp goes first", "Third goes first"]);
    expect(dlg.querySelector(".dock-bar .request-primary")).toBeNull();

    blur();
    keydown("Enter");
    expect(FakeSocket.last!.actions("choose_starting_player")).toHaveLength(0);
  });

  it("sends 'I go first' at once", async () => {
    const c = await mountGame(rollTable(chosen));
    click(buttonNamed(c, "I go first")!);
    const sent = FakeSocket.last!.actions("choose_starting_player");
    expect(sent).toHaveLength(1);
    expect(sent[0].payload?.player).toBe(ME);
    expect(sent[0].payload?.params).toEqual({ seat: 0 });
  });

  it("asks before giving the first turn away, and Confirm takes no Enter", async () => {
    const c = await mountGame(rollTable(chosen));
    click(buttonNamed(c, "Opp goes first")!);
    expect(FakeSocket.last!.actions("choose_starting_player")).toHaveLength(0);
    expect(dialogNamed(c, "choose who takes the first turn")).toBeNull();
    const confirm = dialogNamed(dockOf(c)!, "Let Opp take the first turn?")!;
    expect(confirm).not.toBeNull();
    const primary = confirm.querySelector<HTMLButtonElement>(".dock-bar .request-primary")!;
    expect(nameOf(primary)).toBe("Confirm");
    expect(primary.getAttribute("aria-keyshortcuts")).toBeNull();
    expect(
      [...confirm.querySelectorAll<HTMLButtonElement>(".dock-bar .secondary")].map(nameOf),
    ).toEqual(["Cancel"]);

    keydown("Enter", confirm);
    keydown("Enter");
    expect(FakeSocket.last!.actions("choose_starting_player")).toHaveLength(0);

    click(primary);
    const sent = FakeSocket.last!.actions("choose_starting_player");
    expect(sent).toHaveLength(1);
    expect(sent[0].payload?.params).toEqual({ seat: 1 });
  });

  it("goes back to the sheet on Cancel and on Escape", async () => {
    const c = await mountGame(rollTable(chosen));
    click(buttonNamed(c, "Third goes first")!);
    click(buttonNamed(dialogNamed(c, "Let Third take the first turn?")!, "Cancel")!);
    expect(dialogNamed(c, "Let Third take the first turn?")).toBeNull();
    expect(dialogNamed(dockOf(c)!, "choose who takes the first turn")).not.toBeNull();

    click(buttonNamed(c, "Opp goes first")!);
    expect(dialogNamed(c, "Let Opp take the first turn?")).not.toBeNull();
    blur();
    keydown("Escape");
    expect(dialogNamed(c, "Let Opp take the first turn?")).toBeNull();
    expect(dialogNamed(dockOf(c)!, "choose who takes the first turn")).not.toBeNull();
    expect(FakeSocket.last!.actions("choose_starting_player")).toHaveLength(0);
  });

  it("tells everyone else who is choosing, in the dock's status line", async () => {
    const c = await mountGame(rollTable({ ...chosen, chooser: 1 }));
    const dock = dockOf(c)!;
    expect(dock.querySelector('[role="dialog"]')).toBeNull();
    expect(dock.querySelector(".dock-status")!.textContent).toContain(
      "Opp is choosing who goes first",
    );
  });
});

describe("the strip during and after the opening roll", () => {
  it("shows 'opening roll' with a chip per seat instead of 'opening hand decisions'", async () => {
    const c = await mountGame(
      rollTable({
        rounds: [
          {
            seats: [0, 1, 2],
            rolls: [
              { seat: 0, result: 15 },
              { seat: 1, result: 15 },
              { seat: 2, result: 6 },
            ],
          },
          { seats: [0, 1], rolls: [{ seat: 1, result: 11 }] },
        ],
      }),
    );
    const strip = stripOf(c);
    expect(strip.querySelector('[aria-label="opening hand decisions"]')).toBeNull();
    const banner = strip.querySelector<HTMLElement>('[role="group"][aria-label="opening roll"]')!;
    expect(banner).not.toBeNull();
    const chips = [...banner.querySelectorAll<HTMLElement>(".chip")];
    expect(chips.map((ch) => ch.querySelector(".result")!.textContent)).toEqual([
      "rolling…",
      "11",
      "—",
    ]);
    // The round's tied leaders are outlined.
    expect(chips.map((ch) => ch.classList.contains("tied"))).toEqual([true, true, false]);
    // The roll's own dice raise no strip cue (only the banner shows them).
    expect(strip.querySelector(".random-line")).toBeNull();
  });

  it("holds a chip's result, and the chooser's mark, until its die settles", async () => {
    const queue = new DiceQueue();
    const die: LogEvent = {
      seq: 7,
      kind: "roll",
      seat: 1,
      text: "Opp rolled a d20: 13",
      sides: 20,
      results: [13],
    };
    const before = rollTable({ rounds: [{ seats: [0, 1], rolls: [{ seat: 0, result: 2 }] }] });
    const after = rollTable(
      {
        rounds: [
          {
            seats: [0, 1],
            rolls: [
              { seat: 0, result: 2 },
              { seat: 1, result: 13 },
            ],
          },
        ],
        chooser: 1,
      },
      { log: [die] },
    );
    queue.prime([]);
    const r = render(OpeningRollBanner as never, { view: before, dice: queue } as never);
    const chip = () => r.container.querySelectorAll<HTMLElement>(".chip")[1];
    expect(chip().querySelector(".result")!.textContent).toBe("rolling…");

    const t0 = Date.now();
    queue.ingest(rollsFromLogs(after.log), t0, { motion: true, speed: 1 });
    r.setProps({ view: after } as never);
    expect(chip().querySelector(".result")!.textContent).toBe("rolling…");
    expect(chip().classList.contains("chooser")).toBe(false);

    await new Promise((res) => setTimeout(res, DICE_TUMBLE_MS + 50));
    flushSync();
    expect(chip().querySelector(".result")!.textContent).toBe("13");
    expect(chip().classList.contains("chooser")).toBe(true);
    queue.dispose();
  });

  it("marks the chooser once known", async () => {
    const c = await mountGame(
      rollTable({ rounds: [{ seats: [0, 1, 2], rolls: [{ seat: 2, result: 19 }] }], chooser: 2 }),
    );
    const chips = [...stripOf(c).querySelectorAll<HTMLElement>(".chip")];
    expect(chips.map((ch) => ch.classList.contains("chooser"))).toEqual([false, false, true]);
    expect(chips[2].textContent).toContain("chooses");
  });

  it("goes back to 'opening hand decisions' after the choice, its pill read from starting_player", async () => {
    const log: LogEvent[] = [
      { seq: 1, kind: "roll", seat: 0, text: "Me rolled a d20: 4", sides: 20, results: [4] },
      { seq: 2, kind: "roll", seat: 1, text: "Opp rolled a d20: 9", sides: 20, results: [9] },
      { seq: 3, kind: "roll", seat: 2, text: "Third rolled a d20: 20", sides: 20, results: [20] },
      {
        seq: 4,
        kind: "opening_roll",
        seat: 2,
        label: "won",
        results: [20],
        text: "Third won the opening roll with 20",
      },
      {
        seq: 5,
        kind: "starting_player",
        seat: 2,
        target_seat: 1,
        text: "Third chose Opp to take the first turn",
      },
    ];
    const c = await mountGame(rollTable(undefined, { log }));
    const strip = stripOf(c);
    expect(strip.querySelector('[aria-label="opening roll"]')).toBeNull();
    const banner = strip.querySelector('[aria-label="opening hand decisions"]')!;
    expect(banner).not.toBeNull();
    expect(banner.querySelector(".opening-roll")!.textContent).toContain(
      "Third won the d20 roll with 20 and chose Opp to go first",
    );
  });
});

describe("spectators", () => {
  it("see the banner and no dock", async () => {
    session.set(spectatorSession());
    const c = await mountGame(rollTable(ROUND_ONE));
    expect(dockOf(c)).toBeNull();
    expect(dialogNamed(c, "roll for the first turn")).toBeNull();
    expect(stripOf(c).querySelector('[aria-label="opening roll"]')).not.toBeNull();
  });
});
