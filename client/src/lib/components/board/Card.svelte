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
  import { showsCardBack } from "../../cardBack";
  import { hoveredCard } from "../../cardTypes";
  import { animateTap } from "../../animations";
  import { play } from "../../sounds";
  import { settings } from "../../settings";
  import { targeting, isLegalCardTarget, isPicked } from "../../targeting";
  import { autoTapHighlight } from "../../dragCast";
  import { noUntapAppliesToController } from "../../noUntap";
  import { openCardMenu } from "../../contextMenu";
  import CounterPips from "./CounterPips.svelte";
  import KeywordBadgeRow from "./KeywordBadgeRow.svelte";
  import ManaAbilityMenu from "./ManaAbilityMenu.svelte";
  import RoomDoorStrip from "./RoomDoorStrip.svelte";
  import { displayName } from "../../faces";
  import {
    NO_LEGAL_ACTIONS,
    NO_PIPS,
    pipCount,
    type LegalActions,
    type ReadyPips,
  } from "../../legalActions";

  interface Props {
    card: CardView;
    faceDown?: boolean;
    // Combat / selection visual states. Translate to coloured rings.
    selected?: boolean;
    attacking?: boolean;
    blocking?: boolean;
    // ADR 0105 (#1789): the viewer has a legal action with this card
    // right now — the cyan "ready" ring. The caller decides (from
    // legalActions.ts, and only while highlights are live); the card
    // only draws it.
    ready?: boolean;
    // ADR 0105 §2 (sub-PR 3): the pips on a permanent: a bolt for
    // live activated abilities (with a count from two up), a drop for
    // a mana ability worth marking (§4). The caller reads them off the
    // lookup (legalActions.ts readyPips), as it does `ready`. The card
    // only draws them. NO_PIPS draws none.
    pips?: ReadyPips;
    // ADR 0105: the lookups the ability popover reads for this card.
    // `legal` is what may be highlighted: ready rows take the accent
    // and sort first. `legalGate` is the frame's full lookup, for the
    // popover's sorcery-speed gate. The highlight setting never
    // touches it. Both default to "no information".
    legal?: LegalActions;
    legalGate?: LegalActions;
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
    // battlefield cards the viewer controls, and since #1228 on the
    // viewer's own hand cards — a Spirit Guide's "Exile this card
    // from your hand: Add {R}" is a mana ability that functions there
    // (CR 113.6) and rides `zone_mana_abilities`. Undefined
    // suppresses the menu entirely (opponent permanents, zones where
    // activations aren't meaningful).
    onActivateManaAbility?: (abilityIndex: number) => void;
    // #1438: a left-click on a mana source now taps it FOR mana, so
    // the menu carries the plain tap as "Tap (no mana)". Set by
    // BattlefieldRow on the viewer's own permanents; undefined hides
    // the row (hand cards, opponents).
    onRawTap?: () => void;
    // S21 sub-PR 2: same menu, CR 602 activated abilities. Set by
    // parents for battlefield permanents the viewer controls, and
    // since #660 by Hand.svelte for the viewer's own hand — a card in
    // hand offers the abilities that function THERE (cycling), which
    // ride `zone_abilities` rather than `activated_abilities`. One
    // callback for both: the index means the same thing on the wire.
    onActivateAbility?: (abilityIndex: number) => void;
    // S31: why the CR 307.1 sorcery-speed window is shut, or "" when
    // it is open. Passed straight through to ManaAbilityMenu, which
    // since ADR 0105 uses it only as the WORDS for a row the server
    // has shut (`timing_closed`, the digest). Card has no snapshot
    // of its own, and computing this per card would be wasteful —
    // the window is a property of the turn, so PlayerPanel derives it
    // once and hands it down.
    sorcerySpeedBlocked?: string;
    // #1695: the paying player's current life, for the mana popover's
    // life-cost check. Same threading as sorcerySpeedBlocked — a
    // battlefield row only ever holds one seat's own permanents, so
    // BattlefieldRow hands down that seat's life once per row.
    // Undefined leaves the popover's life-cost check unblocked, same
    // as every card surface this prop hasn't reached yet.
    payerLife?: number;
    // S24 (ADR 0036 decision 14 item 3): the name of the player this
    // permanent enchants, for a Curse. A card attached to a PLAYER has
    // no host card to be drawn behind, so without this the board shows
    // a Curse of Opulence sitting in its controller's row with nothing
    // to say who it is cursing — which is the entire card. Supplied by
    // BattlefieldRow; undefined for everything else.
    enchantedPlayer?: string;
    // ADR 0104 (owner decision 6): the owner's name when another player
    // controls this permanent — a stolen creature, or the permanent a
    // stolen spell became. Supplied by BattlefieldRow; undefined when
    // the controller is the owner.
    takenFrom?: string;
    // #33: request this card's art with fetchpriority="high". Opt-in,
    // set only by Hand.svelte for the viewer's own hand — the art
    // that is above the fold and latency-visible. Card is shared by
    // hand, battlefield, command zone and attachment stacks, and
    // marking every card on the table high is the same as marking
    // none of them, so the default is no hint at all.
    priority?: boolean;
    // #1724: a token group's card stands for every member of its
    // half, so the targeting ring asks about all of them — a group
    // whose drawn member is not a legal target but another member is
    // still lights up, and never hides a legal target.
    memberIDs?: readonly string[];
    // ADR 0103: the viewing seat, so a Room's door strip offers its
    // unlock buttons only to the Room's controller (CR 709.5e). Set by
    // BattlefieldRow; undefined everywhere else, which offers none.
    viewerID?: string | null;
    onClick?: (card: CardView, ev: MouseEvent) => void;
  }

  // S20: while a cast-targeting prompt is live, cards in the legal
  // set get a ring so the player can see what they may click.
  const targetable = $derived.by(() => {
    const t = $targeting;
    if (t === null) return false;
    return (memberIDs ?? [card.instance_id]).some((id) => isLegalCardTarget(t, id));
  });
  // S20 sub-PR 5: already picked in a multi-target prompt.
  const picked = $derived.by(() => {
    const t = $targeting;
    if (t === null) return false;
    return (memberIDs ?? [card.instance_id]).some((id) => isPicked(t, id));
  });

  const {
    card,
    faceDown = false,
    selected = false,
    attacking = false,
    blocking = false,
    ready = false,
    pips = NO_PIPS,
    legal = NO_LEGAL_ACTIONS,
    legalGate = NO_LEGAL_ACTIONS,
    size = "small",
    showManaCost = false,
    onActivateManaAbility,
    onRawTap,
    onActivateAbility,
    sorcerySpeedBlocked = "",
    payerLife,
    enchantedPlayer,
    takenFrom,
    priority = false,
    memberIDs,
    viewerID,
    onClick,
  }: Props = $props();

  // manaMenuOpen — Card-local state driving the ManaAbilityMenu
  // pop-over. Flipped true by oncontextmenu when the card has at
  // least one mana ability and a parent wired onActivateManaAbility.
  // Dismissed on selection, Escape (handled inside the menu), or
  // click elsewhere (the window-level onclick handler below).
  let manaMenuOpen = $state(false);
  // #660: a card projects EITHER list, never both — the server
  // filters by the zone the card is in (CR 113.6) — so one menu reads
  // whichever is present and the indices stay the card's own.
  const menuAbilities = $derived(card.activated_abilities ?? card.zone_abilities ?? []);
  // #1228: and the same sentence for the CR 605 list. A permanent
  // publishes `mana_abilities`; a card in hand whose mana ability
  // functions there (a Spirit Guide) publishes `zone_mana_abilities`,
  // and the index means the same thing on the wire either way.
  const menuManaAbilities = $derived(card.mana_abilities ?? card.zone_mana_abilities ?? []);
  const hasManaAbilities = $derived(
    (!!onActivateManaAbility && menuManaAbilities.length > 0) ||
      (!!onActivateAbility && menuAbilities.length > 0),
  );
  // ADR 0105: a pip is drawn only where the popover it points at is
  // wired. A pip on a card whose abilities this viewer cannot open is
  // worse than none (ADR 0105, Context, fact 3).
  const boltPips = $derived(onActivateAbility ? pips.abilities : 0);
  const dropPip = $derived(!!onActivateManaAbility && pips.mana);

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

  // S13.5 / #95 — render the back for a face-down card the viewer
  // doesn't know, for an explicit `known_by_you: false`, or when the
  // parent says so (Hand.svelte for opponent cards). The rule, and
  // why "doesn't know" is `!== true` rather than `=== false`, lives
  // in cardBack.ts.
  const showBack = $derived(showsCardBack(card, faceDown));

  // #1199 / CR 702.26: a phased-out permanent. It arrives in
  // GameView.phased_out rather than in the battlefield zone
  // (ADR 0084), and Board.svelte folds it back onto its controller's
  // row so the player can see WHERE it was. Dimmed, badged and inert:
  // "treated as though it does not exist" means nothing may be done
  // to it, so the click affordance is withheld here rather than by
  // every parent remembering to withhold it.
  const phasedOut = $derived(card.phased_out === true);
  const interactive = $derived(!!onClick && !phasedOut);

  // ADR 0069 — a face-down object the viewer IS allowed to look at:
  // the controller of their own morph or manifest (CR 708.5), the
  // owner of their own foretold card (CR 702.143d). They see the real
  // face, because hiding their own card from them helps nobody, plus
  // a badge saying the table sees a back. `face_visible` is the
  // server's answer to that permission; the `face_down` fallback keeps
  // the badge on a locally-built CardView that predates the field.
  const showFaceDownBadge = $derived(
    !showBack && (card.face_visible === true || card.face_down === true),
  );
  // The kind is public, so a card back can say WHAT it is rather than
  // just that something is there.
  const faceDownLabel = $derived(
    card.face_down_kind ? card.face_down_kind.toUpperCase() : "FACE DOWN",
  );

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
  const showNoUntap = $derived(noUntapAppliesToController(card));

  // ADR 0071 — the designation badge. A Class's level (CR 716.2) and
  // a Case's solved marker (CR 719.3) say WHICH of the printed lines
  // on the card are live right now, which is the one thing a player
  // cannot read off the art. Both are public, and the server sends
  // `class_level` only for a Class on the battlefield, so presence is
  // the whole test.
  //
  // One badge slot, not two: no printed permanent is both a Class and
  // a Case, so they cannot collide, and giving them one slot keeps
  // the top edge of the card readable next to CMD and GOAD.
  //
  // ADR 0090 adds a third tenant: a preparation creature's PREPARED
  // designation (CR 722.3a), which says its prepare spell is waiting
  // in exile to be cast. A Class or a Case is never a preparation
  // card, so the slot still holds one badge at most.
  //
  // ADR 0071's addendums add a fourth and fifth tenant, #1705:
  // HARNESSED (CR 701.64) and MONSTROUS (CR 701.37b). Unlike Class /
  // Case / preparation, neither carries a subtype gate — any
  // permanent can be harnessed, any creature can become monstrous —
  // so in principle a card could someday carry two of these at once
  // (a harnessed monstrous creature). No printed card does today, and
  // this slot still shows at most one badge; if that combination ever
  // ships, this priority chain is where to widen it.
  const designationBadge = $derived(
    card.solved
      ? "SOLVED"
      : (card.class_level ?? 0) > 0
        ? `LVL ${card.class_level}`
        : card.prepared
          ? "PREPARED"
          : card.harnessed
            ? "HARNESSED"
            : card.monstrous
              ? "MONSTROUS"
              : "",
  );
  const designationTitle = $derived(
    card.solved
      ? "this Case is solved"
      : card.prepared
        ? "prepared — you may cast a copy of its spell from exile"
        : card.harnessed
          ? "harnessed — its ∞ ability lines are on (CR 701.64)"
          : card.monstrous
            ? 'monstrous — its "as long as this creature is monstrous" lines are on (CR 701.37b)'
            : `Class level ${card.class_level ?? 1}`,
  );

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
    // CR 702.26b: a phased-out permanent "can't affect or be affected
    // by anything else in the game", so there is nothing to click.
    if (phasedOut) return;
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
  class:autotap-planned={$autoTapHighlight.has(card.instance_id)}
  class:attacking
  class:blocking
  class:ready
  class:clickable={interactive}
  class:phased-out={phasedOut}
  class:menu-open={manaMenuOpen}
  data-instance-id={card.instance_id}
  data-tapped={card.tapped ? "true" : "false"}
  role={interactive ? "button" : "img"}
  tabindex={interactive ? 0 : undefined}
  aria-label={showBack ? "face-down card" : displayName(card)}
  title={showBack ? "" : displayName(card)}
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
    {#if card.face_down_kind}
      <!-- The kind is public (ADR 0069): the table can see that a
           permanent is a morph and that an exiled card is foretold,
           even though nobody may look at the face. -->
      <span class="badge face-down">{faceDownLabel}</span>
    {/if}
  {:else if imgSrc}
    <!-- use:cardArt (#33): retry once, then a click-to-retry pip.
         Front face only — the back above is a bundled asset. -->
    <img
      src={imgSrc}
      alt={displayName(card)}
      loading="lazy"
      decoding="async"
      draggable="false"
      fetchpriority={priority ? "high" : undefined}
      use:cardArt={imgSrc}
    />
    {#if showFaceDownBadge}
      <!-- ADR 0069: the viewer may look at this face (CR 708.5 for a
           permanent they control, CR 702.143d for their own foretold
           card), so they get the real art — and the badge, because
           everyone else is looking at a card back. -->
      <span
        class="badge face-down"
        title="face down — only you may look at this card"
        aria-label={`face down: ${faceDownLabel}`}
      >
        {faceDownLabel}
      </span>
    {/if}
    {#if phasedOut}
      <!-- CR 702.26. Without it the board simply stops showing the
           permanent, which reads identically to one that died. -->
      <span
        class="badge phased"
        title="phased out — treated as though it doesn't exist"
        aria-label="phased out"
      >
        PHASED
      </span>
    {/if}
    {#if card.is_commander}
      <span class="badge cmd" aria-hidden="true">CMD</span>
    {/if}
    {#if card.goaded_by}
      <span class="badge goad" title="goaded" aria-label="goaded">GOAD</span>
    {/if}
    {#if card.must_attack}
      <span class="badge must-attack" title="must attack this combat" aria-label="must attack"
        >MUST ATTACK</span
      >
    {/if}
    {#if card.must_block}
      <span class="badge must-attack" title="must block this combat" aria-label="must block"
        >MUST BLOCK</span
      >
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
    {#if takenFrom}
      <!-- ADR 0104: a permanent another player controls says whose it
           is — a stolen creature, or the permanent a stolen spell became. -->
      <span
        class="badge taken"
        title={`owned by ${takenFrom} — another player controls it`}
        aria-label={`taken from ${takenFrom}`}
      >
        TAKEN FROM {takenFrom}
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
    <KeywordBadgeRow
      abilities={card.abilities}
      chosenColor={card.chosen_color}
      namedTribe={card.named_tribe}
      chosenOption={card.chosen_option}
      chosenName={card.chosen_name}
      protection={card.protection}
    />
    {#if (card.damage_marked ?? 0) > 0}
      <span class="badge damage" title={`${card.damage_marked} damage marked`} aria-label="damage">
        {card.damage_marked}
      </span>
    {/if}
    {#if (card.regeneration_shields ?? 0) > 0}
      <span
        class="badge regen"
        title={`${card.regeneration_shields} regeneration shield${(card.regeneration_shields ?? 0) === 1 ? "" : "s"} — replaces the next destruction this turn`}
        aria-label="regeneration shield"
      >
        REGEN{(card.regeneration_shields ?? 0) > 1 ? ` x${card.regeneration_shields}` : ""}
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
    <span class="name-fallback">{displayName(card)}</span>
    {#if card.token_text}
      <!-- ADR 0083: a token has no printing, so no art and no oracle
           text — the server sends what the token prints. A trigger or
           a static has no control to read it off, unlike an activated
           ability's menu row, so without this the words are nowhere. -->
      <span class="token-text" title={card.token_text}>{card.token_text}</span>
    {/if}
    {#if showFaceDownBadge}
      <!-- ADR 0069: the viewer may look at this face (CR 708.5 for a
           permanent they control, CR 702.143d for their own foretold
           card), so they get the real art — and the badge, because
           everyone else is looking at a card back. -->
      <span
        class="badge face-down"
        title="face down — only you may look at this card"
        aria-label={`face down: ${faceDownLabel}`}
      >
        {faceDownLabel}
      </span>
    {/if}
    {#if card.goaded_by}
      <span class="badge goad" title="goaded" aria-label="goaded">GOAD</span>
    {/if}
    {#if card.must_attack}
      <span class="badge must-attack" title="must attack this combat" aria-label="must attack"
        >MUST ATTACK</span
      >
    {/if}
    {#if card.must_block}
      <span class="badge must-attack" title="must block this combat" aria-label="must block"
        >MUST BLOCK</span
      >
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
    {#if takenFrom}
      <!-- ADR 0104: a permanent another player controls says whose it
           is — a stolen creature, or the permanent a stolen spell became. -->
      <span
        class="badge taken"
        title={`owned by ${takenFrom} — another player controls it`}
        aria-label={`taken from ${takenFrom}`}
      >
        TAKEN FROM {takenFrom}
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
    <KeywordBadgeRow
      abilities={card.abilities}
      chosenColor={card.chosen_color}
      namedTribe={card.named_tribe}
      chosenOption={card.chosen_option}
      chosenName={card.chosen_name}
      protection={card.protection}
    />
    {#if (card.damage_marked ?? 0) > 0}
      <span class="badge damage" title={`${card.damage_marked} damage marked`} aria-label="damage">
        {card.damage_marked}
      </span>
    {/if}
    {#if (card.regeneration_shields ?? 0) > 0}
      <span
        class="badge regen"
        title={`${card.regeneration_shields} regeneration shield${(card.regeneration_shields ?? 0) === 1 ? "" : "s"} — replaces the next destruction this turn`}
        aria-label="regeneration shield"
      >
        REGEN{(card.regeneration_shields ?? 0) > 1 ? ` x${card.regeneration_shields}` : ""}
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
  {#if showNoUntap}
    <span class="badge no-untap" title="won't untap" aria-label="won't untap">WON'T UNTAP</span>
  {/if}
  {#if designationBadge}
    <!-- ADR 0071: outside the art / name-fallback branches on
         purpose — a Class or a Case says the same thing whether or
         not its art loaded. -->
    <span class="badge designation" title={designationTitle} aria-label={designationTitle}>
      {designationBadge}
    </span>
  {/if}
  {#if boltPips > 0 || dropPip}
    <!-- ADR 0105 §2 (#1789): what this permanent can do right now, by
         shape: a bolt for an activated ability, a drop for a mana
         ability worth marking (§4). Display-only in this sub-PR, so
         the right-click and the click land on the card. Sub-PR 4 makes
         a pip tap-open the popover, and sub-PR 6 gives it accessible
         names. -->
    <span class="ready-pips" aria-hidden="true">
      {#if boltPips > 0}
        <span class="ready-pip bolt" data-pip="bolt">
          <svg viewBox="0 0 24 24" focusable="false"
            ><path d="M13.5 2 4 13.5h6.5L9.5 22 20 9.5h-6.5z" /></svg
          >
          {#if pipCount(boltPips)}<span class="pip-count">{pipCount(boltPips)}</span>{/if}
        </span>
      {/if}
      {#if dropPip}
        <span class="ready-pip drop" data-pip="drop">
          <svg viewBox="0 0 24 24" focusable="false"
            ><path d="M12 2.5S5 10.4 5 15.2a7 7 0 0 0 14 0C19 10.4 12 2.5 12 2.5z" /></svg
          >
        </span>
      {/if}
    </span>
  {/if}
  {#if card.doors && !showBack}
    <!-- ADR 0103: a Room's doors; the unlock buttons are its
         controller's (CR 709.5e). -->
    <RoomDoorStrip {card} canUnlock={!!viewerID && card.controller === viewerID} />
  {/if}
  {#if manaMenuOpen && hasManaAbilities}
    <div class="mana-menu-anchor">
      <ManaAbilityMenu
        abilities={onActivateManaAbility ? menuManaAbilities : []}
        tapped={!!card.tapped}
        onActivate={(idx) => onActivateManaAbility?.(idx)}
        activated={onActivateAbility ? menuAbilities : []}
        onActivateAbility={(idx) => onActivateAbility?.(idx)}
        summoningSick={!!card.summoning_sick}
        timingReason={sorcerySpeedBlocked}
        cardID={card.instance_id}
        {legal}
        {legalGate}
        {payerLife}
        onRawTap={onRawTap && onActivateManaAbility && menuManaAbilities.length > 0 && !card.tapped
          ? onRawTap
          : undefined}
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
       fan and its top-55% peek, the untapped land strip. Not an
       attachment tucked behind its host: once either of them is
       tapped, the host covers most of that edge, so BattlefieldRow
       moves an attachment's pip to its bottom-left corner. 22px
       clears the top badge row (CMD on the left; a GOAD or cost chip
       wide enough to reach across a narrow tile). z-index 5 keeps it
       above the counter column (4): on a tile under about 90px wide a
       wide chip (a two-digit count) reaches under the pip, which
       covers the chip's left end.
       The pip's z-index only counts inside this tile — the transform
       makes the tile its own stacking context — so a later tile that
       overlaps it always paints over it. boardArtPip.test.ts checks
       the rows and the attachment stacks. */
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
  .card.phased-out {
    /* The board-freeze idiom (Board.svelte .board-disabled), because
       it says the same thing: this is here and you may not act on it.
       The badge is the message; the dimming is the tone. Hover-zoom
       still works — a player looking for what phased out wants to
       read it — and only the click is withheld, in the script. */
    opacity: 0.45;
    filter: grayscale(0.7) brightness(0.9);
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
  /* ADR 0083. Clamped rather than scrolled: the full text is in the
     title attribute, and a token card is small. `white-space:
     pre-line` so the printed line breaks the server sends survive. */
  .token-text {
    display: -webkit-box;
    -webkit-line-clamp: 4;
    line-clamp: 4;
    -webkit-box-orient: vertical;
    overflow: hidden;
    padding: 0 4px 6px;
    font-size: 8px;
    line-height: 1.2;
    text-align: center;
    white-space: pre-line;
    color: #aab4cc;
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
  .badge.face-down {
    /* ADR 0069. Top-right like GOAD, and the two never co-occur: a
       face-down permanent has no text, so nothing can goad it. Cool
       slate rather than gold so it reads as "hidden state", not as a
       property of the card. */
    left: auto;
    right: 3px;
    max-width: calc(100% - 6px);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: #b9d8ff;
    background: rgba(12, 22, 44, 0.9);
    border-color: rgba(145, 195, 255, 0.55);
    font-size: 7px;
  }
  .badge.phased {
    /* #1199. Bottom-left, clear of FACE DOWN / GOAD / CMD at the top
       and of the P/T pip at the bottom right, because a phased-out
       permanent keeps every one of those and can wear them at once.
       The same cool slate as FACE DOWN: both say "state", not
       "property of the card". */
    top: auto;
    bottom: 3px;
    color: #b9d8ff;
    background: rgba(12, 22, 44, 0.9);
    border-color: rgba(145, 195, 255, 0.55);
    font-size: 7px;
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
  .badge.must-attack {
    /* #1571: a requirement the declaration still owes. Bottom edge, so
       it clears the CMD and GOAD badges on the top corners. */
    top: auto;
    bottom: 3px;
    left: 50%;
    right: auto;
    transform: translateX(-50%);
    white-space: nowrap;
    color: var(--danger);
    background: rgba(60, 0, 0, 0.85);
    border-color: rgba(255, 122, 122, 0.5);
  }
  .badge.no-untap {
    top: 24px;
    left: 50%;
    right: auto;
    transform: translateX(-50%);
    max-width: calc(100% - 6px);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: #b9d8ff;
    background: rgba(12, 35, 70, 0.9);
    border-color: rgba(145, 195, 255, 0.55);
    letter-spacing: 0.01em;
    font-size: 7px;
  }
  .badge.designation {
    /* ADR 0071 — top-centre, between the CMD pip (top-left) and the
       GOAD / face-down pip (top-right), so a levelled Class that is
       also somebody's commander reads cleanly. Cool blue rather than
       gold: like WON'T UNTAP, it is a state the card is IN, not a
       property printed on it. */
    top: 3px;
    left: 50%;
    right: auto;
    transform: translateX(-50%);
    max-width: calc(100% - 44px);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: #b9d8ff;
    background: rgba(12, 35, 70, 0.9);
    border-color: rgba(145, 195, 255, 0.55);
    font-size: 7px;
  }
  .badge.taken {
    /* ADR 0104. A player's NAME, so full width and truncating like the
       Curse badge — but one line higher, so a stolen Curse and a
       phased-out stolen permanent can wear both without overlap. The
       rose of "not yours" rather than gold's "property of the card". */
    top: auto;
    bottom: 17px;
    left: 3px;
    right: 3px;
    max-width: calc(100% - 6px);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    text-align: center;
    color: #ffc9d2;
    background: rgba(44, 12, 22, 0.9);
    border-color: rgba(255, 145, 170, 0.55);
    font-size: 7px;
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
  .badge.regen {
    /* Top-right, clear of the bottom-right damage / P-T stack: a
       shield is a fact about the NEXT destruction, not about the
       creature's current numbers. */
    top: 3px;
    left: auto;
    right: 3px;
    color: #9fe8a8;
    background: rgba(0, 48, 16, 0.9);
    border-color: rgba(120, 220, 140, 0.5);
    font-size: 9px;
    letter-spacing: 0.04em;
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
  /* ADR 0105 (#1789): the "ready" ring. Static on purpose — forty
     cards animating a box-shadow is the cost §5 rules out. Two parts:
     an OUTSET outline (outlines are not clipped by the card's
     overflow: hidden, and they follow its radius), and a soft glow
     drawn by a pseudo-element just inside the frame. Both scale with
     the card, so the ring reads at every card size and in a thumb.
     Declared before every other ring so a target prompt, a pick, a
     selection or a combat role wins where they meet; and the focus
     ring below replaces it while the card has focus rather than
     drawing over it. */
  .card.ready {
    outline: max(2px, calc(var(--card-w, 80px) * 0.022)) solid var(--ready);
    outline-offset: max(1px, calc(var(--card-w, 80px) * 0.015));
  }
  .card.ready::after {
    content: "";
    position: absolute;
    inset: 0;
    z-index: 3;
    border-radius: inherit;
    pointer-events: none;
    box-shadow: inset 0 0 calc(var(--card-w, 80px) * 0.14) var(--ready-glow);
  }
  /* ADR 0105 §2/§7 (#1789): the pips. They sit on the upper-left edge,
     below the top badge row (CMD) and the failed-art pip (22px): that
     is the part of a tile still visible where tiles overlap (the land
     strip's piles, the hand's top-55% peek), and the one edge with no
     always-on badge. The kind is carried by shape, and the colour is
     the one --ready. A pip scales with the card and never draws under
     16px (§7). It does not take pointer events yet: sub-PR 4 makes it
     the touch route into the popover. */
  .ready-pips {
    --pip: max(16px, calc(var(--card-w, 80px) * 0.17));
    position: absolute;
    top: max(38px, 30%);
    left: 3px;
    z-index: 4;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 3px;
    pointer-events: none;
  }
  .ready-pip {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 1px;
    box-sizing: border-box;
    min-width: var(--pip);
    height: var(--pip);
    padding: 0 calc(var(--pip) * 0.12);
    border-radius: 999px;
    background: var(--ready);
    color: var(--ready-ink);
    border: 1px solid rgba(0, 0, 0, 0.55);
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.55);
    line-height: 1;
  }
  .ready-pip svg {
    width: calc(var(--pip) * 0.68);
    height: calc(var(--pip) * 0.68);
    flex: 0 0 auto;
    fill: currentColor;
  }
  .pip-count {
    font-family: ui-monospace, Menlo, monospace;
    font-size: calc(var(--pip) * 0.62);
    font-weight: 900;
    font-variant-numeric: tabular-nums;
    padding-right: calc(var(--pip) * 0.08);
  }
  /* §7: at the small card size the count goes and the pip shows alone. */
  :global(:root[data-card-size="small"]) .pip-count {
    display: none;
  }
  .card.ready.targetable,
  .card.ready.picked,
  .card.ready.selected,
  .card.ready.attacking,
  .card.ready.blocking {
    outline: none;
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
  /* #1508: a mana source the auto-tapper would spend on the card being
     dragged out of the hand. A light dashed ring, deliberately quieter
     than a target or a selection — it is a preview, not a prompt. */
  .card.autotap-planned {
    outline: 2px dashed rgba(241, 211, 138, 0.85);
    outline-offset: 2px;
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
