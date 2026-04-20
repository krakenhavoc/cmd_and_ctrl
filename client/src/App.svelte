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
  import { route, navigate } from "./lib/router";
  import { session } from "./lib/session";

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
