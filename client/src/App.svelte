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
  import { session } from "./lib/session";
  import { settings } from "./lib/settings";

  // Enforce the auth gate as a side effect of routing. Running this
  // inside $effect ensures it re-evaluates on hash change + session
  // change without manual subscription plumbing.
  $effect(() => {
    const r = $route;
    const s = $session;
    const isPublic = r.name === "login" || r.name === "adminLogin" || r.name === "join";
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

  // Apply the subset of settings that hang off :root as CSS
  // variables. Other settings (mute, animations.enabled) are
  // consumed by the modules that care about them via the shared
  // store; the ones below need a single apply-to-document seam
  // because they drive stylesheet values.
  $effect(() => {
    const s = $settings;
    const root = document.documentElement;
    root.style.setProperty("--font-scale", String(s.accessibility.textScale));
    root.dataset.theme = s.display.theme;
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
  <Game gameID={$route.gameID} />
{/if}

<!-- Settings modal lives at the app shell so it overlays every
     route and the keyboard shortcut / reactive store works from
     lobby, game, join, and login. The component self-renders
     based on settingsOpen — mounting it here just plugs it in. -->
<Settings />
