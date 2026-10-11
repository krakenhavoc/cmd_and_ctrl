<script lang="ts">
  // TutorialScrim — the spotlight (ADR 0076 §2.3, #1079). ONE element: a
  // transparent rect with a 9999px spread shadow, so the hole is the rect
  // and everything else darkens.
  //
  // It is pointer-events: none, always. Dimming is a suggestion, never a
  // lock: the board under it stays fully playable at every step, inside
  // the hole and outside it. This is not an implementation detail to
  // optimise away later; tutorialCoach.render.test.ts pins it.
  //
  // Fixed to the viewport, because the rect comes from
  // getBoundingClientRect. z 56: over the action dock (55), so the dock
  // dims with the rest of the table unless it is the hole; under the
  // coach card (57), the card-local menus (60+), the modals (200) and
  // the hover zoom (300), so a right-click menu or a zoomed card opened
  // during a step is never dimmed.

  import type { AnchorRect } from "../../tutorialAnchor";

  interface Props {
    /** One hole per rect; a single rect is the plain box-shadow spotlight. */
    rects: AnchorRect[];
    /** Room between the anchor and the hole's edge, in px. */
    pad?: number;
  }

  const { rects, pad = 6 }: Props = $props();
  const multi = $derived(rects.length > 1);
  // The shadow trick darkens everything outside ONE rect, so two of them
  // would each dim the other's hole. With several holes the dimming is a
  // single full-viewport layer with the holes masked out of it.
  const maskID = `tutorial-holes-${Math.random().toString(36).slice(2, 8)}`;
</script>

{#if multi}
  <svg class="tutorial-dim" aria-hidden="true" data-holes={rects.length}>
    <defs>
      <mask id={maskID}>
        <rect width="100%" height="100%" fill="white" />
        {#each rects as r, i (i)}
          <rect
            x={r.left - pad}
            y={r.top - pad}
            width={r.width + pad * 2}
            height={r.height + pad * 2}
            rx="12"
            fill="black"
          />
        {/each}
      </mask>
    </defs>
    <rect width="100%" height="100%" fill="rgba(7, 6, 5, 0.62)" mask={`url(#${maskID})`} />
  </svg>
{/if}
{#each rects as rect, i (i)}
  <div
    class="tutorial-scrim"
    class:part={multi}
    aria-hidden="true"
    style:left={`${rect.left - pad}px`}
    style:top={`${rect.top - pad}px`}
    style:width={`${rect.width + pad * 2}px`}
    style:height={`${rect.height + pad * 2}px`}
  ></div>
{/each}

<style>
  .tutorial-scrim {
    position: fixed;
    z-index: 56;
    pointer-events: none;
    border-radius: 12px;
    box-shadow: 0 0 0 9999px rgba(7, 6, 5, 0.62);
    outline: 1px solid color-mix(in srgb, var(--accent) 40%, transparent);
    outline-offset: 3px;
    transition:
      left 140ms var(--ease),
      top 140ms var(--ease),
      width 140ms var(--ease),
      height 140ms var(--ease);
  }
  .tutorial-scrim.part {
    box-shadow: none;
  }
  .tutorial-dim {
    position: fixed;
    inset: 0;
    width: 100%;
    height: 100%;
    z-index: 56;
    pointer-events: none;
  }
  @media (prefers-reduced-motion: reduce) {
    .tutorial-scrim {
      transition: none;
    }
  }
</style>
