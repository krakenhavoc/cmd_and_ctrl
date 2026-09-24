// #1533 — "attack with all" under a CR 508.1c count limit (Silent
// Arbiter, Crawlspace; ADR 0045 Decisions 44-46). The pure half of the
// flow: reading the server's room off the view, the picker's cap rules,
// and which refusal belongs to the bulk swing. The markup half is
// attackLimitPicker.render.test.ts.

import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { get } from "svelte/store";

import {
  attackLimitBinds,
  attackLimitOn,
  attackLimitSentence,
  bulkAttackRefusal,
  offersAttackPicker,
  planAttackAll,
  seedAttackSelection,
  toggleAttackSelection,
} from "./attackAll";
import { PROTOCOL_VERSION } from "./protocol";
import type { AttackTargetView, CardView, GameView, PlayerView, ZoneView } from "./protocol";
import { GameClient } from "./ws";

function zone(kind: string, owner: string | undefined, cards: CardView[]): ZoneView {
  return { kind, owner, count: cards.length, cards };
}

function creature(id: string, controller: string): CardView {
  return { instance_id: id, name: id, owner: controller, controller, type_line: "Creature — Test" };
}

function seat(id: string, name: string, n: number): PlayerView {
  return {
    id,
    name,
    seat: n,
    life: 40,
    library: zone("library", id, []),
    hand: zone("hand", id, []),
    graveyard: zone("graveyard", id, []),
    command: zone("command", id, []),
    commander_damage: {},
    life_history: [],
  };
}

// limitedView: Alice (the viewer) with `n` idle creatures, Bob and
// Cara to attack, and the server's attack_targets rows as given.
function limitedView(n: number, rows: AttackTargetView[]): GameView {
  const bf = Array.from({ length: n }, (_, i) => creature(`bear${i + 1}`, "a"));
  return {
    id: "g1",
    state: "active",
    seats: [seat("a", "Alice", 0), seat("b", "Bob", 1), seat("c", "Cara", 2)],
    battlefield: zone("battlefield", undefined, bf),
    stack: zone("stack", undefined, []),
    exile: zone("exile", undefined, []),
    turn: {
      seq: 1,
      number: 1,
      active_seat: 0,
      priority_holder: 0,
      phase: "combat",
      step: "declare_attackers",
      attack_targets: rows,
    },
    mulligans_open: false,
  };
}

describe("attackLimitOn — the server's room, read not derived", () => {
  it("reads attack_limit for the named seat", () => {
    const v = limitedView(3, [
      { kind: "player", id: "b", attack_limit: 2 },
      { kind: "player", id: "c" },
    ]);
    expect(attackLimitOn(v, "b")).toBe(2);
    expect(attackLimitOn(v, "c")).toBeNull();
  });

  it("keeps 0: a used-up limit is not 'no limit'", () => {
    const v = limitedView(3, [{ kind: "player", id: "b", attack_limit: 0 }]);
    expect(attackLimitOn(v, "b")).toBe(0);
  });

  it("reads only player rows, and nothing from a view without the field", () => {
    // A planeswalker row cannot stand in for its controller's seat.
    const v = limitedView(3, [{ kind: "planeswalker", id: "b", attack_limit: 1 }]);
    expect(attackLimitOn(v, "b")).toBeNull();
    expect(attackLimitOn(limitedView(3, []), "b")).toBeNull();
    expect(attackLimitOn(null, "b")).toBeNull();
  });
});

describe("attackLimitBinds / offersAttackPicker — when the picker is offered up front", () => {
  it("binds only when the room is smaller than the eligible set", () => {
    const v = limitedView(3, [
      { kind: "player", id: "b", attack_limit: 2 },
      { kind: "player", id: "c", attack_limit: 3 },
    ]);
    const plan = planAttackAll(v, "a");
    expect(attackLimitBinds(v, plan, "b")).toBe(true);
    expect(attackLimitBinds(v, plan, "c")).toBe(false);
    expect(offersAttackPicker(v, plan, "b")).toBe(true);
    expect(offersAttackPicker(v, plan, "c")).toBe(false);
  });

  it("offers nothing to pick under a used-up limit", () => {
    const v = limitedView(3, [{ kind: "player", id: "b", attack_limit: 0 }]);
    const plan = planAttackAll(v, "a");
    expect(attackLimitBinds(v, plan, "b")).toBe(true);
    expect(offersAttackPicker(v, plan, "b")).toBe(false);
  });

  it("still offers it for a tax alone (#1162)", () => {
    const v = limitedView(3, [{ kind: "player", id: "b", tax: "{2}" }]);
    expect(offersAttackPicker(v, planAttackAll(v, "a"), "b")).toBe(true);
  });
});

