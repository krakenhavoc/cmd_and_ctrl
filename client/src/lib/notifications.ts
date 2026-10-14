// notifications.ts — Discord DM notifications (ADR 0162): the three
// switches behind GET and PUT /me/notifications, the store Settings and
// the table's nudge share, and the nudge's one-time dismissal.
//
// The server sends the DMs; this file only reads and writes the
// switches. A 503 (no bot token: every environment but production) and
// a 403 (not a signed-in person) both mean "not available here", and
// every surface that reads the store hides itself.

import { get, type Writable } from "svelte/store";
import { guardedWritable } from "./guardedStore";
import { withSeen } from "./hints/seen";
import { signedInUserID } from "./myGames";
import { authFetch, LobbyApiError, session } from "./session";
import { settings } from "./settings";

/** One switch: the body field and what the Settings row says. */
export type NotificationKind = "turn" | "table_ready" | "table_filled";

/** The body of GET and PUT /me/notifications. */
export interface NotificationPrefs {
  available: boolean;
  turn: boolean;
  table_ready: boolean;
  table_filled: boolean;
  /** Discord refused the last DM (code 50007); nothing is sent until a switch is turned on. */
  dm_failed: boolean;
}

/** OFF is what every surface shows until the server says otherwise. */
export const NOTIFICATIONS_OFF: NotificationPrefs = Object.freeze({
  available: false,
  turn: false,
  table_ready: false,
  table_filled: false,
  dm_failed: false,
});

/** The rows Settings shows, in order. */
export const NOTIFICATION_SWITCHES: readonly {
  kind: NotificationKind;
  label: string;
  help: string;
}[] = [
  {
    kind: "turn",
    label: "Your turn",
    help: "When a game has waited on you for 2 minutes and you are not looking at it.",
  },
  {
    kind: "table_ready",
    label: "Table ready",
    help: "When a table you sit at fills up or starts while you are away.",
  },
  {
    kind: "table_filled",
    label: "Table filled",
    help: "When a table you host fills up while you are away.",
  },
];

/** The Discord-refused hint, word for word from ADR 0162 §4. */
export const DM_FAILED_HINT =
  "Discord refused the last message. Check that you allow DMs from the CMD & CTRL server's members.";

/**
 * The nudge's id in `settings.help.seen` (ADR 0125 §4), the synced map
 * of things a person has dismissed. Not a hint: nothing in lib/hints
 * knows it, so "show tips again" does not bring it back, and the map
 * keeps ids it does not know.
 */
export const NUDGE_ID = "notify.dm-nudge";
const NUDGE_VERSION = 1;

/** The switches as the server last said, OFF until it has. */
export const notificationPrefs: Writable<NotificationPrefs> = guardedWritable(
  NOTIFICATIONS_OFF,
  "notificationPrefs",
);

function normalize(raw: unknown): NotificationPrefs {
  const r = (raw ?? {}) as Partial<NotificationPrefs>;
  return {
    available: r.available === true,
    turn: r.turn === true,
    table_ready: r.table_ready === true,
    table_filled: r.table_filled === true,
    dm_failed: r.dm_failed === true,
  };
}

/**
 * loadNotifications reads the switches into the store. A caller that is
 * not a signed-in person is not asked: the server would answer 403.
 * Any refusal leaves the store OFF, which hides everything.
 */
export async function loadNotifications(): Promise<NotificationPrefs> {
  if (signedInUserID(get(session)) === null) {
    notificationPrefs.set(NOTIFICATIONS_OFF);
    return NOTIFICATIONS_OFF;
  }
  try {
    const res = await authFetch("/me/notifications", { cache: "no-store" });
    const got = normalize(await res.json());
    notificationPrefs.set(got);
    return got;
  } catch (e) {
    if (!(e instanceof LobbyApiError && (e.status === 503 || e.status === 403))) {
      console.warn("could not read the notification settings", e);
    }
    notificationPrefs.set(NOTIFICATIONS_OFF);
    return NOTIFICATIONS_OFF;
  }
}

/**
 * setNotification turns one switch on or off. It resolves to the
 * server's answer, already in the store, and rejects with the server's
 * message when the write is refused.
 */
export async function setNotification(
  kind: NotificationKind,
  on: boolean,
): Promise<NotificationPrefs> {
  const res = await authFetch("/me/notifications", {
    method: "PUT",
    body: JSON.stringify({ [kind]: on }),
  });
  const got = normalize(await res.json());
  notificationPrefs.set(got);
  return got;
}

/** allOff reports whether no switch is on. */
export function allOff(p: NotificationPrefs): boolean {
  return !p.turn && !p.table_ready && !p.table_filled;
}

/** nudgeDismissed reports whether this person has answered the nudge. */
export function nudgeDismissed(seen: Record<string, number>): boolean {
  return (seen[NUDGE_ID] ?? 0) >= NUDGE_VERSION;
}

/**
 * shouldNudge is ADR 0162 §1's rule: the feature is on here, every
 * switch is off, and the person has not answered the nudge before.
 */
export function shouldNudge(p: NotificationPrefs, seen: Record<string, number>): boolean {
  return p.available && allOff(p) && !nudgeDismissed(seen);
}

/** dismissNudge remembers the answer in the synced settings, for good. */
export function dismissNudge(): void {
  settings.update((prev) => {
    const seen = withSeen(prev.help.seen, NUDGE_ID, NUDGE_VERSION);
    return seen === prev.help.seen ? prev : { ...prev, help: { ...prev.help, seen } };
  });
}
