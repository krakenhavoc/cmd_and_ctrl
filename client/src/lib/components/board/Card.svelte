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

  import type { CardView, GameView } from "../../protocol";
  import type { CantAttackChip } from "../../cantAttack";
  import { cardImageURL, tableImageSize } from "../../cardImage";
  import { cardArt } from "../../cardArt";
  import { showsCardBack } from "../../cardBack";
  import { hoveredCard } from "../../cardTypes";
  import { animateTap } from "../../animations";
  import { play } from "../../sounds";
  import { settings } from "../../settings";
  import { targeting, isLegalCardTarget, isPicked } from "../../targeting";
  import { autoTapHighlight } from "../../dragCast";
  import { noUntapAppliesToController } from "../../noUntap";
  import { deathMarkBadge } from "../../deathMarks";
  import { landTypeBadge } from "../../landTypes";
  import { damageBadge } from "../../damageBadge";
  import { KEYWORD_ICONS } from "../../keywordIcons";
  import { openCardMenu } from "../../contextMenu";
  import {
    acrossFor,
    castAnywayItem,
    manualLoyaltyRows,
    menuAbilityRows,
    menuManaRows,
    specialActionItems,
    type MenuAction,
  } from "../../contextMenu.logic";
  import {
    abilityPopover,
    closeAbilityPopover,
    openAbilityPopover,
    popoverDrawnOn,
    popoverSurface,
  } from "../../abilityPopover";
  import { manaRowNeedsPicker } from "../../manaSource";
  import { openManaSourcePicker } from "../../manaSourcePicker";
  import CounterPips from "./CounterPips.svelte";
  import ManaCost from "./ManaCost.svelte";
  import KeywordBadgeRow from "./KeywordBadgeRow.svelte";
  import ManaAbilityMenu from "./ManaAbilityMenu.svelte";
  import RoomDoorStrip from "./RoomDoorStrip.svelte";
  import { displayName } from "../../faces";
  import { ICONS } from "../../icons";
  import { ringBearerTitle } from "../../ringEmblem";
  import {
    DROP_PIP_LABEL,
    NO_LEGAL_ACTIONS,
    NO_PIPS,
    boltPipLabel,
    pipCount,
    readyCardLabel,
    readyPhrases,
    specialPipTitle,
    starPipLabel,
    type CombatPip,
    type LegalActions,
    type ReadyPips,
    type ReadyZone,
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
    // #1918: set with `ready` when every cast the card has is legal but
    // would do nothing right now (legalActions.ts idleReadyHint). The
    // ring is drawn muted and this, the server's sentence, is the
    // tooltip. Clicking still casts exactly as a full ring does.
    idleHint?: string;
    // ADR 0105 sub-PR 5: this card is what the selected creature may be
    // declared against: a planeswalker or battle it may attack, or an
    // attacker it may block. Set with `ready`. An attacker is always
    // wearing the red attacking ring, which `ready` yields to, so this
    // draws the ready ring OUTSIDE it and both read.
    combatTarget?: boolean;
    // ADR 0105 §2 (sub-PR 3): the pips on a card: a bolt for live
    // activated abilities (with a count from two up), a drop for a
    // mana ability worth marking (§4), and (sub-PR 4) a star for a
    // live special action. The caller reads them off the lookup
    // (legalActions.ts readyPips), as it does `ready`. The card only
    // draws them. NO_PIPS draws none.
    //
    // A pip is a button: tapping or clicking it, or Enter / Space on
    // it, opens the same popover a right-click opens (§7, owner
    // decision 6). It is the touch route into the popover.
    pips?: ReadyPips;
    // ADR 0105 §7 (sub-PR 6): the zone the card is drawn in, so a
    // ready card's accessible name can say WHY it is ready
    // ("castable", "playable land", "can attack") the way the ring and
    // pips show it (legalActions.ts readyPhrases). The phrases read
    // `legal`, so a caller that sets `ready` hands `legal` down too.
    readyZone?: ReadyZone;
    // ADR 0105 §7 (sub-PR 6): a combat candidate's pip, a sword (may
    // attack) or a shield (may block), beside its ready ring. It is
    // decorative: no popover sits behind a declaration, so it is not a
    // button, and the accessible name already carries "can attack" /
    // "can block". It also tells "may block" from "is blocking" by
    // shape where the cyan and pale-blue rings are hard to tell apart.
    combatPip?: CombatPip | null;
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
    // #1954 / #2209: draw an art tile — the art crop with a name strip
    // — instead of the full card. BattlefieldRow sets it from
    // `display.battlefieldArt` (on by default) and Hand, for the
    // viewer's own cards, from `display.handArt` (off by default); the
    // hover zoom, stack and every other Card leave it off. A face-down
    // card never becomes a tile (cardImage.ts tableImageSize).
    artOnly?: boolean;
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
    // ADR 0117 §3: the popover's Sandbox row, Tap or Untap, on every
    // permanent the viewer controls ("Tap (no mana)" on a mana source,
    // #1438). Set by BattlefieldRow on the viewer's own panel; undefined
    // hides the row (hand cards, opponents). Shown only when `viewerID`
    // controls the card, so an opponent's Aura drawn on the viewer's
    // creature never offers it.
    onRawTap?: () => void;
    // ADR 0118 §2: the popover's "Cast anyway (don't pay)" row, on a card
    // the viewer could cast (a hand card, a strip card, the command-zone
    // panel's commander) while strict payment is on. It opens the dock's
    // confirmation; the parent closes over the card and its zone. It
    // counts toward `hasMenu`, so a plain spell's right-click opens the
    // popover. It draws no pip and does not light the ready ring.
    // Undefined: no row.
    onCastAnyway?: () => void;
    // Why that row is greyed (timing.ts castAnywayBlocked), or "" when it
    // is live.
    castAnywayBlocked?: string;
    // ADR 0117 §3: sends a manual loyalty row's `activate_loyalty`, for
    // an uncatalogued planeswalker the viewer controls. Its rows need
    // `view` to judge the window. Set by BattlefieldRow on the viewer's
    // own panel.
    onMenuAction?: (action: MenuAction) => void;
    // The frame, for the manual loyalty rows' window and a
    // planeswalker's rows. BattlefieldRow passes it; elsewhere absent.
    view?: GameView | null;
    // ADR 0117 §1: the panel's click rule says a left-click on this card
    // does nothing right now, so it drops the `clickable` class: no
    // pointer cursor, no hover lift. It keeps role="button" and its tab
    // stop (a keyboard player still opens the popover from it, and the
    // e2e suite selects cards by role). A live targeting ring wins.
    inert?: boolean;
    // S21 sub-PR 2: same menu, CR 602 activated abilities. Set by
    // parents for battlefield permanents the viewer controls, and
    // since #660 by Hand.svelte for the viewer's own hand — a card in
    // hand offers the abilities that function THERE (cycling), which
    // ride `zone_abilities` rather than `activated_abilities`. One
    // callback for both: the index means the same thing on the wire.
    onActivateAbility?: (abilityIndex: number) => void;
    // ADR 0105 sub-PR 4 (#1789): sends a CR 116.2 special action
    // (foretell, suspend, plot, turn face up) chosen from the popover.
    // When set, the card's `special_actions` become rows in the
    // popover, built by the admin menu's own row builder so the payload
    // is the one that menu sends. Set by Hand and BattlefieldRow on the
    // viewer's own cards that offer one. Undefined: no rows, no star.
    onSpecialAction?: (action: MenuAction) => void;
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
    // ADR 0106 §2 (owner decision 3): whom this creature can't attack —
    // its owner, for Xantcha — drawn as a CAN'T ATTACK chip beside the
    // greyed target ring. Supplied by BattlefieldRow from the card
    // view's attack_target_restrictions; undefined for nearly every card.
    cantAttack?: CantAttackChip;
    // ADR 0114 owner decision 1: the controller's name when this
    // permanent is their Ring-bearer, for the marker's title ("Alice's
    // Ring-bearer"). Supplied by BattlefieldRow; the marker itself
    // follows `card.ring_bearer` and reads plain "Ring-bearer" without
    // it.
    ringBearerOf?: string;
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
    idleHint,
    combatTarget = false,
    pips = NO_PIPS,
    readyZone = "battlefield",
    combatPip = null,
    legal = NO_LEGAL_ACTIONS,
    legalGate = NO_LEGAL_ACTIONS,
    size = "small",
    artOnly = false,
    showManaCost = false,
    onActivateManaAbility,
    onRawTap,
    onCastAnyway,
    castAnywayBlocked = "",
    onMenuAction,
    view,
    inert = false,
    onActivateAbility,
    onSpecialAction,
    sorcerySpeedBlocked = "",
    payerLife,
    enchantedPlayer,
    takenFrom,
    cantAttack,
    ringBearerOf,
    priority = false,
    memberIDs,
    viewerID,
    onClick,
  }: Props = $props();

  // manaMenuOpen — whether this card's ManaAbilityMenu popover is
  // open. ADR 0117 §1: it lives in the abilityPopover store, keyed by
  // instance ID, because a LEFT-click opens it too and that click is
  // routed in PlayerPanel. A right-click or a pip writes the same store.
  // Dismissed on selection, Escape (handled inside the menu), or a
  // click on the card.
  //
  // ADR 0120 §3: with a seat's board expanded over the table, two Cards
  // carry this instance ID. The popover records which surface opened
  // it, and only the Card on that surface draws it.
  const surface = popoverSurface();
  const manaMenuOpen = $derived(popoverDrawnOn($abilityPopover, card.instance_id, surface));
  // #660: a card projects EITHER list, never both — the server
  // filters by the zone the card is in (CR 113.6) — so one menu reads
  // whichever is present and the indices stay the card's own.
  //
  // ADR 0106 §1 decision 6 (#1793): on a permanent the viewer does not
  // control, only its "Any player may activate this ability" rows
  // (CR 602.2); every other row is its controller's, and the server
  // would refuse it (contextMenu.logic.ts menuAbilityRows). A card with
  // no viewer in scope (no `viewerID` prop) lists every row as before.
  const menuAbilities = $derived(menuAbilityRows(card, viewerID));
  // Those rows are the viewer's to activate on another player's
  // permanent, and the popover greys any the exact digest leaves out.
  const across = $derived(acrossFor(card, viewerID));
  // #1228: and the same sentence for the CR 605 list. A permanent
  // publishes `mana_abilities`; a card in hand whose mana ability
  // functions there (a Spirit Guide) publishes `zone_mana_abilities`,
  // and the index means the same thing on the wire either way.
  // ADR 0117: the controller's alone (contextMenu.logic.ts menuManaRows).
  const menuManaAbilities = $derived(menuManaRows(card, viewerID));
  // ADR 0105 sub-PR 4: the special-action rows, from the same builder
  // the admin card menu uses, with the actor it would use: a face-down
  // permanent is turned up by its controller (CR 708.6), and a hand
  // card's controller is its owner. `legal` lights and lifts a ready
  // row; `legalGate` greys an available row the exact digest refuses.
  const specialRows = $derived(
    onSpecialAction
      ? specialActionItems(card, card.controller || card.owner, legal, legalGate)
      : [],
  );
  // ADR 0117 §3: an uncatalogued planeswalker's manual loyalty rows,
  // in the activated section. Only with the frame to judge them.
  const loyaltyRows = $derived(
    onMenuAction && onActivateAbility ? manualLoyaltyRows(card, view, viewerID ?? null) : [],
  );
  // ADR 0117 §3: the Sandbox row, on every permanent the viewer
  // controls. A Card with no viewer in scope trusts its parent.
  const sandbox = $derived(
    !!onRawTap && (viewerID == null || (card.controller || card.owner) === viewerID),
  );
  // ADR 0118 §2: the Sandbox section's Cast anyway row, one builder with
  // the admin menu's (contextMenu.logic.ts castAnywayItem).
  const castAnywayRow = $derived(onCastAnyway ? castAnywayItem(castAnywayBlocked) : undefined);
  // hasMenu: the popover has at least one row to show. Since ADR 0117
  // §3 that is every permanent the viewer controls (its Sandbox row),
  // so a right-click on a vanilla creature opens the popover with Tap.
  const hasMenu = $derived(
    (!!onActivateManaAbility && menuManaAbilities.length > 0) ||
      (!!onActivateAbility && menuAbilities.length > 0) ||
      specialRows.length > 0 ||
      loyaltyRows.length > 0 ||
      sandbox ||
      !!castAnywayRow,
  );
  // ADR 0105: a pip is drawn only where the popover it points at is
  // wired. A pip on a card whose abilities this viewer cannot open is
  // worse than none (ADR 0105, Context, fact 3).
  const boltPips = $derived(onActivateAbility ? pips.abilities : 0);
  const dropPip = $derived(!!onActivateManaAbility && pips.mana);
  const starPip = $derived(specialRows.length > 0 && pips.special > 0);
  const anyPip = $derived(hasMenu && (boltPips > 0 || dropPip || starPip));
  // ADR 0105 §7: the combat pip goes with the ring it explains.
  const swordPip = $derived(ready && combatPip === "attack");
  const shieldPip = $derived(ready && combatPip === "block");

  // ADR 0105 §7: a ready card's accessible name gains a phrase saying
  // what it is ready FOR. Built only while the ring is drawn.
  //
  // ADR 0114: a Ring-bearer says so in its name too. The marker's own
  // label sits inside the card's role, whose children assistive tech
  // does not read, so the name is the route that reaches it.
  const ringTitle = $derived(ringBearerTitle(ringBearerOf));
  const accessibleName = $derived.by(() => {
    if (showBack) return card.ring_bearer ? `face-down card, ${ringTitle}` : "face-down card";
    const name = card.ring_bearer ? `${displayName(card)}, ${ringTitle}` : displayName(card);
    if (!ready) return name;
    return readyCardLabel(name, true, readyPhrases(legal, card, readyZone, combatTarget));
  });

  // cardImageURL defaults to the card's ACTIVE face, so a modal DFC
  // played as its land half — or, later, a transformed permanent —
  // shows the side that is actually up without this component
  // knowing faces exist.
  const imgSize = $derived(tableImageSize(card, size, artOnly));
  const imgSrc = $derived(cardImageURL(card, imgSize));
  // #2209: the art tile. The crop has no frame, so the tile draws the
  // name itself in a strip along the top, and every top-anchored mark
  // (CMD, GOAD, counters, the designation, the pips) steps down below
  // it through --face-top. Everything else on the tile — P/T or
  // loyalty, counters, status marks, keyword chips, the ADR 0105 pips
  // — is the same markup a full card draws. A card with no art (a
  // token with no printing) keeps the name fallback, as before.
  const artTile = $derived(imgSize === "art_crop" && !!imgSrc && !showsCardBack(card, faceDown));

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
  // ADR 0108: "exiled if it dies this turn" / "can't be regenerated
  // this turn" — what killing this permanent does is already decided.
  const deathMark = $derived(deathMarkBadge(card));
  // ADR 0109 §1: a resolved effect changed this land's land types
  // ("Island until end of turn — Tidal Warrior").
  const landTypeMark = $derived(landTypeBadge(card));
  // #2257: the damage badge, and on an indestructible creature the
  // shield that says why the damage doesn't destroy it.
  const damage = $derived(damageBadge(card));
  const interactive = $derived(!!onClick && !phasedOut);
  // ADR 0117 §1: the pointer affordance follows what a click would do.
  const clickable = $derived(interactive && (!inert || targetable));

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
      closeAbilityPopover();
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
      if (manaMenuOpen) closeAbilityPopover();
      openCardMenu({ card, x: ev.clientX, y: ev.clientY });
      return;
    }
    if (!hasMenu) return;
    ev.preventDefault();
    ev.stopPropagation();
    if (manaMenuOpen) closeAbilityPopover();
    else openAbilityPopover(card.instance_id, surface);
  }

  // ADR 0105 §7 (owner decision 6): a pip is the touch route into the
  // popover. It opens what a right-click on this card opens: the
  // override menu under admin overrides, otherwise the ability popover.
  // The event stops here, so the card's own click (cast, tap, select)
  // and its Enter key never fire as well.
  //
  // It OPENS rather than toggles. Enter and Space are handled on
  // keydown, where preventDefault cancels the native button click in
  // every engine we target; if one ever let that click through as
  // well, a toggle would open and shut in one press, and an open
  // cannot. The popover closes as it always has: Escape, choosing a
  // row, or a click on the card.
  function openFromPip(ev: Event): void {
    ev.preventDefault();
    ev.stopPropagation();
    if (phasedOut) return;
    if ($settings.gameplay.adminOverrides) {
      const r = (ev.currentTarget as HTMLElement | null)?.getBoundingClientRect();
      if (manaMenuOpen) closeAbilityPopover();
      openCardMenu({ card, x: r?.right ?? 0, y: r?.top ?? 0 });
      return;
    }
    if (hasMenu) openAbilityPopover(card.instance_id, surface);
  }

  // ADR 0117 §4, "Right-click": the popover's mana row for an ability
  // with a colour choice opens the anchored picker on that one ability
  // (its colour buttons, or its stepper for two or more picking slots)
  // instead of activating with no colours and leaving the server to ask
  // once per slot. A permanent's own `mana_abilities` only: the picker
  // is a battlefield picker, and a hand card's mana ability (a Spirit
  // Guide) has no colour choice.
  function activateManaRow(index: number): void {
    const a = (card.mana_abilities ?? []).find((m) => m.index === index);
    if (a && manaRowNeedsPicker(a) && cardEl) {
      const r = cardEl.getBoundingClientRect();
      openManaSourcePicker({
        cardID: card.instance_id,
        anchor: { left: r.left, top: r.top, right: r.right, bottom: r.bottom },
        abilityIndex: index,
      });
      return;
    }
    onActivateManaAbility?.(index);
  }

  function handlePipKeydown(ev: KeyboardEvent): void {
    if (ev.key !== "Enter" && ev.key !== " ") return;
    openFromPip(ev);
  }

  function handleKeydown(ev: KeyboardEvent): void {
    if (ev.key !== "Enter" && ev.key !== " ") return;
    // ADR 0117: Enter on the card can open its popover, whose rows are
    // buttons inside this element. Their own Enter must reach them, not
    // be taken here as another click on the card.
    if (ev.target !== ev.currentTarget) return;
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
  class:art-tile={artTile}
  class:tapped={card.tapped}
  class:selected
  class:targetable
  class:picked
  class:autotap-planned={$autoTapHighlight.has(card.instance_id)}
  class:attacking
  class:blocking
  class:ready
  class:ready-idle={ready && !!idleHint}
  class:combat-target={combatTarget}
  class:clickable
  class:phased-out={phasedOut}
  class:ring-bearer={!!card.ring_bearer}
  class:menu-open={manaMenuOpen}
  data-instance-id={card.instance_id}
  data-tapped={card.tapped ? "true" : "false"}
  role={interactive ? "button" : "img"}
  tabindex={interactive ? 0 : undefined}
  aria-label={accessibleName}
  title={showBack
    ? ""
    : ready && idleHint
      ? `${displayName(card)} — ${idleHint}`
      : displayName(card)}
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
    {#if artTile}
      <!-- #2209: the art crop has no title bar. Hidden from assistive
           tech because the card's own accessible name already is it. -->
      <span class="art-name" aria-hidden="true">{displayName(card)}</span>
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
    {#if card.echo_due}
      <!-- ADR 0108 §5: its echo triggers at its controller's next upkeep
           (CR 702.30a). -->
      <span
        class="badge echo-due"
        title="echo due — at its controller's next upkeep, pay its echo cost or sacrifice it"
        aria-label="echo due">ECHO DUE</span
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
    {#if cantAttack}
      <!-- ADR 0106 §2: read off the server's restriction, never derived. -->
      <span
        class="badge cant-attack"
        title={cantAttack.title}
        aria-label={`can't attack ${cantAttack.label}`}
      >
        CAN'T ATTACK {cantAttack.label}
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
      <!-- #2231: the printed cost as pips; the badge keeps a spoken name. -->
      <span class="badge cost" title={`mana cost ${card.mana_cost}`}>
        <ManaCost cost={card.mana_cost} size={15} />
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
      riotHaste={card.riot_haste}
    />
    {#if damage}
      <span
        class="badge damage"
        class:survives={damage.survives}
        title={damage.title}
        aria-label={damage.survives ? "damage, indestructible" : "damage"}
      >
        {damage.text}
        {#if damage.survives}
          <span class="survives-icon" aria-hidden="true">
            <!-- eslint-disable-next-line svelte/no-at-html-tags -->
            {@html KEYWORD_ICONS.indestructible}
          </span>
        {/if}
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
    {#if deathMark}
      <span class="badge death-mark" title={deathMark.title} aria-label="death mark"
        >{deathMark.text}</span
      >
    {/if}
    {#if landTypeMark}
      <span class="badge land-type" title={landTypeMark.title} aria-label="land type"
        >{landTypeMark.text}</span
      >
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
    {#if card.echo_due}
      <!-- ADR 0108 §5: its echo triggers at its controller's next upkeep
           (CR 702.30a). -->
      <span
        class="badge echo-due"
        title="echo due — at its controller's next upkeep, pay its echo cost or sacrifice it"
        aria-label="echo due">ECHO DUE</span
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
    {#if cantAttack}
      <!-- ADR 0106 §2: read off the server's restriction, never derived. -->
      <span
        class="badge cant-attack"
        title={cantAttack.title}
        aria-label={`can't attack ${cantAttack.label}`}
      >
        CAN'T ATTACK {cantAttack.label}
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
      riotHaste={card.riot_haste}
    />
    {#if damage}
      <span
        class="badge damage"
        class:survives={damage.survives}
        title={damage.title}
        aria-label={damage.survives ? "damage, indestructible" : "damage"}
      >
        {damage.text}
        {#if damage.survives}
          <span class="survives-icon" aria-hidden="true">
            <!-- eslint-disable-next-line svelte/no-at-html-tags -->
            {@html KEYWORD_ICONS.indestructible}
          </span>
        {/if}
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
    {#if deathMark}
      <span class="badge death-mark" title={deathMark.title} aria-label="death mark"
        >{deathMark.text}</span
      >
    {/if}
    {#if landTypeMark}
      <span class="badge land-type" title={landTypeMark.title} aria-label="land type"
        >{landTypeMark.text}</span
      >
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
  {#if card.ring_bearer}
    <!-- ADR 0114 owner decision 1: the Ring-bearer's own marker, apart
         from the designation slot above, because a Ring-bearer can also
         be monstrous or harnessed. Outside the art / back branches on
         purpose: the designation was chosen in public, so a face-down
         Ring-bearer shows it too (ADR 0114 §9). -->
    <span class="ring-marker" role="img" aria-label="Ring-bearer" title={ringTitle}>
      <svg viewBox="0 0 24 24" focusable="false" aria-hidden="true">
        {#each ICONS.ring as prim, i (i)}
          {#if prim.t === "path"}<path d={prim.d} />{/if}
        {/each}
      </svg>
    </span>
  {/if}
  {#if anyPip || swordPip || shieldPip}
    <!-- ADR 0105 §2/§7 (#1789): what this card can do right now, by
         shape: a star for a special action (any kind, including one
         this client has no name for), a bolt for an activated ability,
         a drop for a mana ability worth marking (§4), and a sword or
         shield for a combat candidate. The first three are buttons
         that open the popover a right-click opens (§7, sub-PR 4), each
         named for what it opens onto. The combat pips are drawing
         only: a declaration has no popover behind it, and the card's
         own name already says "can attack" / "can block". -->
    <span class="ready-pips">
      {#if starPip}
        {@const kinds = legal.readySpecialActions(card.instance_id)}
        <button
          type="button"
          class="ready-pip star"
          data-pip="star"
          aria-label={starPipLabel(kinds)}
          aria-haspopup="menu"
          aria-expanded={manaMenuOpen}
          title={specialPipTitle(kinds)}
          onclick={openFromPip}
          onkeydown={handlePipKeydown}
        >
          <svg viewBox="0 0 24 24" focusable="false" aria-hidden="true"
            ><path
              d="M12 2.5l2.9 6.2 6.8.8-5 4.7 1.3 6.8L12 17.6 6 21l1.3-6.8-5-4.7 6.8-.8z"
            /></svg
          >
        </button>
      {/if}
      {#if boltPips > 0}
        <button
          type="button"
          class="ready-pip bolt"
          data-pip="bolt"
          aria-label={boltPipLabel(boltPips)}
          aria-haspopup="menu"
          aria-expanded={manaMenuOpen}
          onclick={openFromPip}
          onkeydown={handlePipKeydown}
        >
          <svg viewBox="0 0 24 24" focusable="false" aria-hidden="true"
            ><path d="M13.5 2 4 13.5h6.5L9.5 22 20 9.5h-6.5z" /></svg
          >
          {#if pipCount(boltPips)}<span class="pip-count" aria-hidden="true"
              >{pipCount(boltPips)}</span
            >{/if}
        </button>
      {/if}
      {#if dropPip}
        <button
          type="button"
          class="ready-pip drop"
          data-pip="drop"
          aria-label={DROP_PIP_LABEL}
          aria-haspopup="menu"
          aria-expanded={manaMenuOpen}
          onclick={openFromPip}
          onkeydown={handlePipKeydown}
        >
          <svg viewBox="0 0 24 24" focusable="false" aria-hidden="true"
            ><path d="M12 2.5S5 10.4 5 15.2a7 7 0 0 0 14 0C19 10.4 12 2.5 12 2.5z" /></svg
          >
        </button>
      {/if}
      {#if swordPip}
        <span class="ready-pip combat" data-pip="sword" aria-hidden="true" title="can attack">
          <svg viewBox="0 0 24 24" focusable="false"
            ><path
              d="M19 3h2v2l-9.5 9.5-2-2zM6.5 12.5 8 11l5 5-1.5 1.5zM8.5 15.5l1 1-4 4-1-1zM2.3 20.2a1.6 1.6 0 1 0 3.2 0 1.6 1.6 0 1 0-3.2 0z"
            /></svg
          >
        </span>
      {/if}
      {#if shieldPip}
        <span class="ready-pip combat" data-pip="shield" aria-hidden="true" title="can block">
          <svg viewBox="0 0 24 24" focusable="false"
            ><path d="M12 2l8 3v6c0 5-3.5 9.3-8 11-4.5-1.7-8-6-8-11V5z" /></svg
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
  {#if manaMenuOpen && hasMenu}
    <div class="mana-menu-anchor">
      <ManaAbilityMenu
        special={specialRows}
        onSpecialAction={(action) => onSpecialAction?.(action)}
        abilities={onActivateManaAbility ? menuManaAbilities : []}
        tapped={!!card.tapped}
        onActivate={activateManaRow}
        activated={onActivateAbility ? menuAbilities : []}
        onActivateAbility={(idx) => onActivateAbility?.(idx)}
        summoningSick={!!card.summoning_sick}
        timingReason={sorcerySpeedBlocked}
        cardID={card.instance_id}
        {legal}
        {legalGate}
        {payerLife}
        {across}
        {card}
        {view}
        {viewerID}
        manualLoyalty={loyaltyRows}
        onMenuAction={(action) => onMenuAction?.(action)}
        onRawTap={sandbox ? onRawTap : undefined}
        castAnyway={castAnywayRow}
        {onCastAnyway}
        onClose={closeAbilityPopover}
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
    /* #2209: how far the top-anchored marks sit below the tile's top
       edge. 0 on a full card, the name strip's height on an art tile.
       CounterPips reads it too, through inheritance. */
    --face-top: 0px;
  }
  /* #2209: the art tile. It keeps the full card's slot — the same
     --card-w × --card-h, 5:7 portrait — so no row, pile, fan or
     attachment offset reflows when the setting changes; the landscape
     crop is centred and cropped at the sides (object-fit: cover). The
     strip's type scales with the card and is clamped so it stays
     legible on an opponent's smallest compact row and does not shout
     on a large one. */
  .card.art-tile {
    --art-name-size: clamp(9px, calc(var(--card-w, 80px) * 0.085), 13px);
    --art-strip-h: calc(var(--art-name-size) + 7px);
    --face-top: var(--art-strip-h);
  }
  /* The failed-art pip steps under the strip with everything else. A
     tapped tile keeps its own 58% (below). */
  .card.art-tile:not(.tapped) {
    --art-error-top: calc(22px + var(--art-strip-h));
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
  /* #2209: the name strip. Drawn over art, never over the page, so its
     colours are fixed rather than themed: near-white type on a
     near-opaque dark band reads over any art in every theme. Ellipsis,
     never a wrap; the full name is the card's title and the hover zoom. */
  .art-name {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    z-index: 2;
    box-sizing: border-box;
    height: var(--art-strip-h);
    /* BattlefieldRow sets the inset on a land pile, whose count badge
       overhangs this corner. */
    padding: 0 5px 0 var(--art-name-inset, 5px);
    font-size: var(--art-name-size);
    font-weight: 700;
    line-height: var(--art-strip-h);
    letter-spacing: 0.01em;
    color: #f6f1e4;
    background: linear-gradient(rgba(6, 9, 18, 0.92), rgba(6, 9, 18, 0.8));
    border-bottom: 1px solid rgba(255, 255, 255, 0.14);
    text-shadow: 0 1px 1px rgba(0, 0, 0, 0.9);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    pointer-events: none;
  }
  :global(:root[data-theme="high-contrast"]) .art-name {
    color: #fff;
    background: #000;
    border-bottom-color: #fff;
    text-shadow: none;
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
    top: calc(3px + var(--face-top, 0px));
    left: 3px;
    background: rgba(10, 14, 26, 0.88);
    color: var(--accent);
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
  .badge.echo-due {
    /* ADR 0108 §5: above the bottom edge, so a creature that must
       attack AND owes its echo (Tectonic Fiend) shows both. */
    top: auto;
    bottom: 22px;
    left: 50%;
    right: auto;
    transform: translateX(-50%);
    white-space: nowrap;
    color: #ffd27a;
    background: rgba(50, 35, 0, 0.85);
    border-color: rgba(255, 210, 122, 0.5);
  }
  .badge.no-untap {
    top: calc(24px + var(--face-top, 0px));
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
    top: calc(3px + var(--face-top, 0px));
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
  .badge.cant-attack {
    /* ADR 0106 §2 (#1794): a player's NAME, so full width and
       truncating like TAKEN FROM, one row above it — Xantcha wears
       both at once (its owner is the one it was taken from and the one
       it can't attack). The red of a combat restriction, like MUST
       ATTACK. */
    top: auto;
    bottom: 31px;
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
    color: var(--accent);
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
  .badge.damage.survives {
    /* #2257: damage an indestructible creature shrugs off. The shield
       is the keyword badge's own icon, so the two read as one fact. */
    display: inline-flex;
    align-items: center;
    gap: 2px;
  }
  .badge.damage .survives-icon {
    display: inline-flex;
    width: 11px;
    height: 11px;
    color: #ffd07a;
  }
  .badge.damage .survives-icon :global(svg) {
    width: 100%;
    height: 100%;
  }
  .badge.death-mark {
    /* ADR 0108. Top-centre, one row under WON'T UNTAP: a fact about
       what happens when the creature dies, like a regeneration shield,
       in the ash-grey of exile rather than the shield's green. */
    top: calc(38px + var(--face-top, 0px));
    left: 50%;
    right: auto;
    transform: translateX(-50%);
    max-width: calc(100% - 6px);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: #e6dfd3;
    background: rgba(40, 34, 30, 0.9);
    border-color: rgba(210, 196, 176, 0.55);
    letter-spacing: 0.01em;
    font-size: 7px;
  }
  .badge.land-type {
    /* ADR 0109 §1. Top-centre, a row under the death mark: what a
       resolved effect has made this land for now, in a sea-blue that
       reads as a temporary overlay rather than a warning. */
    top: calc(52px + var(--face-top, 0px));
    left: 50%;
    right: auto;
    transform: translateX(-50%);
    max-width: calc(100% - 6px);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: #d6ecff;
    background: rgba(12, 38, 64, 0.9);
    border-color: rgba(120, 180, 230, 0.6);
    letter-spacing: 0.01em;
    font-size: 7px;
  }
  .badge.regen {
    /* Top-right, clear of the bottom-right damage / P-T stack: a
       shield is a fact about the NEXT destruction, not about the
       creature's current numbers. */
    top: calc(3px + var(--face-top, 0px));
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
       with the goad / damage battlefield badges. Drawn as pips (#2231). */
    left: auto;
    right: 3px;
    top: calc(3px + var(--face-top, 0px));
    display: inline-flex;
    align-items: center;
    padding: 2px 3px;
    border-radius: 999px;
    background: rgba(10, 14, 26, 0.92);
    border: 1px solid rgba(200, 168, 106, 0.5);
    max-width: 90%;
    overflow: hidden;
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
  /* #1918: castable, but every cast would do nothing right now (an
     overloaded Counterflux with no spell to counter). Still ringed,
     since the cast is legal, but dashed, faint and without the glow,
     so it does not read as a play worth making. */
  .card.ready.ready-idle {
    outline-style: dashed;
    outline-color: color-mix(in srgb, var(--ready) 45%, transparent);
  }
  .card.ready.ready-idle::after {
    box-shadow: none;
  }
  /* ADR 0114 owner decision 1: the Ring-bearer's marker. A gold disc
     with the ring glyph on the left edge, one badge row down — under
     CMD, which a commander Ring-bearer also wears, and clear of the
     designation slot at top centre, so MONSTROUS or HARNESSED shows
     beside it. The left edge of an untapped tile is the part that
     stays visible where tiles overlap (see the failed-art pip above),
     and the ready pips start below it. A tapped tile turns, so the
     marker follows the failed-art pip to 58% to stay out from under a
     tapped neighbour. Gold like the Ring chip on the player panel; the
     same dark ring around it on every theme, because it sits on card
     art, not on the page. */
  .card.ring-bearer {
    --ring-size: max(15px, calc(var(--card-w, 80px) * 0.16));
    /* The failed-art pip shares the marker's row; it steps right of it. */
    --art-error-left: calc(var(--ring-size) + 7px);
  }
  .ring-marker {
    position: absolute;
    top: calc(21px + var(--face-top, 0px));
    left: 3px;
    z-index: 4;
    box-sizing: border-box;
    width: var(--ring-size, 15px);
    height: var(--ring-size, 15px);
    display: grid;
    place-items: center;
    border-radius: 50%;
    background: radial-gradient(circle at 35% 30%, #ffe9a8, #d9b45c 60%, #8a6a1e);
    color: #2a1d00;
    border: 1px solid rgba(20, 14, 0, 0.8);
    box-shadow:
      0 0 0 1px rgba(255, 220, 140, 0.45),
      0 1px 4px rgba(0, 0, 0, 0.6);
    pointer-events: auto;
  }
  .card.tapped .ring-marker {
    top: 58%;
  }
  .ring-marker svg {
    width: 78%;
    height: 78%;
    fill: none;
    stroke: currentColor;
    stroke-width: 2.4;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
  :global(:root[data-theme="high-contrast"]) .ring-marker {
    background: #ffd400;
    color: #000;
    border-color: #000;
    box-shadow: 0 0 0 1px #fff;
  }
  /* ADR 0105 §2/§7 (#1789): the pips. They sit on the upper-left edge,
     below the top badge row (CMD) and the failed-art pip (22px): that
     is the part of a tile still visible where tiles overlap (the land
     strip's piles, the hand's top-55% peek), and the one edge with no
     always-on badge. The kind is carried by shape, and the colour is
     the one --ready. A pip scales with the card and never draws under
     16px (§7). Each pip is a button and the touch route into the
     popover (sub-PR 4), so the pip itself takes pointer events and
     the column between pips does not: a press in the gap still lands
     on the card. */
  .ready-pips {
    --pip: max(16px, calc(var(--card-w, 80px) * 0.17));
    position: absolute;
    top: calc(max(38px, 30%) + var(--face-top, 0px));
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
    margin: 0;
    padding: 0 calc(var(--pip) * 0.12);
    border-radius: 999px;
    background: var(--ready);
    color: var(--ready-ink);
    border: 1px solid rgba(0, 0, 0, 0.55);
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.55);
    font: inherit;
    line-height: 1;
    cursor: pointer;
    pointer-events: auto;
    /* A tap on a pip is a tap, never the start of a pan or a hand drag. */
    touch-action: manipulation;
    -webkit-tap-highlight-color: transparent;
  }
  .ready-pip:hover {
    filter: brightness(1.12);
  }
  /* ADR 0105 §7 (sub-PR 6): the sword and shield are drawing only. A
     press on one is a press on the card, which is the declaration. */
  .ready-pip.combat {
    cursor: inherit;
    pointer-events: none;
  }
  .ready-pip.combat:hover {
    filter: none;
  }
  .ready-pip:focus-visible {
    outline: 2px solid var(--ready-ink);
    outline-offset: 1px;
    box-shadow:
      0 0 0 3px var(--ready),
      0 1px 4px rgba(0, 0, 0, 0.55);
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
  /* ADR 0105 sub-PR 5: an attacker the selected blocker may block. It
     is attacking by definition, so the rule above would hide the one
     ring that says "you can block this". The red ring is a box-shadow
     hugging the frame and the ready ring an outline, so the outline
     steps out past the red one and both read. A target prompt, a pick
     or a selection still wins. */
  .card.ready.combat-target:is(.attacking, .blocking):not(.targetable, .picked, .selected) {
    outline: max(2px, calc(var(--card-w, 80px) * 0.022)) solid var(--ready);
    outline-offset: max(5px, calc(var(--card-w, 80px) * 0.05));
  }
  /* ADR 0105 §7 (sub-PR 6): the high-contrast theme draws the ring as
     a solid 3px outline and no glow. The theme's tokens already make
     --ready-glow transparent; dropping the pseudo-element says so
     outright. Dormant until App.svelte applies data-theme, which it
     does not yet (the theme select is disabled; see App.svelte). The
     second selector out-specifies the combat-target rule above. */
  :global(:root[data-theme="high-contrast"]) .card.ready,
  :global(:root[data-theme="high-contrast"]) .card.ready.combat-target:is(.attacking, .blocking) {
    outline-width: 3px;
  }
  :global(:root[data-theme="high-contrast"]) .card.ready::after {
    display: none;
  }
  .card.selected {
    box-shadow:
      0 0 0 2px var(--accent),
      0 0 22px color-mix(in srgb, var(--pick) 60%, transparent),
      0 10px 22px rgba(0, 0, 0, 0.5),
      inset 0 0 0 1px rgba(255, 255, 255, 0.08);
  }
  .card.targetable {
    box-shadow:
      0 0 0 2px var(--target),
      0 0 18px color-mix(in srgb, var(--target) 60%, transparent),
      0 6px 16px rgba(0, 0, 0, 0.5);
    cursor: crosshair;
  }
  /* #1508: a mana source the auto-tapper would spend on the card being
     dragged out of the hand. A light dashed ring, deliberately quieter
     than a target or a selection — it is a preview, not a prompt. */
  .card.autotap-planned {
    outline: 2px dashed color-mix(in srgb, var(--accent-strong) 85%, transparent);
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
      0 0 18px color-mix(in srgb, var(--attack) 55%, transparent),
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
