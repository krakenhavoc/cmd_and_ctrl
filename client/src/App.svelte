<script lang="ts">
  // App is the router shell. Each top-level view lives in its own
  // component under src/routes; this file just picks the right one
  // based on the current `route` and the presence of a session.
  //
  // The routing rules are intentionally small: most unauthenticated
  // routes redirect to /login. The one public route is /games/:id/join,
  // which is how new players onboard via an invite link.

  import Login from "./routes/Login.svelte";
  import Lobby from "./routes/Lobby.svelte";
  import Join from "./routes/Join.svelte";
  import Reclaim from "./routes/Reclaim.svelte";
  import Game from "./routes/Game.svelte";
  import Catalog from "./routes/Catalog.svelte";
  import MyGames from "./routes/MyGames.svelte";
  import Settings from "./lib/components/Settings.svelte";
  import ShortcutLayer from "./lib/components/ShortcutLayer.svelte";
  import UpdatePrompt from "./lib/components/UpdatePrompt.svelte";
  import EnvBadge from "./lib/components/EnvBadge.svelte";
  import { route, navigate } from "./lib/router";
  import { session, sessionFromOAuth, setSession } from "./lib/session";
  import { settings } from "./lib/settings";
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
  // change without manual subscription plumbing.
  $effect(() => {
    const r = $route;
    const s = $session;
    const isPublic =
      r.name === "login" ||
      r.name === "adminLogin" ||
      r.name === "join" ||
      // The whole point of a reclaim link is that the holder has no
      // session yet — gating it behind one would bounce them to the
      // login page they cannot get past.
      r.name === "reclaim" ||
      r.name === "oauthComplete";
    if (!s && !isPublic) {
      navigate("#/login");
    }
    // Landed on login with a live session? Kick to the lobby so the
    // reload-after-login flow doesn't leave you staring at a login
    // form you don't need.
    //
    // An identity session is the one exception: a Discord sign-in
    // that hasn't claimed a seat belongs ON the login page, because
    // that is where the invite-code box lives. Bouncing it to the
    // lobby would strand the user one step short of a table.
    if (s && r.name === "login" && s.principal.role !== "identified") {
      navigate("#/lobby");
    }
  });

  // oauth-complete handoff (S12.5). /auth/discord/callback on the
  // server 302s here with the session in the URL fragment. Install
  // it and move on; the fragment doesn't survive the navigate, and
  // the session is already persisted to localStorage.
  //
  // The fragment's shape says which flow this was. With game +
  // player_id the seat is already claimed, so go to the table. With
  // neither, this is an identity-only session from the login page —
  // land back on login, where the invite-code box is waiting.
  $effect(() => {
    const r = $route;
    if (r.name !== "oauthComplete") return;
    // The server's /me returns the full principal, but we don't block
    // navigation on fetching it — game_id + player_id is enough for
    // Game, and the identity variant needs only the name and user id
    // it was handed.
    const s = sessionFromOAuth(r);
    setSession(s);
    navigate(s.principal.role === "player" ? `#/games/${r.gameID}` : "#/login");
  });

  // Apply the subset of settings that hang off :root as CSS
  // variables. Other settings (mute, animations.enabled) are
  // consumed by the modules that care about them via the shared
  // store; the ones below need a single apply-to-document seam
  // because they drive stylesheet values.
  //
  // Note on theme: the schema carries display.theme but we
  // deliberately do NOT push it to root.dataset.theme yet.
  // PR #119 landed the :root[data-theme=...] palette scaffold,
  // but most components in the table chrome (PlayerPanel, Card,
  // Game.svelte styles) still hardcode hex colours rather than
  // reading var(--bg) / var(--fg). Toggling light or high-
  // contrast right now would change the page bg without flipping
  // any of those panels, producing a broken-looking mix. Until
  // the per-component var() migration ships, the theme setting
  // persists but is inert — the Settings panel disables the
  // select with a "coming soon" note so users know.
  $effect(() => {
    const s = $settings;
    const root = document.documentElement;
    root.style.setProperty("--font-scale", String(s.accessibility.textScale));
    root.dataset.cardSize = s.display.cardSize;
    root.dataset.tableLayout = s.display.tableLayout;
    root.dataset.reduceMotion = s.accessibility.reduceMotion ? "1" : "0";
    root.dataset.alwaysShowFocus = s.accessibility.alwaysShowFocus ? "1" : "0";
  });
</script>

{#if $route.name === "login"}
  <Login />
{:else if $route.name === "adminLogin"}
  <Login admin />
{:else if $route.name === "lobby"}
  <Lobby />
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
