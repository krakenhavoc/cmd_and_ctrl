import { describe, expect, it } from "vitest";

import type { GameMeta, SeatInfo } from "./api";
import { L } from "./labels";
import {
  NO_MEMORY,
  READY_MOMENT_MS,
  READY_PRIMARY_DELAY_MS,
  READY_STAGGER_MS,
  observeReady,
  readyAnnouncement,
  readyDelayMs,
  readyMotion,
  type ReadyMemory,
  type ReadyMoment,
} from "./readyMoment";
import { allDecksIn, lobbyPrimary, readiness, type LobbyViewer } from "./tablesPage";

function seat(n: number, extra: Partial<SeatInfo> = {}): SeatInfo {
  return { player_id: `p${n}`, name: `P${n}`, seat: n, deck_uploaded: true, ...extra };
}

const choosing = (n: number): SeatInfo => seat(n, { deck_uploaded: false });
const bot = (n: number): SeatInfo =>
  seat(n, { player_id: `bot${n}`, is_bot: true, deck_uploaded: false });

function table(players: SeatInfo[], extra: Partial<GameMeta> = {}): GameMeta {
  return {
    id: "g1",
    name: "Friday pod",
    created_at: "2026-10-30T12:00:00Z",
    players,
    state: "lobby",
    ...extra,
  };
}

// A client's life as a run of reads: the moments each read produced.
function reads(...tables: GameMeta[][]): ReadyMoment[][] {
  let memory: ReadyMemory = NO_MEMORY;
  return tables.map((list) => {
    const seen = observeReady(memory, list);
    memory = seen.memory;
    return seen.moments;
  });
}

describe("allDecksIn (the state the moment fires on entering)", () => {
  it("needs two seated and every one ready, in the lobby", () => {
    expect(allDecksIn(table([seat(0), seat(1)]))).toBe(true);
    expect(allDecksIn(table([seat(0), bot(1)]))).toBe(true); // a bot arrives with its deck
    expect(allDecksIn(table([seat(0), choosing(1)]))).toBe(false);
    expect(allDecksIn(table([seat(0)]))).toBe(false);
    expect(allDecksIn(table([]))).toBe(false);
    expect(allDecksIn(table([seat(0), seat(1)], { state: "active" }))).toBe(false);
    expect(allDecksIn(table([seat(0), seat(1)], { state: "ended" }))).toBe(false);
  });

  it("agrees with the readiness line and the primary on every table", () => {
    const viewer = (me: SeatInfo | null): LobbyViewer => ({
      me,
      heldElsewhere: false,
      takesSeat: false,
      mayStart: true,
      spectator: false,
    });
    const lists: SeatInfo[][] = [
      [],
      [seat(0)],
      [choosing(0)],
      [seat(0), seat(1)],
      [seat(0), choosing(1)],
      [choosing(0), choosing(1)],
      [seat(0), seat(1), seat(2)],
      [seat(0), seat(1), choosing(2)],
      [seat(0), seat(1), bot(2)],
      [seat(0), seat(1), seat(2), seat(3)],
    ];
    for (const players of lists) {
      const g = table(players);
      const me = players[0] ?? null;
      const line = readiness(g, me, { bots: true });
      const primary = lobbyPrimary(g, viewer(me));
      expect(line.status.toLowerCase() === "every deck is in").toBe(allDecksIn(g));
      expect(primary.kind === "start").toBe(allDecksIn(g) && me !== null);
    }
  });
});

