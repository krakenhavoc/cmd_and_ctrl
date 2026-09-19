import {
  authFetch,
  currentSession,
  setSession,
  LobbyApiError,
  type ApiViolation,
  type Session,
} from "./session";
import {
  redactBugContext,
  type BugLogEntry,
  type BugReportContext,
  type BugReportKind,
} from "./bugReport";
import { redactSecrets } from "./redact";
import type { PrebuiltDecksResponse } from "./prebuiltDecks";
import type { AutoTapCastParams } from "./castPreview";

// Re-export the violation shape so consumers of api.ts don't also
// have to import from session.ts. ApiViolation is the canonical
// name; DeckViolation is kept as an alias for existing callers.
export type DeckViolation = ApiViolation;
export type { ApiViolation };

// GameMeta mirrors lobby.GameMeta in server/internal/lobby/lobby.go.
export interface GameMeta {
  id: string;
  name: string;
  created_at: string;
  invite_token?: string;
  // Per-game spectator invite. Distinct from invite_token — sharing
  // the player invite with a spectator would let them claim a seat.
  // Only emitted to the admin and seated players; stripped from list
  // responses + spectator-session responses. Added in S11.
  spectator_invite?: string;
  players: SeatInfo[];
  state: "lobby" | "active" | "ended";
  // Set when an admin has archived the table: it drops out of the
  // default listing (GET /games) and appears under GET
  // /games?archived=1 instead. Nothing is deleted — unarchiving puts
  // it back, replay and all.
  archived_at?: string;
}

export interface SeatInfo {
  player_id: string;
  name: string;
  seat: number;
  deck_name?: string;
  deck_uploaded: boolean;
  // Discord OAuth seats (S12.5): the client builds the cached avatar
  // URL from discord_id + discord_avatar_hash; display_name carries
  // the friendly global_name. All empty for name-form joins.
  discord_id?: string;
  discord_avatar_hash?: string;
  display_name?: string;
  // Bot seats (S31). is_bot marks a seat driven by a server-side
  // policy runner; bot_tier is its difficulty tier and bot_deck the
  // curated deck ID it was seated with.
  is_bot?: boolean;
  bot_tier?: string;
  bot_deck?: string;
}

// BotTierInfo mirrors aiseat.TierInfo. Every declared tier is listed,
// including the ones this server cannot play — `available` is false
// for those and the picker greys them out rather than pretending the
// difficulty slider has a single notch.
//
// Availability is a property of the SERVER, not of the build: the
// model-backed tiers need a model endpoint configured, and `reason`
// is the one-line explanation of what is missing when they are off.
export interface BotTierInfo {
  tier: string;
  label: string;
  description: string;
  available: boolean;
  // reason is present only on an unavailable tier. Safe to render
  // verbatim next to the greyed-out option.
  reason?: string;
}

// BotDeckInfo mirrors aiseat.DeckInfo — one curated bot deck.
export interface BotDeckInfo {
  id: string;
  name: string;
  description?: string;
  colors?: string[];
  commander?: string;
}

// BotOptions is GET /bot/options: what the Add-bot picker renders.
// `enabled` is false on a server with no bot host at all, in which
// case the lobby hides the control instead of offering a button that
// 503s.
export interface BotOptions {
  tiers: BotTierInfo[];
  decks: BotDeckInfo[];
  enabled: boolean;
}

// AddBotResponse mirrors lobby.addBotResponse.
export interface AddBotResponse {
  game: GameMeta;
  player_id: string;
  deck_name: string;
  warnings?: ApiViolation[];
  // Same meaning as on UploadDeckResponse: cards in the seated deck
  // whose printed rules the engine will not carry out. Always empty
  // for a curated deck (their coverage is build-tested), so this only
  // ever fills on the raw-decklist path.
  unimplemented?: string[];
}

