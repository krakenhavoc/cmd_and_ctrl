<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { GameClient, type ConnectionStatus, type LogEntry } from "./lib/ws";

  const wsURL = (location.protocol === "https:" ? "wss://" : "ws://") + location.host + "/ws";

  const client = new GameClient(wsURL);

  let status: ConnectionStatus = $state("disconnected");
  let entries: LogEntry[] = $state([]);
  let msg = $state("hello");

  const unsubStatus = client.status.subscribe((v) => (status = v));
  const unsubLog = client.log.subscribe((v) => (entries = v));

  onMount(() => client.connect());
  onDestroy(() => {
    client.disconnect();
    unsubStatus();
    unsubLog();
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
    <span class={`tag tag-${status}`} data-testid="status">
      {status}
    </span>
  </p>

  <p>
    <input
      type="text"
      bind:value={msg}
      placeholder="ping message"
      disabled={status !== "connected"}
    />
    <button onclick={sendPing} disabled={status !== "connected"}> send ping </button>
    <button onclick={reconnect}>reconnect</button>
  </p>

  <div class="log" data-testid="log">
    {#each entries as entry (entry.id)}
      <div class={`log-entry ${entry.direction}`}>
        <span class="log-time">{entry.at.toLocaleTimeString()}</span>
        {entry.text}
      </div>
    {/each}
    {#if entries.length === 0}
      <div class="log-entry">(no activity yet)</div>
    {/if}
  </div>
</main>
