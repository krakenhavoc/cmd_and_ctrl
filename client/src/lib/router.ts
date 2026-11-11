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
  // Discord Linked Roles (ADR 0163 §3): where the server's verification
  // URL lands, with how the link ended. Public: linking mints no
  // session, and "signin" is for someone with no account.
  | { name: "linkedRole"; status: LinkedRoleStatus }
  // The public roadmap (PR 2 of the same plan): what the rules engine
  // automates today, what's partial, what's missing, and what's next.
  // Public, like the catalogue (ADR 0092).
  | { name: "roadmap" }
  // Which sets can be drafted (ADR 0222 §4, #5437), at #/draft/sets.
  // Public: it is card data, like the catalogue.
  | { name: "draftSets" }
  // A booster draft (ADR 0222 §5, #5438). #/draft (and #/draft/new) is
  // the create form; `set` is ?set=<code>, the set the sets page's
  // "Draft this set" link chose. #/draft/<id> is one draft: its lobby,
  // the picks and the finished pool. `invite` is ?invite=<code>, the
  // link a host shares, which opens the join page for a browser that
  // holds no seat there. Both are public: a visitor drafts against bots
  // with no sign-in, as they play a practice game.
  | { name: "draftNew"; set?: string }
  // A signed-in person's finished drafts (ADR 0222 §9, #5440): the list
  // at #/drafts and one draft, read-only, at #/drafts/<id>.
  | { name: "myDrafts"; id?: string }
  | { name: "draft"; id: string; invite?: string }
  // The privacy notice (#2983, ADR 0149): what is stored, why, for how
  // long, who it is shared with, and how a signed-in person deletes
  // their account. Public: a stranger deciding whether to sign in has
  // to be able to read it first, and the page holds no account data.
  | { name: "privacy" }
  // The terms of service (#3358): who runs the site, who may use it, and
  // the rules for using it. Public, like the privacy notice: a stranger
  // reads them before signing in.
  | { name: "terms" }
  // The funding page (#3558): how CMD & CTRL is funded. Public, like the
  // terms and the privacy notice.
  | { name: "funding" }
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
  // Your own game statistics (ADR 0165 §7, #3366): record, average
  // place, by deck and by commander, recent games. Signed-in only.
  | { name: "stats" }
  // Friends (ADR 0213 §9, slice 3): requests sent, blocks and the two
  // social settings, beyond what the header's drawer shows. Signed-in only.
  | { name: "friends" }
  // Pods (ADR 0220 §11, slice 3): #/pods, a pod at #/pods/<id>, and the
  // invite link at #/pods/join/<token>. The list and a pod are signed-in
  // only; the link page is public, because a signed-out visitor is told
  // to sign in there and nothing about the pod is shown (§3b).
  | { name: "pods" }
  | { name: "pod"; id: string }
  | { name: "podStats"; id: string }
  | { name: "podJoin"; token: string }
  // A person's shared stats (ADR 0217 §6): #/stats/shared/<token>, a
  // read-only copy of their stats pages. Public: the link is the
  // credential, and it is made to be posted in Discord.
  | { name: "statsShared"; token: string }
  // One deck's advanced stats (ADR 0217 §8): #/stats/deck/<deck_id>, and
  // the same page inside a shared view, #/stats/shared/<token>/deck/<deck_id>.
  | { name: "statsDeck"; id: string }
  | { name: "statsSharedDeck"; token: string; id: string }
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
    case "linked-role":
      return { name: "linkedRole", status: linkedRoleStatus(params.get("status")) };
    case "request-access":
      return params.get("retry") === "1"
        ? { name: "requestAccess", retry: true }
        : { name: "requestAccess" };
    case "roadmap":
      return { name: "roadmap" };
    case "draft": {
      // "sets" and "new" are pages of their own, so neither is ever a
      // draft's id (the server's ids are uuids).
      if (parts[1] === "sets") return { name: "draftSets" };
      if (!parts[1] || parts[1] === "new") {
        const set = params.get("set");
        return set ? { name: "draftNew", set } : { name: "draftNew" };
      }
      const invite = params.get("invite");
      return invite ? { name: "draft", id: parts[1], invite } : { name: "draft", id: parts[1] };
    }
    case "drafts":
      return parts[1] ? { name: "myDrafts", id: parts[1] } : { name: "myDrafts" };
    case "privacy":
      return { name: "privacy" };
    case "terms":
      return { name: "terms" };
    case "funding":
      return { name: "funding" };
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
    case "stats":
      if (parts[1] === "shared" && parts[2]) {
        if (parts[3] === "deck" && parts[4])
          return { name: "statsSharedDeck", token: parts[2], id: parts[4] };
        return { name: "statsShared", token: parts[2] };
      }
      if (parts[1] === "deck" && parts[2]) return { name: "statsDeck", id: parts[2] };
      return { name: "stats" };
    case "friends":
      return { name: "friends" };
    case "pods":
      if (parts[1] === "join")
        return parts[2] ? { name: "podJoin", token: parts[2] } : { name: "pods" };
      if (parts[1] && parts[2] === "stats") return { name: "podStats", id: parts[1] };
      if (parts[1]) return { name: "pod", id: parts[1] };
      return { name: "pods" };
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

/** How a Discord Linked Roles link ended (ADR 0163 §3); the server names it. */
export type LinkedRoleStatus = "linked" | "signin" | "unavailable" | "failed";

/** linkedRoleStatus reads ?status=; anything the server does not send is "failed". */
export function linkedRoleStatus(v: string | null): LinkedRoleStatus {
  return v === "linked" || v === "signin" || v === "unavailable" ? v : "failed";
}