// UploadDeckResponse mirrors lobby.uploadDeckResponse.
export interface UploadDeckResponse {
  game: GameMeta;
  deck_name: string;
  card_count: number;
  // The pre-built deck that was installed; absent for an uploaded
  // list. Lets the picker confirm the seat is holding the deck the
  // player pressed, rather than inferring it from the name.
  deck_id?: string;
  commanders: string[];
  warnings?: ApiViolation[];
  // Distinct names of the accepted deck's cards that print rules the
  // engine will not carry out. Not an error and not a warning — the
  // deck is legal and the game will run; those cards behave as
  // manual sandbox cards. Told here because deck upload is the best
  // moment there is to say it: once, before the game, instead of
  // once per surprise mid-combat.
  unimplemented?: string[];
}

interface SessionResponse {
  token: string;
  expires_at: string;
  principal: Session["principal"];
  game?: GameMeta;
  player_id?: string;
}

// adminLogin exchanges the shared admin token for a session and
// stores it. Throws LobbyApiError on wrong token.
export async function adminLogin(token: string): Promise<Session> {
  const res = await authFetch("/admin/login", {
    method: "POST",
    body: JSON.stringify({ token }),
  });
  const body = (await res.json()) as SessionResponse;
  const s: Session = {
    token: body.token,
    expiresAt: body.expires_at,
    principal: body.principal,
  };
  setSession(s);
  return s;
}

// listGames returns the active tables. Pass { archived: true } for
// the retired ones — a separate view, not a mixed list, so the
// default lobby never grows without bound as old games pile up.
export async function listGames(opts: { archived?: boolean } = {}): Promise<GameMeta[]> {
  const res = await authFetch(opts.archived ? "/games?archived=1" : "/games", { method: "GET" });
  const body = (await res.json()) as { games: GameMeta[] };
  // Defensive normalisation: a buggy server may emit `players: null`
  // for seat-less games (see lobby.copyMeta regression). Iterating
  // `g.players` then throws mid-render and Svelte silently bails on
  // the subtree — "no games yet" keeps showing. Coercing to [] here
  // means the UI degrades to a visible empty-seats row instead of a
  // vanished list.
  return body.games.map((g) => ({ ...g, players: g.players ?? [] }));
}

// GamePreview mirrors lobby.previewResponse: what an invite holder
// may see before joining. Seats come back scrubbed (no player IDs,
// no Discord identity) and both invite tokens are stripped.
export interface GamePreview {
  game: GameMeta;
  invite: "player" | "spectator";
  max_seats: number;
}

// previewGame is unauthenticated — the invite token is the
// credential — so it uses plain fetch: authFetch would read the
// 401 for a bad invite as an expired session and clear it.
export async function previewGame(id: string, inviteToken: string): Promise<GamePreview> {
  const res = await fetch(`/games/${id}/preview?t=${encodeURIComponent(inviteToken)}`, {
    headers: { Accept: "application/json" },
    credentials: "same-origin",
  });
  if (!res.ok) {
    let message = `${res.status} ${res.statusText}`;
    try {
      const body = (await res.json()) as { error?: string };
      if (body.error) message = body.error;
    } catch {
      // not JSON — keep the status line
    }
    throw new LobbyApiError(res.status, message);
  }
  return (await res.json()) as GamePreview;
}

export async function getGame(id: string): Promise<GameMeta> {
  const res = await authFetch(`/games/${id}`, { method: "GET" });
  return (await res.json()) as GameMeta;
}

export async function createGame(name: string): Promise<GameMeta> {
  const res = await authFetch("/games", {
    method: "POST",
    body: JSON.stringify({ name }),
  });
  return (await res.json()) as GameMeta;
}

