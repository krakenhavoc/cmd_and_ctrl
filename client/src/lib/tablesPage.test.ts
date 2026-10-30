import { describe, expect, it } from "vitest";

import type { GameMeta, SeatInfo } from "./api";
import type { MyGame } from "./myGames";
import {
  botChipText,
  contextLine,
  filterHistory,
  historyRows,
  lobbyPrimary,
  lobbySlots,
  moreCount,
  nowTables,
  readiness,
  seatLine,
  seatPreview,
  seatTileArt,
  seatTileText,
  statusChip,
  type LobbyViewer,
} from "./tablesPage";
import type { PickerCommit } from "./deckPicker";

const HOUR = 3_600_000;
const NOW = Date.parse("2026-10-13T12:00:00Z");

function seat(n: number, extra: Partial<SeatInfo> = {}): SeatInfo {
  return { player_id: `p${n}`, name: `P${n}`, seat: n, deck_uploaded: true, ...extra };
}

function meta(extra: Partial<GameMeta> = {}): GameMeta {
  return {
    id: "g",
    name: "T",
    created_at: new Date(NOW - 2 * HOUR).toISOString(),
    players: [],
    state: "lobby",
    ...extra,
  };
}

function mine(id: string, extra: Partial<MyGame> = {}): MyGame {
  return {
    id,
    name: id,
    state: "ended",
    seat: 0,
    winner_seat: null,
    created_at: NOW - HOUR,
    started_at: null,
    ended_at: null,
    archived_at: null,
    others: [],
    ...extra,
  };
}

describe("the Now card's head (ADR 0158 §1 item 2)", () => {
  it("names the status, with the turn once one has begun", () => {
    expect(statusChip(meta())).toEqual({ label: "Lobby", live: false });
    expect(statusChip(meta({ state: "active", turn: 7 })).label).toBe("In play · turn 7");
    expect(statusChip(meta({ state: "active" })).label).toBe("In play");
  });

  it("says who hosts, your seat, and how old a lobby table is", () => {
    const players = [seat(0, { is_host: true, display_name: "Slysloth" }), seat(1)];
    expect(contextLine(meta({ players }), { seat: 0 }, NOW)).toBe("you host · created 2h ago");
    expect(contextLine(meta({ players, state: "active" }), { seat: 1 }, NOW)).toBe(
      "Slysloth hosts · your seat 2",
    );
    expect(contextLine(meta({ is_creator: true }), { seat: null }, NOW)).toBe(
      "you created it · created 2h ago",
    );
  });

  it("gives a seat its deck in the lobby and its deck and life in play", () => {
    const name = (id?: string) => (id === "raid" ? "Raid" : "");
    expect(seatLine(meta(), seat(0, { deck_name: "Atraxa" }), name).text).toBe("Atraxa");
    expect(seatLine(meta(), seat(0, { deck_name: "Atraxa", deck_power: 3.5 }), name)).toEqual({
      text: "Atraxa",
      pending: false,
      power: 3.5,
    });
    expect(
      seatLine(
        meta({ state: "active" }),
        seat(0, { deck_name: "Atraxa", deck_power: 4, life: 31 }),
        name,
      ),
    ).toEqual({ text: "Atraxa · 31 life", pending: false, power: 4 });
    expect(seatLine(meta(), seat(0, { deck_uploaded: false }), name)).toEqual({
      text: "deck pending",
      pending: true,
      power: 0,
    });
    expect(
      seatLine(meta({ state: "active" }), seat(0, { deck_name: "Atraxa", life: 31 }), name).text,
    ).toBe("Atraxa · 31 life");
    expect(
      seatLine(meta(), seat(1, { is_bot: true, bot_tier: "heuristic", bot_deck: "raid" }), name)
        .text,
    ).toBe("heuristic · Raid");
    expect(
      seatLine(
        meta(),
        seat(2, { is_bot: true, bot_tier: "assisted", bot_deck: "raid", bot_own_key: true }),
        name,
      ).text,
    ).toBe("assisted (own key) · Raid");
  });

  it("keeps ended and archived tables out of Now", () => {
    const list = [
      meta({ id: "a" }),
      meta({ id: "b", state: "active" }),
      meta({ id: "c", state: "ended" }),
      meta({ id: "d", archived_at: "2026-10-01T00:00:00Z" }),
    ];
    expect(nowTables(list).map((g) => g.id)).toEqual(["a", "b"]);
  });
});

