<script lang="ts">
  // Card is the visual primitive for one Magic card in the new HTML/
  // CSS board. Replaces the Pixi CardTile from client/src/lib/card-tile.ts.
  //
  // Conventions:
  //   - Tap rotation is 90° clockwise via CSS transform.
  //   - Face-down cards never request art, never write to the hover-
  //     zoom store, and render as a solid back rectangle.
  //   - The browser's image cache + the server's `/cards/:id/image`
  //     endpoint replace the Pixi Assets cache; second hover of the
  //     same card paints from cache instantly.
  //   - Click vs. tap: the parent supplies onClick. The card itself
  //     just dispatches the event with the underlying CardView so
  //     parents (Hand, BattlefieldRow, BattlefieldColumn) can decide
  //     whether the click means "play", "tap", "select for combat",
  //     etc. without each path re-deriving from instance_id.
  //
  // Drag-to-reposition is intentionally deferred. The Pixi version
  // stamped a free-form (battle_x, battle_y) per card; with typed
  // rows that 2-D placement is meaningless. A follow-up can wire
  // within-row reordering when the UX is designed for it.

  import type { CardView } from "../../protocol";
  import { hoveredCard } from "../../cardTypes";
  import { animateTap } from "../../animations";
  import { play } from "../../sounds";
  import { settings } from "../../settings";
  import CounterPips from "./CounterPips.svelte";

  interface Props {
    card: CardView;
    faceDown?: boolean;
    // Combat / selection visual states. Translate to coloured rings.
    selected?: boolean;
    attacking?: boolean;
    blocking?: boolean;
    // size: which Scryfall-resolved image to request from the server.
    // "small" is the default (~146×204) and is what we use everywhere
    // on the table; the hover zoom overlay requests "normal".
    size?: "small" | "normal";
    onClick?: (card: CardView, ev: MouseEvent) => void;
  }

  const {
    card,
    faceDown = false,
    selected = false,
    attacking = false,
    blocking = false,
    size = "small",
    onClick,
  }: Props = $props();

  const imgSrc = $derived(
    card.scryfall_id ? `/cards/${card.scryfall_id}/image?size=${size}` : null,
  );

  // S13.5 — render the back when the wire says face-down OR when
  // the card is one the viewer doesn't know (server redacted the
  // characteristics, so we have nothing meaningful to render face-
  // up). known_by_you is omitted when true, so the inverse is
  // "explicit false" — `=== false` distinguishes "redacted card"
  // from "older client / pre-S13.5 wire / face-up by default".
  const showBack = $derived(faceDown || card.known_by_you === false);

  // Hover delay (settings.display.hoverDelayMs) defers the write to
  // the hoveredCard store until the user has rested on the card for
  // the configured duration. Defaults to 300ms so a fast mouse-over
  // sweep doesn't flash the zoom panel on every card in its path.
  // A per-card timer is cancelled on leave / click / unmount.
  let hoverTimer: ReturnType<typeof setTimeout> | null = null;

  function cancelHoverTimer(): void {
    if (hoverTimer !== null) {
      clearTimeout(hoverTimer);
      hoverTimer = null;
    }
  }

  function handleEnter(): void {
    if (faceDown) return;
    cancelHoverTimer();
    const delay = $settings.display.hoverDelayMs;
    if (delay <= 0) {
      hoveredCard.set(card);
      return;
    }
    hoverTimer = setTimeout(() => {
      hoverTimer = null;
      hoveredCard.set(card);
    }, delay);
  }

  function handleLeave(): void {
    // Cancel any pending delayed-show so the zoom doesn't pop up
    // after the cursor has already left.
    cancelHoverTimer();
    // Only clear if we still own the slot — guards against a snapshot
    // rebuild that swaps the hovered card out from under us before the
    // pointerleave fires.
    hoveredCard.update((c) => (c?.instance_id === card.instance_id ? null : c));
  }

  function handleClick(ev: MouseEvent): void {
    onClick?.(card, ev);
  }

  function handleKeydown(ev: KeyboardEvent): void {
    if (ev.key !== "Enter" && ev.key !== " ") return;
    ev.preventDefault();
    onClick?.(card, ev as unknown as MouseEvent);
  }

  // Drive the tap rotation with GSAP so it eases instead of snapping
  // and so future combat / damage effects can sequence against it.
  // The rotation lives in --tap-rot (animated) while the hover lift
  // lives in --hover-lift (CSS-only); composing via custom properties
  // keeps the two effects independent.
  let cardEl: HTMLDivElement | undefined = $state();
  // Fire the tap SFX only on the untapped→tapped transition, not on
  // initial mount (for cards that arrive already tapped via snapshot)
  // and not on the tapped→untapped direction (the turn-start `untap_all`
  // cue covers the bulk untap, and individual manual untaps don't need
  // a dedicated sound).
  let prevTapped = false;
  let tapEffectHasRun = false;
  $effect(() => {
    if (!cardEl) return;
    const tapped = !!card.tapped;
    animateTap(cardEl, tapped);
    if (tapEffectHasRun && tapped && !prevTapped) play("tap");
    prevTapped = tapped;
    tapEffectHasRun = true;
  });
</script>

