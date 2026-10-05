<script lang="ts">
  // App is the router shell. Each top-level view lives in its own
  // component under src/routes; this file just picks the right one
  // based on the current `route` and the presence of a session.
  //
  // The routing rules are intentionally small, and live in
  // lib/signedInHome.ts: most unauthenticated routes redirect to
  // /login, and a session on /login goes to the Lobby.

  import Login from "./routes/Login.svelte";
  import Lobby from "./routes/Lobby.svelte";
  import Join from "./routes/Join.svelte";
  import Reclaim from "./routes/Reclaim.svelte";
  import Game from "./routes/Game.svelte";
  import Catalog from "./routes/Catalog.svelte";
  import MyGames from "./routes/MyGames.svelte";
  import Decks from "./routes/Decks.svelte";
  import Home from "./routes/Home.svelte";
  import Roadmap from "./routes/Roadmap.svelte";
  import Practice from "./routes/Practice.svelte";
  import Settings from "./lib/components/Settings.svelte";
  import ShortcutLayer from "./lib/components/ShortcutLayer.svelte";
  import UpdatePrompt from "./lib/components/UpdatePrompt.svelte";
  import SettingsSyncToast from "./lib/components/SettingsSyncToast.svelte";
  import SiteFooter from "./lib/components/SiteFooter.svelte";
  import EnvBadge from "./lib/components/EnvBadge.svelte";
  import { route, navigate } from "./lib/router";
  import { session, sessionFromOAuth, setSession } from "./lib/session";
  import { armAdminLapse, loadAdminStatus, needsAdminCheck, onVisibleAgain } from "./lib/admin";
  import { oauthCompleteTarget, routeRedirect } from "./lib/signedInHome";
  import { takeAfterSignIn } from "./lib/decksPage";
  import { settings } from "./lib/settings";
  import { applyRootSettings } from "./lib/rootSettings";
  import { armMusicOnFirstGesture } from "./lib/music";
  import { loadAppConfig } from "./lib/env";

  // Ask the server which deployment this is, once, at shell mount.
  // Deliberately not awaited: every consumer defaults to production
  // (no dev features, no badge), so a slow or missing /config delays
  // nothing and degrades to the safe answer. See lib/env.ts.
  loadAppConfig();

  // Ambient music spans the entire app shell (login → lobby →
  // game), so arm it here rather than in Game.svelte where SFX
  // live. The first pointerdown / keydown anywhere unlocks playback;
  // music.ts no-ops on subsequent calls.
  armMusicOnFirstGesture();

  // Enforce the auth gate as a side effect of routing. Running this
  // inside $effect ensures it re-evaluates on hash change + session
  // change without manual subscription plumbing. The rules live in
  // lib/signedInHome.ts (ADR 0112 §1): a signed-out visitor on a gated
  // route goes to #/login, and every session on #/login goes to the
  // Lobby, the signed-in home. Invite, spectator and reclaim links and
  // the Discord round trip are public and never redirected.
  $effect(() => {
    const to = routeRedirect($route, $session);
    if (to) navigate(to);
  });

  // Ask GET /me whether a newly installed signed-in session is an admin
  // (ADR 0110 §3 item 4): the server's allowlist is not in the token, so
  // this is the only way the client learns it. Once per session per page
  // load; the answer lands on the session as `admin` (lib/admin.ts).
  $effect(() => {
    const s = $session;
    if (needsAdminCheck(s)) void loadAdminStatus(s);
  });

  // Admin mode lapses 12 hours after it was switched on (ADR 0112 §2,
  // owner answer 1). The lapse timer follows the installed session's
  // end time, and a hidden tab that becomes visible asks /me again, so
  // a switch made in another tab or on another device is caught up.
  $effect(() => {
    armAdminLapse($session);
  });
  $effect(() => {
    document.addEventListener("visibilitychange", onVisibleAgain);
    return () => document.removeEventListener("visibilitychange", onVisibleAgain);
  });

  // oauth-complete handoff (S12.5). /auth/discord/callback on the
  // server 302s here with the session in the URL fragment. Install
  // it and move on; the fragment doesn't survive the navigate, and
  // the session is already persisted to localStorage.
  //
  // The fragment's shape says which flow this was. With game +
  // player_id the seat is already claimed, so go to the table. With
  // neither, this is an identity-only session from the login page,
  // which lands on the Lobby with its join box (ADR 0112 §1 item 2), or
  // back on the decks page when the sign-in started there (§3 item 7).
  // takeAfterSignIn reads and clears that stored route either way.
  $effect(() => {
    const r = $route;
    if (r.name !== "oauthComplete") return;
    // The server's /me returns the full principal, but we don't block
    // navigation on fetching it — game_id + player_id is enough for
    // Game, and the identity variant needs only the name and user id
    // it was handed.
    const s = sessionFromOAuth(r);
    setSession(s);
    navigate(oauthCompleteTarget(s, takeAfterSignIn()));
  });

  // Apply the subset of settings that hang off :root as CSS
  // variables. Other settings (mute, animations.enabled) are
  // consumed by the modules that care about them via the shared
  // store; the ones below need a single apply-to-document seam
  // because they drive stylesheet values.
  // That includes the skin (display.theme → data-theme) and a custom
  // accent; the browser chrome colour follows the skin's --bg.
  $effect(() => {
    applyRootSettings(document.documentElement, $settings);
    const bg = getComputedStyle(document.documentElement).getPropertyValue("--bg").trim();
    if (bg) document.querySelector('meta[name="theme-color"]')?.setAttribute("content", bg);
  });