// ADR 0212: a table in the lobby as a character select.
describe("a lobby table's seat tiles (ADR 0212 §1)", () => {
  it("lays the four seats out by seat number, open ones as null", () => {
    const slots = lobbySlots(meta({ players: [seat(1), seat(0)] }));
    expect(slots.map((p) => p?.seat ?? null)).toEqual([0, 1, null, null]);
    // A seat out of range is kept, after the four.
    expect(lobbySlots(meta({ players: [seat(0), seat(6)] }))).toHaveLength(5);
  });

  it("names a person's deck once, until the seats carry commanders", () => {
    expect(seatTileText(seat(0, { deck_name: "Kenrith's court" }), undefined)).toEqual({
      what: "Kenrith's court",
      second: "",
      colors: [],
      ready: true,
    });
    expect(seatTileText(seat(0, { deck_uploaded: false }), undefined)).toMatchObject({
      what: "No deck yet",
      ready: false,
    });
  });

  it("gives a bot its pre-built deck's commander, the deck, and its colours", () => {
    const bot = seat(3, { is_bot: true, bot_tier: "heuristic", bot_deck: "elves" });
    const deck = {
      id: "elves",
      name: "Elf ball",
      commander: "Marwyn, the Nurturer",
      colors: ["g"],
    };
    expect(seatTileText(bot, deck)).toEqual({
      what: "Marwyn, the Nurturer",
      second: "Pre-built · Elf ball",
      colors: ["G"],
      ready: true,
    });
    // A bot is ready even before GET /bot/options has answered.
    expect(seatTileText({ ...bot, deck_name: "Elf ball" }, undefined)).toMatchObject({
      what: "Elf ball",
      ready: true,
    });
  });

  it("draws a bot's pre-built deck's commander art, and a person's seat none (phase B1)", () => {
    const bot = seat(3, { is_bot: true, bot_tier: "heuristic", bot_deck: "elves" });
    const deck = {
      id: "elves",
      name: "Elf ball",
      commander: "Marwyn, the Nurturer",
      colors: ["G"],
      commander_cards: [
        { name: "Marwyn, the Nurturer", scryfall_id: "m1", color_identity: ["G"] },
        { name: "No printing", color_identity: null },
      ],
    };
    // Only a commander with a printing is drawn; the rest of the seat is the wash.
    expect(seatTileArt(bot, deck)).toEqual([{ name: "Marwyn, the Nurturer", scryfallId: "m1" }]);
    expect(seatTileArt(bot, { ...deck, commander_cards: undefined })).toEqual([]);
    expect(seatTileArt(bot, undefined)).toEqual([]);
    // A person's seat takes no art from a deck; their preview supplies it (seatPreview).
    expect(seatTileArt(seat(0), deck)).toEqual([]);
  });

  it("names a bot's tier, and a player's own key", () => {
    const tiers = [{ tier: "heuristic", label: "Medium", description: "", available: true }];
    const bot = seat(3, { is_bot: true, bot_tier: "heuristic" });
    expect(botChipText(bot, tiers)).toBe("Medium");
    expect(botChipText({ ...bot, bot_own_key: true }, tiers)).toBe("Medium (own key)");
    expect(botChipText({ ...bot, bot_tier: "model" }, tiers)).toBe("model");
  });
});

describe("a lobby table's readiness line (ADR 0212 §2)", () => {
  it("counts the seated players, not four, and says what it waits on", () => {
    const me = seat(0, { deck_uploaded: false });
    const players = [me, seat(1), seat(3, { is_bot: true, deck_uploaded: false })].map((p, i) => ({
      ...p,
      seat: i,
    }));
    const r = readiness(meta({ players }), me, { bots: true });
    expect([r.ready, r.seated]).toEqual([2, 3]);
    expect(r.segments.map((s) => s.state)).toEqual(["choosing", "ready", "ready", "open"]);
    expect([r.lead, r.status]).toEqual(["1 seat open", "waiting on your deck"]);
  });

  it("names one player still choosing, or counts several", () => {
    const one = meta({ players: [seat(0), seat(1, { deck_uploaded: false, name: "Bob" })] });
    expect(readiness(one, seat(0), { bots: true }).status).toBe("waiting on Bob's deck");
    const two = meta({
      players: [seat(0), seat(1, { deck_uploaded: false }), seat(2, { deck_uploaded: false })],
    });
    expect(readiness(two, seat(0), { bots: true }).status).toBe("waiting on 2 decks");
  });

  it("says every deck is in, capitalised when nothing comes before it", () => {
    const full = meta({ players: [seat(0), seat(1), seat(2), seat(3)] });
    expect(readiness(full, seat(0), { bots: true })).toMatchObject({
      lead: "",
      status: "Every deck is in",
    });
  });

  it("asks for a second player below two, offering a bot only to who may add one", () => {
    const alone = meta({ players: [seat(0)] });
    expect(readiness(alone, seat(0), { bots: true }).status).toBe("invite someone or add a bot");
    expect(readiness(alone, seat(0), { bots: false }).status).toBe("invite someone");
  });
});

