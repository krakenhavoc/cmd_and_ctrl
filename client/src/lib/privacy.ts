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
 *
 * 26 (#3337, ADR 0177) adds seats.stand_in / stand_in_since: the bot
 * tier playing an away player's seat, and since when. Neither names a
 * person, and nothing on the notice changes.
 *
 * 27 (#3671) adds decks.power_level: an admin's estimate of a saved
 * deck's power, shown at the table. It is about the deck, goes with
 * the deck row, and the "Your decks" row now names it.
 *
 * 28 (#4135) adds deck_request_cards: which cards each deck's GitHub
 * issue lists, under a deck key and an issue number. It names no person
 * (the requester is in deck_request_asks, which account deletion
 * removes) and is kept, like deck_requests; the "Deck requests" row now
 * says the server remembers the cards an issue listed.
 *
 * 29 (#4880, ADR 0217) is game statistics version 2: new columns on
 * game_stats and seat_stats, and seat_turn_stats, seat_target_stats and
 * seat_card_seen, worked out from a game's log and its replay. Like
 * migration 0023's tables they are keyed by game and seat, name no one,
 * and are kept with the game. The "Game statistics" row now says what
 * the replay adds.
 *
 * 30 (#4880, ADR 0217 slice 3) adds stats_prefs: whether advanced stats
 * are on for a person, and whether they or an admin set it (plus empty
 * columns for the later sharing slices). The new "Stats settings" row
 * names it, and account deletion removes the row. It also adds
 * stats_settings, the server's advanced-stats default, which names no
 * one.
 *
 * 31 (#4562, ADR 0211) adds games.last_active_at and games.closed_reason:
 * when a table was last used, and why the server closed it for being
 * idle. Both are about the table and name no one, and nothing on the
 * notice changes.
 *
 * 32 (#3313, ADR 0213 slice 3) adds friendships, friend_requests,
 * user_blocks and social_prefs: a person's friends, the friend requests
 * they sent and were sent (with whether the receiver ignored one), the
 * people they blocked, and their two social settings. The new "Friends
 * and blocks" row names them, with the 30-day request expiry and the
 * friends cap; account deletion deletes every row with the person on
 * either side.
 *
 * 33 (#4880, ADR 0220 slice 2) adds pods, pod_members, pod_invites and
 * pod_suggestion_dismissals: the pods a person made or joined (a pod's
 * name, its owner and the SHA-256 of its invite link), who is in each,
 * the invites they sent and were sent (with whether they declined one),
 * and, from slice 4, their Not now on a suggested pod. The new "Pods"
 * row names them, with the 30-day invite expiry; account deletion passes
 * a pod they own to the member who joined first and deletes their rows.
 *
 * 34 (#3321, ADR 0215 slice 2) adds scheduled_tables and scheduled_rsvps:
 * the tables a person scheduled (a name, a start, their time zone, the
 * players and the create options) and the list of people with a seat or
 * on the waitlist, in the order they reserved, keyed by Discord id. The
 * new "Scheduled tables" row names them, with the 24-hour keep; account
 * deletion deletes the plans they host and every RSVP of theirs.
 * 35 (#3323, ADR 0163 §4) adds identities.role_connection_token: the
 * sealed refresh token of a person's Discord Linked Roles grant, kept
 * only once they link CMD & CTRL on Discord. The new "Your stats on your
 * Discord profile" row names it and what Discord is sent (display name,
 * games played, games won, the date joined), and the Discord paragraph
 * says so; account deletion empties the Discord connection, then
 * deletes the token with the identity row.
 */
export const PRIVACY_SCHEMA_VERSION = 35;

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
  /** social.RequestTTL, in days: an unanswered friend request is deleted after this long. */
  friendRequestDays: 30,
  /** social.MaxFriends: friends one person can have. */
  friendCap: 200,
  /** pods.InviteTTL, in days: an unanswered pod invite is deleted after this long. */
  podInviteDays: 30,
  /** pods.DismissTTL, in days: how long a Not now on a suggested pod is kept. */
  podDismissDays: 180,
  /** pods.MaxMembers: people one pod can hold. */
  podMemberCap: 12,
  /** pods.MaxPodsPerPerson: pods one person can be in. */
  podsPerPerson: 10,
  /** schedule.Keep, in hours: a scheduled table and its list are kept this long after the start, or after a cancel. */
  scheduledKeepHours: 24,
  /** ADR 0123 decision 7: how long the monitoring stack keeps server logs. */
  logDays: 30,
  /** scripts/backup-offsite.sh: restic keeps this many daily snapshots. */
  backupDaily: 7,
  /** ... this many weekly ones. */
  backupWeekly: 4,
  /** ... and this many monthly ones. */
  backupMonthly: 6,
} as const;