<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
<div
  bind:this={cardEl}
  class="card"
  class:face-down={showBack}
  class:tapped={card.tapped}
  class:selected
  class:attacking
  class:blocking
  class:clickable={!!onClick}
  data-instance-id={card.instance_id}
  data-tapped={card.tapped ? "true" : "false"}
  role={onClick ? "button" : "img"}
  tabindex={onClick ? 0 : undefined}
  aria-label={showBack ? "face-down card" : card.name}
  title={showBack ? "" : card.name}
  onpointerenter={handleEnter}
  onpointerleave={handleLeave}
  onclick={handleClick}
  onkeydown={handleKeydown}
>
  {#if showBack}
    <div class="back"></div>
  {:else if imgSrc}
    <img src={imgSrc} alt={card.name} loading="lazy" decoding="async" draggable="false" />
    {#if card.is_commander}
      <span class="badge cmd" aria-hidden="true">CMD</span>
    {/if}
    {#if card.goaded_by}
      <span class="badge goad" title="goaded" aria-label="goaded">GOAD</span>
    {/if}
    <CounterPips counters={card.counters} />
    {#if (card.damage_marked ?? 0) > 0}
      <span class="badge damage" title={`${card.damage_marked} damage marked`} aria-label="damage">
        {card.damage_marked}
      </span>
    {/if}
  {:else}
    <span class="name-fallback">{card.name}</span>
    {#if card.goaded_by}
      <span class="badge goad" title="goaded" aria-label="goaded">GOAD</span>
    {/if}
    <CounterPips counters={card.counters} />
    {#if (card.damage_marked ?? 0) > 0}
      <span class="badge damage" title={`${card.damage_marked} damage marked`} aria-label="damage">
        {card.damage_marked}
      </span>
    {/if}
  {/if}
</div>

<style>
  .card {
    /* Sizes are driven by inherited CSS vars so a parent panel can
       cascade smaller dimensions (e.g. opponent panels at the top of
       the board) without each Card needing a per-call prop. Defaults
       in the var() fallback match the prior intrinsic size. */
    width: var(--card-w, 80px);
    height: var(--card-h, 112px);
    border-radius: 6px;
    border: 1px solid #4a5270;
    background: #1e2638;
    overflow: hidden;
    position: relative;
    box-sizing: border-box;
    flex: 0 0 auto;
    transition: box-shadow 80ms ease;
    transform-origin: center center;
    user-select: none;
    -webkit-user-select: none;
    /* Compose tap rotation (animated by GSAP via --tap-rot) with the
       CSS-only hover lift (--hover-lift). Stacking through CSS vars
       lets each effect tween independently without one stomping the
       other's transform string. */
    transform: rotate(var(--tap-rot, 0deg)) translateY(var(--hover-lift, 0px));
  }
  .card.clickable {
    cursor: pointer;
  }
  .card.clickable:hover {
    /* --hover-lift snaps rather than tweens; smoothly transitioning a
       custom property requires @property registration which Svelte's
       scoped CSS doesn't expose. The 6px snap is small enough not to
       read as a jump cut, and the box-shadow + GSAP tap rotation
       still animate. */
    --hover-lift: -6px;
    box-shadow: 0 6px 14px rgba(0, 0, 0, 0.45);
    z-index: 5;
  }
  .card.face-down {
    background: #1a1f35;
  }
  .back {
    width: 100%;
    height: 100%;
    background: repeating-linear-gradient(45deg, #1a1f35, #1a1f35 6px, #232842 6px, #232842 12px);
  }
  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
    pointer-events: none;
  }
  .name-fallback {
    display: -webkit-box;
    -webkit-line-clamp: 4;
    line-clamp: 4;
    -webkit-box-orient: vertical;
    overflow: hidden;
    padding: 6px 4px;
    font-size: 10px;
    line-height: 1.15;
    text-align: center;
    color: #e0e6f5;
  }
  .badge {
    position: absolute;
    top: 2px;
    left: 2px;
    background: rgba(0, 0, 0, 0.7);
    color: #ffd07a;
    font-size: 8px;
    font-weight: 700;
    padding: 1px 3px;
    border-radius: 2px;
    letter-spacing: 0.05em;
  }
  .badge.goad {
    /* Top-right so it doesn't collide with the CMD badge on legendary
       commanders that get goaded back at their owner. */
    left: auto;
    right: 2px;
    color: #ff7a7a;
    background: rgba(80, 0, 0, 0.85);
  }
  .badge.damage {
    /* Bottom-right so it stays clear of the goad / CMD badges and
       sits next to the card's printed P/T conceptually. S13.2 — the
       lethal-damage SBA reads damage_marked. */
    top: auto;
    bottom: 2px;
    left: auto;
    right: 2px;
    color: #ff9090;
    background: rgba(80, 0, 0, 0.9);
    font-size: 11px;
  }
  .card.selected {
    box-shadow:
      0 0 0 3px #ffd07a,
      0 6px 14px rgba(0, 0, 0, 0.45);
  }
  .card.attacking {
    box-shadow: 0 0 0 2px #ff7a7a;
  }
  .card.blocking {
    box-shadow: 0 0 0 2px #9ec7ff;
  }
  .card:focus-visible {
    outline: 2px solid #5fb0ff;
    outline-offset: 2px;
  }
</style>