describe("observeReady: when the moment fires", () => {
  it("fires once, on the read where the last deck lands", () => {
    const out = reads([table([seat(0), choosing(1)])], [table([seat(0), seat(1)])]);
    expect(out[0]).toEqual([]);
    expect(out[1]).toEqual([{ gameID: "g1", landed: 1 }]);
  });

  it("says nothing on the first read of a table that is already ready", () => {
    const out = reads([table([seat(0), seat(1)])]);
    expect(out[0]).toEqual([]);
  });

  it("says nothing on a page's first read of every table, however many are ready", () => {
    const ready = table([seat(0), seat(1)], { id: "a" });
    const also = table([seat(0), seat(1), seat(2)], { id: "b" });
    expect(reads([ready, also])[0]).toEqual([]);
  });

  it("does not fire again on the polls that follow", () => {
    const out = reads(
      [table([seat(0), choosing(1)])],
      [table([seat(0), seat(1)])],
      [table([seat(0), seat(1)])],
      [table([seat(0), seat(1)])],
      [table([seat(0), seat(1)])],
    );
    expect(out.map((m) => m.length)).toEqual([0, 1, 0, 0, 0]);
  });

  it("does not fire when the same table is seen again after a patch from the sit-down answer", () => {
    // The sit-down answer patches the seats, then the list is re-read:
    // two writes of the same ready table.
    const before = table([seat(0), choosing(1)]);
    const after = table([seat(0), seat(1)]);
    expect(reads([before], [after], [after]).map((m) => m.length)).toEqual([0, 1, 0]);
  });

  it("does not fire when a seat leaves or joins ready while the table stays startable", () => {
    const out = reads(
      [table([seat(0), seat(1), seat(2)])],
      [table([seat(0), seat(1)])], // the third seat left
      [table([seat(0), seat(1), seat(2)])], // and another sat down ready (a bot, say)
      [table([seat(0), seat(1), seat(2), seat(3)])],
    );
    expect(out.map((m) => m.length)).toEqual([0, 0, 0, 0]);
  });

  it("fires again when the table leaves all-ready and comes back", () => {
    const out = reads(
      [table([seat(0), seat(1)])],
      [table([seat(0), seat(1), choosing(2)])], // someone sat down with no deck yet
      [table([seat(0), seat(1), seat(2)])], // their deck landed
    );
    expect(out.map((m) => m.length)).toEqual([0, 0, 1]);
    expect(out[2][0]).toEqual({ gameID: "g1", landed: 2 });
  });

  it("fires again when the table drops below two seated and gets a second back", () => {
    const out = reads(
      [table([seat(0), seat(1)])],
      [table([seat(0)])],
      [table([seat(0), seat(1)])], // the seat came back, deck and all
    );
    expect(out.map((m) => m.length)).toEqual([0, 0, 1]);
    expect(out[2][0].landed).toBe(1);
  });

  it("does not fire for a seat that drops and re-readies between two reads", () => {
    // Slysloth's deck is swapped, or they leave and re-sit with a deck,
    // inside one 5 s window: both reads are all-ready, so nothing changed
    // about whether the table can start.
    const out = reads(
      [table([seat(0), seat(1)])],
      [table([seat(0), seat(1, { player_id: "p9" })])],
    );
    expect(out.map((m) => m.length)).toEqual([0, 0]);
  });

  it("fires, with no seat to ring, when the one still choosing walks away", () => {
    const out = reads([table([seat(0), seat(1), choosing(2)])], [table([seat(0), seat(1)])]);
    expect(out[1]).toEqual([{ gameID: "g1", landed: null }]);
  });

  it("fires when a bot joins the only ready player, and rings the bot's seat", () => {
    const out = reads([table([seat(0)])], [table([seat(0), bot(1)])]);
    expect(out[1]).toEqual([{ gameID: "g1", landed: 1 }]);
  });

  it("rings the highest-numbered seat when two decks land in one read", () => {
    const out = reads(
      [table([seat(0), choosing(1), choosing(2)])],
      [table([seat(0), seat(1), seat(2)])],
    );
    expect(out[1]).toEqual([{ gameID: "g1", landed: 2 }]);
  });

  it("follows each table on its own", () => {
    const a = (players: SeatInfo[]) => table(players, { id: "a" });
    const b = (players: SeatInfo[]) => table(players, { id: "b" });
    const out = reads(
      [a([seat(0), choosing(1)]), b([seat(0), choosing(1)])],
      [a([seat(0), seat(1)]), b([seat(0), choosing(1)])],
      [a([seat(0), seat(1)]), b([seat(0), seat(1)])],
    );
    expect(out.map((m) => m.map((x) => x.gameID))).toEqual([[], ["a"], ["b"]]);
  });

  it("says nothing when the table starts, and nothing for a table in play", () => {
    const out = reads(
      [table([seat(0), choosing(1)])],
      [table([seat(0), seat(1)], { state: "active" })],
      [table([seat(0), seat(1)], { state: "active" })],
    );
    expect(out.map((m) => m.length)).toEqual([0, 0, 0]);
  });

  it("forgets a table that leaves the list, so its return is a first read", () => {
    const out = reads(
      [table([seat(0), choosing(1)])],
      [], // not listed for a read (the admin's filter, a failed rejoin lookup)
      [table([seat(0), seat(1)])],
    );
    expect(out.map((m) => m.length)).toEqual([0, 0, 0]);
  });

  it("keeps the memory it is given untouched", () => {
    const first = observeReady(NO_MEMORY, [table([seat(0), choosing(1)])]);
    const snapshot = [...first.memory.entries()];
    observeReady(first.memory, [table([seat(0), seat(1)])]);
    expect([...first.memory.entries()]).toEqual(snapshot);
  });
});

