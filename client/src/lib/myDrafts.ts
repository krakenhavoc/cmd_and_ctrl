// myDrafts.ts — a signed-in person's finished drafts (ADR 0222 §9,
// #5440): GET /me/drafts and GET /me/drafts/{id}, and the words the two
// pages say about a record.
//
// A record is one person's own look back at one draft: the set, the
// date, the cards they picked, the deck they built and how each round
// went. It names no other person except, for a round against a
// signed-in opponent, the display name that account has now.

import { authFetch } from "./session";

/** One card the person picked. */
export interface MyDraftPick {
  name: string;
  scryfall_id?: string;
  set?: string;
  collector_number?: string;
  rarity?: string;
  colors?: string[];
  mana_cost?: string;
  mana_value?: number;
  type_line?: string;
  basic?: boolean;
  /** Their deck played it; the rest was their sideboard. */
  in_deck: boolean;
}

export type OpponentKind = "person" | "guest" | "bot" | "bye";

/** One round the person played. */
export interface MyDraftMatch {
  round: number;
  opponent_kind: OpponentKind;
  /** The opponent's display name now; absent when there is none to show. */
  opponent_name?: string;
  result: "win" | "loss" | "draw" | "";
  games_won: number;
  games_lost: number;
}

/** One finished draft. `picks` and `basics` are absent from the list. */
export interface MyDraft {
  id: string;
  name: string;
  set: string;
  set_name: string;
  mode: string;
  seats: number;
  best_of: number;
  seat: number;
  drafted_at: string;
  picks?: MyDraftPick[];
  basics?: Record<string, number>;
  matches: MyDraftMatch[];
  wins: number;
  losses: number;
  draws: number;
}

/** The hash of the list, and of one draft. */
export const MY_DRAFTS_HASH = "#/drafts";
export function myDraftHash(id: string): string {
  return `${MY_DRAFTS_HASH}/${encodeURIComponent(id)}`;
}

/** `GET /me/drafts`: newest first, without cards. */
export async function fetchMyDrafts(signal?: AbortSignal): Promise<MyDraft[]> {
  const res = await authFetch("/me/drafts", { signal });
  const body = (await res.json()) as { drafts?: MyDraft[] };
  return body.drafts ?? [];
}

/** `GET /me/drafts/{id}`: one draft, with its cards. */
export async function fetchMyDraft(id: string, signal?: AbortSignal): Promise<MyDraft> {
  const res = await authFetch(`/me/drafts/${encodeURIComponent(id)}`, { signal });
  return ((await res.json()) as { draft: MyDraft }).draft;
}

/** A draft's record as "2–1", with draws as a third number when it has any. */
export function draftRecord(d: Pick<MyDraft, "wins" | "losses" | "draws">): string {
  return d.draws > 0 ? `${d.wins}–${d.losses}–${d.draws}` : `${d.wins}–${d.losses}`;
}

/** Who a round was against, in words. */
export function opponentText(m: MyDraftMatch): string {
  switch (m.opponent_kind) {
    case "bye":
      return "nobody";
    case "bot":
      return "a bot";
    case "guest":
      return "a guest";
    default:
      return m.opponent_name || "a player";
  }
}

/** One round, as a sentence. */
export function roundText(m: MyDraftMatch): string {
  if (m.opponent_kind === "bye") return "Sat the round out";
  const who = opponentText(m);
  const score = `${m.games_won}–${m.games_lost}`;
  switch (m.result) {
    case "win":
      return `Won ${score} against ${who}`;
    case "loss":
      return `Lost ${score} to ${who}`;
    case "draw":
      return `Drew ${score} with ${who}`;
    default:
      return `Played ${who}`;
  }
}

/** The date a draft was played, for the list: "11 Nov 2026". */
export function draftDate(iso: string, locale?: string): string {
  const t = Date.parse(iso);
  if (Number.isNaN(t)) return "";
  return new Date(t).toLocaleDateString(locale, {
    day: "numeric",
    month: "short",
    year: "numeric",
  });
}

/** A record's picks as draft cards, for the pool component: the deck, then the rest. */
export function splitPicks(d: Pick<MyDraft, "picks">): {
  deck: (MyDraftPick & { id: string })[];
  side: (MyDraftPick & { id: string })[];
} {
  const deck: (MyDraftPick & { id: string })[] = [];
  const side: (MyDraftPick & { id: string })[] = [];
  (d.picks ?? []).forEach((p, i) => (p.in_deck ? deck : side).push({ ...p, id: `p${i}` }));
  return { deck, side };
}

/** The basic lands a deck added, as "9 Forest, 8 Island"; "" for none. */
export function basicsText(basics: Record<string, number> | undefined): string {
  return ["Plains", "Island", "Swamp", "Mountain", "Forest"]
    .filter((name) => (basics?.[name] ?? 0) > 0)
    .map((name) => `${basics?.[name]} ${name}`)
    .join(", ");
}
