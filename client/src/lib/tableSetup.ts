// tableSetup.ts — creating a table as a player, and the last setup
// (ADR 0110 §5 items 1–4, Delivery PR 7).
//
// Mirrors server/internal/lobby/setups.go (GET /me/setup, POST
// /games/{id}/setup, POST /games's `setup`) and player_tables.go.
// Hand-maintained, like myDecks.ts — update both sides in lockstep.
//
// The decisions live here rather than in Lobby.svelte for the reason
// myDecks.ts gives: "who sees the create form", "what does the last
// setup say" and "in what order are tablemates offered" are worth a
// test, and a `.svelte` file is the awkward place to write one.

import { isAdmin } from "./admin";
import { signedInUserID } from "./myGames";
import type { Session } from "./session";
import type { Tablemate } from "./tablemates";

/** One bot seat in a setup, mirroring tablesetups.Bot. */
export interface SetupBot {
  tier: string;
  /** The curated deck id; empty for a bot that played a pasted list. */
  deck_id: string;
  name: string;
}

/** The settings a setup carries: a complete PATCH /games/{id}/settings body. */
export interface SetupSettings {
  undo_limit?: number;
  undo_scope?: string;
  starting_life?: number;
  commander_damage?: number;
  bot_pace?: string;
  allow_spawn?: boolean;
}

/** A remembered setup, as GET /me/setup serves it. */
export interface TableSetup {
  settings: SetupSettings;
  bots: SetupBot[];
  /** User ids of the other signed-in people who sat at that table. */
  tablemates: string[];
}

/** GET /me/setup. `setup` is null when the caller has none yet. */
export interface MySetupResponse {
  setup: TableSetup | null;
  game_id?: string;
  /** Unix milliseconds. */
  updated_at?: number;
}

/** One part of a setup that was not applied. */
export interface SetupSkip {
  name: string;
  reason: string;
}

/** What applying a setup did (POST /games's `setup`, POST /games/{id}/setup). */
export interface SetupResult {
  settings: boolean;
  bots_added: number;
  skipped: SetupSkip[];
}

/** The most open lobby tables one person may have (owner answer 2). */
export const MAX_OPEN_TABLES = 3;

/**
 * canCreateTables reports whether the lobby offers this session the
 * create form: any signed-in person (owner answer 2), or an admin. The
 * server decides (POST /games refuses a guest with 403); this only
 * keeps a guest from being offered a form that always fails.
 */
export function canCreateTables(s: Session | null | undefined): boolean {
  return isAdmin(s) || signedInUserID(s) !== null;
}

/** The defaults a new table starts with (game.DefaultTableSettings). */
const DEFAULTS = {
  starting_life: 40,
  commander_damage: 21,
  bot_pace: "normal",
  allow_spawn: false,
};

/**
 * setupSummary is the one line the create form shows beside "Use my
 * last setup": the bots, then any setting that differs from a new
 * table's default. `deckName` turns a curated deck id into its name.
 */
export function setupSummary(
  setup: TableSetup,
  deckName: (id: string) => string = (id) => id,
): string {
  const parts: string[] = [];
  const bots = setup.bots ?? [];
  if (bots.length === 0) {
    parts.push("no bots");
  } else {
    const names = bots.map((b) => {
      const deck = b.deck_id ? deckName(b.deck_id) : "";
      return deck ? `${b.name} (${b.tier}, ${deck})` : `${b.name} (${b.tier})`;
    });
    parts.push(`${bots.length} bot${bots.length === 1 ? "" : "s"}: ${names.join(", ")}`);
  }
  const s = setup.settings ?? {};
  if (s.starting_life !== undefined && s.starting_life !== DEFAULTS.starting_life) {
    parts.push(`starting life ${s.starting_life}`);
  }
  if (s.commander_damage !== undefined && s.commander_damage !== DEFAULTS.commander_damage) {
    parts.push(`commander damage ${s.commander_damage}`);
  }
  if (s.bot_pace !== undefined && s.bot_pace !== DEFAULTS.bot_pace) {
    parts.push(`${s.bot_pace} bots`);
  }
  if (s.allow_spawn === true) parts.push("spawning on");
  return parts.join(" · ");
}

/**
 * setupResultMessage says what applying a setup did, in one line, with
 * every skipped part and its reason. Empty when there is nothing to say.
 */
export function setupResultMessage(res: SetupResult | null | undefined): string {
  if (!res) return "";
  const done: string[] = [];
  if (res.settings) done.push("table settings applied");
  if (res.bots_added > 0)
    done.push(`${res.bots_added} bot${res.bots_added === 1 ? "" : "s"} added`);
  const skipped = (res.skipped ?? []).map((s) => `${s.name}: ${s.reason}`);
  const out: string[] = [];
  if (done.length > 0) out.push(`Last setup: ${done.join(", ")}.`);
  if (skipped.length > 0) out.push(`Skipped ${skipped.join("; ")}.`);
  return out.join(" ");
}

/** One row of the create flow's tablemate list. */
export interface OrderedTablemate {
  mate: Tablemate;
  /** They sat at the table the last setup came from. */
  atLastTable: boolean;
}

/**
 * orderTablemates puts the people from the last setup first, marked
 * "at your last table", then everyone else in the server's order
 * (newest shared table first). Each group keeps the server's order.
 */
export function orderTablemates(
  mates: Tablemate[],
  lastTable: readonly string[] | null | undefined,
): OrderedTablemate[] {
  const first = new Set(lastTable ?? []);
  const top: OrderedTablemate[] = [];
  const rest: OrderedTablemate[] = [];
  for (const mate of mates) {
    if (first.has(mate.user_id)) top.push({ mate, atLastTable: true });
    else rest.push({ mate, atLastTable: false });
  }
  return [...top, ...rest];
}
