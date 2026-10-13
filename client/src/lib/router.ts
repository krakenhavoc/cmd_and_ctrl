import { type Readable } from "svelte/store";

import { parseAdminView, type AdminView } from "./adminViews";
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
  // The admin views (ADR 0124 §7): Live now, Games, one table, Accounts
  // and one account, under #/admin/<view>. #/admin alone stays the
  // shared token's form (adminLogin). `view` carries the filters the
  // hash names; a value the server does not know is dropped here, so
  // it is never sent (lib/adminViews.ts).
  | { name: "adminViews"; view: AdminView }
  // #/home was the site portal (#1386, ADR 0092). ADR 0156 §4 retired
  // it: the front page serves visitors and the header everyone else.
  // The route survives only so old links land somewhere: the router
  // sends it on (signedInHome.ts's routeRedirect).
  | { name: "home" }
  // The gate page (ADR 0156 §5): where the Discord callback sends a
  // sign-in the role gate refused. `retry` is ?retry=1, the callback's
  // word that Discord did not answer the check. Public.
  | { name: "requestAccess"; retry?: true }
  // The public roadmap (PR 2 of the same plan): what the rules engine
  // automates today, what's partial, what's missing, and what's next.
  // Public, like the catalogue (ADR 0092).
  | { name: "roadmap" }
  // The privacy notice (#2983, ADR 0149): what is stored, why, for how
  // long, who it is shared with, and how a signed-in person deletes
  // their account. Public: a stranger deciding whether to sign in has
  // to be able to read it first, and the page holds no account data.
  | { name: "privacy" }
  // The terms of service (#3358): who runs the site, who may use it, and
  // the rules for using it. Public, like the privacy notice: a stranger
  // reads them before signing in.
  | { name: "terms" }
  // The About page (#2981): what this project is, what the engine does and
  // does not do for you, where to send feedback, and the credits. Public:
  // it is what a stranger reads before deciding to sign in.
  | { name: "about" }
  // The getting-started guide: getting in, finding a pod, a table, the
  // C2 Discord's channels and commands. Public like About, for the same
  // reader; the Discord's #welcome posts say the same things.
  | { name: "guide" }
  // The one decks page (ADR 0112 §3): check a deck by link or paste,
  // its coverage report, "Request missing cards", "Save to my decks",
  // the saved library and the pre-built decks. Public like the roadmap:
  // the report carries names, oracle IDs and caveat sentences, never
  // art or oracle text, and the library shows only to its owner.
  //
  // `url` is #/decks?url=<link>, which pre-fills the link field and
  // runs the check on load.
  //
  // #/deck-check (ADR 0095 §5's page) parses to this same route,
  // permanently: the Discord bot's `/c2-deck-check` replies link
  // #/deck-check?url=<link> for the full report, and every reply
  // already posted must keep working. router.test.ts pins it.
  | { name: "decks"; url?: string }
  // Tables, the signed-in home. `show: "history"` scrolls to its History
  // section (ADR 0158 §1 item 5): #/lobby?show=history, where the old
  // #/my-games redirects.
  | { name: "lobby"; show?: "history" }
  // The tutorial's practice table (ADR 0076 §2.2): opening this route
  // opens a fresh practice game for the signed-in session and moves on
  // to it (routes/Practice.svelte, lib/practiceTable.ts). It is the
  // one door the lobby offer and Settings' replay (sub-PR 5) both use.
  | { name: "practice" }
  // serverFull is set only when the Discord callback sent the browser
  // back here because the server is at its connection cap (ADR 0148,
  // #3103): `?error=server_full`. Absent otherwise, so a plain join
  // route compares equal to what it always did.
  | { name: "join"; gameID: string; inviteToken: string; spectator: boolean; serverFull?: true }
  // Seat reclaim: an admin-minted, single-use, short-lived link that
  // puts a disconnected player back in their OWN seat at a table that
  // has already started. Public like /join — the ticket is the
  // credential, because the holder by definition has no session.
  | { name: "reclaim"; gameID: string; ticket: string }
  | { name: "game"; gameID: string }
  // The card catalogue: every card the engine automates, and how
  // completely. Public, like the server's /catalog route (ADR 0042):
  // with sign-up gated on a Discord role (ADR 0156), a session here
  // would put Scryfall data behind a chat server (S70, #2985).
  //
  // `query` is #/catalog?q=<text>, which pre-fills the search box. The
  // roadmap links a signed-in viewer's example card names here.
  | { name: "catalog"; query?: string }
  // The retired "My games" page (ADR 0051 decision 4, S34). Its list is
  // Tables' History now (ADR 0158 §1 item 5), and #/my-games only
  // redirects there (signedInHome.routeRedirect), so a bookmark lands.
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

// "login" is the front page (ADR 0156 §4): #/ and #/login both open it.
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
    case "admin": {
      // #/admin alone, or a path under it that names no view, is the
      // token's form, as every #/admin/… was before ADR 0124.
      const view = parseAdminView(parts.slice(1), params);
      return view ? { name: "adminViews", view } : { name: "adminLogin" };
    }
    case "home":
      return { name: "home" };
    case "request-access":
      return params.get("retry") === "1"
        ? { name: "requestAccess", retry: true }
        : { name: "requestAccess" };
    case "roadmap":
      return { name: "roadmap" };
    case "privacy":
      return { name: "privacy" };
    case "terms":
      return { name: "terms" };
    case "about":
      return { name: "about" };
    case "guide":
      return { name: "guide" };
    // #/deck-check is a permanent alias of #/decks (see the Route type).
    case "decks":
    case "deck-check": {
      const u = params.get("url");
      return u ? { name: "decks", url: u } : { name: "decks" };
    }
    case "lobby":
      return params.get("show") === "history"
        ? { name: "lobby", show: "history" }
        : { name: "lobby" };
    case "practice":
      return { name: "practice" };
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
          ...(params.get("error") === "server_full" ? { serverFull: true as const } : {}),
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
