<script lang="ts">
  import { GameClient } from "../lib/ws";
  import { navigate } from "../lib/router";
  import { session } from "../lib/session";

  interface Props {
    gameID: string;
  }
  const { gameID }: Props = $props();

  // Build the WS URL from the stored session + route. Query string
  // carries the session token (browsers can't send Authorization on
  // WS upgrade) and identifies which seat to render. If the session
  // has a player_id bound to this game, we pin it; admin sessions
  // may omit ?player= and fall through to the spectator view.
  const baseURL = (location.protocol === "https:" ? "wss://" : "ws://") + location.host + "/ws";
  const sess = $derived($session);
  const wsURL = $derived.by(() => {
    const params = new URLSearchParams();
    params.set("game", gameID);
    if (sess?.token) params.set("token", sess.token);
    if (sess?.playerID && sess.gameID === gameID) params.set("player", sess.playerID);
    return `${baseURL}?${params.toString()}`;
  });

  // One GameClient per component instance. Start with an empty URL
  // — the $effect below installs the real URL on first run and
  // reconnects whenever wsURL changes. This preserves the store
  // subscribers across reactive reruns (vs. replacing the client
  // instance, which would strand subscriptions on the old object).
  const client = new GameClient("");
  const { status, snapshot, lastSeq } = client;

  $effect(() => {
    client.disconnect();
    client.setURL(wsURL);
    client.connect();
    return () => client.disconnect();
  });

  function back(): void {
    navigate("#/lobby");
  }
</script>

<section>
  <header>
    <button onclick={back}>← lobby</button>
    <h1>game {gameID.slice(0, 8)}</h1>
    <span class={`tag tag-${$status}`}>{$status}</span>
  </header>

  <p class="muted">seq: {$lastSeq}</p>

  {#if $snapshot}
    <pre>{JSON.stringify($snapshot, null, 2)}</pre>
  {:else}
    <p class="muted">waiting for snapshot…</p>
  {/if}
</section>

<style>
  section {
    max-width: 960px;
    margin: 1rem auto;
    padding: 1rem;
  }
  header {
    display: flex;
    gap: 0.75rem;
    align-items: baseline;
  }
  pre {
    background: #f4f4f4;
    padding: 0.75rem;
    overflow: auto;
    max-height: 70vh;
    font-size: 0.85em;
  }
  .muted {
    color: #666;
  }
  .tag {
    padding: 0.1rem 0.4rem;
    border-radius: 3px;
    font-size: 0.8em;
  }
  .tag-connected {
    background: #cfc;
  }
  .tag-connecting {
    background: #ffc;
  }
  .tag-disconnected {
    background: #fcc;
  }
</style>
