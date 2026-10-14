// privacy.ts — the numbers and links the privacy notice (routes/Privacy.svelte,
// #2983, ADR 0149) states, in one place.
//
// THE NOTICE MUST BE TRUE OF THE CODE. Every number below is a server
// constant or a deployment setting, and server/internal/lobby/
// privacy_facts_test.go reads this file and fails when one drifts from
// its source. If you change a retention, a lifetime or a limit on the
// server, that test names the line here to change, and the page text
// follows from it. If you add a table, column or file that holds
// something about a person, accountdelete.Inventory fails first; then
// update the page in the same change (ADR 0149 §6).

import { LEGAL_CONTACT_EMAIL, LEGAL_CONTACT_URL } from "./legalNotice";

/** The hash route of the notice. */
export const PRIVACY_HASH = "#/privacy";

/** What a person types to confirm a deletion; the server checks the same words. */
export const DELETE_PHRASE = "delete my account";

/**
 * The contact the notice gives: the operator's legal address (#3358),
 * which lives in lib/legalNotice.ts with the operator's name so the
 * terms of service and this notice cannot name different ones.
 */
export const CONTACT_URL = LEGAL_CONTACT_URL;
export const CONTACT_TEXT = LEGAL_CONTACT_EMAIL;

/**
 * The database schema version (the highest server/internal/db/migrations
 * number) this notice was last checked against. privacy_facts_test.go
 * fails when a newer migration exists, so a change to what is stored
 * cannot ship without someone reading this page again (ADR 0149 §6).
 * Bump it when you have.
 *
 * 18 (#3071) renames a pre-built deck ID inside users.last_deck and
 * table_setups, which the notice already covers; it stores nothing new.
 *
 * 19 (#3179, ADR 0156 §1) adds users.access_checked_at and
 * users.access_result, the sign-up gate's last answer; the notice's
 * "Your Discord sign-in" row names them, and the guilds.members.read
 * scope the gate asks for.
 *
 * 20 (#3223, ADR 0159 §5) adds games.voice_room and
 * games.voice_channel_id: a table's choice of a Discord voice room and
 * the room's channel id. Neither is about a person; the notice's voice
 * paragraph comes with the voice manager (ADR 0159 §8), which is what
 * sends seated players' Discord ids to Discord.
 *
 * 21 (#3319, ADR 0161 §6) adds games.lfg and games.lfg_message_id: a
 * table's choice of the Discord #lfg board and its post's message id.
 * Neither is about a person; the Discord row now says what a post shows
 * the community server (the host's Discord account among it) and when
 * it is deleted (ADR 0161 §8).
 *
 * 22 (#3320, ADR 0162 §5) adds notification_prefs: a person's three
 * Discord DM notification switches and when Discord last refused one.
 * The notice's "Your Discord notifications" row names it, and the
 * Discord paragraph says what the messages carry (§6).
 *
 * 23 (#3364, ADR 0165) adds game_stats, seat_stats and seat_card_stats:
 * numbers and card names derived from a finished game, keyed by game
 * and seat, with no names or user ids. The notice's "Game statistics"
 * row covers them.
 *
 * 24 (#3322, ADR 0174) adds games.results: the host's choice of a
 * results post in the Discord server. The column names no one; the
 * post it leads to does, and the notice's "Results post in Discord"
 * row covers it.
 *
 * 25 (#3460, ADR 0182) adds model_keys, a person's own model API key
 * for the AI seats they add (sealed, last four characters, provider,
 * model, limits, dates, last error), and seats.bot_key_user_id /
 * bot_key_model on a bot that ran on it. The notice's "Your own model
 * key for AI opponents" row covers them; account deletion deletes the
 * key row.
 */
export const PRIVACY_SCHEMA_VERSION = 25;

/**
 * The numbers the notice states. privacy_facts_test.go (Go) checks each
 * against the code it names; keep this literal one `key: number` per line.
 */
export const PRIVACY_FACTS = {
  /** ws.EndedTableRetention: an ended table's restore point is rebuilt at boot for this long. */
  endedTableDays: 7,
  /** bugstore.DefaultRetention: how long a report's screenshots and pinned files survive. */
  bugArtifactDays: 90,
  /** bugstore.DefaultMaxStoreBytes, in MiB: the store's size ceiling, oldest reports first. */
  bugArtifactMiB: 512,
  /** bugstore.MaxImages: screenshots per report. */
  bugImages: 4,
  /** lobby identityTTL: a Discord sign-in lasts this long. */
  signInDays: 30,
  /** lobby sessionTTL: a guest seat or spectator link lasts this long. */
  guestSessionHours: 12,
  /** lobby deckRequestLimit: deck requests per person per window. */
  deckRequestsPerWindow: 3,
  /** lobby deckRequestWindow, in hours. */
  deckRequestWindowHours: 24,
  /** playmat.MaxSlots: saved playmats per account. */
  playmatSlots: 3,
  /** lobby.ReportRetention, in days: how long a player report is kept. */
  reportDays: 90,
  /** lobby.ReportMaxRows: reports kept at most; the oldest go first. */
  reportMax: 500,
  /** lobby reportExcerptLines: recent chat lines a report carries. */
  reportChatLines: 10,
  /** lobby reportBurst: reports one person can file in a burst. */
  reportsPerWindow: 3,
  /** lobby reportEvery, in minutes: then one more per this long. */
  reportEveryMinutes: 5,
  /** ws.RecentChatLimit: chat lines a table remembers in memory, only to attach to a report. */
  chatRecentLines: 20,
  /** ADR 0123 decision 7: how long the monitoring stack keeps server logs. */
  logDays: 30,
  /** scripts/backup-offsite.sh: restic keeps this many daily snapshots. */
  backupDaily: 7,
  /** ... this many weekly ones. */
  backupWeekly: 4,
  /** ... and this many monthly ones. */
  backupMonthly: 6,
} as const;