// joinGame is the public invite-link flow: call with the invite
// token and a display name, receive a fresh RolePlayer session bound
// to (gameID, playerID). The session is installed in the session
// store on success so subsequent API calls use it.
export async function joinGame(
  gameID: string,
  inviteToken: string,
  name: string,
): Promise<Session> {
  const res = await authFetch(`/games/${gameID}/join`, {
    method: "POST",
    body: JSON.stringify({ invite_token: inviteToken, name }),
  });
  const body = (await res.json()) as SessionResponse;
  const s: Session = {
    token: body.token,
    expiresAt: body.expires_at,
    principal: body.principal,
    playerID: body.player_id,
    gameID: body.game?.id,
  };
  setSession(s);
  return s;
}

// joinByCode is the login-page flow: the caller holds an invite code
// but no game id, so the server resolves the table from the code
// itself (POST /join). When a Discord identity session is installed,
// authFetch attaches it and the seat takes its name and avatar from
// Discord; `name` is only needed on the manual path, where there is
// no identity to read them from.
export async function joinByCode(inviteToken: string, name = ""): Promise<Session> {
  const res = await authFetch("/join", {
    method: "POST",
    body: JSON.stringify({ invite_token: inviteToken, name }),
  });
  const body = (await res.json()) as SessionResponse;
  const s: Session = {
    token: body.token,
    expiresAt: body.expires_at,
    principal: body.principal,
    playerID: body.player_id,
    gameID: body.game?.id,
  };
  setSession(s);
  return s;
}

// spectateGame is the read-only counterpart to joinGame: posts the
// per-game spectator invite, receives a RoleSpectator session bound
// to the game (no player_id). The session is installed in the
// session store; the Game route uses session.principal.role to hide
// action affordances. Added in S11.
export async function spectateGame(
  gameID: string,
  inviteToken: string,
  name: string,
): Promise<Session> {
  const res = await authFetch(`/games/${gameID}/spectate`, {
    method: "POST",
    body: JSON.stringify({ invite_token: inviteToken, name }),
  });
  const body = (await res.json()) as SessionResponse;
  const s: Session = {
    token: body.token,
    expiresAt: body.expires_at,
    principal: body.principal,
    playerID: body.player_id,
    gameID: body.game?.id,
  };
  setSession(s);
  return s;
}

// archiveGame retires a table (admin only): it leaves the lobby
// listing, its bots stop, and anyone still connected is dropped —
// but nothing on disk is removed and unarchiveGame is the undo.
export async function archiveGame(id: string): Promise<GameMeta> {
  const res = await authFetch(`/games/${id}/archive`, { method: "POST" });
  return (await res.json()) as GameMeta;
}

// unarchiveGame returns a retired table to the listing, restarting
// its bot runners if it was mid-game.
export async function unarchiveGame(id: string): Promise<GameMeta> {
  const res = await authFetch(`/games/${id}/archive`, { method: "DELETE" });
  return (await res.json()) as GameMeta;
}

// deleteGame destroys a table (admin only) — the lobby entry, the
// engine snapshot, the restore point AND the replay JSONL. There is
// no undo; archiveGame is the reversible one. 204, no body.
export async function deleteGame(id: string): Promise<void> {
  await authFetch(`/games/${id}`, { method: "DELETE" });
}

// ReclaimTicket mirrors lobby.reclaimTicketResponse. `ticket` is a
// bearer credential for one player's seat: anyone holding it can
// play as that player until it is redeemed or expires. Show the
// expiry next to the link so whoever hands it out knows exactly what
// they just gave away.
export interface ReclaimTicket {
  ticket: string;
  game_id: string;
  player_id: string;
  seat: number;
  player_name: string;
  expires_at: string;
  ttl_seconds: number;
  single_use: boolean;
}

// mintSeatReclaim asks the server for a one-shot link that puts a
// disconnected player back in their own seat. Admin only — the
// server enforces it; this just fails loudly if called otherwise.
export async function mintSeatReclaim(gameID: string, playerID: string): Promise<ReclaimTicket> {
  const res = await authFetch(`/games/${gameID}/seats/${playerID}/reclaim`, { method: "POST" });
  return (await res.json()) as ReclaimTicket;
}