describe("readyMotion: the three signals of DESIGN.md §4", () => {
  it("plays only when none of them says reduce", () => {
    expect(readyMotion({ enabled: true, reduceMotion: false })).toBe("full");
    expect(readyMotion({ enabled: true, reduceMotion: false, osReduceMotion: false })).toBe("full");
  });

  it("is still when the animations switch is off", () => {
    expect(readyMotion({ enabled: false, reduceMotion: false })).toBe("still");
  });

  it("is still under the setting accessibility.reduceMotion", () => {
    expect(readyMotion({ enabled: true, reduceMotion: true })).toBe("still");
  });

  it("is still under the OS's prefers-reduced-motion", () => {
    expect(readyMotion({ enabled: true, reduceMotion: false, osReduceMotion: true })).toBe("still");
  });
});

describe("the moment's timing", () => {
  it("staggers the checks by seat and holds the ring for the last one", () => {
    expect(readyDelayMs(0)).toBe(0);
    expect(readyDelayMs(1)).toBe(READY_STAGGER_MS);
    expect(readyDelayMs(3)).toBe(3 * READY_STAGGER_MS);
    // a seat number out of range does not stretch the wave
    expect(readyDelayMs(9)).toBe(3 * READY_STAGGER_MS);
    expect(readyDelayMs(-1)).toBe(0);
    expect(READY_PRIMARY_DELAY_MS).toBeGreaterThan(readyDelayMs(3));
  });

  it("keeps the classes on until the last keyframe has played", () => {
    // wave + primary delay + the ring's 1.5 s
    expect(READY_MOMENT_MS).toBeGreaterThanOrEqual(READY_PRIMARY_DELAY_MS + 1500);
  });
});

describe("readyAnnouncement: the line the live region says once", () => {
  const g = table([seat(0, { is_host: true, name: "Luke" }), seat(1)]);

  it("tells whoever may start that they can", () => {
    expect(readyAnnouncement(g, { startable: true, named: false })).toBe(
      "Every deck is in. You can start the game.",
    );
  });

  it("tells everyone else who they wait for", () => {
    expect(readyAnnouncement(g, { startable: false, named: false })).toBe(
      "Every deck is in. Waiting for Luke to start.",
    );
  });

  it("names the table when the person is at more than one", () => {
    expect(readyAnnouncement(g, { startable: true, named: true })).toBe(
      "Every deck is in at Friday pod. You can start the game.",
    );
  });

  it("uses the host's shown name, and 'the host' when none is marked", () => {
    const shown = table([
      seat(0, { is_host: true, name: "luke_w", display_name: "Luke" }),
      seat(1),
    ]);
    expect(readyAnnouncement(shown, { startable: false, named: false })).toContain(
      "Waiting for Luke to",
    );
    const none = table([seat(0), seat(1)]);
    expect(readyAnnouncement(none, { startable: false, named: false })).toBe(
      "Every deck is in. Waiting for the host to start.",
    );
  });

  it("is the registered label's line", () => {
    expect(readyAnnouncement(g, { startable: true, named: false })).toBe(
      L.everyDeckIn(true, "Luke"),
    );
  });
});
