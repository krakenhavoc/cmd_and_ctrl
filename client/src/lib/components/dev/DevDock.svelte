<script lang="ts">
  // The develop environment's tool dock (ADR 0023).
  //
  // One fixed corner strip owns every dev tool, because they are all
  // "a panel anchored to the bottom-left" and two of them competing
  // for that spot would overlap. Each tab is independently gated on
  // its own feature flag, so a dev deployment that switches one off
  // loses the tab, not the dock.
  //
  // The whole dock renders nothing unless at least one tool is
  // enabled — which on production is always, because /config reports
  // every feature false there.
  import { devFeature } from "../../env";
  import type { GameClient } from "../../ws";
  import type { GameView } from "../../protocol";
  import FrameInspector from "./FrameInspector.svelte";
  import CardSpawner from "../CardSpawner.svelte";
  import SeatSwitcher from "./SeatSwitcher.svelte";
  import ReplayScrubber from "./ReplayScrubber.svelte";
  import type { ReplayFrame } from "../../replay";
  import { canSwapSeats } from "../../gameURL";
  import { session } from "../../session";

  interface Props {
    client: GameClient;
    gameID: string;
    snapshot: GameView | null;
    // Seat currently being viewed, or null for the spectator view.
    seat: string | null;
    onseatchange: (seatID: string | null) => void;
    // Index of the replay frame being rendered, or null for live.
    replayIndex: number | null;
    onreplayselect: (frame: ReplayFrame | null, index: number | null) => void;
  }
  const { client, gameID, snapshot, seat, onseatchange, replayIndex, onreplayselect }: Props =
    $props();

  const showFrames = devFeature("frame_inspector");
  const showSpawn = devFeature("card_spawn");
  const seatSwapFlag = devFeature("seat_swap");
  const showReplay = devFeature("replay_scrubber");
  // Only offered to admin sessions: for anyone else the server
  // resolves the seat from the principal and ignores the request, so
  // the control would silently do nothing. See gameURL.canSwapSeats.
  const showSeats = $derived($seatSwapFlag && canSwapSeats($session));

  const frames = $derived(client.frames);

  type Tab = "frames" | "spawn" | "replay";
  let open: Tab | null = $state(null);

  function toggle(tab: Tab) {
    open = open === tab ? null : tab;
  }

  // If the tool that is open gets disabled underneath us, close the
  // panel rather than leaving an empty shell.
  $effect(() => {
    if (open === "frames" && !$showFrames) open = null;
    if (open === "spawn" && !$showSpawn) open = null;
    if (open === "replay" && !$showReplay) open = null;
  });
</script>

{#if $showFrames || $showSpawn || $showReplay || showSeats}
  <div class="dock">
    {#if open}
      <div class="panel">
        {#if open === "frames"}
          <FrameInspector {client} />
        {:else if open === "spawn"}
          <CardSpawner {gameID} {snapshot} />
        {:else if open === "replay"}
          <ReplayScrubber {gameID} index={replayIndex} onselect={onreplayselect} />
        {/if}
      </div>
    {/if}

    <div class="tabs" role="group" aria-label="Developer tools">
      {#if showSeats}
        <SeatSwitcher seats={snapshot?.seats ?? []} current={seat} onchange={onseatchange} />
      {/if}
      {#if $showSpawn}
        <button
          class:active={open === "spawn"}
          aria-expanded={open === "spawn"}
          onclick={() => toggle("spawn")}
        >
          SPAWN
        </button>
      {/if}
      {#if $showReplay}
        <button
          class:active={open === "replay"}
          aria-expanded={open === "replay"}
          onclick={() => toggle("replay")}
        >
          REPLAY
          {#if replayIndex !== null}<span class="count">past</span>{/if}
        </button>
      {/if}
      {#if $showFrames}
        <button
          class:active={open === "frames"}
          aria-expanded={open === "frames"}
          onclick={() => toggle("frames")}
        >
          FRAMES
          <span class="count">{$frames.length}</span>
        </button>
      {/if}
    </div>
  </div>
{/if}

<style>
  .dock {
    position: fixed;
    left: 0;
    bottom: 0;
    z-index: 9998;
    display: flex;
    flex-direction: column;
    font-family: var(--font-mono, ui-monospace, "SF Mono", Menlo, monospace);
    font-size: 0.7rem;
    color: var(--fg, #e6ecff);
  }

  .panel {
    width: min(46rem, 100vw);
    height: min(24rem, 50vh);
    display: flex;
    flex-direction: column;
    min-height: 0;
    border: 1px solid var(--border-strong, #3a4570);
    border-left: none;
    border-bottom: none;
    background: var(--bg, #070b14);
    box-shadow: var(--shadow-lg, 0 12px 32px rgba(0, 0, 0, 0.55));
  }

  .tabs {
    display: flex;
    gap: 1px;
  }
  .tabs button {
    display: flex;
    align-items: center;
    gap: 0.5em;
    padding: 0.3rem 0.7rem;
    border: 1px solid var(--border-strong, #3a4570);
    border-bottom: none;
    background: var(--bg-1, #0b1220);
    color: var(--fg-muted, #9aa5cd);
    font: inherit;
    letter-spacing: 0.1em;
    cursor: pointer;
  }
  .tabs button:first-child {
    border-left: none;
    border-top-right-radius: var(--radius, 8px);
  }
  .tabs button:hover {
    color: var(--fg, #e6ecff);
  }
  .tabs button.active {
    border-color: var(--gold, #ffd07a);
    color: var(--gold, #ffd07a);
  }
  .count {
    padding: 0 0.4em;
    border-radius: var(--radius-sm, 4px);
    background: var(--surface-raised, #17213a);
    color: var(--accent, #7aa7ff);
  }
</style>
