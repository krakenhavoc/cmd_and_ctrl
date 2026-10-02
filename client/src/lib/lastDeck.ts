// lastDeck.ts — the deck a player last seated, for the deck panel to
// preselect (ADR 0110 §5 item 5, Delivery PR 7).
//
// A signed-in person's last deck is on the account (users.last_deck),
// written by the server whenever they seat one and read from GET
// /me/last-deck. A guest's is in this browser, pre-built decks only
// (a guest has no library): localStorage["cmdctrl.lastDeck"].
//
// It is only ever PRESELECTED. Nothing seats it automatically: a seated
// deck is visible to the table, and a stale choice should cost a click,
// not a mulligan. A deck that has gone (deleted from the library, or a
// pre-built deck retired) is simply not preselected.

import { fetchLastDeck } from "./api";
import { signedInUserID } from "./myGames";
import type { Session } from "./session";

export type LastDeckKind = "library" | "prebuilt";

/** users.last_deck, as GET /me/last-deck serves it. */
export interface LastDeck {
  kind: LastDeckKind;
  id: string;
}

/** Where a guest's last pre-built deck is kept. */
export const GUEST_LAST_DECK_KEY = "cmdctrl.lastDeck";

/** readGuestLastDeck is the guest's remembered pre-built deck, or null. */
export function readGuestLastDeck(): LastDeck | null {
  let raw: string | null;
  try {
    raw = localStorage.getItem(GUEST_LAST_DECK_KEY);
  } catch {
    return null;
  }
  if (!raw) return null;
  try {
    const v = JSON.parse(raw) as Partial<LastDeck>;
    if (v.kind === "prebuilt" && typeof v.id === "string" && v.id !== "") {
      return { kind: "prebuilt", id: v.id };
    }
  } catch {
    // A value this build cannot read is no preselection.
  }
  return null;
}

/** rememberGuestLastDeck records a pre-built deck a guest just seated. */
export function rememberGuestLastDeck(prebuiltID: string): void {
  if (!prebuiltID) return;
  try {
    localStorage.setItem(GUEST_LAST_DECK_KEY, JSON.stringify({ kind: "prebuilt", id: prebuiltID }));
  } catch {
    // Private window or blocked storage: no memory, and no harm.
  }
}

/**
 * loadLastDeck is the deck to preselect for this session: the
 * account's for a signed-in person, the browser's for anyone else. A
 * failure is no preselection, never an error.
 */
export async function loadLastDeck(s: Session | null | undefined): Promise<LastDeck | null> {
  if (signedInUserID(s) === null) return readGuestLastDeck();
  try {
    return await fetchLastDeck();
  } catch {
    return null;
  }
}

/**
 * preselectFor returns the id a picker of `kind` should preselect: the
 * last deck when it is of that kind and still in the picker's list,
 * otherwise null (the picker keeps its own default).
 */
export function preselectFor(
  kind: LastDeckKind,
  ids: readonly string[],
  last: LastDeck | null | undefined,
): string | null {
  if (!last || last.kind !== kind) return null;
  return ids.includes(last.id) ? last.id : null;
}