// redeemSeatReclaim is the public half, called by the Reclaim route
// when a player follows the link. Plain fetch rather than authFetch
// for the same reason previewGame uses one: the caller has no
// session (that is the whole problem), and a 401 here means "bad
// ticket", not "your session expired".
export async function redeemSeatReclaim(gameID: string, ticket: string): Promise<Session> {
  const res = await fetch(`/games/${gameID}/reclaim`, {
    method: "POST",
    headers: { "Content-Type": "application/json", Accept: "application/json" },
    credentials: "same-origin",
    body: JSON.stringify({ ticket }),
  });
  if (!res.ok) {
    let message = `${res.status} ${res.statusText}`;
    try {
      const body = (await res.json()) as { error?: string };
      if (body.error) message = body.error;
    } catch {
      // not JSON — keep the status line
    }
    throw new LobbyApiError(res.status, message);
  }
  const body = (await res.json()) as SessionResponse;
  const s: Session = {
    token: body.token,
    expiresAt: body.expires_at,
    principal: body.principal,
    playerID: body.player_id,
    gameID: body.game?.id,
  };
  setSession(s);
  return s;
}

export async function startGame(id: string): Promise<GameMeta> {
  const res = await authFetch(`/games/${id}/start`, { method: "POST" });
  return (await res.json()) as GameMeta;
}

// fetchBotOptions reads the tier + curated-deck catalog for the
// Add-bot picker. Game-independent: the same answer for every table,
// so the lobby fetches it once on mount. Added in S31.
export async function fetchBotOptions(): Promise<BotOptions> {
  const res = await authFetch("/bot/options");
  return (await res.json()) as BotOptions;
}

// addBotSeat seats a bot at an unstarted table. Allowed for any
// player already seated there, and for admin — bots take real seats,
// so a seated human can add at most three. Added in S31.
export async function addBotSeat(
  gameID: string,
  body: { tier: string; deck?: string; name?: string; format?: string; source?: string },
): Promise<AddBotResponse> {
  const res = await authFetch(`/games/${gameID}/seats/bot`, {
    method: "POST",
    body: JSON.stringify(body),
  });
  return (await res.json()) as AddBotResponse;
}

// removeBotSeat unseats a bot while the table is still in the lobby.
// Only bot seats can be removed this way. Added in S31.
export async function removeBotSeat(gameID: string, playerID: string): Promise<GameMeta> {
  const res = await authFetch(`/games/${gameID}/seats/bot/${playerID}`, { method: "DELETE" });
  return (await res.json()) as GameMeta;
}

// replayURL returns an authenticated download URL for a game's
// replay JSONL. Ships the session token as ?token= because browser
// downloads can't set Authorization headers. Consumed by the lobby
// "download replay" link via a plain <a href>. Added in S11.
export function replayURL(gameID: string): string | null {
  const s = currentSession();
  if (!s?.token) return null;
  return `/games/${gameID}/replay?token=${encodeURIComponent(s.token)}`;
}

// uploadDeck ships a decklist (plain text or Moxfield JSON) to the
// server for parse + validate + install. Format can be omitted — the
// server auto-detects by checking the first non-whitespace byte for
// '{' (Moxfield) vs anything else (text).
//
// On validation failure (422), authFetch throws a LobbyApiError whose
// `.violations` carries the full structured list (one per offending
// card) and `.warnings` carries non-fatal advisories (e.g. sideboard
// ignored). Callers that want to highlight individual rows should
// read `.violations`; the plain `.message` is the human-readable
// summary.
export async function uploadDeck(
  gameID: string,
  playerID: string,
  source: string,
  format?: "text" | "moxfield",
): Promise<UploadDeckResponse> {
  const res = await authFetch(`/games/${gameID}/decks`, {
    method: "POST",
    body: JSON.stringify({ player_id: playerID, source, format }),
  });
  return (await res.json()) as UploadDeckResponse;
}

