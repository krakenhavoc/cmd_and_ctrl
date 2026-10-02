// guestName.ts — "remember my name", for guests only (ADR 0110 §5 item
// 6, owner answer 8).
//
// A guest types a name to sit down; the browser remembers the last one
// and pre-fills the Join page and the login page's code box with it. A
// signed-in seat keeps the Discord display name (ADR 0051 sub-PR 4) and
// is never asked, so nothing here runs for one.
//
// Browser-only, and best effort: a private window or blocked storage
// just means an empty field, as before.

/** Where the last guest name is kept. */
export const GUEST_NAME_KEY = "cmdctrl.guestName";

/** The server trims a seat name to 40 characters (lobby.AddBot / Join). */
const MAX_NAME = 40;

/** loadGuestName is the name a guest last typed, or "". */
export function loadGuestName(): string {
  try {
    return (localStorage.getItem(GUEST_NAME_KEY) ?? "").slice(0, MAX_NAME);
  } catch {
    return "";
  }
}

/** rememberGuestName keeps a name a guest just joined with. Blank is ignored. */
export function rememberGuestName(name: string): void {
  const v = name.trim().slice(0, MAX_NAME);
  if (!v) return;
  try {
    localStorage.setItem(GUEST_NAME_KEY, v);
  } catch {
    // No storage: nothing remembered.
  }
}
