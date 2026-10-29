<script lang="ts">
  // HandoffChip — #4434, ADR 0167 §9. While the viewer plays another
  // seat's turn, the board stays on the board they chose (it switches by
  // itself only when the control begins). When the next decision is owed
  // on the other board, this small chip, drawn over the action dock,
  // says so and jumps there:
  //
  //   Your decision on Bob's board · Jump
  //
  // It is not a dialog and does not take focus: it appears quietly, pulses
  // once as it does (not on every frame, and not at all under reduced
  // motion), and a click on it leaves the focus where it was, so the
  // board's shortcuts keep working. It is also a tab stop, for the
  // keyboard. lib/playingAs.ts handoffChip decides when it shows.
  import { L } from "../../labels";

  interface Props {
    // "Bob's" or "your own": the board the decision is owed on.
    whose: string;
    onJump: () => void;
  }
  const { whose, onJump }: Props = $props();
</script>

<div class="handoff-chip" role="status" aria-label={L.jumpChip(whose)}>
  <span class="text">{L.jumpChip(whose)}</span>
  <!-- mousedown is held back so a click does not move focus off the board. -->
  <button type="button" class="jump" onmousedown={(e) => e.preventDefault()} onclick={onJump}>
    {L.jumpToBoard}
  </button>
</div>

<style>
  /* Over the dock's top edge, right-aligned with it, in the dock's own
     layer (z 55, DESIGN.md §4): the chip belongs to the dock and moves
     with it. A pill, per the chip pattern (DESIGN.md §6). */
  .handoff-chip {
    position: absolute;
    right: 0;
    bottom: calc(100% + 6px);
    max-width: 100%;
    box-sizing: border-box;
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 3px 4px 3px 12px;
    background: var(--surface);
    border: 1px solid color-mix(in srgb, var(--accent) 55%, var(--border-strong));
    border-radius: 999px;
    box-shadow: var(--shadow-sm);
    color: var(--fg);
    font-size: 0.78rem;
    line-height: 1.3;
    animation: handoff-pulse 900ms var(--ease) 1;
  }
  .text {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .jump {
    flex: none;
    padding: 2px 12px;
    border-radius: 999px;
    font-size: 0.74rem;
    font-weight: 700;
    background: var(--accent);
    color: var(--accent-fg);
    border: 1px solid transparent;
    cursor: pointer;
  }
  .jump:focus-visible {
    outline: 2px solid var(--accent-strong);
    outline-offset: 2px;
  }
  /* One pulse as it appears: a ring that grows and fades. The words
     carry the information; the pulse only says "look here". */
  @keyframes handoff-pulse {
    0% {
      box-shadow: 0 0 0 0 color-mix(in srgb, var(--accent) 70%, transparent);
    }
    100% {
      box-shadow: 0 0 0 10px color-mix(in srgb, var(--accent) 0%, transparent);
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .handoff-chip {
      animation: none;
    }
  }
  :global(:root[data-reduce-motion="1"]) .handoff-chip {
    animation: none;
  }
</style>