// fetchPrebuiltDecks reads the pre-built deck catalog and each deck's
// engine-coverage profile (GET /decks). Game-independent — the same
// answer for every table — so callers fetch it once on mount.
//
// authFetch, not plain fetch: the route is session-gated, unlike the
// public /catalog. A logged-out caller gets a 401 and the picker
// simply does not render, which is correct — there is no seat to
// install a deck into.
export async function fetchPrebuiltDecks(): Promise<PrebuiltDecksResponse> {
  const res = await authFetch("/decks");
  return (await res.json()) as PrebuiltDecksResponse;
}

// installPrebuiltDeck seats the caller with one of the pre-built
// decks. Same route, same response and the same 422-with-violations
// failure shape as uploadDeck, because it IS uploadDeck — the server
// expands the deck ID to that deck's decklist text and runs the one
// parse -> resolve -> validate -> install pipeline. There is
// deliberately no second install path.
export async function installPrebuiltDeck(
  gameID: string,
  playerID: string,
  deckID: string,
): Promise<UploadDeckResponse> {
  const res = await authFetch(`/games/${gameID}/decks`, {
    method: "POST",
    body: JSON.stringify({ player_id: playerID, deck: deckID }),
  });
  return (await res.json()) as UploadDeckResponse;
}

// avatarURL builds the cached-Discord-avatar URL for a (discord_id,
// avatar_hash) pair. Returns null when either value is missing so
// callers can use it as a render gate:
//   const url = avatarURL(seat.discord_id, seat.discord_avatar_hash);
//   {#if url}<img src={url}>{/if}
// The endpoint is session-gated and <img> tags can't set an
// Authorization header, so the token rides as ?token= (same pattern
// as replayURL) — the session cookie alone isn't reliable (Secure-
// flag mismatches, cleared cookies with a live localStorage session).
// The /avatars handler 503s when the disk cache is unconfigured;
// callers that care should treat a 503 as "fall back to initials".
export function avatarURL(discordID?: string, avatarHash?: string): string | null {
  if (!discordID || !avatarHash) return null;
  const base = `/avatars/${encodeURIComponent(discordID)}/${encodeURIComponent(avatarHash)}.png`;
  const token = currentSession()?.token;
  return token ? `${base}?token=${encodeURIComponent(token)}` : base;
}

// discordAuthEnabled probes /auth/discord/config and reports whether
// the server has the Discord OAuth three-tuple configured. Used by
// Join.svelte to decide whether to render the "Sign in with Discord"
// button. A 503 / network failure → false; the manual-name fallback
// is always safe, so a down probe shouldn't block the join flow.
export async function discordAuthEnabled(): Promise<boolean> {
  try {
    const res = await fetch("/auth/discord/config", {
      method: "GET",
      credentials: "same-origin",
    });
    if (!res.ok) return false;
    const body = (await res.json()) as { enabled?: boolean };
    return body.enabled === true;
  } catch {
    return false;
  }
}

// discordLoginHref is the login-page entry into the OAuth flow. No
// game, no invite: the server mints an identity-only session and
// bounces back to #/oauth-complete with no game in the fragment.
//
// A plain href rather than a fetch, for the same reason as the Join
// page's variant — the server answers with a 302 to Discord, and
// only a real navigation lands the user on the consent screen.
export function discordLoginHref(): string {
  return "/auth/discord/start";
}

// BugReportConfig mirrors the JSON from GET /bugreport/config.
//
// `attachments` is separate from `enabled` because the two halves fail
// independently: a server with a GitHub token but no data dir (or no
// public base URL) files text reports perfectly well and simply can't
// host screenshots. The modal hides its file picker in that case rather
// than offering an upload that would 503.
export interface BugReportConfig {
  enabled: boolean;
  attachments: boolean;
  maxImages: number;
  maxBytes: number;
}

const BUG_REPORT_DISABLED: BugReportConfig = {
  enabled: false,
  attachments: false,
  maxImages: 0,
  maxBytes: 0,
};