</script>

{#if $route.name === "login"}
  <Login />
{:else if $route.name === "adminLogin"}
  <Login admin />
{:else if $route.name === "home"}
  <Home />
{:else if $route.name === "roadmap"}
  <Roadmap />
{:else if $route.name === "decks"}
  <Decks />
{:else if $route.name === "lobby"}
  <Lobby />
{:else if $route.name === "practice"}
  <Practice />
{:else if $route.name === "catalog"}
  <Catalog />
{:else if $route.name === "myGames"}
  <MyGames />
{:else if $route.name === "join"}
  <Join gameID={$route.gameID} inviteToken={$route.inviteToken} spectator={$route.spectator} />
{:else if $route.name === "reclaim"}
  <Reclaim gameID={$route.gameID} ticket={$route.ticket} />
{:else if $route.name === "game"}
  <!-- Keyed so navigating game A → game B tears down and remounts
       the route (fresh GameClient, fresh per-game local state)
       instead of reusing the old component with a swapped prop. -->
  {#key $route.gameID}
    <Game gameID={$route.gameID} />
  {/key}
{/if}

<!-- Unofficial-fan-content / Scryfall notice (#2191). Every page outside
     the table; never on the game route, which is also where the practice
     table lives (#/practice is only the door that opens it). The oauth hand-off renders nothing, so skip it too. -->
{#if $route.name !== "game" && $route.name !== "oauthComplete"}
  <SiteFooter />
{/if}

<!-- Environment marker. Renders nothing in production; on dev it is a
     fixed, non-dismissible corner badge so no one mistakes this
     deployment for the live table. -->
<EnvBadge />

<!-- The global keymap (ADR 0047). Mounted at the shell because there
     must be exactly ONE global keydown listener for shortcuts — a
     second one is how two features end up both claiming a key. It
     dispatches nothing on its own while a modal is up or a text field
     has focus, and the game route publishes its handlers through
     lib/shortcutRuntime.ts. Also carries the `?` overlay. -->
<ShortcutLayer />

<!-- Settings modal lives at the app shell so it overlays every
     route and the keyboard shortcut / reactive store works from
     lobby, game, join, and login. The component self-renders
     based on settingsOpen — mounting it here just plugs it in. -->
<Settings />

<!-- Service-worker update toast (ADR 0031). Self-hiding until a new
     build is installed and waiting; mounted at the shell so it can
     appear on any route, and deliberately non-modal so it never
     interrupts a game. -->
<UpdatePrompt />

<!-- Account settings (ADR 0110 §4): shown for the rest of the visit
     after a sign-in applied the account's settings over different ones
     in this browser, offering this browser's back. -->
<SettingsSyncToast />
