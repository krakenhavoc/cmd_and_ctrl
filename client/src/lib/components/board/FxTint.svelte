<script lang="ts">
  // The tint a signature leaves on a permanent until the turn ends
  // (fx/tints.ts, ADR 0209 §4, #4416): Toxic Deluge's survivors keep a
  // faint dark-green stain while its -X/-X lasts. Drawn over the card's
  // image and under its badges (no z-index, so the badges after it in the
  // tile paint above). Presentation only; the card's state says the rest.
  import { tableTints } from "../../fx/tints";

  interface Props {
    id: string;
  }

  const { id }: Props = $props();
  const kind = $derived($tableTints.get(id) ?? null);
</script>

{#if kind}
  <span class="fx-tint" data-fx-tint={kind} aria-hidden="true"></span>
{/if}

<style>
  .fx-tint {
    position: absolute;
    inset: 0;
    border-radius: inherit;
    pointer-events: none;
  }
  /* A faint stain, darker at the edges where the sludge clung. It fades
     in as the sludge drains; reduced motion shows it at once (app.css
     cuts the animation under data-reduce-motion, and the media query
     below covers the OS signal). */
  .fx-tint[data-fx-tint="deluge"] {
    background:
      radial-gradient(ellipse at 50% 42%, rgba(20, 58, 22, 0.16) 0%, rgba(10, 34, 12, 0.36) 100%),
      linear-gradient(to top, rgba(8, 26, 10, 0.4) 0%, rgba(8, 26, 10, 0) 45%);
    box-shadow: inset 0 0 calc(var(--card-w, 80px) * 0.16) rgba(22, 64, 20, 0.85);
    animation: fx-tint-in 900ms var(--ease) both;
  }
  @keyframes fx-tint-in {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .fx-tint {
      animation: none;
    }
  }
</style>
