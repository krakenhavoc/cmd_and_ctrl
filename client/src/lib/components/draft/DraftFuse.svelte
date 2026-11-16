<script lang="ts">
  // DraftFuse — the pick clock as a fuse (#5848 §7): a bar under the
  // site header that burns from the left, with a spark at the burning
  // end. It is blue at first, orange by a third of the time left, red
  // by a seventh, and it flashes in the last eight seconds.
  //
  // It runs from the server's deadline, the same one the text clock
  // counts (draft.ts, secondsLeft), and moves every animation frame, not
  // once a second. It is decoration: the clock beside it, and the
  // announcements at 30 and 10 seconds, say the same thing in words, so
  // it is hidden from a screen reader. With reduced motion the colour
  // still changes, but the spark holds still and nothing flashes.
  //
  // It sticks under the site header, so it is in view as the page
  // scrolls. The room draws it only while a pick is being made.
  import { onMount } from "svelte";

  import { msLeftAfter, type DraftView } from "../../draft";
  import { fuseColour, fuseFlashing, fuseFraction } from "../../draftFuse";

  interface Props {
    /** The view's server clock: `deadline` and `now` as the frame was built. */
    clock: Pick<DraftView, "deadline" | "now">;
    /** `performance.now()` when that frame arrived. */
    frameAt: number;
    /** How long a pick lasts, in seconds: what a full fuse is. */
    totalSeconds: number;
  }
  const { clock, frameAt, totalSeconds }: Props = $props();

  let bar = $state<HTMLElement | null>(null);
  let flashing = $state(false);
  // Under the site header, whatever its height is on this screen.
  let top = $state(61);

  function paint(): void {
    if (!bar) return;
    const left = msLeftAfter(clock, performance.now() - frameAt);
    const f = left === null ? 0 : fuseFraction(left, totalSeconds * 1000);
    bar.style.width = `${(f * 100).toFixed(2)}%`;
    bar.style.setProperty("--fuse", fuseColour(f));
    flashing = fuseFlashing(left);
  }

  onMount(() => {
    const head = document.querySelector<HTMLElement>(".site-header");
    const measure = (): void => {
      if (head) top = head.offsetHeight;
    };
    measure();
    const ro = head && typeof ResizeObserver !== "undefined" ? new ResizeObserver(measure) : null;
    if (head) ro?.observe(head);

    let id = 0;
    const frame = (): void => {
      paint();
      id = requestAnimationFrame(frame);
    };
    id = requestAnimationFrame(frame);
    return () => {
      cancelAnimationFrame(id);
      ro?.disconnect();
    };
  });
</script>

<div
  class="fuse"
  class:flash={flashing}
  data-testid="draft-fuse"
  aria-hidden="true"
  style:top={`${top}px`}
>
  <div class="burn" bind:this={bar}><i class="spark"></i></div>
</div>

<style>
  /* Under the site header (z 30), above the page. */
  .fuse {
    position: sticky;
    z-index: 29;
    height: 8px;
    /* Clear of the page head above it and the room below. */
    margin: -8px 0 16px;
    border-radius: 4px;
    background: var(--surface-sunken);
    box-shadow: inset 0 0 0 1px var(--border);
  }
  /* It burns from the left: the bar is anchored on the right and shrinks,
     so its left end, where the spark is, moves toward the right. */
  .burn {
    position: absolute;
    top: 0;
    right: 0;
    bottom: 0;
    width: 100%;
    border-radius: 4px;
    background: var(--fuse, rgb(61, 139, 255));
    box-shadow: 0 0 10px var(--fuse, rgb(61, 139, 255));
  }
  .spark {
    position: absolute;
    left: -6px;
    top: 50%;
    width: 14px;
    height: 14px;
    margin-top: -7px;
    border-radius: 50%;
    background: radial-gradient(
      circle,
      #fff 0 22%,
      #ffd27a 42%,
      rgba(255, 160, 60, 0.5) 58%,
      transparent 72%
    );
    animation: sparkle 0.22s infinite alternate;
  }
  @keyframes sparkle {
    from {
      transform: scale(0.75);
    }
    to {
      transform: scale(1.3);
    }
  }
  .flash .burn {
    animation: fuse-flash 0.5s steps(2, jump-none) infinite;
  }
  @keyframes fuse-flash {
    from {
      opacity: 1;
    }
    to {
      opacity: 0.2;
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .spark,
    .flash .burn {
      animation: none;
    }
  }
</style>