// fetchBugReportConfig probes whether the server can file GitHub issues
// and whether it can store attachments. Used by Game.svelte to decide
// whether to render the report-a-bug button at all. Same posture as
// discordAuthEnabled: any failure → disabled, the button simply
// doesn't render.
export async function fetchBugReportConfig(): Promise<BugReportConfig> {
  try {
    const res = await fetch("/bugreport/config", {
      method: "GET",
      credentials: "same-origin",
    });
    if (!res.ok) return BUG_REPORT_DISABLED;
    const body = (await res.json()) as {
      enabled?: boolean;
      attachments?: boolean;
      max_images?: number;
      max_bytes?: number;
    };
    return {
      enabled: body.enabled === true,
      attachments: body.attachments === true,
      maxImages: body.max_images ?? 0,
      maxBytes: body.max_bytes ?? 0,
    };
  } catch {
    return BUG_REPORT_DISABLED;
  }
}

// BugReportResult mirrors the 201 body of POST /bugreport: the
// created issue's URL + number, so the modal can link straight to it.
// `reportID` is present only when the report stored artifacts — it is
// the key the operator uses to pull the pinned replay.
export interface BugReportResult {
  url: string;
  number: number;
  reportID?: string;
  // The label the issue actually landed with. Absent when the label
  // was rejected upstream and the server re-filed unlabelled — so the
  // modal says "filed as enhancement" only when that is true.
  label?: string;
}

// BugReportDraft is everything a report can carry.
export interface BugReportDraft {
  title: string;
  // kind picks the GitHub label and what the report attaches. Omitted
  // → the server files a bug, which is what clients predating the
  // picker send.
  kind?: BugReportKind;
  description: string;
  context?: BugReportContext;
  log?: BugLogEntry[];
  images?: File[];
}

// submitBugReport files an in-app bug report. The server renders the
// issue body and talks to GitHub; the browser never sees a token.
//
// Two encodings, chosen by whether there are images: multipart when
// there are (files can't ride in JSON without base64 inflating them by
// a third), plain JSON when there aren't. The JSON shape is the
// documented one and the one a curl-wielding operator will reach for,
// so it stays the default rather than becoming a legacy path.
//
// Throws LobbyApiError on 400 (validation), 413 (too large), 429 (rate
// limit), 502 (GitHub upstream failure), 503 (feature disabled).
//
// Every text field is redacted before it leaves the browser (#721): a
// reporter who pastes their invite link into the description, or a log
// line that slipped past the buffers, must not publish a credential in
// the issue. The server redacts again, for clients older than this.
export async function submitBugReport(draft: BugReportDraft): Promise<BugReportResult> {
  const report = {
    title: redactSecrets(draft.title),
    kind: draft.kind,
    description: redactSecrets(draft.description),
    context: redactBugContext(draft.context),
    log:
      draft.log && draft.log.length > 0
        ? draft.log.map((e) => ({ ...e, text: redactSecrets(e.text) }))
        : undefined,
  };
  let body: BodyInit;
  if (draft.images && draft.images.length > 0) {
    const form = new FormData();
    form.append("report", JSON.stringify(report));
    for (const f of draft.images) {
      form.append("image", f, f.name);
    }
    body = form;
  } else {
    body = JSON.stringify(report);
  }
  const res = await authFetch("/bugreport", { method: "POST", body });
  const parsed = (await res.json()) as {
    url: string;
    number: number;
    report_id?: string;
    label?: string;
  };
  return {
    url: parsed.url,
    number: parsed.number,
    reportID: parsed.report_id,
    label: parsed.label,
  };
}

// AutoTapPreview mirrors the JSON returned by
// `GET /games/:id/auto-tap-preview` (server/internal/lobby/http.go).
// `ok=false` means the auto-tapper couldn't satisfy the cost; the
// caller should keep showing the missing list and disable the
// "Auto-tap & cast" submit affordance. `plan` is the ordered list
// of permanent UUIDs the server proposes to tap; the modal renders
// them in the order returned (colored requirements first, generic
// recruits second). Added in S15 sub-PR 5.
export interface AutoTapPreview {
  ok: boolean;
  plan?: string[];
  missing?: string[];
  cost: string;
}

