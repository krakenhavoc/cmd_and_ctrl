// myDecks.ts — the signed-in player's deck library (ADR 0051 decision
// 7, S34 sub-PR 5): types, the "signed in" check, and the one line the
// picker puts under each saved deck's name.
//
// Mirrors GET /me/decks (server/internal/lobby/http.go's
// myDeckInfo/myDecksResponse). Hand-maintained, like prebuiltDecks.ts
// and protocol.ts — update both sides in lockstep.
//
// The logic lives here rather than in the component for the same
// reason prebuiltDecks.ts's summariser does: this project has no
// jsdom, so a `.svelte` file cannot be unit-tested, and both "is this
// caller signed in" and "what does this row say" are worth pinning
// down with a test.

/** One saved deck, as GET /me/decks reports it. */
export interface MyDeckInfo {
  id: string;
  name: string;
  commanders: string[];
  card_count: number;
  updated_at: string;
}

export interface MyDecksResponse {
  decks: MyDeckInfo[];
}

// ZERO_USER_ID is the all-zero uuid a Go uuid.UUID marshals to when
// it's the zero value. Principal.user_id (server:
// server/internal/auth/auth.go) carries `json:"...,omitempty"`, but
// `omitempty` never actually omits it: encoding/json only calls an
// array "empty" at length zero, and uuid.UUID is a fixed [16]byte, so
// the field always renders as a string rather than being left out.
// This is true of every uuid.UUID field on Principal (admin_id,
// game_id, player_id too), not something specific to user_id — see
// ADR 0051's sub-PR 5 implementation notes.
export const ZERO_USER_ID = "00000000-0000-0000-0000-000000000000";

/**
 * isSignedIn reports whether a principal's user_id names a real
 * users(id) row — i.e. whether "your decks" has anywhere to read
 * from. Deliberately NOT a truthiness check on user_id: the field is
 * always present (see ZERO_USER_ID above), so `!!user_id` is always
 * true and would show the picker to every guest.
 */
export function isSignedIn(userID: string | undefined): boolean {
  return !!userID && userID !== ZERO_USER_ID;
}

/**
 * deckSubtitle is the "commander(s) · N cards" line under a saved
 * deck's name. Partners/backgrounds join with " / ", matching how a
 * player would read two commanders named together; an empty
 * commanders list (parsing hadn't resolved one, or the format never
 * carries one) just leaves the count.
 */
export function deckSubtitle(deck: MyDeckInfo): string {
  const commanders = deck.commanders.filter((c) => c.trim() !== "").join(" / ");
  const count = `${deck.card_count} card${deck.card_count === 1 ? "" : "s"}`;
  return [commanders, count].filter((s) => s !== "").join(" · ");
}