describe("a lobby table's one primary (ADR 0212 §2)", () => {
  const viewer = (v: Partial<LobbyViewer> = {}): LobbyViewer => ({
    me: null,
    heldElsewhere: false,
    takesSeat: false,
    mayStart: false,
    spectator: false,
    ...v,
  });
  const host = seat(0, { is_host: true, display_name: "Luke" });

  it("reclaims a seat held elsewhere first, then seats the creator", () => {
    const g = meta({ players: [host] });
    expect(lobbyPrimary(g, viewer({ heldElsewhere: true, takesSeat: true })).label).toBe(
      "Open table",
    );
    expect(lobbyPrimary(g, viewer({ takesSeat: true })).kind).toBe("take");
  });

  it("asks a seat without a deck to choose one", () => {
    const me = seat(1, { deck_uploaded: false });
    const p = lobbyPrimary(meta({ players: [host, me] }), viewer({ me }));
    expect(p).toMatchObject({ kind: "choose", label: "Choose your deck", disabled: false });
  });

  it("waits, disabled, on the players still choosing", () => {
    const g = meta({ players: [host, seat(1, { deck_uploaded: false }), seat(2)] });
    expect(lobbyPrimary(g, viewer({ me: host, mayStart: true }))).toMatchObject({
      kind: "wait",
      label: "Waiting for 1 player",
      disabled: true,
    });
    const two = meta({
      players: [host, seat(1, { deck_uploaded: false }), seat(2, { deck_uploaded: false })],
    });
    expect(lobbyPrimary(two, viewer({ me: host })).label).toBe("Waiting for 2 players");
  });

  it("waits for players below two", () => {
    const p = lobbyPrimary(meta({ players: [host] }), viewer({ me: host, mayStart: false }));
    expect(p).toMatchObject({ label: "Waiting for players", disabled: true });
    expect(p.title).toMatch(/invite someone or add a bot/);
  });

  it("offers Start game to who may start, naming the count while a seat is open", () => {
    const three = meta({ players: [host, seat(1), seat(2, { is_bot: true })] });
    expect(lobbyPrimary(three, viewer({ me: host, mayStart: true }))).toMatchObject({
      kind: "start",
      label: "Start game · 3 players",
      disabled: false,
    });
    const four = meta({ players: [host, seat(1), seat(2), seat(3)] });
    expect(lobbyPrimary(four, viewer({ me: host, mayStart: true })).label).toBe("Start game");
  });

  it("names the host to everyone else once every deck is in", () => {
    const me = seat(1);
    const g = meta({ players: [host, me] });
    expect(lobbyPrimary(g, viewer({ me }))).toMatchObject({
      label: "Waiting for Luke to start",
      disabled: true,
    });
  });

  it("gives the admin Start when the table can start, and Open table when not", () => {
    const ready = meta({ players: [host, seat(1)] });
    expect(lobbyPrimary(ready, viewer({ mayStart: true })).kind).toBe("start");
    const choosing = meta({ players: [host, seat(1, { deck_uploaded: false })] });
    expect(lobbyPrimary(choosing, viewer({ mayStart: true })).label).toBe("Open table");
    expect(lobbyPrimary(meta(), viewer()).label).toBe("Open table");
  });

  it("lets a spectator of the table watch", () => {
    expect(lobbyPrimary(meta({ players: [host] }), viewer({ spectator: true }))).toMatchObject({
      kind: "watch",
      label: "Watch",
    });
  });

  // Rule 4, and owner decision 8: the pickers under the seats commit
  // from the footer, so the card keeps one primary.
  const sit: PickerCommit = {
    label: "Sit down with this deck",
    disabled: false,
    title: "sit down with Elf ball",
    busy: false,
  };
  const emptyPaste: PickerCommit = {
    label: "Upload deck",
    disabled: true,
    title: "paste a decklist or a deck link first",
    busy: false,
    empty: true,
  };

  it("seats the picker's deck from the footer while the picker holds one", () => {
    const me = seat(1, { deck_uploaded: false });
    const g = meta({ players: [host, me] });
    expect(lobbyPrimary(g, viewer({ me, picking: sit }))).toMatchObject({
      kind: "sit",
      label: "Sit down with this deck",
      disabled: false,
      title: "sit down with Elf ball",
    });
    // Nothing selected: Choose your deck, which brings the picker into view.
    expect(lobbyPrimary(g, viewer({ me, picking: null })).kind).toBe("choose");
    // Changing a deck already in: the commit still wins over Start.
    const ready = meta({ players: [host, seat(1)] });
    expect(lobbyPrimary(ready, viewer({ me: host, mayStart: true, picking: sit })).kind).toBe(
      "sit",
    );
  });

  it("lets an empty Paste a list box lead only a seat with no deck", () => {
    const me = seat(1, { deck_uploaded: false });
    expect(
      lobbyPrimary(meta({ players: [host, me] }), viewer({ me, picking: emptyPaste })),
    ).toMatchObject({ kind: "sit", label: "Upload deck", disabled: true });
    const ready = meta({ players: [host, seat(1)] });
    expect(
      lobbyPrimary(ready, viewer({ me: host, mayStart: true, picking: emptyPaste })).kind,
    ).toBe("start");
  });

  it("adds the bot picker's bot from the footer", () => {
    const add: PickerCommit = { label: "Add bot", disabled: false, title: "seat it", busy: false };
    const g = meta({ players: [host, seat(1, { deck_uploaded: false })] });
    expect(lobbyPrimary(g, viewer({ me: host, addingBot: add, picking: sit }))).toMatchObject({
      kind: "bot",
      label: "Add bot",
    });
    // The admin, not seated, adds bots too.
    expect(lobbyPrimary(g, viewer({ mayStart: true, addingBot: add })).kind).toBe("bot");
    // A seat held elsewhere is reclaimed first.
    expect(lobbyPrimary(g, viewer({ heldElsewhere: true, addingBot: add })).kind).toBe("open");
  });
});

