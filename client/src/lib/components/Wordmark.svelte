<script lang="ts">
  // The brand: the mark (four Cs, four players, around a diamond, the
  // commander at the centre of the table) and the wordmark (the name as
  // two keycaps, cmd & ctrl). The mark is drawn in --accent, so it
  // follows the skin and a custom accent.
  //
  // `size` sets the type size; everything else is in em. "hero" stacks
  // the mark over the keys (the login page); "bar" and "header" sit
  // inline. The keys are aria-hidden and the name is spoken once, as
  // "CMD & CTRL", so a heading or link wrapping this keeps that name.
  //
  // The icons in public/icons draw the same mark; keep the geometry in
  // step (tools/gen-icons.sh rasterises them).
  let { size = "header" }: { size?: "bar" | "header" | "hero" } = $props();

  // One C, the top-left one, on a grid centred on the diamond; the
  // other three are its mirror images.
  const C =
    "M-10-40V-80L-30-100H-52A48 48 0 0 0-100-52V-30L-80-10H-40L-33-17L-48-32H-71A7 7 0 0 1-78-39V-52A26 26 0 0 1-52-78H-39A7 7 0 0 1-32-71V-48L-17-33Z";
  const MIRRORS = ["", "scale(-1 1)", "scale(1 -1)", "scale(-1 -1)"];
</script>

<span class="wordmark-art {size}">
  <svg class="mark" viewBox="-104 -104 208 208" aria-hidden="true" focusable="false">
    {#each MIRRORS as t (t)}
      <path d={C} transform={t} />
    {/each}
    <path d="M0-38L38 0L0 38L-38 0Z" />
  </svg>
  <span class="keys" aria-hidden="true">
    <span class="key">cmd</span><span class="amp">&amp;</span><span class="key">ctrl</span>
  </span>
  <span class="sr-only">CMD &amp; CTRL</span>
</span>

<style>
  .wordmark-art {
    display: inline-flex;
    align-items: center;
    gap: 0.5em;
    font-family: var(--font-display);
    font-weight: 700;
    line-height: 1;
    color: var(--fg);
  }
  .bar {
    font-size: 12px;
  }
  .header {
    font-size: 14px;
  }
  .hero {
    flex-direction: column;
    gap: 18px;
    font-size: 38px;
  }
  .mark {
    width: 1.5em;
    height: 1.5em;
    flex: 0 0 auto;
    overflow: visible;
  }
  .hero .mark {
    width: 76px;
    height: 76px;
    filter: drop-shadow(0 10px 26px color-mix(in srgb, var(--accent) 30%, transparent));
  }
  .mark path {
    fill: var(--accent);
  }
  .keys {
    display: inline-flex;
    align-items: center;
    gap: 0.3em;
  }
  /* A keycap: a raised face lit from above, a darker lip under it. */
  .key {
    display: inline-flex;
    align-items: center;
    padding: 0.14em 0.4em 0.18em;
    border-radius: 0.28em;
    letter-spacing: -0.01em;
    background: linear-gradient(
      180deg,
      color-mix(in srgb, var(--surface-raised) 85%, var(--overlay-ink) 7%),
      var(--surface-raised)
    );
    border: 1px solid color-mix(in srgb, var(--overlay-ink) 15%, transparent);
    box-shadow:
      0 0.12em 0 color-mix(in srgb, var(--bg) 60%, var(--shadow-ink)),
      inset 0 1px 0 color-mix(in srgb, var(--overlay-ink) 10%, transparent);
  }
  .amp {
    color: var(--fg-muted);
    font-weight: 600;
  }
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }
</style>
