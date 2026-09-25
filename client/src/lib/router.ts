import { type Readable } from "svelte/store";

import { guardedWritable } from "./guardedStore";

// Route is the discriminated union of client-side views. The router
// derives one of these from location.hash and re-derives on
// hashchange. Components read `route` as a readable store.
//
// We pick hash routing over path routing because the client is served
// as a single static SPA — every path (invite links included) funnels
// through Vite's index.html and the hash selects the view client-side.
// That sidesteps a server-side router and keeps deep-linking cheap.
export type Route =
  | { name: "login" }
  | { name: "adminLogin" }
  // The site portal (#1386, ADR 0092): a public sitemap of link cards
  // to everywhere else on the site, grouped by what you're trying to
  // do. Exists alongside SiteHeader as a second entry point — the
  // owner will pick between the two later.
  | { name: "home" }
  // The public roadmap (PR 2 of the same plan): what the rules engine
  // automates today, what's partial, what's missing, and what's next.
  // Public like /catalog was meant to be closed — see ADR 0092 for why
  // this one is the opposite call.
  | { name: "roadmap" }
  // The public deck coverage checker (ADR 0095 §5, PR 3 of the same
  // plan): how much of a decklist the engine automates, plus a
  // "Request these cards" filing button for a signed-in Discord
  // session. Public like the roadmap — the body carries names, oracle
  // IDs and caveat sentences, never art or oracle text.
  //
  // `url` is #/deck-check?url=<link>, which pre-fills the link field
  // and runs the check on load. The bot's `/c2-deck-check` reply and
  // the "up to about 15 names" summary link here for the full report.
  | { name: "deckCheck"; url?: string }
  | { name: "lobby" }
  | { name: "join"; gameID: string; inviteToken: string; spectator: boolean }
  // Seat reclaim: an admin-minted, single-use, short-lived link that
  // puts a disconnected player back in their OWN seat at a table that
  // has already started. Public like /join — the ticket is the
  // credential, because the holder by definition has no session.
  | { name: "reclaim"; gameID: string; ticket: string }
  | { name: "game"; gameID: string }
  // The card catalogue: every card the engine automates, and how
  // completely. Session-gated, like the server's /catalog route —
  // it was built as an anonymous showcase and deliberately closed,
  // because AGENTS.md §1/§8 describe this project as private and
  // personal-use and serving card art to signed-out visitors is a
  // different posture from the one the repo states.
  //
  // `query` is #/catalog?q=<text>, which pre-fills the search box. The
  // roadmap links a signed-in viewer's example card names here.
  | { name: "catalog"; query?: string }
  // "My games" (ADR 0051 decision 4, S34): every seat the signed-in
  // person holds, with a way back into the open ones. Session-gated;
  // the page itself explains what a guest or admin session is missing.
  | { name: "myGames" }
  // S12.5: /auth/discord/callback (server-side) redirects here with
  // the session details in the URL fragment. App.svelte's effect
  // reads them, installs the session, and navigates onward.
  //
  // Two shapes, matching the two ways the round-trip can start.
  // From an invite link the fragment carries game + player_id and
  // the user goes straight to the table. From the login page it
  // carries neither — the session is identity-only — and the user
  // lands back on login to type an invite code, with displayName
  // there to show who they signed in as.
  | {
      name: "oauthComplete";
      token: string;
      expiresAt: string;
      gameID?: string;
      playerID?: string;
      displayName?: string;
      // userID is the server's users row (S34), present when it has a
      // user database. It gates "sign out everywhere" and "My games".
      userID?: string;
    };

const defaultRoute: Route = { name: "login" };

