<script lang="ts">
  // HintCard — a first-use hint on screen (ADR 0125 §3.6).
  //
  // A card and a ring, no scrim. The card is about 280px wide and sits
  // beside its anchor where lib/hints/place.ts put it; the anchor gets a
  // 2px ring that takes no clicks. Nothing dims, and nothing under the
  // card loses its clicks except what the card itself covers.
  //
  //   - It never takes focus and never swallows a key. Go to the tip
  //     (`i`) moves focus here; Tab then moves between its buttons, and
  //     Escape dismisses it (as Got it) and puts focus back.
  //   - It is `complementary "tip"`, so landmark navigation finds it.
  //     HintLayer reads it out once through its polite live region.
  //   - On a phone it is a one-line strip on the bottom edge (on the
  //     dock bar at the table); a tap opens it to its two lines and
  //     buttons, like the coach card's strip (ADR 0076 §2.3).
  //   - With reduced motion it simply appears: no fade, no slide, and
  //     the ring does not pulse.
  //
  // Presentational: HintLayer (or HintSlot, inside the Settings dialog)
  // decides what shows and where.

  import { L } from "../../labels";
  import type { Rect } from "../../hints/place";
  import type { ActiveTip } from "../../hints/runtime";

  interface Props {
    title: string;
    body: string;
    /** The id of the body, for the anchor's aria-describedby. */
    bodyId: string;
    placement: ActiveTip["placement"];
    /** The anchor's rect, ringed; null draws no ring. */
    ring?: Rect | null;
    stripBottom?: number;
    reduceMotion?: boolean;
    action?: { label: string; href: string };
    onGotIt?: () => void;
    onHideTips?: () => void;
    onAction?: () => void;
    /** Focus entered the card from `el`: where Escape sends it back. */
    onFocusFrom?: (el: Element | null) => void;
  }

  const {
    title,
    body,
    bodyId,
    placement,
    ring = null,
    stripBottom = 6,
    reduceMotion = false,
    action,
    onGotIt,
    onHideTips,
    onAction,
    onFocusFrom,
  }: Props = $props();

  const strip = $derived(placement.kind === "strip");
  // The phone strip opens on a tap; a desktop card is always open.
  let opened = $state(false);
  const open = $derived(!strip || opened);

  let root: HTMLElement | null = $state(null);

  function onKeydown(e: KeyboardEvent): void {
    if (e.key !== "Escape") return;
    // The card's own key: Escape here is Got it, and goes no further
    // (a Settings dialog under it must not close too).
    e.preventDefault();
    e.stopPropagation();
    onGotIt?.();
  }

  function onFocusIn(e: FocusEvent): void {
    const from = e.relatedTarget as Element | null;
    if (from && root && !root.contains(from)) onFocusFrom?.(from);
  }
</script>

