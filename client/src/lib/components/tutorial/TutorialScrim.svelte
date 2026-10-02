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
    rect: AnchorRect;
    /** Room between the anchor and the hole's edge, in px. */
    pad?: number;
  }

  const { rect, pad = 6 }: Props = $props();
</script>

<div
  class="tutorial-scrim"
  aria-hidden="true"
  style:left={`${rect.left - pad}px`}
  style:top={`${rect.top - pad}px`}
  style:width={`${rect.width + pad * 2}px`}
  style:height={`${rect.height + pad * 2}px`}
></div>

<style>
  .tutorial-scrim {
    position: fixed;
    z-index: 56;
    pointer-events: none;
    border-radius: 12px;
    box-shadow: 0 0 0 9999px rgba(7, 6, 5, 0.62);
    outline: 1px solid rgba(217, 180, 92, 0.4);
    outline-offset: 3px;
    transition:
      left 140ms var(--ease),
      top 140ms var(--ease),
      width 140ms var(--ease),
      height 140ms var(--ease);
  }
  @media (prefers-reduced-motion: reduce) {
    .tutorial-scrim {
      transition: none;
    }
  }
</style>