// parseHash is exported for tests; everything else reads `route`.
export function parseHash(hash: string): Route {
  const raw = hash.replace(/^#\/?/, "");
  if (!raw) return defaultRoute;

  const [pathPart, queryPart = ""] = raw.split("?");
  const parts = pathPart.split("/").filter(Boolean);
  const params = new URLSearchParams(queryPart);

  switch (parts[0]) {
    case "login":
      return { name: "login" };
    case "admin":
      return { name: "adminLogin" };
    case "home":
      return { name: "home" };
    case "roadmap":
      return { name: "roadmap" };
    case "deck-check": {
      const u = params.get("url");
      return u ? { name: "deckCheck", url: u } : { name: "deckCheck" };
    }
    case "lobby":
      return { name: "lobby" };
    case "catalog": {
      const q = params.get("q");
      return q ? { name: "catalog", query: q } : { name: "catalog" };
    }
    case "my-games":
      return { name: "myGames" };
    case "games":
      // /games/:id/join?t=<token> → Join
      // /games/:id                → Game
      if (parts.length >= 3 && parts[2] === "join") {
        // ?spectator=1 toggles the page from a player join (claims a
        // seat) to a spectator join (read-only watch). Both flows
        // share the same Join.svelte route — the UI swaps copy + the
        // outbound API call based on this flag.
        return {
          name: "join",
          gameID: parts[1],
          inviteToken: params.get("t") ?? "",
          spectator: params.get("spectator") === "1",
        };
      }
      if (parts.length >= 3 && parts[2] === "reclaim") {
        return { name: "reclaim", gameID: parts[1], ticket: params.get("t") ?? "" };
      }
      if (parts.length >= 2) {
        return { name: "game", gameID: parts[1] };
      }
      return { name: "lobby" };
    case "oauth-complete": {
      // token + expires_at are the minimum for any session. game and
      // player_id arrive together on the invite-link variant and are
      // both absent on the login-page one; one without the other is
      // a malformed fragment, so it falls through to login rather
      // than installing a half-populated session.
      const token = params.get("token");
      const expiresAt = params.get("expires_at");
      const gameID = params.get("game");
      const playerID = params.get("player_id");
      if (!token || !expiresAt) return defaultRoute;
      if (Boolean(gameID) !== Boolean(playerID)) return defaultRoute;
      return {
        name: "oauthComplete",
        token,
        expiresAt,
        gameID: gameID ?? undefined,
        playerID: playerID ?? undefined,
        displayName: params.get("name") ?? undefined,
        userID: params.get("user_id") ?? undefined,
      };
    }
    default:
      return defaultRoute;
  }
}

const store = guardedWritable<Route>(
  parseHash(typeof location === "undefined" ? "" : location.hash),
  "route",
);

if (typeof window !== "undefined") {
  window.addEventListener("hashchange", () => {
    store.set(parseHash(location.hash));
  });
}

export const route: Readable<Route> = store;

// navigate updates the URL hash, triggering the hashchange listener
// above. Kept as a helper so components don't reach into location
// directly; makes the set of navigation targets auditable.
export function navigate(hash: string): void {
  if (!hash.startsWith("#")) hash = "#" + hash;
  if (location.hash === hash) {
    // Force re-parse even if hash is unchanged (e.g. clicking the
    // same link twice).
    store.set(parseHash(hash));
    return;
  }
  location.hash = hash;
}

// inviteURL builds the full invite URL for sharing. Used by the
// lobby UI's "copy invite link" button.
export function inviteURL(gameID: string, token: string): string {
  return `${location.origin}${location.pathname}#/games/${gameID}/join?t=${encodeURIComponent(token)}`;
}

// spectatorInviteURL builds the spectator-flow variant of inviteURL.
// Same Join.svelte route, but the ?spectator=1 query flag flips the
// page into the read-only join path (calls /games/{id}/spectate
// instead of /games/{id}/join).
export function spectatorInviteURL(gameID: string, token: string): string {
  return `${location.origin}${location.pathname}#/games/${gameID}/join?t=${encodeURIComponent(token)}&spectator=1`;
}

// reclaimURL builds the seat-reclaim link an admin hands to a player
// who lost their session. Same hash-route shape as the invite links,
// with the ticket in ?t= — but note the difference in kind: an
// invite admits anyone who holds it to an OPEN seat, while this one
// hands over a SPECIFIC player's seat, hidden information included.
// Treat it like a password, give it to one person, and let it
// expire.
export function reclaimURL(gameID: string, ticket: string): string {
  return `${location.origin}${location.pathname}#/games/${gameID}/reclaim?t=${encodeURIComponent(ticket)}`;
}