{#if ring}
  <div
    class="hint-ring"
    class:pulse={!reduceMotion}
    aria-hidden="true"
    style:left="{ring.left - 3}px"
    style:top="{ring.top - 3}px"
    style:width="{ring.width + 6}px"
    style:height="{ring.height + 6}px"
  ></div>
{/if}

<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<aside
  bind:this={root}
  class="hint-card"
  class:strip
  class:open
  class:animate={!reduceMotion}
  aria-label={L.tip}
  tabindex="-1"
  style:left={placement.kind === "card" ? `${placement.left}px` : undefined}
  style:top={placement.kind === "card" ? `${placement.top}px` : undefined}
  style:bottom={strip ? `${stripBottom}px` : undefined}
  onkeydown={onKeydown}
  onfocusin={onFocusIn}
>
  {#if strip && !open}
    <button
      type="button"
      class="hint-strip-head"
      aria-expanded="false"
      onclick={() => (opened = true)}
    >
      <span class="hint-eyebrow">Tip</span>
      <span class="hint-title hint-one-line">{title}</span>
    </button>
    <p class="hint-sr" id={bodyId}>{body}</p>
  {:else}
    <p class="hint-eyebrow">Tip</p>
    <p class="hint-title">{title}</p>
    <p class="hint-body" id={bodyId}>{body}</p>
    <div class="hint-bar">
      <button type="button" class="hint-quiet" onclick={() => onHideTips?.()}>Hide tips</button>
      <span class="hint-grow"></span>
      {#if action}
        <a class="hint-action" href={action.href} onclick={() => onAction?.()}>{action.label}</a>
      {/if}
      <button type="button" class="hint-primary" onclick={() => onGotIt?.()}
        >{action ? "Not now" : "Got it"}</button
      >
    </div>
  {/if}
</aside>

<style>
  /* z 58: over the coach (57) and the dock (55); under the card-local
     menus (60+), the modals (200) and the hover zoom (300). Fixed, so it
     sits where place.ts put it in viewport coordinates. */
  .hint-ring {
    position: fixed;
    z-index: 58;
    pointer-events: none;
    border: 2px solid var(--accent);
    border-radius: 10px;
    box-sizing: border-box;
  }
  .hint-ring.pulse {
    animation: hint-pulse 1.6s ease-in-out 2;
  }
  @keyframes hint-pulse {
    50% {
      box-shadow: 0 0 0 4px color-mix(in srgb, var(--accent) 35%, transparent);
    }
  }
  .hint-card {
    position: fixed;
    z-index: 58;
    box-sizing: border-box;
    width: 280px;
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 12px 14px 10px;
    border-radius: 12px;
    background: var(--surface);
    border: 1px solid var(--border-strong);
    box-shadow: var(--shadow-lg);
    color: var(--fg);
  }
  .hint-card:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
  .hint-card.animate {
    animation: hint-in 160ms ease-out;
  }
  @keyframes hint-in {
    from {
      opacity: 0;
      transform: translateY(4px);
    }
  }
  .hint-eyebrow {
    margin: 0;
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--accent);
  }
  .hint-title {
    margin: 0;
    font-family: var(--font-display);
    font-size: 15px;
    line-height: 1.25;
    font-weight: 700;
    color: var(--fg);
  }
  .hint-body {
    margin: 0;
    font-size: 13px;
    line-height: 1.45;
    color: var(--fg-muted);
  }
  .hint-bar {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 2px;
  }
  .hint-grow {
    flex: 1 1 auto;
  }
  .hint-quiet {
    background: none;
    border: 0;
    padding: 6px 2px;
    font: 500 12px var(--font-ui);
    color: var(--fg-muted);
    cursor: pointer;
  }
  .hint-quiet:hover {
    color: var(--fg);
  }
  .hint-primary,
  .hint-action {
    display: inline-flex;
    align-items: center;
    height: 30px;
    padding: 0 14px;
    border-radius: var(--radius);
    font: 600 12.5px var(--font-ui);
    cursor: pointer;
    text-decoration: none;
  }
  .hint-primary {
    border: 0;
    background: var(--accent);
    color: var(--accent-fg);
  }
  .hint-primary:hover {
    background: var(--accent-strong);
  }
  .hint-action {
    border: 1px solid var(--border-strong);
    background: transparent;
    color: var(--fg);
  }
  .hint-sr {
    position: absolute;
    width: 1px;
    height: 1px;
    margin: -1px;
    padding: 0;
    overflow: hidden;
    clip: rect(0 0 0 0);
    white-space: nowrap;
    border: 0;
  }
  /* ADR 0125 §3.6 on a phone: a one-line strip on the bottom edge, in
     the same 6px gutter as the coach's strip; a tap opens it. */
  .hint-card.strip {
    left: 6px;
    right: 6px;
    width: auto;
    padding: 8px 12px;
  }
  .hint-card.strip.open {
    padding: 10px 12px;
  }
  .hint-strip-head {
    display: flex;
    align-items: center;
    gap: 10px;
    min-height: 32px;
    width: 100%;
    padding: 0;
    border: 0;
    background: none;
    color: inherit;
    text-align: left;
    cursor: pointer;
  }
  .hint-one-line {
    flex: 1 1 auto;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 13px;
  }
  @media (max-width: 599px) {
    .hint-quiet,
    .hint-primary,
    .hint-action {
      min-height: 44px;
    }
  }
</style>