// fetchAutoTapPreview asks the server which permanents the auto-
// tapper would tap to cast `cardID` right now. Read-only — calling
// it does not mutate game state. `excluded` lets the caller pass
// the lock-tap UI's reservation list. `xValue` is the announced
// X for spells with {X} in their cost (defaults to 0), and
// `phyrexianLife` the announced Phyrexian-symbol count (#916).
export async function fetchAutoTapPreview(
  gameID: string,
  cardID: string,
  opts: {
    xValue?: number;
    excluded?: string[];
    abilityIndex?: number;
    phyrexianLife?: number;
    // #696: the announce-time half of the cast being previewed, as
    // castPreviewParams builds it. Every field changes the PRICE, so
    // a preview that omits them answers about a different cast than
    // the one the confirm button will send — a flashback cast priced
    // at the printed cost, an airbent permanent priced at {5}{R}{R}
    // instead of the grant's {2}.
    cast?: AutoTapCastParams;
  } = {},
): Promise<AutoTapPreview> {
  const params = new URLSearchParams({ card: cardID });
  if (opts.xValue && opts.xValue > 0) {
    params.set("x", String(opts.xValue));
  }
  const cast = opts.cast;
  if (cast) {
    if (cast.fromZone) params.set("from_zone", cast.fromZone);
    if (cast.alternativeCost) params.set("alternative_cost", cast.alternativeCost);
    if (cast.optionalCosts && cast.optionalCosts.length > 0) {
      params.set("optional_costs", cast.optionalCosts.join(","));
    }
    if (cast.tapIDs && cast.tapIDs.length > 0) {
      params.set("tap_ids", cast.tapIDs.join(","));
    }
    if (cast.face) params.set("face", String(cast.face));
  }
  // `abilityIndex` prices a CR 602 activated ability's own mana
  // component instead of the card's printed cast cost. Without it
  // the X picker an ability opens would report on the {4} in Helm of
  // Obedience's corner rather than on the {X} being announced.
  if (opts.abilityIndex !== undefined) {
    params.set("ability", String(opts.abilityIndex));
  }
  // #916: the Phyrexian symbols the announcement will pay with 2 life
  // each. The server strikes them before planning, so the preview
  // reports on the MANA the cast or activation still owes rather than
  // tapping a land for a pip the player just said they would buy.
  if (opts.phyrexianLife && opts.phyrexianLife > 0) {
    params.set("phyrexian", String(opts.phyrexianLife));
  }
  if (opts.excluded && opts.excluded.length > 0) {
    params.set("exclude", opts.excluded.join(","));
  }
  const res = await authFetch(`/games/${gameID}/auto-tap-preview?${params.toString()}`, {
    method: "GET",
  });
  return (await res.json()) as AutoTapPreview;
}

// logout revokes the current session server-side and clears local
// state. Best-effort: a network failure still clears the store so
// the user isn't stranded in a half-logged-out UI. We bypass
// authFetch because its 401 handler would double-clear the session
// and throw — /logout accepts stale credentials and always returns
// 204, so there's nothing to interpret from the body.
export async function logout(): Promise<void> {
  const s = currentSession();
  try {
    await fetch("/logout", {
      method: "POST",
      headers: s?.token ? { Authorization: `Bearer ${s.token}` } : {},
      credentials: "same-origin",
    });
  } catch {
    // Swallow network errors — we still want to drop the local
    // session so the UI recovers.
  }
  setSession(null);
}

