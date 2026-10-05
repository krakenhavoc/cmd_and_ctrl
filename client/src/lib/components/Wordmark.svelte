<script lang="ts">
  // The brand: the mark (a card and a second card turned sideways,
  // i.e. tapped) and the wordmark (the name as two keycaps, cmd & ctrl).
  // Drawn in the skin's colours: the upright card is --fg and the tapped
  // one --accent, so it follows the skin and a custom accent.
  //
  // `size` sets the type size; everything else is in em. "hero" stacks
  // the mark over the keys (the login page); "bar" and "header" sit
  // inline. The keys are aria-hidden and the name is spoken once, as
  // "CMD & CTRL", so a heading or link wrapping this keeps that name.
  //
  // The icons in public/icons draw the same mark; keep the geometry in
  // step (tools/gen-icons.sh rasterises them).
  let { size = "header" }: { size?: "bar" | "header" | "hero" } = $props();
  const uid = $props.id();
  const cut = `wm-cut-${uid}`;
</script>

<span class="wordmark-art {size}">
  <svg class="mark" viewBox="0 0 64 64" aria-hidden="true" focusable="false">
    <defs>
      <!-- The tapped card is cut out of the upright one with a 4px
           gap, so the mark sits on any background without a halo. -->
      <mask id={cut}>
        <rect width="64" height="64" fill="white" />
        <rect x="19" y="26" width="42" height="32" rx="8" fill="black" />
      </mask>
    </defs>
    <rect x="7" y="10" width="26" height="36" rx="5" class="up" mask="url(#{cut})" />
    <rect x="23" y="30" width="34" height="24" rx="5" class="tapped" />
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
  .up {
    fill: var(--fg);
  }
  .tapped {
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
