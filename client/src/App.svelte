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
  import Game from "./routes/Game.svelte";
  import Settings from "./lib/components/Settings.svelte";
  import { route, navigate } from "./lib/router";
  import { session, setSession } from "./lib/session";
  import { settings } from "./lib/settings";
  import { armMusicOnFirstGesture } from "./lib/music";

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
      r.name === "oauthComplete";
    if (!s && !isPublic) {
      navigate("#/login");
    }
    // Landed on login with a live session? Kick to the lobby so the
    // reload-after-login flow doesn't leave you staring at a login
    // form you don't need.
    if (s && r.name === "login") {
      navigate("#/lobby");
    }
  });

  // oauth-complete handoff (S12.5). /auth/discord/callback on the
  // server 302s here with token / game / player_id / expires_at in
  // the URL fragment. Install the session, clear the fragment so
  // it doesn't survive a reload (we've already persisted the
  // session to localStorage), and send the user into the game.
  $effect(() => {
    const r = $route;
    if (r.name !== "oauthComplete") return;
    setSession({
      token: r.token,
      expiresAt: r.expiresAt,
      principal: {
        // The server's /me returns the full principal, but we
        // don't block navigation on fetching it — a RolePlayer
        // session with game_id + player_id is enough for Game.
        role: "player",
        game_id: r.gameID,
        player_id: r.playerID,
        issued_at: new Date().toISOString(),
        expires_at: r.expiresAt,
      },
      playerID: r.playerID,
      gameID: r.gameID,
    });
    navigate(`#/games/${r.gameID}`);
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
    root.dataset.reduceMotion = s.accessibility.reduceMotion ? "1" : "0";
    root.dataset.alwaysShowFocus = s.accessibility.alwaysShowFocus ? "1" : "0";
  });
</script>

{#if $route.name === "login"}
  <Login />
{:else if $route.name === "adminLogin"}
  <Login />
{:else if $route.name === "lobby"}
  <Lobby />
{:else if $route.name === "join"}
  <Join gameID={$route.gameID} inviteToken={$route.inviteToken} spectator={$route.spectator} />
{:else if $route.name === "game"}
  <!-- Keyed so navigating game A → game B tears down and remounts
       the route (fresh GameClient, fresh per-game local state)
       instead of reusing the old component with a swapped prop. -->
  {#key $route.gameID}
    <Game gameID={$route.gameID} />
  {/key}
{/if}

<!-- Settings modal lives at the app shell so it overlays every
     route and the keyboard shortcut / reactive store works from
     lobby, game, join, and login. The component self-renders
     based on settingsOpen — mounting it here just plugs it in. -->
<Settings />