// logoutEverywhere withdraws every session the signed-in user holds,
// in every browser, and then drops this one (POST /logout/everywhere,
// ADR 0051 decision 6). Only meaningful when canSignOutEverywhere is
// true: the server refuses a session with no user.
//
// Unlike logout, a failure is reported: the user asked for their
// OTHER browsers to be signed out, and clearing the local session
// quietly would tell them that happened when it may not have. The
// local session is still dropped on success and on a 401, since a 401
// means this session is already dead (revoked from another browser,
// or expired). It bypasses authFetch for the same reason logout does:
// a 401 here is an answer, not a reason to throw "session expired".
export async function logoutEverywhere(): Promise<void> {
  const s = currentSession();
  const res = await fetch("/logout/everywhere", {
    method: "POST",
    headers: s?.token ? { Authorization: `Bearer ${s.token}` } : {},
    credentials: "same-origin",
  });
  if (res.ok || res.status === 401) {
    setSession(null);
    return;
  }
  let message = `${res.status} ${res.statusText}`;
  try {
    const body = (await res.json()) as { error?: string };
    if (body.error) message = body.error;
  } catch {
    // not JSON — keep the status line
  }
  throw new LobbyApiError(res.status, message);
}

// --- develop-environment card spawner (ADR 0023) ---------------------
//
// These call routes that exist only on a dev deployment. In
// production they 404, which is the intended answer: the UI that
// reaches them is gated on the card_spawn feature from /config, so
// nothing should be calling them there in the first place.

export interface DevCardResult {
  id: string;
  name: string;
  type_line: string;
  mana_cost: string;
  set: string;
}

export type DevSpawnZone = "battlefield" | "hand" | "graveyard" | "exile" | "library" | "command";

// searchDevCards queries the Scryfall index by name.
//
// Swallows every failure into an empty list: this runs on a debounce
// as the user types, so a 404 (wrong environment), a 503 (index not
// loaded yet), or an aborted in-flight request are all "nothing to
// show right now" rather than something to interrupt typing with.
// The spawn call is where a real error gets surfaced.
export async function searchDevCards(q: string, signal?: AbortSignal): Promise<DevCardResult[]> {
  try {
    const res = await authFetch(`/dev/cards?q=${encodeURIComponent(q)}`, {
      method: "GET",
      signal,
    });
    const body = (await res.json()) as { cards?: DevCardResult[] };
    return body.cards ?? [];
  } catch {
    return [];
  }
}

export interface DevSpawnRequest {
  scryfallID?: string;
  name?: string;
  playerID: string;
  zone: DevSpawnZone;
  count?: number;
  commander?: boolean;
}

export interface DevSpawnResult {
  spawned: string[];
  name: string;
  zone: string;
  count: number;
}

// spawnDevCard puts cards into a zone. Lets authFetch's LobbyApiError
// propagate so the panel can render the server's own message —
// "card index not loaded", "game must be active to spawn cards" and
// "spawn count must be between 1 and 20" are all things the user
// needs to read verbatim rather than as a generic failure.
export async function spawnDevCard(gameID: string, req: DevSpawnRequest): Promise<DevSpawnResult> {
  const res = await authFetch(`/games/${encodeURIComponent(gameID)}/dev/spawn`, {
    method: "POST",
    body: JSON.stringify({
      scryfall_id: req.scryfallID,
      name: req.name,
      player_id: req.playerID,
      zone: req.zone,
      count: req.count,
      commander: req.commander,
    }),
  });
  return (await res.json()) as DevSpawnResult;
}

// fetchReplay downloads the per-game replay log (JSONL, one snapshot
// per line — see lib/replay.ts). The route is not dev-only: it has
// existed since S11 and is admin-gated for a game still in progress,
// player-accessible once the game has ended. The dev scrubber is just
// its first interactive consumer.
//
// A 204 means the route is there but the game has produced no Apply
// yet; returns an empty string so the caller renders "no frames"
// rather than an error.
export async function fetchReplay(gameID: string): Promise<string> {
  const res = await authFetch(`/games/${encodeURIComponent(gameID)}/replay`, { method: "GET" });
  if (res.status === 204) return "";
  return await res.text();
}
