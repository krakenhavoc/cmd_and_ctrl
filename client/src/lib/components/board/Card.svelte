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
  import ManaAbilityMenu from "./ManaAbilityMenu.svelte";

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
    // showManaCost renders the S15 cost-chip overlay bottom-left.
    // Enabled by Hand.svelte for the viewer's own hand so they can
    // see what each spell costs without hover-zooming. Hidden on the
    // battlefield (no value there) and on opponents' hands (would
    // leak identity even when the card is face-down, since the wire
    // redacts mana_cost for non-knowers anyway).
    showManaCost?: boolean;
    // onActivateManaAbility — when supplied, right-click / context-menu
    // opens the ManaAbilityMenu for the card's mana_abilities and
    // this callback fires with the chosen index. Parents set it on
    // battlefield cards the viewer controls; undefined suppresses
    // the menu entirely (hand cards, opponent permanents, zones
    // where activations aren't meaningful).
    onActivateManaAbility?: (abilityIndex: number) => void;
    onClick?: (card: CardView, ev: MouseEvent) => void;
  }

  const {
    card,
    faceDown = false,
    selected = false,
    attacking = false,
    blocking = false,
    size = "small",
    showManaCost = false,
    onActivateManaAbility,
    onClick,
  }: Props = $props();

  // manaMenuOpen — Card-local state driving the ManaAbilityMenu
  // pop-over. Flipped true by oncontextmenu when the card has at
  // least one mana ability and a parent wired onActivateManaAbility.
  // Dismissed on selection, Escape (handled inside the menu), or
  // click elsewhere (the window-level onclick handler below).
  let manaMenuOpen = $state(false);
  const hasManaAbilities = $derived(
    !!onActivateManaAbility && !!card.mana_abilities && card.mana_abilities.length > 0,
  );

  const imgSrc = $derived(
    card.scryfall_id ? `/cards/${card.scryfall_id}/image?size=${size}` : null,
  );

  // Real MTG card back bundled as a static asset under client/public.
  // Two sizes to keep hand/battlefield thumbnails snappy while the
  // hover-zoom overlay gets a crisper back. Served at /card-back*.jpg
  // by Vite and in production by whoever serves the built client.
  const backSrc = $derived(size === "normal" ? "/card-back.jpg" : "/card-back-small.jpg");

  // S13.5 — render the back when the wire says face-down. The
  // `|| card.known_by_you === false` arm is a belt-and-braces
  // fallback: the server's CardView uses `omitempty` on KnownByYou,
  // so a revealed card sends `true` and an unrevealed card omits
  // the field entirely (opponents never see an explicit `false`
  // from the wire). If a future code path were to build a CardView
  // locally with an explicit `{known_by_you: false}`, this check
  // would keep it rendering as a back. Parents that know the zone
  // (Hand.svelte for opponent cards) still set `faceDown` directly.
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
    // Suppress hover-zoom for any back-rendered card: the explicit
    // faceDown prop AND the server-redacted case (known_by_you ===
    // false). Either way the viewer doesn't have the characteristics
    // to reveal in the overlay.
    if (showBack) return;
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
    // Close any open mana-ability menu on a regular click; the
    // outer-click dismiss fires BEFORE the menu receives its
    // button click because the menu's onclick uses stopPropagation.
    if (manaMenuOpen) {
      manaMenuOpen = false;
      return;
    }
    onClick?.(card, ev);
  }

  function handleContextMenu(ev: MouseEvent): void {
    if (!hasManaAbilities) return;
    ev.preventDefault();
    ev.stopPropagation();
    manaMenuOpen = !manaMenuOpen;
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
  class:menu-open={manaMenuOpen}
  data-instance-id={card.instance_id}
  data-tapped={card.tapped ? "true" : "false"}
  role={onClick ? "button" : "img"}
  tabindex={onClick ? 0 : undefined}
  aria-label={showBack ? "face-down card" : card.name}
  title={showBack ? "" : card.name}
  onpointerenter={handleEnter}
  onpointerleave={handleLeave}
  onclick={handleClick}
  oncontextmenu={handleContextMenu}
  onkeydown={handleKeydown}
>
  {#if showBack}
    <img
      class="back-img"
      src={backSrc}
      alt="card back"
      loading="lazy"
      decoding="async"
      draggable="false"
    />
  {:else if imgSrc}
    <img src={imgSrc} alt={card.name} loading="lazy" decoding="async" draggable="false" />
    {#if card.is_commander}
      <span class="badge cmd" aria-hidden="true">CMD</span>
    {/if}
    {#if card.goaded_by}
      <span class="badge goad" title="goaded" aria-label="goaded">GOAD</span>
    {/if}
    {#if card.auto}
      <span
        class="badge auto"
        title="auto-resolving card — effect fires on resolve"
        aria-label="auto-resolving"
      >
        AUTO
      </span>
    {/if}
    {#if showManaCost && card.mana_cost}
      <span class="badge cost" title={`mana cost ${card.mana_cost}`} aria-label="mana cost">
        {card.mana_cost}
      </span>
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
    {#if card.auto}
      <span
        class="badge auto"
        title="auto-resolving card — effect fires on resolve"
        aria-label="auto-resolving"
      >
        AUTO
      </span>
    {/if}
    <CounterPips counters={card.counters} />
    {#if (card.damage_marked ?? 0) > 0}
      <span class="badge damage" title={`${card.damage_marked} damage marked`} aria-label="damage">
        {card.damage_marked}
      </span>
    {/if}
  {/if}
  {#if manaMenuOpen && hasManaAbilities && onActivateManaAbility && card.mana_abilities}
    <div class="mana-menu-anchor">
      <ManaAbilityMenu
        abilities={card.mana_abilities}
        tapped={!!card.tapped}
        onActivate={(idx) => onActivateManaAbility(idx)}
        onClose={() => (manaMenuOpen = false)}
      />
    </div>
  {/if}
</div>

<style>
  .card {
    /* Sizes are driven by inherited CSS vars so a parent panel can
       cascade smaller dimensions (e.g. opponent panels at the top of
       the board) without each Card needing a per-call prop. */
    width: var(--card-w, 80px);
    height: var(--card-h, 112px);
    border-radius: 8px;
    border: 1px solid #0a0e1a;
    background: #0d1220;
    overflow: hidden;
    position: relative;
    box-sizing: border-box;
    flex: 0 0 auto;
    box-shadow:
      0 2px 4px rgba(0, 0, 0, 0.5),
      inset 0 0 0 1px rgba(255, 255, 255, 0.05);
    transition:
      box-shadow 140ms var(--ease),
      filter 140ms var(--ease);
    transform-origin: center center;
    user-select: none;
    -webkit-user-select: none;
    /* Compose tap rotation (animated by GSAP via --tap-rot) with the
       CSS-only hover lift (--hover-lift). */
    transform: rotate(var(--tap-rot, 0deg)) translateY(var(--hover-lift, 0px));
  }
  .card.clickable {
    cursor: pointer;
  }
  .card.clickable:hover {
    --hover-lift: -8px;
    box-shadow:
      0 14px 28px rgba(0, 0, 0, 0.55),
      0 0 0 1px rgba(122, 167, 255, 0.35),
      inset 0 0 0 1px rgba(255, 255, 255, 0.08);
    filter: brightness(1.06);
    z-index: 5;
  }
  .card.face-down {
    background: #0f1428;
  }
  .back-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
    pointer-events: none;
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
    top: 3px;
    left: 3px;
    background: rgba(10, 14, 26, 0.88);
    color: var(--gold);
    font-size: 8px;
    font-weight: 800;
    padding: 2px 5px;
    border-radius: 999px;
    letter-spacing: 0.06em;
    border: 1px solid rgba(255, 208, 122, 0.45);
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.4);
    backdrop-filter: blur(4px);
  }
  .badge.goad {
    /* Top-right so it doesn't collide with the CMD badge on legendary
       commanders that get goaded back at their owner. */
    left: auto;
    right: 3px;
    color: var(--danger);
    background: rgba(60, 0, 0, 0.85);
    border-color: rgba(255, 122, 122, 0.5);
  }
  .badge.auto {
    /* Bottom-left so the AUTO pip sits opposite the damage badge
       (bottom-right) and below the CMD / GOAD badges (top). */
    top: auto;
    bottom: 3px;
    left: 3px;
    color: var(--gold);
    background: rgba(60, 44, 0, 0.85);
    border: 1px solid rgba(200, 168, 106, 0.7);
  }
  .badge.damage {
    /* Bottom-right so it stays clear of the goad / CMD badges. */
    top: auto;
    bottom: 3px;
    left: auto;
    right: 3px;
    color: #ff9090;
    background: rgba(60, 0, 0, 0.9);
    border-color: rgba(255, 122, 122, 0.5);
    font-size: 11px;
  }
  .card.menu-open {
    /* When the mana-ability menu is open, let the popover extend
       past the card frame. The menu itself carries its own border /
       shadow, so loosening the clip here doesn't fight other art. */
    overflow: visible;
  }
  .mana-menu-anchor {
    /* Float the menu below the card. Anchored via the .card's
       position: relative; .card.menu-open disables overflow:hidden
       so the pop-over extends past the card frame without
       needing a portal. */
    position: absolute;
    top: 100%;
    left: 0;
    margin-top: 4px;
    z-index: 60;
  }
  .badge.cost {
    /* Top-right to mirror the printed-card convention. Only shown in
       hand-zone presentations via the showManaCost prop, so no clash
       with the goad / damage battlefield badges. Monospace so cost
       strings like "{W}{U}{B}{R}{G}" stay legible at small sizes. */
    left: auto;
    right: 3px;
    top: 3px;
    font-family: ui-monospace, Menlo, monospace;
    font-size: 9px;
    letter-spacing: 0;
    color: var(--gold);
    background: rgba(10, 14, 26, 0.92);
    border: 1px solid rgba(200, 168, 106, 0.5);
    max-width: 72%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .card.selected {
    box-shadow:
      0 0 0 2px var(--gold),
      0 0 22px rgba(255, 208, 122, 0.6),
      0 10px 22px rgba(0, 0, 0, 0.5),
      inset 0 0 0 1px rgba(255, 255, 255, 0.08);
  }
  .card.attacking {
    box-shadow:
      0 0 0 2px var(--danger),
      0 0 18px rgba(255, 122, 122, 0.55),
      0 6px 16px rgba(0, 0, 0, 0.5);
  }
  .card.blocking {
    box-shadow:
      0 0 0 2px #9ec7ff,
      0 0 18px rgba(158, 199, 255, 0.5),
      0 6px 16px rgba(0, 0, 0, 0.5);
  }
  .card:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 3px;
  }
</style>
