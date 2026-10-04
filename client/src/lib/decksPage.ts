// decksPage.ts is the pure half of ADR 0112 §3, the one decks page
// (#/decks, with #/deck-check as its permanent alias): who may request,
// save and see a library, the report's "as printed" line, the save
// button, and the route a sign-in returns to. No fetch and no DOM
// beyond sessionStorage, so the decision table is unit-tested
// (decksPage.test.ts) and routes/Decks.svelte only renders the answers.

import type { CoverageReport, DeckCheckRequest } from "./deckcheck";
import { signedInUserID } from "./myGames";
import type { MyDeckInfo } from "./myDecks";
import { parseHash } from "./router";
import type { Session } from "./session";

// DeckAction is what the page shows for one of request, save or the
// library (§3 item 3):
//
//   - "allowed": the control itself.
//   - "signIn": a signed-out visitor, offered "Sign in with Discord".
//   - "linkDiscord": a guest seat or guest spectator. A Discord sign-in
//     from the login flow would replace the browser's session, and a
//     guest seat cannot be got back without a reclaim ticket, so their
//     way to an account is "Link Discord" at their table (ADR 0051).
//   - "hidden": nothing at all. The admin token is not a person: it has
//     no library, and it requests through the bot's path only.
export type DeckAction = "allowed" | "signIn" | "linkDiscord" | "hidden";

export interface DecksAccess {
  request: DeckAction;
  save: DeckAction;
  library: DeckAction;
}

// decksAccess is §3 item 3's table. A signed-in person may do
// everything in either admin mode: the library is theirs, and saving
// and requesting are not admin rights.
//
// One row the ADR's table does not list: a Discord sign-in on a server
// with no user database (role "identified", no user_id). POST
// /deck-requests accepts it (it has a Discord identity), but there is
// no library to save to, so save and the library are hidden.
export function decksAccess(s: Session | null | undefined): DecksAccess {
  if (!s) return { request: "signIn", save: "signIn", library: "signIn" };
  if (signedInUserID(s) !== null) {
    return { request: "allowed", save: "allowed", library: "allowed" };
  }
  const { role } = s.principal;
  if (role === "identified") return { request: "allowed", save: "hidden", library: "hidden" };
  if (role === "player" || role === "spectator") {
    return { request: "linkDiscord", save: "linkDiscord", library: "linkDiscord" };
  }
  return { request: "hidden", save: "hidden", library: "hidden" };
}

// reportDeckSize is every copy the report accounts for: the buckets by
// copies plus the copies the card index could not find. A Commander
// deck totals 100 (#2220).
export function reportDeckSize(r: CoverageReport): number {
  const c = r.copies;
  return c.manual + c.unreviewed + c.caveats + c.automated + c.no_effect + r.unknown_copies;
}

// reportAsPrinted is the check report's headline, in the library's
// words (ADR 0110 §6 item 2): automated plus nothing-to-automate, over
// the deck's size. Counted by copies, as the server's
// Report.AsPrintedCopies counts them, so thirty Forests are thirty.
export function reportAsPrinted(r: CoverageReport): string {
  const size = reportDeckSize(r);
  if (size === 0) return "";
  return `${r.copies.automated + r.copies.no_effect} of ${size} cards play as printed`;
}

// reportNotFound is the "K not found" note beside the bar: names the
// index could not resolve, in copies. Empty when there are none.
export function reportNotFound(r: CoverageReport): string {
  return r.unknown_copies > 0 ? `${r.unknown_copies} not found` : "";
}

// defaultSaveName fills the save field: the deck's name, then its first
// commander. Empty leaves the server's own fallback ("Untitled deck").
export function defaultSaveName(r: CoverageReport): string {
  const name = r.deck_name.trim();
  if (name) return name;
  return (r.commanders.find((c) => c.trim() !== "") ?? "").trim();
}

export const SAVE_LABEL = "Save to my decks";

// saveButtonLabel says "Replace ‹name›" before anything is sent when
// the caller already has a deck by that name, since the library's
// update rule (same owner, same name) replaces it in place. The server
// compares names exactly after trimming, and so does this.
export function saveButtonLabel(name: string, library: readonly MyDeckInfo[]): string {
  const n = name.trim();
  if (n && library.some((d) => d.name === n)) return `Replace ${n}`;
  return SAVE_LABEL;
}

