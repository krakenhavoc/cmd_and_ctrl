// community.ts — the C2 Discord server, as the client learns it from
// GET /community (ADR 0156 §6): the public invite link and whether
// sign-up is gated by a Discord role.
//
// Fetched once per page load, at shell mount. Every failure (a 404 from
// a server that has no route yet, a 500, a network error, a body that is
// not what we expect) resolves to "no invite, no gate", and every
// Discord placement hides when `invite_url` is null. So the client works
// before, and without, the server half.

import { type Readable } from "svelte/store";

import { guardedWritable } from "./guardedStore";

export interface Community {
  /** The public invite (`https://discord.gg/…`), or null: hide every Discord placement. */
  invite_url: string | null;
  /** Accounts need the Player role on the C2 Discord (ADR 0156 §1). */
  gate: boolean;
  /** A host may give a table a Discord voice room (ADR 0159 §7). */
  voice?: boolean;
  /** A host may post a table to the Discord's #lfg board (ADR 0161 §7). */
  lfg?: boolean;
  /** A host may opt a table in to a results post in #game-results (ADR 0174 §6). */
  results?: boolean;
}

/** The initial value, and the value every failure resolves to. */
export const NO_COMMUNITY: Community = {
  invite_url: null,
  gate: false,
  voice: false,
  lfg: false,
  results: false,
};

const store = guardedWritable<Community>(NO_COMMUNITY, "community");

/** community is the C2 Discord as the server reports it. Only loadCommunity writes it. */
export const community: Readable<Community> = { subscribe: store.subscribe };

/**
 * parseCommunity narrows an untrusted body onto Community. Only an
 * absolute https URL is an invite: anything else would be a link the
 * header renders into every page.
 */
export function parseCommunity(body: unknown): Community {
  if (typeof body !== "object" || body === null) return NO_COMMUNITY;
  const b = body as {
    invite_url?: unknown;
    gate?: unknown;
    voice?: unknown;
    lfg?: unknown;
    results?: unknown;
  };
  return {
    invite_url: inviteURL(b.invite_url),
    gate: b.gate === true,
    voice: b.voice === true,
    lfg: b.lfg === true,
    results: b.results === true,
  };
}

function inviteURL(v: unknown): string | null {
  if (typeof v !== "string" || v === "") return null;
  try {
    return new URL(v).protocol === "https:" ? v : null;
  } catch {
    return null;
  }
}

let inflight: Promise<Community> | null = null;

/** loadCommunity fetches GET /community once per page load; concurrent calls share it. */
export async function loadCommunity(): Promise<Community> {
  if (inflight) return inflight;
  inflight = (async () => {
    try {
      const res = await fetch("/community", { method: "GET", credentials: "same-origin" });
      if (!res.ok) return NO_COMMUNITY;
      return parseCommunity(await res.json());
    } catch {
      return NO_COMMUNITY;
    }
  })();
  const c = await inflight;
  store.set(c);
  return c;
}

/** _resetCommunityForTests clears the memoised fetch and the store. */
export function _resetCommunityForTests(): void {
  inflight = null;
  store.set(NO_COMMUNITY);
}

/** _setCommunityForTests puts a value in the store, as a load would. */
export function _setCommunityForTests(c: Community): void {
  store.set(c);
}
