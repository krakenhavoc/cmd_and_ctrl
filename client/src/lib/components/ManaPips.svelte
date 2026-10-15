<script lang="ts">
  // ManaPips — Magic's five colours, W U B R G, as five small squares: C2's
  // mark on site pages, beside a page's eyebrow. Drawn from the colour-
  // identity tokens (app.css, ADR 0158 §2), which are tuned to glow on the
  // dark page. Decorative.
  const COLOURS = ["w", "u", "b", "r", "g"] as const;
</script>

<span class="pips" aria-hidden="true">
  {#each COLOURS as c (c)}
    <span class="pip" style="--pip: var(--identity-{c})"></span>
  {/each}
</span>

<style>
  .pips {
    display: inline-flex;
    gap: 3px;
    vertical-align: middle;
  }
  .pip {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--pip);
  }
</style>
