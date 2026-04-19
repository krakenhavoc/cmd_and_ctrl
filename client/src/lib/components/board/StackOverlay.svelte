<script lang="ts">
  // StackOverlay is the floating display of the shared stack zone,
  // anchored at (0.2, 0.2) of the board area — top-left quadrant.
  // Mirrors the placement convention of HoverZoomOverlay (top-right
  // at 0.8, 0.2) so both overlays share visual hierarchy without
  // colliding. Visible only when the stack is non-empty; the wire
  // protocol clears stack.cards on resolve, so the visibility toggle
  // is automatic.
  //
  // Each entry on the stack is a small Card stacked from bottom to
  // top with a fixed offset, so the topmost (most recently cast)
  // sits on top of the pile and the original casting order reads
  // bottom-up.

  import type { ZoneView } from "../../protocol";
  import Card from "./Card.svelte";

  interface Props {
    stack: ZoneView;
  }

  const { stack }: Props = $props();
</script>

{#if stack.count > 0}
  <div class="overlay" aria-label={`stack: ${stack.count} on the stack`}>
    <span class="label">stack · {stack.count}</span>
    <div class="cards">
      {#each stack.cards as c, i (c.instance_id)}
        <div class="stack-slot" style:top="{i * 18}px" style:z-index={i + 1}>
          <Card card={c} />
        </div>
      {/each}
    </div>
  </div>
{/if}

<style>
  .overlay {
    position: absolute;
    left: 20%;
    top: 20%;
    transform: translate(-50%, -50%);
    background: #0b1220;
    border: 1px solid #4a5270;
    border-radius: 12px;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.6);
    z-index: 40;
    padding: 10px 10px 14px;
    box-sizing: border-box;
    pointer-events: none;
  }
  .label {
    display: block;
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: #6c7a99;
    margin-bottom: 6px;
  }
  .cards {
    position: relative;
    width: 80px;
    /* Height grows with stack count; base + per-card offset accounts
       for the cascade so the topmost card stays fully visible. */
    height: calc(112px + var(--extra-h, 0px));
    pointer-events: auto;
  }
  .stack-slot {
    position: absolute;
    left: 0;
  }
</style>