describe("your seat while you browse the picker (ADR 0212 §1)", () => {
  const pv = {
    name: "Elf ball",
    text: {
      what: "Marwyn, the Nurturer",
      second: "Pre-built · Elf ball",
      colors: ["G"],
      ready: false,
    },
    art: [{ name: "Marwyn, the Nurturer", scryfallId: "m1" }],
  };

  it("previews the selected deck, not ready, with its art", () => {
    const p = seatPreview(seat(0, { deck_uploaded: false }), pv);
    expect(p).toEqual({ text: pv.text, art: pv.art, preview: true });
    // Another deck than the one seated is a preview too.
    expect(seatPreview(seat(0, { deck_name: "Raid" }), pv)?.preview).toBe(true);
  });

  it("draws the deck already in the seat as ready, and nothing with nothing selected", () => {
    const p = seatPreview(seat(0, { deck_name: "elf ball" }), pv);
    expect(p?.preview).toBe(false);
    expect(p?.text.ready).toBe(true);
    expect(p?.art).toEqual(pv.art);
    expect(seatPreview(seat(0), null)).toBeNull();
  });
});

describe("History (ADR 0158 §1 item 3)", () => {
  const rows = historyRows(
    [
      mine("won", { winner_seat: 0, ended_at: NOW - 3 * HOUR }),
      mine("lost", {
        winner_seat: 1,
        ended_at: NOW - 2 * HOUR,
        others: [{ seat: 1, name: "Bob" }],
      }),
      mine("shelved", { state: "active", archived_at: NOW - HOUR }),
      mine("now", { state: "active" }),
    ],
    [
      meta({
        id: "listed",
        state: "ended",
        created_at: new Date(NOW - 5 * HOUR).toISOString(),
      }),
      meta({ id: "won", state: "ended" }),
    ],
    new Set(["now"]),
  );

  it("is GET /me/games minus Now, plus the ended tables listed, most recent first", () => {
    expect(rows.map((r) => r.id)).toEqual(["shelved", "lost", "won", "listed"]);
    expect(rows.map((r) => r.result.label)).toEqual(["Archived", "Bob won", "You won", "Ended"]);
  });

  it("filters", () => {
    expect(filterHistory(rows, "won").map((r) => r.id)).toEqual(["won"]);
    expect(filterHistory(rows, "friends").map((r) => r.id)).toEqual(["lost"]);
    expect(filterHistory(rows, "archived").map((r) => r.id)).toEqual(["shelved"]);
    expect(filterHistory(rows, "all")).toHaveLength(4);
  });

  it("shows twenty more, or what is left", () => {
    expect(moreCount(27, 6)).toBe(20);
    expect(moreCount(27, 26)).toBe(1);
    expect(moreCount(6, 6)).toBe(0);
  });
});
