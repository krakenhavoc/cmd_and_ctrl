<script lang="ts">
  // IdleCloseNotice — ADR 0211 §6. Once a started game has had no move by
  // a person for 36 hours, the server warns that it will close it twelve
  // hours on, and every frame carries `idle_close_at` until somebody
  // moves. This one line, drawn over the action dock's top edge, says so
  // in the viewer's own clock:
  //
  //   ⚠ This game closes at 21:14 on Friday unless someone moves
  //
  // A status, not a question: it takes no focus and no keys, and goes
  // away with the next move. lib/idleClose.ts words it.
  import { onDestroy } from "svelte";

  import Icon from "../Icon.svelte";
  import { idleCloseNotice } from "../../idleClose";

  interface Props {
    // The frame's idle_close_at (RFC 3339).
    idleCloseAt: string;
    // The handoff chip is showing in the same spot: sit above it.
    raised?: boolean;
  }
  const { idleCloseAt, raised = false }: Props = $props();

  // "today" and "tomorrow" move at midnight; a minute is fine enough.
  let now = $state(Date.now());
  const timer = setInterval(() => (now = Date.now()), 60_000);
  onDestroy(() => clearInterval(timer));

  const text = $derived(idleCloseNotice(idleCloseAt, now));
</script>

{#if text}
  <div class="idle-close" class:raised role="status">
    <Icon name="warn" size={14} />
    <span class="text">{text}</span>
  </div>
{/if}

<style>
  /* Over the dock's top edge, right-aligned with it, in the dock's own
     layer (z 55, DESIGN.md §4), as the handoff chip is: the line belongs
     to the dock and moves with it. A pill, per the chip pattern (DESIGN.md
     §6), that may wrap to a second line on a narrow dock rather than
     lose the time. */
  .idle-close {
    position: absolute;
    right: 0;
    bottom: calc(100% + 6px);
    max-width: 100%;
    box-sizing: border-box;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 4px 12px;
    background: var(--surface);
    border: 1px solid var(--border-strong);
    border-radius: 999px;
    box-shadow: var(--shadow-sm);
    color: var(--fg);
    font-size: 0.78rem;
    line-height: 1.3;
  }
  .idle-close :global(.icon) {
    flex: none;
    color: var(--fg-muted);
  }
  /* Above the handoff chip (about 30px tall) when both are up. */
  .idle-close.raised {
    bottom: calc(100% + 42px);
  }
  .text {
    min-width: 0;
  }
</style>
