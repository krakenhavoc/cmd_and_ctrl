<script lang="ts">
  import { GameClient } from "./lib/ws";

  const wsURL = (location.protocol === "https:" ? "wss://" : "ws://") + location.host + "/ws";

  const client = new GameClient(wsURL);
  // Extract the stores for auto-subscription via `$status` / `$log` in
  // the template. This is the idiomatic Svelte 5 way to read stores
  // without writing manual `.subscribe` boilerplate.
  const { status, log } = client;

  let msg = $state("hello");

  // $effect handles mount + cleanup through the Svelte 5 effect
  // lifecycle. The returned function runs when the effect is torn down
  // (component unmount, HMR replacement, etc.).
  $effect(() => {
    client.connect();
    return () => client.disconnect();
  });

  function sendPing(): void {
    client.sendPing(msg);
  }

  function reconnect(): void {
    client.disconnect();
    client.connect();
  }
</script>

<main>
  <h1>cmd_and_ctrl</h1>
  <p>
    S01 — Go server + client scaffold ·
    <span class={`tag tag-${$status}`} data-testid="status">
      {$status}
    </span>
  </p>

  <p>
    <input
      type="text"
      bind:value={msg}
      placeholder="ping message"
      disabled={$status !== "connected"}
    />
    <button onclick={sendPing} disabled={$status !== "connected"}> send ping </button>
    <button onclick={reconnect}>reconnect</button>
  </p>

  <div class="log" data-testid="log">
    {#each $log as entry (entry.id)}
      <div class={`log-entry ${entry.direction}`}>
        <span class="log-time">{entry.at.toLocaleTimeString()}</span>
        {entry.text}
      </div>
    {/each}
    {#if $log.length === 0}
      <div class="log-entry">(no activity yet)</div>
    {/if}
  </div>
</main>