describe("seedAttackSelection / toggleAttackSelection — the cap", () => {
  const eligible = ["x", "y", "z"].map((id) => creature(id, "a"));

  it("opens at every eligible creature without a limit, and the first N under one", () => {
    expect(seedAttackSelection(eligible, null)).toEqual(["x", "y", "z"]);
    expect(seedAttackSelection(eligible, 2)).toEqual(["x", "y"]);
    expect(seedAttackSelection(eligible, 0)).toEqual([]);
    expect(seedAttackSelection(eligible, 9)).toEqual(["x", "y", "z"]);
  });

  it("refuses to check one past the cap, and always allows unchecking", () => {
    expect(toggleAttackSelection(["x"], "y", 1)).toEqual(["x"]);
    expect(toggleAttackSelection(["x"], "x", 1)).toEqual([]);
    expect(toggleAttackSelection([], "y", 1)).toEqual(["y"]);
    expect(toggleAttackSelection(["x", "y"], "z", null)).toEqual(["x", "y", "z"]);
  });
});

describe("attackLimitSentence — the picker's reason without a refusal", () => {
  it("names the number and the seat", () => {
    expect(attackLimitSentence(1, "Bob")).toBe("Only 1 more creature can attack Bob this combat.");
    expect(attackLimitSentence(2, "Bob")).toBe("Only 2 more creatures can attack Bob this combat.");
    expect(attackLimitSentence(0, "Bob")).toBe("No more creatures can attack Bob this combat.");
  });
});

describe("bulkAttackRefusal — only the bulk swing's own refusal opens a picker", () => {
  const attempt = { defenderSeatID: "b", frameID: "f-1" };
  const limit = { code: "illegal_attack", reason: "attack_limit", replyTo: "f-1" };

  it("claims an attack_limit refusal of that frame", () => {
    expect(bulkAttackRefusal(limit, attempt)).toEqual({ kind: "limit", defenderSeatID: "b" });
  });

  it("claims an attack-tax refusal of that frame (#1162, unchanged)", () => {
    expect(
      bulkAttackRefusal({ code: "attack_tax_unpaid", reason: "{2}{2}", replyTo: "f-1" }, attempt),
    ).toEqual({ kind: "tax", defenderSeatID: "b" });
  });

  it("leaves a refusal of another action to the generic toast", () => {
    // A single click-declared attacker refused by Silent Arbiter after
    // an earlier bulk swing: same code, different frame.
    expect(bulkAttackRefusal({ ...limit, replyTo: "f-2" }, attempt)).toBeNull();
    expect(bulkAttackRefusal({ ...limit, replyTo: undefined }, attempt)).toBeNull();
    expect(bulkAttackRefusal(limit, null)).toBeNull();
    expect(bulkAttackRefusal(null, attempt)).toBeNull();
  });

  it("does not claim an illegal_attack with another reason", () => {
    expect(bulkAttackRefusal({ ...limit, reason: "something_else" }, attempt)).toBeNull();
    expect(bulkAttackRefusal({ code: "bad_request", replyTo: "f-1" }, attempt)).toBeNull();
  });
});

// The correlation above rests on the ws client stamping the refused
// frame's id onto lastError. Driven through a real GameClient on a fake
// socket, the way ws.redact.test.ts does.
class FakeSocket {
  static readonly OPEN = 1;
  static last: FakeSocket | null = null;
  readyState = FakeSocket.OPEN;
  private listeners: Record<string, ((ev: unknown) => void)[]> = {};
  constructor(readonly url: string) {
    FakeSocket.last = this;
  }
  addEventListener(type: string, fn: (ev: unknown) => void): void {
    (this.listeners[type] ??= []).push(fn);
  }
  removeEventListener(): void {}
  send(): void {}
  close(): void {}
  emit(type: string, ev: unknown): void {
    for (const fn of this.listeners[type] ?? []) fn(ev);
  }
}

describe("GameClient.lastError.replyTo — the answered frame's id", () => {
  let realWebSocket: unknown;
  beforeEach(() => {
    realWebSocket = (globalThis as Record<string, unknown>).WebSocket;
    (globalThis as Record<string, unknown>).WebSocket = FakeSocket;
  });
  afterEach(() => {
    (globalThis as Record<string, unknown>).WebSocket = realWebSocket;
  });

  it("carries the id the refused action was sent with", () => {
    const client = new GameClient("ws://test/ws?game=g1");
    client.connect();
    const socket = FakeSocket.last!;
    socket.emit("open", {});
    const frameID = client.sendAction("declare_attackers", undefined, { attackers: [] });
    expect(frameID).toBeTruthy();
    socket.emit("message", {
      data: JSON.stringify({
        v: PROTOCOL_VERSION,
        kind: "error",
        id: frameID,
        payload: {
          code: "illegal_attack",
          reason: "attack_limit",
          message: "No more than one creature can attack each combat (Silent Arbiter).",
        },
      }),
    });
    const err = get(client.lastError);
    expect(err?.replyTo).toBe(frameID);
    expect(bulkAttackRefusal(err, { defenderSeatID: "b", frameID: frameID! })).toEqual({
      kind: "limit",
      defenderSeatID: "b",
    });
    client.disconnect();
  });
});