/** POST /me/decks' body: the GET /me/decks entry, and whether it replaced one. */
export interface SaveDeckResponse {
  deck: MyDeckInfo;
  replaced: boolean;
}

export function savedMessage(res: SaveDeckResponse): string {
  return res.replaced
    ? `Replaced ${res.deck.name} in your decks.`
    : `Saved ${res.deck.name} to your decks.`;
}

// libraryDeckRequestable decides whether a library deck shows "Request
// missing cards" (§3 item 5): every deck, link or pasted, unless its
// coverage shows nothing to ask for, the same rule the check report
// uses (canRequestCards). A deck whose coverage the server could not
// compute keeps the button, and the server's answer is the final word.
export function libraryDeckRequestable(d: MyDeckInfo): boolean {
  const c = d.coverage;
  if (!c) return true;
  return c.counts.manual > 0 || c.counts.unreviewed > 0;
}

// --- the sign-in return route (§3 item 7) ---------------------------

/** sessionStorage key for the route a Discord sign-in returns to. */
export const AFTER_SIGN_IN_KEY = "cmdctrl.afterSignIn";
/** sessionStorage key for the pasted list that route checks again. */
export const AFTER_SIGN_IN_TEXT_KEY = "cmdctrl.afterSignIn.text";

// isDecksReturn accepts a stored return route only if it is a hash
// route that parses as the decks page, so a stored value can never send
// the browser anywhere else.
export function isDecksReturn(hash: string | null | undefined): hash is string {
  if (!hash || !hash.startsWith("#/")) return false;
  return parseHash(hash).name === "decks";
}

// returnHashFor is the route that shows the same deck again: a checked
// link comes back as #/decks?url=<link>, which runs the check on load;
// a pasted list comes back as #/decks, with the list stored beside it.
export function returnHashFor(checked: DeckCheckRequest | null): string {
  if (checked && "url" in checked) return `#/decks?url=${encodeURIComponent(checked.url)}`;
  return "#/decks";
}

function storage(): Storage | null {
  try {
    return typeof sessionStorage === "undefined" ? null : sessionStorage;
  } catch {
    return null;
  }
}

// rememberAfterSignIn stores the return route, and a pasted list, just
// before a signed-out visitor follows "Sign in with Discord". It is
// client-only: the server's OAuth state carries nothing new.
export function rememberAfterSignIn(checked: DeckCheckRequest | null): void {
  const st = storage();
  if (!st) return;
  try {
    st.setItem(AFTER_SIGN_IN_KEY, returnHashFor(checked));
    if (checked && "text" in checked) st.setItem(AFTER_SIGN_IN_TEXT_KEY, checked.text);
    else st.removeItem(AFTER_SIGN_IN_TEXT_KEY);
  } catch {
    // Storage full or blocked: the sign-in still works, and lands on
    // the Lobby as it would without this.
  }
}

// The pasted list taken by takeAfterSignIn, held for the decks page to
// pick up once it mounts. Module state, not storage: the keys are
// cleared the moment the route is read, and the SPA does not reload
// between the Discord round trip and the page.
let pendingText = "";

// takeAfterSignIn reads and clears both keys, and returns the stored
// route if it is the decks page. Called once by the Discord round
// trip's login-page branch (App.svelte, oauthCompleteTarget).
export function takeAfterSignIn(): string | null {
  const st = storage();
  if (!st) return null;
  let hash: string | null = null;
  let text: string | null = null;
  try {
    hash = st.getItem(AFTER_SIGN_IN_KEY);
    text = st.getItem(AFTER_SIGN_IN_TEXT_KEY);
    st.removeItem(AFTER_SIGN_IN_KEY);
    st.removeItem(AFTER_SIGN_IN_TEXT_KEY);
  } catch {
    return null;
  }
  if (!isDecksReturn(hash)) {
    pendingText = "";
    return null;
  }
  pendingText = text ?? "";
  return hash;
}

// takePendingDeckText hands the decks page the list a sign-in carried
// across, once.
export function takePendingDeckText(): string {
  const t = pendingText;
  pendingText = "";
  return t;
}
