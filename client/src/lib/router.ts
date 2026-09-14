import { writable, type Readable } from "svelte/store";

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
  | { name: "catalog" }
  // S12.5: /auth/discord/callback (server-side) redirects here with
  // the session details in the URL fragment. App.svelte's effect
  // reads them, installs the session, and navigates onward.
  | { name: "oauthComplete"; token: string; gameID: string; playerID: string; expiresAt: string };

const defaultRoute: Route = { name: "login" };

function parseHash(hash: string): Route {
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
    case "lobby":
      return { name: "lobby" };
    case "catalog":
      return { name: "catalog" };
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
    case "oauth-complete":
      // Server fragment carries the four Session fields. Any
      // missing → we fall through to the login route rather than
      // install a half-populated session.
      if (
        params.get("token") &&
        params.get("game") &&
        params.get("player_id") &&
        params.get("expires_at")
      ) {
        return {
          name: "oauthComplete",
          token: params.get("token")!,
          gameID: params.get("game")!,
          playerID: params.get("player_id")!,
          expiresAt: params.get("expires_at")!,
        };
      }
      return defaultRoute;
    default:
      return defaultRoute;
  }
}

const store = writable<Route>(parseHash(location.hash));

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
