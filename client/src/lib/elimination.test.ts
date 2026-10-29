// @vitest-environment jsdom
// #4449 — the elimination cue's rules: which seats a frame says left, the
// burst's pieces, and the commander a tombstone names.

import { beforeEach, describe, expect, it } from "vitest";

import {
  SPARK_COUNT,
  burstShards,
  burstSparks,
  commanderLine,
  commanderNames,
  eliminatedSeats,
  rememberCommanders,
  resetCommanderMemory,
  tombstoneCommanders,
} from "./elimination";
import type { CardView, GameView, LogEvent, PlayerView } from "./protocol";

const out = (seq: number, seat: number, cause = "life"): LogEvent =>
  ({ seq, kind: "eliminated", seat, cause, text: "" }) as LogEvent;

describe("eliminatedSeats", () => {
  it("names each seat that left, once, in log order", () => {
    expect(eliminatedSeats([out(1, 2), out(2, 0), out(3, 2)])).toEqual([2, 0]);
  });

  it("counts a concession as leaving the game", () => {
    expect(eliminatedSeats([out(1, 3, "concede")])).toEqual([3]);
  });

  it("ignores every other kind of entry and a seat that is not one", () => {
    const other = { seq: 1, kind: "damage", seat: 1, text: "" } as LogEvent;
    const over = { seq: 2, kind: "game_over", seat: 0, text: "" } as LogEvent;
    expect(eliminatedSeats([other, over, out(3, -1)])).toEqual([]);
    expect(eliminatedSeats([])).toEqual([]);
  });
});

describe("the burst", () => {
  it("cuts the avatar into eight wedges that fly apart", () => {
    const shards = burstShards();
    expect(shards).toHaveLength(8);
    expect(new Set(shards.map((s) => s.clip)).size).toBe(8);
    for (const s of shards) expect(Math.hypot(s.dx, s.dy)).toBeGreaterThan(0.5);
  });

  it("flings the same sparks for the same seat, and others for another", () => {
    expect(burstSparks("p2")).toEqual(burstSparks("p2"));
    expect(burstSparks("p2")).not.toEqual(burstSparks("p3"));
    expect(burstSparks("p2")).toHaveLength(SPARK_COUNT);
  });

  it("alternates ember and gold sparks", () => {
    const sparks = burstSparks("p2");
    expect(sparks.filter((s) => s.ember)).toHaveLength(SPARK_COUNT / 2);
    for (const s of sparks) expect(s.size).toBeGreaterThan(0);
  });
});

const card = (name: string, extra: Partial<CardView> = {}): CardView =>
  ({ instance_id: name, name, owner: "p1", controller: "p1", ...extra }) as unknown as CardView;

const zone = (cards: CardView[] = []) => ({ kind: "z", count: cards.length, cards });

const seat = (over: Partial<PlayerView> = {}): PlayerView =>
  ({
    id: "p1",
    name: "Pia",
    seat: 1,
    life: 40,
    graveyard: zone(),
    command: zone(),
    ...over,
  }) as unknown as PlayerView;

const gameView = (over: Partial<GameView> = {}): GameView =>
  ({
    id: "g1",
    battlefield: zone(),
    stack: zone(),
    exile: zone(),
    seats: [],
    ...over,
  }) as unknown as GameView;

describe("commanders on a tombstone", () => {
  beforeEach(() => resetCommanderMemory());

  it("reads the commander from the command zone", () => {
    const s = seat({ command: zone([card("Atraxa, Praetors' Voice", { is_commander: true })]) });
    expect(commanderNames(gameView(), s)).toEqual(["Atraxa, Praetors' Voice"]);
  });

  it("leaves a dungeon in the command zone out when no card is flagged", () => {
    const s = seat({
      command: zone([
        card("Tomb of Annihilation", { type_line: "Dungeon" }),
        card("Edea, Possessed Sorceress"),
      ]),
    });
    expect(commanderNames(gameView(), s)).toEqual(["Edea, Possessed Sorceress"]);
  });

  it("finds a commander out on the battlefield, and each partner", () => {
    const view = gameView({
      battlefield: zone([card("Kraum", { is_commander: true })]),
    });
    const s = seat({ command: zone([card("Tymna", { is_commander: true })]) });
    expect(commanderNames(view, s)).toEqual(["Tymna", "Kraum"]);
    expect(commanderLine(["Tymna", "Kraum"])).toBe("Tymna & Kraum");
  });

  it("does not take another seat's commander", () => {
    const view = gameView({
      battlefield: zone([card("Kraum", { is_commander: true, owner: "p2", controller: "p2" })]),
    });
    expect(commanderNames(view, seat())).toEqual([]);
  });

  it("remembers the commander after it has left with the player", () => {
    const here = gameView();
    rememberCommanders(here, seat({ command: zone([card("Atraxa", { is_commander: true })]) }));
    // The frame that eliminates the seat shows no commander.
    expect(tombstoneCommanders(here, seat({ eliminated: true }))).toEqual(["Atraxa"]);
    // And a frame that shows none does not make it forget.
    rememberCommanders(here, seat({ eliminated: true }));
    expect(tombstoneCommanders(here, seat({ eliminated: true }))).toEqual(["Atraxa"]);
  });

  it("keeps each game and each seat apart", () => {
    rememberCommanders(
      gameView(),
      seat({ command: zone([card("Atraxa", { is_commander: true })]) }),
    );
    expect(tombstoneCommanders(gameView({ id: "g2" }), seat())).toEqual([]);
    expect(tombstoneCommanders(gameView(), seat({ id: "p2" }))).toEqual([]);
  });

  it("prefers what the frame shows to what it remembered", () => {
    rememberCommanders(
      gameView(),
      seat({ command: zone([card("Atraxa", { is_commander: true })]) }),
    );
    const s = seat({ command: zone([card("Kenrith", { is_commander: true })]) });
    expect(tombstoneCommanders(gameView(), s)).toEqual(["Kenrith"]);
  });

  it("survives a reload of the page", () => {
    rememberCommanders(
      gameView(),
      seat({ command: zone([card("Atraxa", { is_commander: true })]) }),
    );
    // A reload keeps sessionStorage but not the module's memory.
    const kept = sessionStorage.getItem("cmdctrl.commanders");
    expect(kept).toContain("Atraxa");
    resetCommanderMemory();
    sessionStorage.setItem("cmdctrl.commanders", kept ?? "");
    expect(tombstoneCommanders(gameView(), seat())).toEqual(["Atraxa"]);
  });
});
