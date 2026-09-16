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
  //     parents (Hand, BattlefieldRow) can decide
  //     whether the click means "play", "tap", "select for combat",
  //     etc. without each path re-deriving from instance_id.
  //
  // Drag-to-reposition is intentionally deferred. The Pixi version
  // stamped a free-form (battle_x, battle_y) per card; with typed
  // rows that 2-D placement is meaningless. A follow-up can wire
  // within-row reordering when the UX is designed for it.

  import type { CardView } from "../../protocol";
  import { cardImageURL } from "../../cardImage";
  import { cardArt } from "../../cardArt";
  import { hoveredCard } from "../../cardTypes";
  import { animateTap } from "../../animations";
  import { play } from "../../sounds";
  import { settings } from "../../settings";
  import { targeting, isLegalCardTarget, isPicked } from "../../targeting";
  import { openCardMenu } from "../../contextMenu";
  import CounterPips from "./CounterPips.svelte";
  import KeywordBadgeRow from "./KeywordBadgeRow.svelte";
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
    // S21 sub-PR 2: same menu, CR 602 activated abilities. Set by
    // parents for battlefield permanents the viewer controls.
    onActivateAbility?: (abilityIndex: number) => void;
    // S31: why the CR 307.1 sorcery-speed window is shut, or "" when
    // it is open. Passed straight through to ManaAbilityMenu, which
    // greys `sorcery_speed` abilities with it. Card has no snapshot
    // of its own, and computing this per card would be wasteful —
    // the window is a property of the turn, so PlayerPanel derives it
    // once and hands it down.
    sorcerySpeedBlocked?: string;
    // S24 (ADR 0036 decision 14 item 3): the name of the player this
    // permanent enchants, for a Curse. A card attached to a PLAYER has
    // no host card to be drawn behind, so without this the board shows
    // a Curse of Opulence sitting in its controller's row with nothing
    // to say who it is cursing — which is the entire card. Supplied by
    // BattlefieldRow; undefined for everything else.
    enchantedPlayer?: string;
    // #33: request this card's art with fetchpriority="high". Opt-in,
    // set only by Hand.svelte for the viewer's own hand — the art
    // that is above the fold and latency-visible. Card is shared by
    // hand, battlefield, command zone and attachment stacks, and
    // marking every card on the table high is the same as marking
    // none of them, so the default is no hint at all.
    priority?: boolean;
    onClick?: (card: CardView, ev: MouseEvent) => void;
  }

  // S20: while a cast-targeting prompt is live, cards in the legal
  // set get a ring so the player can see what they may click.
  const targetable = $derived.by(() => {
    const t = $targeting;
    return t !== null && isLegalCardTarget(t, card.instance_id);
  });
  // S20 sub-PR 5: already picked in a multi-target prompt.
  const picked = $derived.by(() => {
    const t = $targeting;
    return t !== null && isPicked(t, card.instance_id);
  });

  const {
    card,
    faceDown = false,
    selected = false,
    attacking = false,
    blocking = false,
    size = "small",
    showManaCost = false,
    onActivateManaAbility,
    onActivateAbility,
    sorcerySpeedBlocked = "",
    enchantedPlayer,
    priority = false,
    onClick,
  }: Props = $props();

  // manaMenuOpen — Card-local state driving the ManaAbilityMenu
  // pop-over. Flipped true by oncontextmenu when the card has at
  // least one mana ability and a parent wired onActivateManaAbility.
  // Dismissed on selection, Escape (handled inside the menu), or
  // click elsewhere (the window-level onclick handler below).
  let manaMenuOpen = $state(false);
  const hasManaAbilities = $derived(
    (!!onActivateManaAbility && !!card.mana_abilities && card.mana_abilities.length > 0) ||
      (!!onActivateAbility && !!card.activated_abilities && card.activated_abilities.length > 0),
  );

  // cardImageURL defaults to the card's ACTIVE face, so a modal DFC
  // played as its land half — or, later, a transformed permanent —
  // shows the side that is actually up without this component
  // knowing faces exist.
  const imgSrc = $derived(cardImageURL(card, size));

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

  // Type-aware P/T overlay (S16): every creature card on the table
  // gets a small bottom-right pip showing its current power/toughness
  // — these are the post-layer effective values from the wire, so an
  // anthem-buffed creature shows the bumped numbers (3/3 instead of
  // 2/2 for a Bear under Glorious Anthem). Planeswalkers swap to a
  // loyalty counter pip; non-permanent or non-creature cards (instants,
  // sorceries, lands, artifacts/enchantments without Creature in the
  // type-line) get nothing — there's no P/T to show.
  const typeLine = $derived(card.type_line ?? "");
  const isCreature = $derived(/\bCreature\b/.test(typeLine));
  const isPlaneswalker = $derived(/\bPlaneswalker\b/.test(typeLine));
  const loyaltyValue = $derived(card.counters?.loyalty ?? 0);
  const showPT = $derived(!showBack && (isCreature || isPlaneswalker));

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

  // Right-click routing (#170). With admin overrides enabled the
  // gesture belongs to the per-card override menu, which subsumes the
  // ability popover — CardContextMenu lists mana / activated
  // abilities as its first section, so nothing is lost. With the
  // setting off (the default) the historic ability popover is the
  // only thing right-click does.
  function handleContextMenu(ev: MouseEvent): void {
    if ($settings.gameplay.adminOverrides) {
      ev.preventDefault();
      ev.stopPropagation();
      manaMenuOpen = false;
      openCardMenu({ card, x: ev.clientX, y: ev.clientY });
      return;
    }
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
  class:targetable
  class:picked
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
    <!-- use:cardArt (#33): retry once, then a click-to-retry pip.
         Front face only — the back above is a bundled asset. -->
    <img
      src={imgSrc}
      alt={card.name}
      loading="lazy"
      decoding="async"
      draggable="false"
      fetchpriority={priority ? "high" : undefined}
      use:cardArt={imgSrc}
    />
    {#if card.is_commander}
      <span class="badge cmd" aria-hidden="true">CMD</span>
    {/if}
    {#if card.goaded_by}
      <span class="badge goad" title="goaded" aria-label="goaded">GOAD</span>
    {/if}
    {#if enchantedPlayer}
      <span
        class="badge curse"
        title={`enchanting ${enchantedPlayer}`}
        aria-label={`enchanting ${enchantedPlayer}`}
      >
        ENCHANTING {enchantedPlayer}
      </span>
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
    <KeywordBadgeRow abilities={card.abilities} />
    {#if (card.damage_marked ?? 0) > 0}
      <span class="badge damage" title={`${card.damage_marked} damage marked`} aria-label="damage">
        {card.damage_marked}
      </span>
    {/if}
    {#if showPT}
      {#if isPlaneswalker}
        <span class="badge loyalty" title={`loyalty ${loyaltyValue}`} aria-label="loyalty">
          {loyaltyValue}
        </span>
      {:else}
        <span
          class="badge pt"
          title={`power/toughness ${card.power ?? 0}/${card.toughness ?? 0}`}
          aria-label="power/toughness"
        >
          {card.power ?? 0}/{card.toughness ?? 0}
        </span>
      {/if}
    {/if}
  {:else}
    <span class="name-fallback">{card.name}</span>
    {#if card.goaded_by}
      <span class="badge goad" title="goaded" aria-label="goaded">GOAD</span>
    {/if}
    {#if enchantedPlayer}
      <span
        class="badge curse"
        title={`enchanting ${enchantedPlayer}`}
        aria-label={`enchanting ${enchantedPlayer}`}
      >
        ENCHANTING {enchantedPlayer}
      </span>
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
    <KeywordBadgeRow abilities={card.abilities} />
    {#if (card.damage_marked ?? 0) > 0}
      <span class="badge damage" title={`${card.damage_marked} damage marked`} aria-label="damage">
        {card.damage_marked}
      </span>
    {/if}
    {#if showPT}
      {#if isPlaneswalker}
        <span class="badge loyalty" title={`loyalty ${loyaltyValue}`} aria-label="loyalty">
          {loyaltyValue}
        </span>
      {:else}
        <span
          class="badge pt"
          title={`power/toughness ${card.power ?? 0}/${card.toughness ?? 0}`}
          aria-label="power/toughness"
        >
          {card.power ?? 0}/{card.toughness ?? 0}
        </span>
      {/if}
    {/if}
  {/if}
  {#if manaMenuOpen && hasManaAbilities}
    <div class="mana-menu-anchor">
      <ManaAbilityMenu
        abilities={onActivateManaAbility ? (card.mana_abilities ?? []) : []}
        tapped={!!card.tapped}
        onActivate={(idx) => onActivateManaAbility?.(idx)}
        activated={onActivateAbility ? (card.activated_abilities ?? []) : []}
        onActivateAbility={(idx) => onActivateAbility?.(idx)}
        summoningSick={!!card.summoning_sick}
        {sorcerySpeedBlocked}
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
    /* The failed-art pip (#33) sits on the left edge, one badge row
       down. The top-right corner is the busiest on the tile — GOAD,
       the hand's cost chip and a counter column that grows downward
       with every counter type — and the left edge of an UNTAPPED tile
       is the part that stays visible where tiles overlap: the hand
       fan and its top-55% peek, the untapped land strip, attachments
       tucked behind their host. 22px clears the top badge row (CMD on
       the left; a GOAD or cost chip wide enough to reach across a
       narrow tile). z-index 5 keeps it above the counter column (4): on
       a tile under about 90px wide a wide chip (a two-digit count)
       reaches under the pip, which covers the chip's left end.
       The pip's z-index only counts inside this tile — the transform
       makes the tile its own stacking context — so a later tile that
       overlaps it always paints over it. boardArtPip.test.ts checks
       the rows. */
    --art-error-top: 22px;
    --art-error-left: 3px;
    --art-error-right: auto;
    --art-error-z: 5;
  }
  .card.tapped {
    /* A tapped tile turns 90° clockwise: its left edge becomes its top
       edge, and the further down the tile the pip sits, the further
       left it ends up. At 22px it lands on the right of the turned
       tile, under the next tapped land in the strip (which overlaps by
       35% of a width) and under a tapped neighbour in a battlefield
       row. Below about 54% of the height it is covered; 58% puts it on
       the uncovered left, clear of the AUTO badge and a single keyword
       row from 64px wide up. Hand tiles are never tapped, so the
       hand's peek keeps 22px. */
    --art-error-top: 58%;
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
  .badge.curse {
    /* A Curse is drawn in its controller's row with no host behind it,
       so the badge is the only thing on the card that names its
       victim. Full width across the bottom rather than a corner pip:
       it carries a player NAME, not a three-letter flag, and it must
       truncate rather than overflow the card. */
    top: auto;
    bottom: 3px;
    left: 3px;
    right: 3px;
    max-width: calc(100% - 6px);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    text-align: center;
    color: var(--danger);
    background: rgba(60, 0, 0, 0.85);
    border-color: rgba(255, 122, 122, 0.5);
    letter-spacing: 0.02em;
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
    /* Bottom-right; stacks above the P/T pip when both are showing
       (P/T is always-on for creatures so this is the common case
       any time a creature has been hit). */
    top: auto;
    bottom: 22px;
    left: auto;
    right: 3px;
    color: #ff9090;
    background: rgba(60, 0, 0, 0.9);
    border-color: rgba(255, 122, 122, 0.5);
    font-size: 11px;
  }
  .badge.pt,
  .badge.loyalty {
    /* Bottom-right (MTG card convention). Always-on for creatures
       and planeswalkers so the player can see effective P/T at a
       glance without hover-zooming — which especially matters once
       layered effects (anthems, CDAs) start mutating the numbers
       relative to the printed art. */
    top: auto;
    bottom: 3px;
    left: auto;
    right: 3px;
    font-family: ui-monospace, Menlo, monospace;
    font-size: 11px;
    letter-spacing: 0.02em;
    color: #f4ead5;
    background: rgba(10, 14, 26, 0.92);
    border: 1px solid rgba(180, 180, 180, 0.45);
  }
  .badge.loyalty {
    /* Loyalty pip — small green-tinted accent so planeswalkers'
       counter is visually distinct from a creature P/T. */
    color: #b8e0b8;
    background: rgba(20, 40, 20, 0.92);
    border-color: rgba(150, 200, 150, 0.5);
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
  .card.targetable {
    box-shadow:
      0 0 0 2px #6fe3a4,
      0 0 18px rgba(111, 227, 164, 0.6),
      0 6px 16px rgba(0, 0, 0, 0.5);
    cursor: crosshair;
  }
  .card.picked {
    box-shadow:
      0 0 0 3px #ffe69a,
      0 0 22px rgba(255, 230, 154, 0.7),
      0 6px 16px rgba(0, 0, 0, 0.5);
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
