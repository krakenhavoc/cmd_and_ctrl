<script lang="ts">
  // Board is the table-level view. Lays out player panels in
  // quadrants — bottom row upright (self + the player next to you),
  // top row rotated 180° (the player across from you, and optionally
  // the one next to them in a 4-player game). Mounts the floating
  // hover-zoom and stack overlays. Replaces the entire PixiJS table
  // renderer (client/src/lib/table.ts).
  //
  // Quadrant assignment lives in cardTypes.ts:seatPlacements. The
  // grid layout here switches template by opponent count so 2- and
  // 3-player games get full-width panels in the unused rows rather
  // than empty quadrants.
  //
  // Battlefield cards arrive as one unified ZoneView from the server.
  // Board groups them by controller before handing each PlayerPanel
  // its own slice; the panel then sub-buckets by card type for the
  // creatures / lands / right-column rows. Likewise, the shared exile
  // zone is filtered by owner so each panel's EXILE pile shows only
  // the cards that player owns.

  import type { Snippet } from "svelte";
  import type {
    ActionPayload,
    ActionType,
    ActivatedAbilityView,
    AdditionalCostView,
    CardView,
    GameView,
    LegalTargetsView,
    ManaAbilityView,
    PlayerView,
    StackItemView,
    ZoneView,
  } from "../../protocol";
  import { answeredOnBoard, listFallback } from "../../boardAnsweredChoice";
  import { seatPlacements, type SeatPosition } from "../../cardTypes";
  import { consideringDelayMs, isResponseWindowFor, responseWindowKey } from "../../considering";
  import PlayerPanel from "./PlayerPanel.svelte";
  import SeatSummary from "./SeatSummary.svelte";
  import {
    decideSeatRendering,
    expansionSettingsFor,
    legalDefenderIDs,
    seatControlsLegalTarget,
    seatHasAttackersOn,
    type SeatDecision,
  } from "../../expansion";
  import { settings } from "../../settings";
  import { opponentsInARow } from "../../tableLayout";
  import HoverZoomOverlay from "./HoverZoomOverlay.svelte";
  import Icon from "../Icon.svelte";
  // CommanderDamageTooltip was folded into HoverZoomOverlay — the
  // damage readout now lives inside the card preview panel instead
  // of the lower-right corner.
  import StackOverlay from "./StackOverlay.svelte";
  import StackLaneHost from "./StackLaneHost.svelte";
  import StackLinger from "./StackLinger.svelte";
  import {
    DEFAULT_STACK_STYLE,
    isStackStyle,
    stackLaneLive,
    type StackLaneStyle,
  } from "../../stackLane";
  import { attentionStrip, pileFallsBack, stripContentBottom } from "../../stackPile";
  import CombatArrows from "./CombatArrows.svelte";
  import CombatStrikes from "./CombatStrikes.svelte";
  import DiceLayer from "./DiceLayer.svelte";
  import StackTargetRings from "./StackTargetRings.svelte";
  import TargetingArrows from "./TargetingArrows.svelte";
  import VotingPanel from "./VotingPanel.svelte";
  import ZoneBrowserModal from "./ZoneBrowserModal.svelte";
  import { zoneBrowser, closeZoneBrowser } from "../../zoneBrowser";
  import { freeCastRequest, freeCastTarget } from "../../freeCastRequest";
  import { phasedOutCards } from "../../phasedOut";
  import { canActivateSorcerySpeedAbility } from "../../timing";
  import CardContextMenu from "./CardContextMenu.svelte";
  import { cardMenu, closeCardMenu } from "../../contextMenu";
  import ManaSourcePicker from "./ManaSourcePicker.svelte";
  import { manaSourcePicker, closeManaSourcePicker } from "../../manaSourcePicker";
  import { abilityPopover, closeAbilityPopover } from "../../abilityPopover";
  import {
    bumpBoardExpandLayout,
    initialExpandState,
    liveExpanded,
    parsePx,
    peekOpenDelay,
    placeOverlay,
    reduceExpand,
    setAvatarExpand,
    type ExpandEvent,
    type Span,
  } from "../../boardExpand";
  import { seatSelector } from "../../boardAnchor";
  import { currentDockRequest, dockKeyFor } from "../../dock";
  import { foreignModalOpen } from "../../modalLayers";
  import { onDestroy, tick, untrack } from "svelte";
  import { get } from "svelte/store";
  import { provideCombatCues } from "../../combatCues.svelte";
  import { manaColorParams } from "../../manaSource";
  import { activatedAbilityRef, manaAbilityRef } from "../../abilityRef";
  import { findCard, locateCard, type MenuActivate } from "../../contextMenu.logic";
  import DockRequest from "./DockRequest.svelte";
  import { castAnywayConfirmRequest } from "../../targetingDock";
  import { castAnywayPending, clearCastAnyway } from "../../castAnyway";
  import { payLifeRequest } from "../../payLifeForMana";
  import { fetchAutoTapPreview } from "../../api";
  import {
    targeting,
    begin as beginTargeting,
    cancel as cancelTargeting,
    beginChoice as beginTargetingChoice,
    beginForAbility as beginTargetingForAbility,
    beginForModes as beginTargetingForModes,
    advance,
    allPicks,
    choiceAnswer,
    hasXCost,
    castLocksXAtZero,
    isModal,
    discardCostOf,
    castAdditionalCost,
    costBranchesOf,
    castSacrificeClause,
    castSacrificeLabel,
    optionalCostsOf,
    castTeamworkOffer,
    castBlightOffer,
    castRevealCost,
    tapCostOf,
    tapCostLimit,
    alternativeCostsOf,
    alternativeCostByKey,
    castTargetOverride,
    modesUnderChoices,
    altCostPayOptions,
    altCostSacrificeClause,
    altCostTapClause,
    applyCastChoices,
    castChoicesBase,
    isLegalCardTarget,
    isLegalPlayerTarget,
    isMultiPick,
    opensTargetPicker,
    togglePick,
    canConfirm,
    setConfirmHandler,
    needsDivision,
    withDivision,
    distributionOf,
    type CastChoices,
    type CastSourceZone,
    type TargetingState,
    type TargetRef,
  } from "../../targeting";
  import { abilityEnergyMaxX, suggestedAbilityX as suggestedAbilityXFor } from "../../abilityX";
  import { castPreviewParams } from "../../castPreview";
  import { castSacrificeRange, orderSacrificeOptions, sacrificeRange } from "../../sacrificeCost";
  import XCostModal from "./XCostModal.svelte";
  import DivideDamageModal from "./DivideDamageModal.svelte";
  import SacrificeCostModal from "./SacrificeCostModal.svelte";
  import CrewCostModal from "./CrewCostModal.svelte";
  import CounterCostModal from "./CounterCostModal.svelte";
  import {
    autoCounterChoice,
    counterCostNeedsPrompt,
    counterPaymentParams,
    hasCounterCost,
    type CounterChoice,
    type CounterPayment,
  } from "../../counterCost";
  import AltCostPaymentModal from "./AltCostPaymentModal.svelte";
  import ModePickerModal from "./ModePickerModal.svelte";
  import DiscardCostModal from "./DiscardCostModal.svelte";
  import CostConfirmModal from "./CostConfirmModal.svelte";
  import {
    manaAbilityNeedsPrompt,
    manaExilePermanentPayment,
    manaTapPayment,
  } from "../../manaAbilityCost";
  import { exileCostNote, exileCostOptionCards, exileCostWhere } from "../../exileCost";
  import { discardedManaValue } from "../../discardCostX";
  import {
    costConfirmLines,
    costConfirmNote,
    needsCostConfirm,
    randomDiscardPool,
    topCostNote,
  } from "../../libraryCost";
  import AlternativeCostModal from "./AlternativeCostModal.svelte";
  import FacePickerModal from "./FacePickerModal.svelte";
  import { cardAsFace, cardAsFused, faceOptions, needsFacePicker } from "../../faces";
  import { unlockParams, unlockRequest } from "../../roomDoors";
  import TapCostModal from "./TapCostModal.svelte";
  import DelveCostModal from "./DelveCostModal.svelte";
  import { delveOptionIDs, hasDelveChoice } from "../../delve";
  import { shouldAskAbilityWaterbend, waterbendLimit } from "../../waterbend";
  import PhyrexianCostModal from "./PhyrexianCostModal.svelte";
  import { NO_LEGAL_ACTIONS, type LegalActions } from "../../legalActions";
  import { L } from "../../labels";
  import {
    onlyGrantedSymbols,
    phyrexianGrantedForAbility,
    phyrexianGrantedForCast,
    phyrexianSymbolsForAbility,
    phyrexianSymbolsForCast,
    shouldAskPhyrexianLife,
  } from "../../phyrexianLife";

  type ActionSender = (type: ActionType, params?: ActionPayload["params"], player?: string) => void;

  interface Props {
    view: GameView;
    viewerID: string | null;
    isAdmin: boolean;
    sendAction: ActionSender;
    combatMode: "idle" | "attack" | "block";
    selectedCombatCardID: string | null;
    onSelectCombatCard: (cardID: string) => void;
    onDeclareAttack: (targetPlayerID: string) => void;
    onDeclareBlock: (attackerCardID: string) => void;
    // #1724: a token group's "Attack <seat> with N" — forwarded to the
    // viewer's own PlayerPanel. See PlayerPanel's prop of the same name.
    onDeclareAttackers?: (attackerIDs: string[], defenderSeatID: string) => void;
    // #519: connectionBanner.ts's actionsDisabled(status), computed by
    // Game.svelte and handed down rather than recomputed here — Board
    // has no socket of its own to ask. Dims the table and turns every
    // card affordance into a no-op instead of a click that silently
    // goes nowhere while the connection is down.
    disabled?: boolean;
    // ADR 0111 §4: Game.svelte has mounted the action dock over the
    // board's bottom-right corner, so the viewer's own panel keeps that
    // corner clear. See PlayerPanel's prop of the same name.
    docked?: boolean;
    // ADR 0076 §2.3 (amended 2026-10-02): the tutorial's coach card sits
    // over the board's bottom-left corner, so the viewer's own panel keeps
    // that corner clear too. See PlayerPanel's prop of the same name.
    coached?: boolean;
    // Game.svelte's live prompts (targeting, combat hint, mulligan
    // roll-call, toasts, game end) render inside the attention strip
    // under the stack card so every "look here" surface shares one
    // anchor over the table.
    attention?: Snippet;
    // #187 / ADR 0053: changes whenever the combat damage beats should
    // prime again instead of cueing what they missed (reconnect, replay
    // toggle). Primes the board's combat damage clock (ADR 0134 §1), and
    // is passed straight to the dice layer and the linger.
    beatsPrimeKey?: string;
    // ADR 0105 (#1789): the frame's legal-action lookup, built once in
    // Game.svelte and already "nothing" while highlights are off or
    // smart autopass is about to pass this frame. Handed to the
    // viewer's own panel and the zone browser for their ready rings,
    // and to the card menu for its ready rows.
    legal?: LegalActions;
    // ADR 0105 sub-PR 3: the frame's FULL lookup, which neither the
    // highlight setting nor autopass touches. Only the ability
    // popover's sorcery-speed gate reads it.
    legalGate?: LegalActions;
  }

  const {
    view,
    viewerID,
    isAdmin,
    sendAction,
    combatMode,
    selectedCombatCardID,
    onSelectCombatCard,
    onDeclareAttack,
    onDeclareBlock,
    onDeclareAttackers,
    disabled = false,
    docked = false,
    coached = false,
    attention,
    beatsPrimeKey,
    legal = NO_LEGAL_ACTIONS,
    legalGate = NO_LEGAL_ACTIONS,
  }: Props = $props();

  // #519: every action this component initiates funnels through here
  // rather than through `sendAction` directly, so a card, an ability
  // row or a context-menu entry stops being clickable the instant the
  // socket goes down — instead of dispatching into a queue nothing is
  // reading. `sendAction` itself still refuses offline sends on its
  // own (see connectionBanner.ts's offlineSendMessage), so a call that
  // slips past this guard is surfaced, not swallowed.
  const guardedSendAction: ActionSender = (type, params, player) => {
    if (disabled) return;
    sendAction(type, params, player);
  };

  // ADR 0134 §1: one combat damage clock for the whole board. The arrows'
  // beat cues and the strike layer's lunges both listen to it, and every
  // Card reads its striking set (through the context) to hide while its
  // copy flies. A prime-key change (reconnect, replay toggle) primes the
  // next frame; declared before the frame effect so a key and a view
  // that change together prime that same frame.
  const combatCues = provideCombatCues();
  onDestroy(() => combatCues.dispose());
  $effect(() => {
    void beatsPrimeKey;
    untrack(() => combatCues.requestReprime());
  });
  $effect(() => {
    const log = view.log;
    untrack(() => combatCues.frame(log, get(settings)));
  });

  // Spectators have no perspective — there's no "self" seat to anchor
  // the around-the-table rotation. Use a uniform grid for them with
  // every panel upright and equally sized; quadrant placements only
  // apply when there's a viewer at the table.
  const isSpectator = $derived(viewerID === null);

  const placements = $derived(isSpectator ? null : seatPlacements(view.seats, viewerID));
  const opponentCount = $derived(
    placements
      ? (["next", "across", "across_next"] as SeatPosition[]).filter((p) => placements[p]).length
      : 0,
  );

  // Group battlefield cards by controller so each panel only sees
  // its own slice. Done once per snapshot.
  const cardsByController = $derived.by(() => {
    const out = new Map<string, CardView[]>();
    for (const c of view.battlefield.cards) {
      const list = out.get(c.controller);
      if (list) list.push(c);
      else out.set(c.controller, [c]);
    }
    // #1199 / CR 702.26, ADR 0084. A phased-out permanent arrives in
    // its own zone and is absent from `battlefield`, which is what
    // makes every rules question the client asks the board come out
    // right. This is the ONE place that puts it back, at the END of
    // its controller's row: the player has to be able to see that
    // their four permanents are coming back rather than gone, and no
    // other surface says so. Card.svelte dims it, badges it PHASED
    // and withholds the click.
    for (const c of phasedOutCards(view)) {
      const list = out.get(c.controller);
      if (list) list.push(c);
      else out.set(c.controller, [c]);
    }
    return out;
  });

  function exileForOwner(ownerID: string): ZoneView {
    const cards = view.exile.cards.filter((c) => c.owner === ownerID);
    return { kind: "exile", owner: ownerID, count: cards.length, cards };
  }

  // S20 sub-PR 3: a starting suggestion for the X prompt — floating
  // mana plus untapped permanents the viewer controls that produce
  // mana (basic lands + anything with a mana ability), minus the
  // coloured part of the cost the prompt itself shows. Rough by
  // design; the modal's live preview is the real check.
  const suggestedX = $derived.by(() => {
    if (!viewerID) return 0;
    const me = view.seats.find((s) => s.id === viewerID);
    const floating = me?.mana_pool?.length ?? 0;
    let sources = 0;
    for (const c of view.battlefield.cards) {
      if (c.controller !== viewerID || c.tapped) continue;
      const t = (c.type_line ?? "").toLowerCase();
      if (t.includes("land") || (c.mana_abilities?.length ?? 0) > 0) sources++;
    }
    return Math.max(0, floating + sources - 1);
  });

  // The ability picker's opening guess. Same mana estimate, divided
  // by the number of {X} slots — Treasure Vault's "{X}{X}" buys half
  // the X the same mana buys a one-slot cost — and never below the
  // printed floor, which the modal also enforces.
  const suggestedAbilityX = $derived.by(() =>
    xAbilityPrompt ? suggestedAbilityXFor(xAbilityPrompt.ability, suggestedX, viewerEnergy) : 0,
  );

  // ADR 0129 §8: the viewer's energy, the ceiling on a "Pay X {E}"
  // ability's X (CR 118.3).
  const viewerEnergy = $derived(view.seats.find((s) => s.id === viewerID)?.counters?.energy ?? 0);

  // #916: the viewer's life total, which is CR 119.4's cap on a
  // Phyrexian life payment. Read off the live snapshot so a life loss
  // while a prompt is open shrinks its ceiling.
  const viewerLife = $derived(view.seats.find((s) => s.id === viewerID)?.life ?? 0);

  const activeSeatID = $derived(view.seats[view.turn.active_seat]?.id ?? null);
  const prioritySeatID = $derived(view.seats[view.turn.priority_holder]?.id ?? null);
  const monarchID = $derived(view.monarch ?? null);
  const initiativeID = $derived(view.initiative ?? null);

  // ---- "considering a response…" chip (#1307) -----------------------
  //
  // See considering.ts's header for the rationale: the chip has to be
  // derived from public state and elapsed time ALONE, so a real hold,
  // a timed bluff, a manual bluff and someone away from the keyboard
  // all read the same way to every other seat. Board owns the timer
  // because it's the one place that sees every seat and mounts every
  // panel; the pure predicate stays in considering.ts so it's testable
  // without a component.
  //
  // The delay is consideringDelayMs: 800 ms, which an automatic pass
  // always beats, or in a stack window the largest stack hold plus
  // 800 ms, because an automatic pass there now waits for the holder's
  // hold (ADR 0119 §2).
  let consideringSeatID = $state<string | null>(null);
  let consideringTimer: ReturnType<typeof setTimeout> | null = null;
  // Plain (non-reactive) watermark: `view` is a brand-new object on
  // every snapshot, so an effect that merely reads it re-runs on every
  // broadcast — a life total changing included. Comparing against the
  // last key this effect actually acted on is what turns that into
  // "only when the response window itself changed".
  let lastConsideringKey = "";

  $effect(() => {
    const key = responseWindowKey(view);
    if (key === lastConsideringKey) return;
    lastConsideringKey = key;

    consideringSeatID = null;
    if (consideringTimer !== null) {
      clearTimeout(consideringTimer);
      consideringTimer = null;
    }

    const holderIdx = view.turn?.priority_holder ?? -1;
    const holder = holderIdx >= 0 ? (view.seats[holderIdx] ?? null) : null;
    // Never the viewer's own seat (nothing to signal to yourself),
    // never a bot (which already has its own thinking chip), never an
    // eliminated seat.
    if (!holder || holder.id === viewerID || holder.is_bot === true || holder.eliminated) {
      return;
    }

    const delay = consideringDelayMs(view);
    consideringTimer = setTimeout(() => {
      consideringTimer = null;
      // Re-read fresh rather than trust the closure: the window this
      // timer was scheduled for might have already moved on to the
      // next one (which would have reset lastConsideringKey and
      // cleared this very timer) — this check is the belt to that
      // brace for a timer that somehow still fires anyway.
      if (responseWindowKey(view) !== key) return;
      if (!isResponseWindowFor(view, holderIdx)) return;
      consideringSeatID = holder.id;
    }, delay);
  });

  // Teardown-only: clears an in-flight timer on unmount (leaving the
  // game, switching tables). Kept dependency-free so it doesn't fire
  // on every snapshot the way the effect above deliberately does.
  $effect(() => {
    return () => {
      if (consideringTimer !== null) clearTimeout(consideringTimer);
    };
  });

  // #1438: the mana picker's store is module-scoped, and since ADR 0117
  // so is the open ability popover, so leaving the table must not leave
  // either waiting for the next one.
  $effect(() => {
    return () => {
      closeManaSourcePicker();
      closeAbilityPopover();
    };
  });

  function handleTapToggle(card: CardView): void {
    guardedSendAction(card.tapped ? "untap" : "tap", { instance_id: card.instance_id });
  }

  // The cast flow is a chain of announce-time prompts, each handing
  // the accumulated CastChoices to the next: alternative cost →
  // discard cost → sacrifice cost → X → modes → targeting → fire.
  // Every stage stashes the choices alongside its own prompt card,
  // because a Svelte modal's confirm arrives on a later tick.
  //
  // Ordering note (S22): the alternative cost goes FIRST, and unlike
  // the rest of the order that is not a UI preference. Overload and
  // cleave rewrite the target clause, so the choice decides what the
  // targeting prompt is even allowed to offer — asking afterwards
  // would mean re-opening it.

  // S20 sub-PR 3: an {X} spell asks for X. The modal's confirm
  // continues into targeting / cast with the chosen value.
  let xPromptCard = $state<CardView | null>(null);
  // $state because the X picker's cost preview reads it: the
  // announcement so far (source zone, alternative cost, optional
  // costs, face) is what the preview prices against (#696), so the
  // template has to see it change when the prompt opens.
  let xPromptChoices = $state<CastChoices>({});
  function confirmX(x: number): void {
    const card = xPromptCard;
    const choices = xPromptChoices;
    xPromptCard = null;
    xPromptChoices = {};
    if (!card) return;
    afterXCost(card, { ...choices, xValue: x });
  }

  // S22: a card that offers a cost paid INSTEAD of its mana cost
  // ("Overload {6}{U}", "Evoke {3}{U}", "Cleave {1}{U}{U}") asks
  // which one before anything else. Declining is a first-class
  // answer — the picker always offers the printed cost — and is what
  // the vast majority of casts of these cards will be.
  let altCostPromptCard = $state<CardView | null>(null);
  let altCostPromptChoices: CastChoices = {};
  function confirmAltCost(
    key: string | undefined,
    optional: number[],
    giftOpponent?: string,
    costBranch?: number,
  ): void {
    const card = altCostPromptCard;
    const choices = altCostPromptChoices;
    altCostPromptCard = null;
    altCostPromptChoices = {};
    if (!card) return;
    // ADR 0073 (#664): one prompt answers both halves of CR 601.2b —
    // the cost paid INSTEAD of the mana cost, and the costs paid ON
    // TOP of whichever that turns out to be. They compose, so neither
    // branch of `key` drops the other.
    let next: CastChoices = key === undefined ? choices : { ...choices, altCost: key };
    if (optional.length > 0) next = { ...next, optionalCosts: optional };
    // #1267: the gift's opponent is part of the same announcement; the
    // modal only hands one back when the gift toggle is on.
    if (giftOpponent !== undefined) next = { ...next, giftOpponent };
    // ADR 0100: the either/or branch is announced in the same step
    // (CR 601.2b); the branch's own picker opens next in the chain.
    if (costBranch !== undefined) next = { ...next, costBranch };
    afterAltCost(card, next);
  }

  // S21 sub-PR 5: a spell with an additional cost ("As an
  // additional cost to cast this spell, discard a card") asks for
  // the payment first. The picked IDs thread through the rest of
  // the cast flow the way xValue does and ride cast_spell as
  // discard_ids; the server validates them at announce and pays
  // them once the spell is on the stack.
  let discardPromptCard = $state<CardView | null>(null);
  let discardPromptChoices: CastChoices = {};
  // ADR 0100: the additional cost this cast pays, settled when the
  // prompt opens — the chosen either/or branch, or the card's own.
  let discardPromptCost = $state<AdditionalCostView | undefined>(undefined);

  // Everything in the caster's hand except the spell itself — CR
  // 601.2a puts it on the stack before costs are paid, so it can't
  // pay for itself.
  const discardCostOptions = $derived.by(() => {
    const card = discardPromptCard;
    if (!card || !viewerID) return [];
    const me = view.seats.find((s) => s.id === viewerID);
    return (me?.hand.cards ?? []).filter((c) => c.instance_id !== card.instance_id);
  });

  function confirmDiscardCost(ids: string[]): void {
    const card = discardPromptCard;
    const choices = discardPromptChoices;
    discardPromptCard = null;
    discardPromptChoices = {};
    if (!card) return;
    afterDiscardCost(card, { ...choices, discardIDs: ids });
  }

  // S21 sub-PR 6: the sacrifice half of an additional cost ("As an
  // additional cost to cast this spell, sacrifice a creature").
  // Reuses SacrificeCostModal, the same picker the CR 602 activated
  // abilities open. No printed card charges both a discard and a
  // sacrifice, but the two prompts chain rather than exclude each
  // other, so one that did would work.
  let sacrificePromptCard = $state<CardView | null>(null);
  let sacrificePromptChoices: CastChoices = {};

  // #747: in the server's payment order, not board order, so the
  // picker's "Choose for me" takes the top of the list.
  // ADR 0073: WHICH clause this cast is paying is settled when the
  // prompt opens, not re-derived while it is open. The announcement
  // cannot change underneath an open picker — the optional costs were
  // claimed two prompts ago — and stashing it keeps the derived option
  // list depending only on the board, which is the thing that CAN
  // change while the player is choosing.
  let sacrificePromptClause = $state<LegalTargetsView | undefined>(undefined);
  let sacrificePromptLabel = $state("a permanent");
  const castSacrificeOptions = $derived.by(() => {
    if (!sacrificePromptCard) return [];
    return orderSacrificeOptions(view.battlefield.cards, sacrificePromptClause?.cards);
  });
  // ADR 0100 §3: the bounds the cast's picker enforces. A fixed clause is
  // N..N as before; "sacrifice any number of creatures" and "sacrifice X
  // lands" run from zero with no ceiling but the board.
  const castSacrificeBounds = $derived(castSacrificeRange(sacrificePromptClause));

  function confirmSacrificeCost(instanceIDs: string[]): void {
    const card = sacrificePromptCard;
    const choices = sacrificePromptChoices;
    const clause = sacrificePromptClause;
    sacrificePromptCard = null;
    sacrificePromptChoices = {};
    if (!card) return;
    afterSacrificeCost(card, withCastSacrifice(choices, clause, instanceIDs));
  }

  // withCastSacrifice records the permanents a cast's sacrifice clause
  // names. ADR 0100 §3: for "sacrifice X …" the number picked IS the
  // announced X (CR 107.3i), so it rides as the x_value and the X prompt
  // never opens — the tap-X rule (ADR 0073 Decision 10).
  function withCastSacrifice(
    choices: CastChoices,
    clause: LegalTargetsView | undefined,
    ids: string[],
  ): CastChoices {
    const out: CastChoices = { ...choices, sacrificeIDs: ids };
    if (clause?.count_from_x) out.xValue = ids.length;
    return out;
  }

  // #1703: a claimed teamwork offer asks which creatures to tap — the
  // crew picker, since CR 702.194a is crew's sentence on a spell — and
  // a claimed blight asks which one creature takes the counters. Both
  // after the sacrifice step and before X, in the order the server
  // validates the announcement.
  let teamworkPrompt = $state<{ card: CardView; n: number; choices: CastChoices } | null>(null);
  let teamworkPromptOptionIDs = $state<string[]>([]);
  const teamworkOptions = $derived.by(() => {
    const ids = new Set(teamworkPromptOptionIDs);
    return view.battlefield.cards.filter((c) => ids.has(c.instance_id));
  });
  let blightPrompt = $state<{
    card: CardView;
    label: string;
    choices: CastChoices;
  } | null>(null);
  let blightPromptOptionIDs = $state<string[]>([]);
  const blightOptions = $derived.by(() =>
    orderSacrificeOptions(view.battlefield.cards, blightPromptOptionIDs),
  );

  // ADR 0100 amendment 2026-10-07: a chosen reveal / behold branch asks
  // which one card to show — a card in your hand or, to behold, a
  // permanent you control. The same single-pick sheet the blight
  // creature uses, over the server's `reveal_options` (hand first, then
  // permanents), resolved against both zones.
  let revealPrompt = $state<{ card: CardView; label: string; choices: CastChoices } | null>(null);
  let revealPromptOptionIDs = $state<string[]>([]);
  const revealOptions = $derived.by(() => {
    const me = view.seats.find((s) => s.id === viewerID);
    return orderSacrificeOptions(
      [...(me?.hand.cards ?? []), ...view.battlefield.cards],
      revealPromptOptionIDs,
    );
  });

  function confirmReveal(ids: string[]): void {
    const p = revealPrompt;
    revealPrompt = null;
    if (!p) return;
    afterSacrificeCost(p.card, { ...p.choices, revealIDs: ids });
  }

  function afterSacrificeCost(card: CardView, choices: CastChoices): void {
    const tw = castTeamworkOffer(card, choices);
    if (tw && choices.teamworkIDs === undefined) {
      teamworkPromptOptionIDs = tw.options;
      teamworkPrompt = { card, n: tw.n, choices };
      return;
    }
    const bl = castBlightOffer(card, choices);
    if (bl && choices.blightIDs === undefined) {
      blightPromptOptionIDs = bl.options;
      blightPrompt = {
        card,
        label:
          bl.n === 0
            ? `a creature you control for ${bl.offer.label ?? "Blight X"} (it gets X -1/-1 counters)`
            : `a creature you control for ${bl.offer.label ?? `Blight ${bl.n}`} (it gets ${bl.n} -1/-1 counter${bl.n === 1 ? "" : "s"})`,
        choices,
      };
      return;
    }
    const rv = castRevealCost(card, choices);
    if (rv && choices.revealIDs === undefined) {
      revealPromptOptionIDs = rv.options;
      revealPrompt = {
        card,
        label: rv.behold ? `${rv.label} (choose or reveal)` : rv.label,
        choices,
      };
      return;
    }
    afterCastCosts(card, choices);
  }

  function confirmTeamwork(ids: string[]): void {
    const p = teamworkPrompt;
    teamworkPrompt = null;
    if (!p) return;
    afterSacrificeCost(p.card, { ...p.choices, teamworkIDs: ids });
  }

  function confirmBlight(ids: string[]): void {
    const p = blightPrompt;
    blightPrompt = null;
    if (!p) return;
    afterSacrificeCost(p.card, { ...p.choices, blightIDs: ids });
  }

  // S22: convoke / waterbend — "you may tap your own untapped
  // permanents to help pay for this". Opens after the X prompt,
  // because a waterbend {X} cost has no size until X is announced,
  // and before targeting, because paying is what makes the spell
  // castable at all. Tapping nothing is always a legal answer.
  let tapPromptCard = $state<CardView | null>(null);
  let tapPromptChoices: CastChoices = {};

  const tapCostOptions = $derived.by(() => {
    const card = tapPromptCard;
    if (!card) return [];
    const ids = new Set(tapCostOf(card)?.options?.cards ?? []);
    return view.battlefield.cards.filter((c) => ids.has(c.instance_id));
  });

  const tapCostCap = $derived.by(() => {
    const tc = tapPromptCard ? tapCostOf(tapPromptCard) : undefined;
    if (!tc) return 0;
    return tapCostLimit(tc, tapPromptChoices.xValue);
  });

  function confirmTapCost(ids: string[]): void {
    const card = tapPromptCard;
    const choices = tapPromptChoices;
    tapPromptCard = null;
    tapPromptChoices = {};
    if (!card) return;
    afterTapCost(card, { ...choices, tapIDs: ids });
  }

  // ADR 0100: delve — "you may exile cards from your graveyard to pay
  // generic mana" (CR 702.66a). Opens after the convoke picker, because
  // the taps change how much generic is left to delve, and before
  // targeting, like the taps. Exiling nothing is always legal. The
  // modal asks the server for the cap; the options are the server's.
  let delvePromptCard = $state<CardView | null>(null);
  let delvePromptChoices = $state<CastChoices>({});

  const delveCostOptions = $derived.by(() => {
    const card = delvePromptCard;
    if (!card || !viewerID) return [];
    const me = view.seats.find((s) => s.id === viewerID);
    const pile = me?.graveyard.cards ?? [];
    const byID = new Map(pile.map((c) => [c.instance_id, c]));
    return delveOptionIDs(card)
      .map((id) => byID.get(id))
      .filter((c): c is CardView => c !== undefined);
  });

  function afterTapCost(card: CardView, choices: CastChoices): void {
    if (hasDelveChoice(card)) {
      delvePromptChoices = choices;
      delvePromptCard = card;
      return;
    }
    continueCast(card, choices);
  }

  function confirmDelve(ids: string[]): void {
    const card = delvePromptCard;
    const choices = delvePromptChoices;
    delvePromptCard = null;
    delvePromptChoices = {};
    if (!card) return;
    continueCast(card, ids.length > 0 ? { ...choices, delveIDs: ids } : choices);
  }

  // S28: the non-mana half of a chosen alternative cost — Force of
  // Will's "exile a blue card from your hand", Daze's "return an
  // Island you control", Solitude's evoke pitch. Opens immediately
  // after the alternative-cost picker, because it pays the cost that
  // picker just chose; every other cost prompt comes after it.
  //
  // The options can live in any of three zones — hand for a pitch,
  // battlefield for a bounce, graveyard for S29's escape exiles — so
  // the lookup searches all of them rather than assuming one. The
  // server has already narrowed the ID set; this only has to find
  // the CardView behind each ID.
  let altPayPromptCard = $state<CardView | null>(null);
  let altPayPromptChoices: CastChoices = {};

  const altPayOffer = $derived(
    altPayPromptCard
      ? alternativeCostByKey(altPayPromptCard, altPayPromptChoices.altCost)
      : undefined,
  );

  const altPayOptions = $derived.by(() => {
    if (!altPayPromptCard || !viewerID) return [];
    const ids = new Set(altCostPayOptions(altPayOffer) ?? []);
    if (ids.size === 0) return [];
    const me = view.seats.find((s) => s.id === viewerID);
    const pool = [
      ...(me?.hand.cards ?? []),
      ...(me?.graveyard.cards ?? []),
      ...view.battlefield.cards,
    ];
    return pool.filter((c) => ids.has(c.instance_id));
  });

  function confirmAltPay(instanceIDs: string[]): void {
    const card = altPayPromptCard;
    const choices = altPayPromptChoices;
    altPayPromptCard = null;
    altPayPromptChoices = {};
    if (!card) return;
    afterAltCostPayment(card, { ...choices, altCostIDs: instanceIDs });
  }

  // #1727: an offer whose card half is a SACRIFICE (Dread Return's
  // flashback, Fireblast) opens the one sacrifice picker every other
  // sacrifice cost uses — payment order, "Choose for me" — instead of
  // AltCostPaymentModal. The clause is stashed when the prompt opens,
  // as the cast's own sacrifice prompt does; the picks pay THIS offer,
  // so they ride alt_cost_ids, not sacrifice_ids.
  let altSacPromptCard = $state<CardView | null>(null);
  let altSacPromptChoices: CastChoices = {};
  let altSacPromptClause = $state<LegalTargetsView | undefined>(undefined);
  let altSacPromptLabel = $state("a permanent");
  const altSacOptions = $derived.by(() => {
    if (!altSacPromptCard) return [];
    return orderSacrificeOptions(view.battlefield.cards, altSacPromptClause?.cards);
  });
  const altSacBounds = $derived(castSacrificeRange(altSacPromptClause));

  function confirmAltSacrifice(instanceIDs: string[]): void {
    const card = altSacPromptCard;
    const choices = altSacPromptChoices;
    altSacPromptCard = null;
    altSacPromptChoices = {};
    altSacPromptClause = undefined;
    if (!card) return;
    afterAltCostPayment(card, { ...choices, altCostIDs: instanceIDs });
  }

  // ADR 0135 §1: an offer whose card half TAPS permanents (Orim's Cure,
  // Battle Screech's flashback) opens the tap picker an ability's "tap an
  // untapped creature you control" uses, the sacrifice picker with the
  // verb "Tap". It is shown even when the board offers exactly the count:
  // tapping a blocker is a choice the player should see. The picks pay
  // the offer, on alt_cost_ids.
  let altTapPromptCard = $state<CardView | null>(null);
  let altTapPromptChoices: CastChoices = {};
  let altTapPromptClause = $state<LegalTargetsView | undefined>(undefined);
  let altTapPromptLabel = $state("an untapped creature you control");
  const altTapOptions = $derived.by(() => {
    if (!altTapPromptCard) return [];
    return orderSacrificeOptions(view.battlefield.cards, altTapPromptClause?.cards);
  });
  const altTapBounds = $derived(sacrificeRange(altTapPromptClause));

  function confirmAltTap(instanceIDs: string[]): void {
    const card = altTapPromptCard;
    const choices = altTapPromptChoices;
    altTapPromptCard = null;
    altTapPromptChoices = {};
    altTapPromptClause = undefined;
    if (!card) return;
    afterAltCostPayment(card, { ...choices, altCostIDs: instanceIDs });
  }

  // afterAltCost / afterDiscardCost / afterCastCosts are the seams
  // between the cost prompts and the rest of the cast flow, so adding
  // a cost kind doesn't mean editing every earlier prompt's confirm.
  //
  // An alternative cost REPLACES the mana cost but not the additional
  // costs (CR 601.2f is evaluated independently), so the chain
  // continues through the discard / sacrifice prompts rather than
  // short-circuiting past them.
  function afterAltCost(card: CardView, choices: CastChoices): void {
    // The chosen offer may charge a card as well as — or instead of —
    // mana. Ask for it before anything else, matching the order the
    // server validates the cast in.
    const offer = alternativeCostByKey(card, choices.altCost);
    const sacClause = altCostSacrificeClause(offer);
    if (sacClause !== undefined) {
      altSacPromptClause = sacClause;
      altSacPromptLabel = offer?.pay_label ?? offer?.label ?? "a permanent";
      altSacPromptChoices = choices;
      altSacPromptCard = card;
      return;
    }
    const tapClause = altCostTapClause(offer);
    if (tapClause !== undefined) {
      altTapPromptClause = tapClause;
      altTapPromptLabel = offer?.pay_label ?? "an untapped creature you control";
      altTapPromptChoices = choices;
      altTapPromptCard = card;
      return;
    }
    if (altCostPayOptions(offer) !== undefined) {
      altPayPromptChoices = choices;
      altPayPromptCard = card;
      return;
    }
    afterAltCostPayment(card, choices);
  }

  function afterAltCostPayment(card: CardView, choices: CastChoices): void {
    // ADR 0100: an either/or cost's discard is the chosen branch's.
    if (discardCostOf(card, choices) > 0) {
      discardPromptCost = castAdditionalCost(card, choices);
      discardPromptChoices = choices;
      discardPromptCard = card;
      return;
    }
    afterDiscardCost(card, choices);
  }

  function afterDiscardCost(card: CardView, choices: CastChoices): void {
    // ADR 0073: a NON-MANA optional cost (Constant Mists'
    // "Buyback—Sacrifice a land") is paid through the same picker the
    // mandatory clause opens. castSacrificeClause picks whichever
    // clause this cast is actually paying, so the modal's options,
    // count and label all come from one place.
    const clause = castSacrificeClause(card, choices);
    // ADR 0100 §3: a count that may be zero, with nothing on the board
    // to pay it, has one answer — none — so there is nothing to ask.
    if (
      clause !== undefined &&
      castSacrificeRange(clause).min === 0 &&
      (clause.cards ?? []).length === 0
    ) {
      afterSacrificeCost(card, withCastSacrifice(choices, clause, []));
      return;
    }
    if (clause !== undefined) {
      sacrificePromptClause = clause;
      sacrificePromptLabel = castSacrificeLabel(card, choices);
      sacrificePromptChoices = choices;
      sacrificePromptCard = card;
      return;
    }
    afterSacrificeCost(card, choices);
  }

  function afterCastCosts(card: CardView, choices: CastChoices): void {
    // CR 107.3b (#831): a free cast of an {X} spell has one legal X
    // and it is 0, so the picker is skipped and nothing is sent. The
    // server refuses a non-zero X on such a cast, which is what makes
    // this a prompt decision rather than a rule the client enforces.
    if (hasXCost(card, choices.altCost) && !castLocksXAtZero(card, choices.altCost)) {
      xPromptChoices = choices;
      xPromptCard = card;
      return;
    }
    afterXCost(card, choices);
  }

  // afterXCost is the seam between the X prompt and the rest of the
  // announcement, and the reason the tap picker isn't in
  // afterCastCosts with the others: a waterbend {X} cost can't size
  // its picker until X is known, so this step has to come after the X
  // prompt rather than before it.
  function afterXCost(card: CardView, choices: CastChoices): void {
    // CR 107.4f (#916): "{U/P} can be paid with either {U} or 2
    // life", and CR 601.2b makes which one part of announcing the
    // spell. Asked after X for the same reason the tap picker is:
    // the stepper's live readout prices the mana half, and an {X}
    // cost has no size until X is announced. Skipped when the cost
    // prints no Phyrexian symbol, and when CR 119.4 leaves the
    // caster unable to buy even one — a prompt whose only answer is
    // 0 is a click, not a choice.
    //
    // ADR 0131 (owner decision 4): when EVERY symbol is one a grant
    // (K'rrik) lets life pay, the stepper would open on nearly every
    // black spell, so it opens only when the mana falls short — asked of
    // the auto-tap preview with nothing claimed — or when the player
    // chose "Pay life for {B}…" from the card's menu. A printed
    // Phyrexian symbol still always asks.
    const symbols = phyrexianSymbolsForCast(card, choices.altCost);
    const granted = phyrexianGrantedForCast(card, choices.altCost);
    const asked = choices.askPhyrexianLife === true;
    if (onlyGrantedSymbols(symbols, granted) && !asked) {
      // A confirmed Cast anyway pays no mana, so mana is never short;
      // with strict payment off the engine charges none either.
      if (choices.forceCast || !$settings.gameplay.strictMana) {
        afterPhyrexianLife(card, choices);
        return;
      }
      fetchAutoTapPreview(view.id, card.instance_id, {
        xValue: choices.xValue,
        cast: castPreviewParams(choices),
      })
        .then((p) => {
          if (shouldAskPhyrexianLife(symbols, viewerLife, granted, !p.ok)) {
            phyrexianPrompt = { card, symbols, granted, suggest: true, choices };
          } else {
            afterPhyrexianLife(card, choices);
          }
        })
        // A preview that failed says nothing: cast on, and the server,
        // the only real answer, refuses a short strict cast.
        .catch(() => afterPhyrexianLife(card, choices));
      return;
    }
    if (shouldAskPhyrexianLife(symbols, viewerLife, granted, true, asked)) {
      phyrexianPrompt = { card, symbols, granted, suggest: false, choices };
      return;
    }
    afterPhyrexianLife(card, choices);
  }

  function afterPhyrexianLife(card: CardView, choices: CastChoices): void {
    if (tapCostOf(card)) {
      tapPromptChoices = choices;
      tapPromptCard = card;
      return;
    }
    afterTapCost(card, choices);
  }

  // #916: the cast's Phyrexian stepper. One reactive object rather
  // than the card / choices pair the older stages use, because the
  // modal reads the stashed choices (the announced X, the symbol
  // count the chosen cost prints) as well as the card — the same
  // shape xAbilityPrompt has, for the same reason.
  let phyrexianPrompt = $state<{
    card: CardView;
    symbols: number;
    // ADR 0131: how many of `symbols` are a grant's, not printed.
    granted: number;
    // Open at the smallest count that makes the mana half payable.
    suggest: boolean;
    choices: CastChoices;
  } | null>(null);

  function confirmPhyrexianLife(n: number): void {
    const p = phyrexianPrompt;
    phyrexianPrompt = null;
    if (!p) return;
    afterPhyrexianLife(p.card, n > 0 ? { ...p.choices, phyrexianLife: n } : p.choices);
  }

  // ADR 0034: a modal double-faced card asks which HALF first —
  // ahead of even the alternative cost, and for a stronger reason
  // than the one that put the alternative cost ahead of targeting.
  // An overload rewrites a spell's target clause; a face choice
  // decides what the card IS. Sea Gate Restoration is a seven-mana
  // sorcery and Sea Gate, Reborn is a land, and every prompt after
  // this one — costs, modes, targets, even whether the card touches
  // the stack at all — depends on which of them the player meant.
  let facePromptCard = $state<CardView | null>(null);
  // #1508: the choices the cast started with — its zone and whether it
  // was dragged — so the face picker's confirm carries both on.
  let facePromptBase: CastChoices = {};
  // ADR 0103: the zone of the cast the picker is open for, as state so
  // the picker re-reads it — which halves it offers depends on it.
  let facePromptZone = $state<CastSourceZone | undefined>(undefined);
  function confirmFace(face: number, fused = false): void {
    const card = facePromptCard;
    const base = facePromptBase;
    facePromptCard = null;
    facePromptBase = {};
    if (!card) return;
    // ADR 0103, CR 702.102: the FUSED cast of a split card with fuse —
    // both halves, the server's fused announce block, and `fuse: true`
    // on the cast.
    if (fused) {
      afterFace(cardAsFused(card), { ...base, fuse: true });
      return;
    }
    // Run the rest of the chain against the CHOSEN face, so the
    // prompts and the cast-timing checks see its type line, its cost
    // and — since #992 — its own announce data: the cost picker, the
    // mode picker and the TARGET picker all read the block the server
    // published for this half. Casting Stomp opens a target picker
    // here; before #992 cardAsFace cleared the front's answers and
    // put nothing back, so the chain fell through to an announce the
    // server refused.
    //
    // Face 0 goes through it too. cardAsFace swaps rather than clears
    // now, and face 0's block is the same answer the card's top-level
    // block carries, so the special case the old code needed is gone
    // — and one path is one path.
    afterFace(cardAsFace(card, face), { ...base, face });
  }

  function afterFace(card: CardView, choices: CastChoices): void {
    // ADR 0073: a card with kicker and no alternative cost opens the
    // same picker with only the add-ons showing — one prompt for one
    // question (CR 601.2b), rather than a second modal asking the
    // other half of it.
    //
    // ADR 0100: and an either/or additional cost's branch radio, the
    // same "what am I paying for this?" question.
    if (
      alternativeCostsOf(card).length > 0 ||
      optionalCostsOf(card).length > 0 ||
      costBranchesOf(card).length > 0
    ) {
      altCostPromptChoices = choices;
      altCostPromptCard = card;
      return;
    }
    afterAltCost(card, choices);
  }

  // handlePlayCard is the head of the chain — the ONE entry point for
  // casting a card, whichever surface the click came from. `fromZone`
  // is undefined for the hand, which is every cast the board's own
  // surfaces fire; S29's zone browser passes "graveyard" so a
  // flashback cast walks the identical prompt chain, and #874 adds
  // "exile", so an impulse cast does too.
  //
  // `face` is a face the CALLER already knows, which is only ever an
  // exile grant naming one — a defeated Siege's back face, or the
  // creature half of an adventure card exiled by its own Adventure
  // (CR 715.4). It is not a default for the picker: a grant that names
  // a face offers no choice, so asking would be asking a question with
  // one answer, and the server ignores the request and uses the
  // grant's face anyway (game/face.go, faceForCastLocked). What
  // passing it buys is that the REST of the chain — the X picker, the
  // targets, the modes — reads the half being cast rather than the
  // half sitting face-up in exile.
  //
  // Which is why the gate is "defined", not "greater than zero". A
  // CR 715.4 grant names face 0, and a `face > 0` test read that as
  // "no face given" and re-opened the picker on a cast with exactly
  // one legal half. Face 0 goes through cardAsFace like any other
  // since #992: the function swaps a face's announce block in rather
  // than clearing the card's, and face 0's block is the same answer
  // the card already carries.
  //
  // The face picker's confirm re-enters at afterFace with its own
  // choices object, so the zone has to be seeded here rather than at
  // the end — otherwise a modal DFC cast out of the graveyard would
  // lose it.
  //
  // #1508: `viaDrag` is set only by the hand's drag-to-cast gesture. It
  // rides CastChoices through every prompt and applyCastChoices turns
  // it into `strict: true, auto_tap: true` on whichever cast_spell the
  // chain finally sends.
  //
  // ADR 0118 §2: `forceCast` is set only by a confirmed "Cast anyway
  // (don't pay)". It rides the chain the same way, and applyCastChoices
  // turns it into `strict: true, force_cast: true`.
  function handlePlayCard(
    card: CardView,
    fromZone?: CastSourceZone,
    face?: number,
    viaDrag = false,
    forceCast = false,
    askPhyrexianLife = false,
  ): void {
    const base = castChoicesBase(fromZone, viaDrag, forceCast, askPhyrexianLife);
    if (face !== undefined) {
      afterFace(cardAsFace(card, face), { ...base, face });
      return;
    }
    if (needsFacePicker(card)) {
      // ADR 0103: a split card may have only one half this zone allows
      // (an aftermath card in hand) — then there is nothing to ask.
      const options = faceOptions(card, fromZone);
      if (options.length === 1) {
        const only = options[0];
        afterFace(only.view, only.fused ? { ...base, fuse: true } : { ...base, face: only.face });
        return;
      }
      facePromptBase = base;
      facePromptZone = base.fromZone;
      facePromptCard = card;
      return;
    }
    afterFace(card, base);
  }

  // ADR 0103: a door button on a Room (Card.svelte's door strip) asks
  // for an unlock through the roomDoors store; the Board is the one
  // place that sends it, as the context menu's rows are.
  $effect(() => {
    const req = $unlockRequest;
    if (!req) return;
    unlockRequest.set(null);
    guardedSendAction("special_action", unlockParams(req.cardID, req.door), viewerID ?? undefined);
  });

  // ADR 0118 §2 (owner decision 6): "Cast anyway (don't pay)" asks
  // first. The row (Hand, the strip, the command zone panel, the admin
  // menu) only sets castAnywayPending; while it is set the dock asks
  // "Cast <card> without paying its mana cost?". Cast starts the ordinary
  // cast chain with `forceCast`; Cancel and Escape send nothing. The
  // question goes away, also sending nothing, once the card has left the
  // zone it was asked about.
  const castAnywayRequest = $derived.by(() => {
    const p = $castAnywayPending;
    if (!p) return null;
    return castAnywayConfirmRequest(p.card.name, {
      onCast: () => {
        clearCastAnyway();
        // The card as the frame has it now, so the chain reads today's
        // targets, modes and costs rather than the right-click's.
        const live = findCard(view, p.card.instance_id) ?? p.card;
        handlePlayCard(live, p.zone === "hand" ? undefined : p.zone, p.face, false, true);
      },
      onCancel: clearCastAnyway,
    });
  });
  $effect(() => {
    const p = $castAnywayPending;
    if (!p) return;
    if (locateCard(view, p.card.instance_id)?.zone !== p.zone) clearCastAnyway();
  });

  // ADR 0131 §4: the card menu's "Pay life for {B}…" row only sets
  // payLifeRequest; the Board starts the ordinary cast chain with the
  // Phyrexian stepper forced open, so a player whose mana is there can
  // still choose to pay life.
  $effect(() => {
    const card = $payLifeRequest;
    if (!card) return;
    payLifeRequest.set(null);
    const live = findCard(view, card.instance_id) ?? card;
    handlePlayCard(live, undefined, undefined, false, false, true);
  });

  // ADR 0099 §7: "Cast it free" on a discover or cascade prompt starts
  // the cast chain for the exiled card as soon as the snapshot carrying
  // its grant arrives. Cancelling the chain leaves the card in exile
  // with its ordinary cast button; passing puts it where the keyword
  // sends it.
  $effect(() => {
    const id = $freeCastRequest;
    if (!id) return;
    const target = freeCastTarget(view.exile?.cards, id);
    if (target === undefined) return;
    freeCastRequest.set(null);
    if (target) handlePlayCard(target, "exile");
  });

  // S20 sub-PR 4: a modal spell asks for its mode(s) after X and
  // before targeting. The picker's confirm continues with the
  // chosen indexes; a targeted option enters targeting with that
  // option's legal set, otherwise the cast fires straight away.
  let modePromptCard = $state<CardView | null>(null);
  let modePromptChoices: CastChoices = {};
  // #764: EVERY chosen bullet contributes its clauses to the walk,
  // in the order they were chosen (CR 608.2c), and a repeated bullet
  // (CR 700.2d) contributes them once per occurrence. The old shape
  // took the FIRST targeted option and dropped the rest, which is
  // why Kolaghan's Command could not be cast.
  function confirmModes(modes: number[]): void {
    const card = modePromptCard;
    const choices = modePromptChoices;
    modePromptCard = null;
    modePromptChoices = {};
    if (!card) return;
    // #2126, CR 702.120a: escalate's cards or creatures are named once
    // the mode count is known, before the targets.
    const owed = escalateOwed(card, modes);
    if (owed) {
      escalatePrompt = { card, modes, choices, stage: owed };
      return;
    }
    finishModes(card, modes, choices);
  }

  function finishModes(card: CardView, modes: number[], choices: CastChoices): void {
    if (beginTargetingForModes(card, modes, choices)) return;
    const params: Record<string, unknown> = { instance_id: card.instance_id, modes };
    applyCastChoices(params, choices);
    guardedSendAction("cast_spell", params, viewerID ?? undefined);
  }

  // Escalate (CR 702.120a, #2126): one discard prompt, then one tap
  // prompt, each for (modes - 1) times the per-mode count. Both reuse
  // DiscardCostModal as an exact-count card picker; the answers ride
  // discard_ids and teamwork_ids.
  let escalatePrompt = $state<{
    card: CardView;
    modes: number[];
    choices: CastChoices;
    stage: "discard" | "tap";
  } | null>(null);

  function escalateOwed(
    card: CardView,
    modes: number[],
    after?: "discard",
  ): "discard" | "tap" | null {
    const esc = card.modes?.escalate;
    if (!esc || modes.length < 2) return null;
    if (after !== "discard" && (esc.discard_cards ?? 0) > 0) return "discard";
    if ((esc.tap_creatures ?? 0) > 0) return "tap";
    return null;
  }

  const escalateNeed = $derived.by(() => {
    const p = escalatePrompt;
    const esc = p?.card.modes?.escalate;
    if (!p || !esc) return 0;
    const per = p.stage === "discard" ? (esc.discard_cards ?? 0) : (esc.tap_creatures ?? 0);
    return per * (p.modes.length - 1);
  });

  const escalateOptions = $derived.by(() => {
    const p = escalatePrompt;
    if (!p) return [];
    if (p.stage === "discard") {
      const me = view.seats.find((s) => s.id === viewerID);
      return (me?.hand.cards ?? []).filter((c) => c.instance_id !== p.card.instance_id);
    }
    const ids = new Set(p.card.modes?.escalate?.tap_options ?? []);
    return view.battlefield.cards.filter((c) => ids.has(c.instance_id));
  });

  function confirmEscalate(ids: string[]): void {
    const p = escalatePrompt;
    escalatePrompt = null;
    if (!p) return;
    if (p.stage === "discard") {
      const choices: CastChoices = { ...p.choices, discardIDs: ids };
      const next = escalateOwed(p.card, p.modes, "discard");
      if (next) {
        escalatePrompt = { card: p.card, modes: p.modes, choices, stage: next };
        return;
      }
      finishModes(p.card, p.modes, choices);
      return;
    }
    finishModes(p.card, p.modes, { ...p.choices, teamworkIDs: ids });
  }

  // continueCast is the post-cost half of the cast flow: pick modes
  // for a modal card, enter targeting for a targeted card, or fire
  // cast_spell straight away.
  //
  // S22: the target clause is whichever one the chosen cost leaves
  // the spell with, and the server has already resolved that — an
  // overloaded Cyclonic Rift's offer carries no target_mode at all,
  // so this falls through to the immediate cast without the client
  // needing to know what "overload" does.
  function continueCast(card: CardView, choices: CastChoices): void {
    if (isModal(card)) {
      modePromptChoices = choices;
      // #1655: a ticked kicker can change how many bullets it may take.
      modePromptCard = modesUnderChoices(card, choices);
      return;
    }
    // S14: if the card declares a target_mode (catalog cards with
    // a target slot — Lightning Bolt, Counterspell), enter the
    // targeting flow and wait for a second click on a legal target.
    // Otherwise fire cast_spell immediately (lands, sorceries with
    // no targets, vanilla permanents).
    //
    // #1267: an alternative cost's clause replaces the card's, and so
    // does a claimed optional cost's that carries one — a promised
    // gift. That includes a card that prints no target at all but
    // gains one with the gift, which is why this reads the override
    // rather than the card first.
    const alt = castTargetOverride(card, choices);
    const mode = alt ? alt.target_mode : card.target_mode;
    // #1211: the allowlist moved into targeting.ts as
    // opensTargetPicker. It was written out here as a chain of `===`
    // and a mode missing from it does not fall back to a picker — it
    // falls THROUGH to an immediate cast_spell with no targets, which
    // the server then refuses, with nothing on screen to say why. One
    // list, next to the type that declares the vocabulary, and a unit
    // test over it.
    if (opensTargetPicker(mode)) {
      beginTargeting(card, mode, choices, alt);
      return;
    }
    const params: Record<string, unknown> = { instance_id: card.instance_id };
    applyCastChoices(params, choices);
    guardedSendAction("cast_spell", params, viewerID ?? undefined);
  }

  // completeTargetedCast fires cast_spell with the resolved target
  // and clears the targeting store. Called by targetable surfaces
  // (player portraits, battlefield creatures, stack items) when the
  // viewer clicks them while a targeting prompt is active.
  function completeTargetedCast(kind: "player" | "card", targetID: string): void {
    const state = $targeting;
    if (!state) return;
    const ref: TargetRef = { kind, id: targetID };
    // S20 sub-PR 5: a multi-target clause accumulates clicks; the
    // banner's Done (confirmTargets) fires the action.
    if (isMultiPick(state)) {
      targeting.set(togglePick(state, ref));
      return;
    }
    stepOrFire({ ...state, picked: [ref] });
  }

  function confirmTargets(): void {
    const state = $targeting;
    if (!state || !canConfirm(state)) return;
    stepOrFire(state);
  }

  // stepOrFire is the two-step picker's hinge (#764): a walk with
  // another clause to ask about opens the next prompt; a finished
  // walk fires the one action carrying every pick, each stamped with
  // the clause it answered.
  //
  // #1563: a divided step with two or more picks asks for the shares
  // first (CR 601.2d — the division is announced with the targets),
  // and the walk carries on from confirmDivision.
  function stepOrFire(state: TargetingState): void {
    if (needsDivision(state)) {
      dividePrompt = state;
      return;
    }
    continueWalk(state);
  }

  // dividePrompt is the walk paused on a divided step's shares; the
  // targeting store stays live underneath so Back returns to the picks.
  let dividePrompt = $state<TargetingState | null>(null);
  const divideTargets = $derived.by(() => {
    const p = dividePrompt;
    if (!p) return [];
    const names = new Map<string, string>();
    for (const c of view.battlefield.cards) names.set(c.instance_id, c.name);
    for (const seat of view.seats) names.set(seat.id, seat.name);
    return p.picked.map((ref) => ({ id: ref.id, name: names.get(ref.id) ?? ref.id.slice(0, 8) }));
  });
  // A walk that ends underneath the prompt — cancelled, or a trigger's
  // pick_target answered elsewhere — takes the prompt with it.
  $effect(() => {
    if (!$targeting && dividePrompt) dividePrompt = null;
  });
  function confirmDivision(dist: Record<string, number>): void {
    const state = dividePrompt;
    dividePrompt = null;
    if (!state) return;
    continueWalk(withDivision(state, dist));
  }

  function continueWalk(state: TargetingState): void {
    const next = advance(state);
    if (next) {
      targeting.set(next);
      return;
    }
    fireTargets(state, allPicks(state));
  }
  $effect(() => {
    setConfirmHandler(confirmTargets);
    return () => setConfirmHandler(null);
  });

  function fireTargets(state: TargetingState, targets: TargetRef[]): void {
    if (state.choiceID) {
      // S20 sub-PR 2: answering a triggered ability's pick_target
      // prompt. The store clears when the next snapshot no longer
      // carries the choice (see the effect below).
      // #1563: a divided trigger's shares ride the same answer.
      // #2394: a card-set pick answered on the board sends the
      // modal's {choice_id, card_ids} instead (choiceAnswer).
      const answer = choiceAnswer(state, targets);
      const dist = distributionOf(state);
      if (dist) answer.distribution = dist;
      guardedSendAction("resolve_choice", answer, viewerID ?? undefined);
      targeting.set(null);
      return;
    }
    // S21 sub-PR 2: the prompt belongs to an activated ability, not
    // a cast — its cost was already paid at announce.
    if (state.ability) {
      const params: Record<string, unknown> = {
        source_card_id: state.card.instance_id,
        ability_index: state.ability.index,
        // ADR 0093: the row the activation meant, so a stale one is refused.
        ...activatedAbilityRef(state.card, state.ability.index),
        sacrifice_ids: state.ability.sacrificeIDs,
        crew_ids: state.ability.crewIDs,
        // #660: the discard picks were made at announce, before the
        // targeting step, and ride the one activate_ability with the
        // rest of the cost.
        discard_ids: abilityDiscardIDs,
        // #1213: same announcement, same message.
        return_ids: abilityReturnIDs,
        // #759: and the station creature.
        tap_ids: abilityTapIDs,
        ...state.ability.counter,
        targets,
      };
      // #1600: the exile-a-permanent picks (Altar of Bhaal), omitted
      // when the cost has no such component.
      if (abilityExilePermanentIDs.length > 0) {
        params.exile_permanent_ids = abilityExilePermanentIDs;
      }
      // #1297: the exile-N-cards picks, made at announce with the
      // discard picks. Their own field — an exiled card is not
      // discarded — and omitted when the cost has no such component.
      if (abilityExileIDs.length > 0) params.exile_ids = abilityExileIDs;
      // ADR 0109 §7: the card put on top of the library, likewise.
      if (abilityTopIDs.length > 0) params.top_ids = abilityTopIDs;
      // #2598: the cards revealed to pay "Reveal X black cards".
      if (abilityRevealIDs.length > 0) params.reveal_ids = abilityRevealIDs;
      if (state.ability.xValue !== undefined) params.x_value = state.ability.xValue;
      // #916, CR 107.4f: announced with the rest of the cost, before
      // these targets, and sent in the same message.
      if (state.ability.phyrexianLife) params.phyrexian_life = state.ability.phyrexianLife;
      if (state.modes !== undefined) params.modes = state.modes;
      // #1563: the division, announced with the targets (CR 602.2b).
      const abilityDist = distributionOf(state);
      if (abilityDist) params.distribution = abilityDist;
      // #1310: the waterbend taps chosen before the targeting step.
      if (abilityWaterbendIDs && abilityWaterbendIDs.length > 0) {
        params.waterbend_ids = abilityWaterbendIDs;
      }
      abilityWaterbendIDs = undefined;
      abilityRevealIDs = [];
      abilityRevealX = undefined;
      abilityDiscardIDs = [];
      abilityExileIDs = [];
      abilityTopIDs = [];
      abilityReturnIDs = [];
      abilityExilePermanentIDs = [];
      abilityTapIDs = [];
      abilitySacrificeX = undefined;
      abilityDiscardX = undefined;
      abilityTapX = undefined;
      guardedSendAction("activate_ability", params, viewerID ?? undefined);
      targeting.set(null);
      return;
    }
    const params: Record<string, unknown> = { instance_id: state.card.instance_id, targets };
    if (state.modes !== undefined) params.modes = state.modes;
    // #1563: the division, announced with the targets (CR 601.2d).
    const castDist = distributionOf(state);
    if (castDist) params.distribution = castDist;
    // #764: a modal activated ability sends its modes beside its
    // targets (CR 602.2b) — handled in the ability branch above.
    applyCastChoices(params, state.choices);
    guardedSendAction("cast_spell", params, viewerID ?? undefined);
    cancelTargeting();
  }

  // --- S21 sub-PR 2: activated abilities ------------------------
  //
  // Order of operations mirrors CR 601.2 / 602.2: choose the mode
  // (n/a today), pay the costs, then choose targets. So a sacrifice
  // cost is picked BEFORE targeting, and both are sent together in
  // one activate_ability — the server pays and announces atomically.
  //
  // The same modal serves both ability kinds: a CR 602 activated
  // ability (which may go on to target) and a CR 605 mana ability
  // whose cost sacrifices another permanent (Ashnod's Altar). The
  // prompt is tagged so confirm knows which action to send — mana
  // abilities never target, so they fire immediately.
  type SacrificePrompt =
    | { kind: "ability"; card: CardView; ability: ActivatedAbilityView }
    | { kind: "mana"; card: CardView; ability: ManaAbilityView };

  let sacrificePrompt = $state<SacrificePrompt | null>(null);

  const sacrificeOptions = $derived.by(() => {
    const p = sacrificePrompt;
    if (!p) return [];
    return orderSacrificeOptions(view.battlefield.cards, p.ability.sacrifice_options?.cards);
  });

  // #1213: the bounds the picker enforces. A fixed clause is N..N and
  // behaves exactly as it did; an open count ("sacrifice one or more
  // artifacts") is a floor with no ceiling, and a clause whose count
  // is the announced X has no printed bounds at all — the number
  // picked becomes the x_value.
  const sacrificeBounds = $derived(sacrificeRange(sacrificePrompt?.ability.sacrifice_options));

  // #1213: "Return a permanent you control to its owner's hand" as a
  // COST (Quirion Ranger, Master Transmuter, Meloku). The same picker
  // the sacrifice cost uses, one verb over, so there is one component
  // and one "Choose for me" rather than two. Carried on the side like
  // the discard picks, because the announce chain already takes eight
  // arguments.
  let abilityReturnPrompt = $state<{
    card: CardView;
    ability: ActivatedAbilityView;
  } | null>(null);
  let abilityReturnIDs: string[] = [];

  const abilityReturnOptions = $derived.by(() => {
    const p = abilityReturnPrompt;
    if (!p) return [];
    return orderSacrificeOptions(view.battlefield.cards, p.ability.return_options?.cards);
  });

  // #1600: "Exile a creature you control" as a cost (The Soul Stone's
  // harness, Altar of Bhaal, City of Shadows) — the return pick one
  // destination over, asked right after it, through the same picker
  // with the verb "Exile". Its own field on the wire,
  // `exile_permanent_ids`, because `exile_ids` names cards in a pile.
  let abilityExilePermanentPrompt = $state<{
    card: CardView;
    ability: ActivatedAbilityView;
  } | null>(null);
  let abilityExilePermanentIDs: string[] = [];

  const abilityExilePermanentOptions = $derived.by(() => {
    const p = abilityExilePermanentPrompt;
    if (!p) return [];
    return orderSacrificeOptions(view.battlefield.cards, p.ability.exile_permanent_options?.cards);
  });

  // #1600: the same component on a MANA ability (Food Chain).
  let manaExilePermanentPrompt = $state<{
    card: CardView;
    ability: ManaAbilityView;
  } | null>(null);
  let manaExilePermanentIDs: string[] = [];

  const manaExilePermanentOptions = $derived.by(() => {
    const p = manaExilePermanentPrompt;
    if (!p) return [];
    return orderSacrificeOptions(view.battlefield.cards, p.ability.exile_permanent_options?.cards);
  });

  // #759: "Tap another untapped creature you control" as a cost — the
  // station ability (CR 702.184a). The same picker again with the verb
  // "Tap", asked right after the return pick and for the same reason:
  // it names permanents at announce (CR 602.2b). Skipped when the board
  // offers exactly the creatures the clause needs.
  let abilityTapPrompt = $state<{
    card: CardView;
    ability: ActivatedAbilityView;
  } | null>(null);
  let abilityTapIDs: string[] = [];

  const abilityTapOptions = $derived.by(() => {
    const p = abilityTapPrompt;
    if (!p) return [];
    return orderSacrificeOptions(view.battlefield.cards, p.ability.tap_others_options?.cards);
  });

  const abilityTapBounds = $derived(sacrificeRange(abilityTapPrompt?.ability.tap_others_options));

  // #758: the same fixed-count TapOthers component on a MANA
  // ability (Springleaf Drum). Kept beside the activated prompt but
  // with its own state because the two actions finish through
  // different payload builders.
  let manaTapPrompt = $state<{
    card: CardView;
    ability: ManaAbilityView;
  } | null>(null);
  let manaTapIDs: string[] = [];

  const manaTapOptions = $derived.by(() => {
    const p = manaTapPrompt;
    if (!p) return [];
    return orderSacrificeOptions(view.battlefield.cards, p.ability.tap_others_options?.cards);
  });

  // #1213: the X a "Sacrifice X Treasures" clause announces. It is
  // the SIZE of the payment rather than a number the player types, so
  // the X stepper is skipped for such an ability — asking twice could
  // only produce an announcement the server refuses.
  let abilitySacrificeX: number | undefined;
  // #2527: and the count picked for "Discard X cards" (Gix), the same
  // way: the number of cards the player discards IS the announcement.
  let abilityDiscardX: number | undefined;
  // #1421: the count picked for "Tap X" is the announcement, just
  // as the sacrifice picker supplies X for "Sacrifice X".
  let abilityTapX: number | undefined;

  // S27: a Vehicle's crew cost. Its own prompt rather than a reuse of
  // the sacrifice picker because crew is a many-pick with a POWER
  // floor, not a single pick — see CrewCostModal.
  let crewPrompt = $state<{ card: CardView; ability: ActivatedAbilityView } | null>(null);

  const crewOptions = $derived.by(() => {
    const p = crewPrompt;
    if (!p) return [];
    const ids = new Set(p.ability.crew_options?.cards ?? []);
    return view.battlefield.cards.filter((c) => ids.has(c.instance_id));
  });

  // The X picker for an ability whose cost carries {X} (CR 602.2b).
  // It sits in the same place in the ability's announcement that it
  // sits in a cast's: after the cost picks that name cards, before
  // targeting, and locked once sent.
  let xAbilityPrompt = $state<{
    card: CardView;
    ability: ActivatedAbilityView;
    sacrificeIDs: string[];
    crewIDs: string[];
    counter?: CounterPayment;
  } | null>(null);

  // #625: a "remove N counters" cost. Asked after the sacrifice / crew
  // picks and before X and targeting — all of them announce-time cost
  // choices (CR 602.2b) — and skipped outright when the server offers
  // exactly one permanent with exactly one kind, which is Dragon's
  // Hoard, Mikaeus, and Heart of Kiran with a single planeswalker.
  let counterPrompt = $state<{
    card: CardView;
    ability: ActivatedAbilityView;
    sacrificeIDs: string[];
    crewIDs: string[];
  } | null>(null);

  // #789: the same question for a MANA ability — Mage-Ring Network's
  // "any number of storage counters", a Vivid land whose kind or
  // permanent is ambiguous. One modal and one payload builder; only
  // the action it ends in differs, because a mana ability has no
  // targeting or X step after the cost.
  let manaCounterPrompt = $state<{
    card: CardView;
    ability: ManaAbilityView;
    sacrificeIDs?: string[];
  } | null>(null);

  // #660: the "Discard a creature card" component of an activated
  // ability's cost (CR 602.2b), and cycling's "Discard this card",
  // which needs no picker at all. Carried on the side rather than
  // threaded through every hop of the announce chain, which already
  // takes five arguments.
  let abilityDiscardPrompt = $state<{
    card: CardView;
    ability: ActivatedAbilityView;
  } | null>(null);
  let abilityDiscardIDs: string[] = [];

  // #2598: the "Reveal X black cards from your hand" component of an
  // activated ability's cost (Martyr of Bones), asked first — before the
  // discard pick — because revealing is the cost nothing else depends on.
  // The cards picked are the announced X for the X form. Carried on the
  // side for the reason the discard picks are.
  let abilityRevealPrompt = $state<{
    card: CardView;
    ability: ActivatedAbilityView;
  } | null>(null);
  let abilityRevealIDs: string[] = [];
  let abilityRevealX: number | undefined;

  const abilityRevealOptions = $derived.by(() => {
    const p = abilityRevealPrompt;
    if (!p || !viewerID) return [];
    const ids = new Set(p.ability.reveal_cost_options ?? []);
    const me = view.seats.find((s) => s.id === viewerID);
    return (me?.hand.cards ?? []).filter((c) => ids.has(c.instance_id));
  });

  // The cards the clause admits, resolved out of the seat's hand. The
  // server already filtered them — a cost does not target, so nothing
  // narrows them further here.
  const abilityDiscardOptions = $derived.by(() => {
    const p = abilityDiscardPrompt;
    if (!p || !viewerID) return [];
    const ids = new Set(p.ability.discard_cost_options ?? []);
    const me = view.seats.find((s) => s.id === viewerID);
    return (me?.hand.cards ?? []).filter((c) => ids.has(c.instance_id));
  });

  // #1297: the "Exile N cards from your graveyard / hand" component of
  // an activated ability's cost (Grim Lavamancer, Holistic Wisdom),
  // asked right after the discard — the order the engine validates in.
  // Carried on the side for the reason the discard picks are.
  let abilityExilePrompt = $state<{
    card: CardView;
    ability: ActivatedAbilityView;
  } | null>(null);
  let abilityExileIDs: string[] = [];

  // The cards the clause admits, resolved out of whichever pile the
  // server says they are in.
  const abilityExileOptions = $derived.by(() => {
    const p = abilityExilePrompt;
    if (!p || !viewerID) return [];
    return exileCostOptionCards(
      view.seats.find((s) => s.id === viewerID),
      p.ability,
    );
  });

  // ADR 0109 §7 (#1902): "Put a card from your hand on top of your
  // library" (Penance, Leashling), asked after the exile pick — the
  // order the engine validates in — with the discard picker. Carried on
  // the side for the reason the discard picks are.
  let abilityTopPrompt = $state<{
    card: CardView;
    ability: ActivatedAbilityView;
  } | null>(null);
  let abilityTopIDs: string[] = [];

  const abilityTopOptions = $derived.by(() => {
    const p = abilityTopPrompt;
    if (!p || !viewerID) return [];
    const ids = new Set(p.ability.top_cost_options ?? []);
    const me = view.seats.find((s) => s.id === viewerID);
    return (me?.hand.cards ?? []).filter((c) => ids.has(c.instance_id));
  });

  // ADR 0109 §7 and owner decision 3: the costs with nothing to pick —
  // "Discard a card at random" and "Exile the top N cards of your
  // library" — are confirmed once the card picks are made, so the
  // confirm can count the hand those picks leave.
  let abilityCostConfirmPrompt = $state<{
    card: CardView;
    ability: ActivatedAbilityView;
  } | null>(null);

  const abilityCostConfirmLines = $derived.by(() => {
    const p = abilityCostConfirmPrompt;
    if (!p || !viewerID) return [];
    const me = view.seats.find((s) => s.id === viewerID);
    const hand = randomDiscardPool(me, p.card.instance_id, [
      ...abilityDiscardIDs,
      ...abilityExileIDs,
      ...abilityTopIDs,
    ]);
    return costConfirmLines(p.ability, hand, me?.library.count ?? 0);
  });

  // #660: a card in hand projects its abilities on `zone_abilities`
  // and a permanent on `activated_abilities` — never both, because
  // the server filters by the zone the card is in (CR 113.6). One
  // lookup reads whichever is there; `index` means the same thing on
  // the wire either way.
  function abilitiesOf(card: CardView): ActivatedAbilityView[] {
    return card.activated_abilities ?? card.zone_abilities ?? [];
  }

  // #1221: the CR 307.1 window, for the zone browser's ability rows.
  // Every keyword that functions from a graveyard prints "only as a
  // sorcery", so without this a browsed unearth row would be
  // clickable during an opponent's combat and come back refused —
  // the failure mode ADR 0033 §1 cites. PlayerPanel derives the same
  // string for the battlefield and the hand; the browser is a
  // top-level modal with no panel above it, so Board derives it here.
  const browsedZoneSorcerySpeedBlocked = $derived(
    canActivateSorcerySpeedAbility(view, viewerID).reason ?? "",
  );

  function handleActivateAbility(card: CardView, index: number): void {
    const ability = abilitiesOf(card).find((a) => a.index === index);
    if (!ability) return;
    // #1310: a fresh announcement asks its own waterbend question; a
    // pick left over from one the player backed out of must not ride
    // along with this one.
    abilityWaterbendIDs = undefined;
    abilitySacrificeX = undefined;
    abilityDiscardX = undefined;
    abilityTapX = undefined;
    abilityRevealIDs = [];
    abilityRevealX = undefined;
    // #2598: "Reveal X black cards from your hand" is asked before the
    // rest. The X form always asks — how many is the question — and an
    // empty set of matching cards pays it at X=0 with nothing to pick; a
    // fixed count skips the picker when the hand holds exactly that many.
    if (ability.reveal_cost_count_from_x) {
      const options = ability.reveal_cost_options ?? [];
      if (options.length > 0) {
        abilityRevealPrompt = { card, ability };
        return;
      }
      afterAbilityRevealCost(card, ability, []);
      return;
    }
    if (ability.reveal_cost_n) {
      const options = ability.reveal_cost_options ?? [];
      if (options.length > ability.reveal_cost_n) {
        abilityRevealPrompt = { card, ability };
        return;
      }
      afterAbilityRevealCost(card, ability, options);
      return;
    }
    askAbilityDiscardCost(card, ability);
  }

  function confirmAbilityRevealCost(ids: string[]): void {
    const p = abilityRevealPrompt;
    abilityRevealPrompt = null;
    if (!p) return;
    afterAbilityRevealCost(p.card, p.ability, ids);
  }

  // afterAbilityRevealCost records the reveal picks and goes on to the
  // discard question. For the X form the number revealed IS the
  // announced X, so the X stepper has nothing left to ask.
  function afterAbilityRevealCost(
    card: CardView,
    ability: ActivatedAbilityView,
    revealIDs: string[],
  ): void {
    abilityRevealIDs = revealIDs;
    abilityRevealX = ability.reveal_cost_count_from_x ? revealIDs.length : undefined;
    askAbilityDiscardCost(card, ability);
  }

  // askAbilityDiscardCost is the head of the card-cost chain.
  function askAbilityDiscardCost(card: CardView, ability: ActivatedAbilityView): void {
    // #660: the discard payment is asked FIRST, as the cast flow asks
    // its own — it is the cost most likely to make a player back out.
    // Skipped when the hand holds exactly the cards the clause
    // demands: a modal with one possible answer is a worse version of
    // no modal.
    // ADR 0109 §7: a random discard names nothing — it is confirmed
    // below, with the library exile, rather than picked here.
    // #2527: "Discard X cards" always asks — how many is the question,
    // so there is no hand size at which the answer is forced. An empty
    // hand pays it at X=0 with nothing to pick.
    if (ability.discard_cost_count_from_x) {
      const options = ability.discard_cost_options ?? [];
      if (options.length > 0) {
        abilityDiscardPrompt = { card, ability };
        return;
      }
      afterAbilityDiscardCost(card, ability, []);
      return;
    }
    if (ability.discard_cost_n && !ability.discard_cost_random) {
      const options = ability.discard_cost_options ?? [];
      if (options.length > ability.discard_cost_n) {
        abilityDiscardPrompt = { card, ability };
        return;
      }
      afterAbilityDiscardCost(card, ability, options);
      return;
    }
    afterAbilityDiscardCost(card, ability, []);
  }

  // afterAbilityDiscardCost is the rest of the announce chain with the
  // discard picks in hand: the sacrifice picker, the crew picker, the
  // counter cost, then X and targeting.
  function afterAbilityDiscardCost(
    card: CardView,
    ability: ActivatedAbilityView,
    discardIDs: string[],
  ): void {
    abilityDiscardIDs = discardIDs;
    // #2527: the cards picked for "Discard X cards" are the announced
    // X, so the X stepper has nothing left to ask.
    abilityDiscardX = ability.discard_cost_count_from_x ? discardIDs.length : undefined;
    // #2190: and the card picked for "Discard a card with mana value X"
    // IS the announcement: its mana value is X, for the target clause
    // that follows as much as for the engine.
    if (ability.discard_cost_mana_value_x) {
      abilityDiscardX = discardedManaValue(
        view.seats.find((s) => s.id === viewerID)?.hand.cards,
        discardIDs,
      );
    }
    // #1297: the exile pick next — the same card-shaped question one
    // component over, skipped the same way when the pile holds exactly
    // what the clause demands.
    if (ability.exile_cost_n) {
      const options = ability.exile_cost_options ?? [];
      if (options.length > ability.exile_cost_n) {
        abilityExilePrompt = { card, ability };
        return;
      }
      afterAbilityExileCost(card, ability, options);
      return;
    }
    afterAbilityExileCost(card, ability, []);
  }

  function confirmAbilityExileCost(ids: string[]): void {
    const p = abilityExilePrompt;
    abilityExilePrompt = null;
    if (!p) return;
    afterAbilityExileCost(p.card, p.ability, ids);
  }

  // afterAbilityExileCost is the rest of the announce chain once the
  // exile pick is made: the put-on-top pick, the confirm for the costs
  // with nothing to pick (ADR 0109 §7), then afterAbilityCardCosts.
  function afterAbilityExileCost(
    card: CardView,
    ability: ActivatedAbilityView,
    exileIDs: string[],
  ): void {
    abilityExileIDs = exileIDs;
    // ADR 0109 §7: the card put on top of the library, skipped the same
    // way when the hand holds exactly what the clause demands.
    if (ability.top_cost_n) {
      const options = ability.top_cost_options ?? [];
      if (options.length > ability.top_cost_n) {
        abilityTopPrompt = { card, ability };
        return;
      }
      afterAbilityTopCost(card, ability, options);
      return;
    }
    afterAbilityTopCost(card, ability, []);
  }

  function confirmAbilityTopCost(ids: string[]): void {
    const p = abilityTopPrompt;
    abilityTopPrompt = null;
    if (!p) return;
    afterAbilityTopCost(p.card, p.ability, ids);
  }

  // afterAbilityTopCost confirms the costs with nothing to pick, then
  // goes on to the permanent-naming picks.
  function afterAbilityTopCost(
    card: CardView,
    ability: ActivatedAbilityView,
    topIDs: string[],
  ): void {
    abilityTopIDs = topIDs;
    if (needsCostConfirm(ability)) {
      abilityCostConfirmPrompt = { card, ability };
      return;
    }
    afterAbilityCardCosts(card, ability);
  }

  function confirmAbilityCostConfirm(): void {
    const p = abilityCostConfirmPrompt;
    abilityCostConfirmPrompt = null;
    if (!p) return;
    afterAbilityCardCosts(p.card, p.ability);
  }

  // afterAbilityCardCosts is the rest of the announce chain once every
  // card-shaped cost is answered: the return, tap, sacrifice and crew
  // pickers, the counter cost, then X and targeting.
  function afterAbilityCardCosts(card: CardView, ability: ActivatedAbilityView): void {
    // #1213: the return-to-hand pick, in the same place the sacrifice
    // pick sits — both are announce-time cost choices (CR 602.2b) and
    // both name permanents. Skipped when the board offers exactly the
    // permanents the clause demands, for the reason the discard
    // picker is: a modal with one possible answer is a worse version
    // of no modal.
    if (ability.return_options) {
      const options = ability.return_options.cards ?? [];
      const need = ability.return_options.max ?? ability.return_options.min ?? 1;
      if (options.length > need) {
        abilityReturnPrompt = { card, ability };
        return;
      }
      abilityReturnIDs = options;
    } else {
      abilityReturnIDs = [];
    }
    askAbilityExilePermanentCost(card, ability);
  }

  // #1600: the exile-a-permanent pick, after the return pick and for
  // the same reason, skipped the same way when the board offers exactly
  // the permanents the clause demands.
  function askAbilityExilePermanentCost(card: CardView, ability: ActivatedAbilityView): void {
    if (ability.exile_permanent_options) {
      const options = ability.exile_permanent_options.cards ?? [];
      const need = ability.exile_permanent_options.max ?? ability.exile_permanent_options.min ?? 1;
      if (options.length > need) {
        abilityExilePermanentPrompt = { card, ability };
        return;
      }
      abilityExilePermanentIDs = options;
    } else {
      abilityExilePermanentIDs = [];
    }
    askAbilityTapCost(card, ability);
  }

  function confirmAbilityExilePermanentCost(ids: string[]): void {
    const p = abilityExilePermanentPrompt;
    abilityExilePermanentPrompt = null;
    if (!p) return;
    abilityExilePermanentIDs = ids;
    askAbilityTapCost(p.card, p.ability);
  }

  // #759: the tap-another pick, then the rest of the chain.
  function askAbilityTapCost(card: CardView, ability: ActivatedAbilityView): void {
    if (ability.tap_others_options) {
      const options = ability.tap_others_options.cards ?? [];
      const need = ability.tap_others_options.max ?? ability.tap_others_options.min ?? 1;
      if (ability.tap_others_options.count_from_x || options.length > need) {
        abilityTapPrompt = { card, ability };
        return;
      }
      abilityTapIDs = options;
    } else {
      abilityTapIDs = [];
    }
    afterAbilityPermanentPicks(card, ability);
  }

  // The chain once the permanent-naming picks (return, tap) are made:
  // the sacrifice picker, the crew picker, the counter cost.
  function afterAbilityPermanentPicks(card: CardView, ability: ActivatedAbilityView): void {
    if (ability.sacrifice_options) {
      sacrificePrompt = { kind: "ability", card, ability };
      return;
    }
    if (ability.crew_cost) {
      crewPrompt = { card, ability };
      return;
    }
    askCounterCost(card, ability, [], []);
  }

  function confirmAbilityTapCost(ids: string[]): void {
    const p = abilityTapPrompt;
    abilityTapPrompt = null;
    if (!p) return;
    abilityTapIDs = ids;
    abilityTapX = p.ability.tap_others_options?.count_from_x ? ids.length : undefined;
    afterAbilityPermanentPicks(p.card, p.ability);
  }

  function confirmAbilityDiscardCost(ids: string[]): void {
    const p = abilityDiscardPrompt;
    abilityDiscardPrompt = null;
    if (!p) return;
    afterAbilityDiscardCost(p.card, p.ability, ids);
  }

  // #1213: the return pick answered. The discard picks are already on
  // the side, so the chain resumes at the sacrifice picker.
  function confirmAbilityReturnCost(ids: string[]): void {
    const p = abilityReturnPrompt;
    abilityReturnPrompt = null;
    if (!p) return;
    abilityReturnIDs = ids;
    askAbilityExilePermanentCost(p.card, p.ability);
  }

  function askCounterCost(
    card: CardView,
    ability: ActivatedAbilityView,
    sacrificeIDs: string[],
    crewIDs: string[],
  ): void {
    if (!hasCounterCost(ability)) {
      continueActivation(card, ability, sacrificeIDs, crewIDs);
      return;
    }
    const auto = autoCounterChoice(ability);
    if (auto) {
      continueActivation(
        card,
        ability,
        sacrificeIDs,
        crewIDs,
        undefined,
        counterPaymentParams(ability, auto),
      );
      return;
    }
    counterPrompt = { card, ability, sacrificeIDs, crewIDs };
  }

  function confirmCounterCost(choices: CounterChoice[]): void {
    const p = counterPrompt;
    const m = manaCounterPrompt;
    counterPrompt = null;
    manaCounterPrompt = null;
    // #789: the same prompt serves a mana ability's counter cost —
    // one component, one picker — so the confirm routes to whichever
    // activation opened it.
    if (m) {
      sendManaAbility(m.card, m.ability, counterPaymentParams(m.ability, choices), m.sacrificeIDs);
      return;
    }
    if (!p) return;
    continueActivation(
      p.card,
      p.ability,
      p.sacrificeIDs,
      p.crewIDs,
      undefined,
      counterPaymentParams(p.ability, choices),
    );
  }

  function confirmCrew(instanceIDs: string[]): void {
    const p = crewPrompt;
    crewPrompt = null;
    if (!p) return;
    askCounterCost(p.card, p.ability, [], instanceIDs);
  }

  function confirmAbilityX(x: number): void {
    const p = xAbilityPrompt;
    xAbilityPrompt = null;
    if (!p) return;
    continueActivation(p.card, p.ability, p.sacrificeIDs, p.crewIDs, x, p.counter);
  }

  // S21, widened in #789: a mana ability whose cost still needs an
  // answer, handed up by PlayerPanel because the pickers are
  // board-wide. Sacrifice first, then counters — the order the engine
  // validates and pays them in.
  function handleManaAbilityCost(
    card: CardView,
    ability: ManaAbilityView,
    colors: string[] = [],
  ): void {
    // #1443: the colour the anchored picker already has, held for the
    // one action at the end of the chain. Set at the chain's start, so
    // a chain that was cancelled cannot leak its colour into the next.
    manaColors = colors;
    // #1213: the discard pick first, as the CR 602 chain asks its own
    // — it is the cost most likely to make a player back out — and
    // skipped when the hand holds exactly what the clause demands.
    if (ability.discard_cost_n) {
      const options = ability.discard_cost_options ?? [];
      if (options.length > ability.discard_cost_n) {
        manaDiscardPrompt = { card, ability };
        return;
      }
      manaDiscardIDs = options;
    } else {
      manaDiscardIDs = [];
    }
    askManaExileCost(card, ability);
  }

  // #1283: the exile-a-card pick (Cadaverous Bloom), after the discard
  // and before the sacrifice — the order the engine validates in —
  // and skipped the same way when the hand holds exactly what the
  // clause demands. The same modal as the discard, with the verb
  // changed, because the question is the same one; the answer rides
  // its own field, `exile_ids`, because the component is not.
  function askManaExileCost(card: CardView, ability: ManaAbilityView): void {
    if (ability.exile_cost_n) {
      const options = ability.exile_cost_options ?? [];
      if (options.length > ability.exile_cost_n) {
        manaExilePrompt = { card, ability };
        return;
      }
      manaExileIDs = options;
    } else {
      manaExileIDs = [];
    }
    afterManaCardCosts(card, ability);
  }

  // The rest of the mana-ability chain once the card-shaped costs are
  // answered: exile-a-permanent, tap-another, sacrifice, then the
  // counter cost.
  function afterManaCardCosts(card: CardView, ability: ManaAbilityView): void {
    // #1600: Food Chain's "Exile a creature you control", skipped when
    // the board offers exactly the permanents the clause demands.
    if (ability.exile_permanent_options) {
      const options = ability.exile_permanent_options.cards ?? [];
      const need = ability.exile_permanent_options.max ?? ability.exile_permanent_options.min ?? 1;
      if (options.length > need) {
        manaExilePermanentPrompt = { card, ability };
        return;
      }
      manaExilePermanentIDs = options;
    } else {
      manaExilePermanentIDs = [];
    }
    askManaTapCost(card, ability);
  }

  function confirmManaExilePermanentCost(ids: string[]): void {
    const p = manaExilePermanentPrompt;
    manaExilePermanentPrompt = null;
    if (!p) return;
    manaExilePermanentIDs = ids;
    askManaTapCost(p.card, p.ability);
  }

  function askManaTapCost(card: CardView, ability: ManaAbilityView): void {
    if (ability.tap_others_options) {
      const options = ability.tap_others_options.cards ?? [];
      const need = ability.tap_others_options.max ?? ability.tap_others_options.min ?? 1;
      if (options.length > need) {
        manaTapPrompt = { card, ability };
        return;
      }
      manaTapIDs = options;
    } else {
      manaTapIDs = [];
    }
    afterManaTapCost(card, ability);
  }

  function confirmManaTapCost(ids: string[]): void {
    const p = manaTapPrompt;
    manaTapPrompt = null;
    if (!p) return;
    manaTapIDs = ids;
    afterManaTapCost(p.card, p.ability);
  }

  function afterManaTapCost(card: CardView, ability: ManaAbilityView): void {
    if (ability.sacrifice_options) {
      sacrificePrompt = { kind: "mana", card, ability };
      return;
    }
    askManaCounterCost(card, ability);
  }

  let manaExilePrompt = $state<{
    card: CardView;
    ability: ManaAbilityView;
  } | null>(null);
  let manaExileIDs: string[] = [];

  const manaExileOptions = $derived.by(() => {
    const p = manaExilePrompt;
    if (!p || !viewerID) return [];
    // #1297: out of whichever pile the clause reads, as the CR 602
    // owner's picker does.
    return exileCostOptionCards(
      view.seats.find((s) => s.id === viewerID),
      p.ability,
    );
  });

  function confirmManaExileCost(ids: string[]): void {
    const p = manaExilePrompt;
    manaExilePrompt = null;
    if (!p) return;
    manaExileIDs = ids;
    afterManaCardCosts(p.card, p.ability);
  }

  // #1213: the mana-ability half of #660's discard picker. The same
  // modal, the same wire field, a different action to end in —
  // Skirge Familiar's "Discard a card: Add {B}".
  let manaDiscardPrompt = $state<{
    card: CardView;
    ability: ManaAbilityView;
  } | null>(null);
  let manaDiscardIDs: string[] = [];
  // #1443: the colours named at the card for the mana ability whose
  // cost chain is open (see handleManaAbilityCost).
  let manaColors: string[] = [];

  const manaDiscardOptions = $derived.by(() => {
    const p = manaDiscardPrompt;
    if (!p || !viewerID) return [];
    const ids = new Set(p.ability.discard_cost_options ?? []);
    const me = view.seats.find((s) => s.id === viewerID);
    return (me?.hand.cards ?? []).filter((c) => ids.has(c.instance_id));
  });

  function confirmManaDiscardCost(ids: string[]): void {
    const p = manaDiscardPrompt;
    manaDiscardPrompt = null;
    if (!p) return;
    manaDiscardIDs = ids;
    askManaExileCost(p.card, p.ability);
  }

  // askManaCounterCost is askCounterCost's mana-ability twin: the same
  // component, the same picker, the same auto-skip when there is only
  // one way to pay (every Vivid land, Ramos). It exists separately
  // only because a mana ability has no targeting or X step after the
  // cost — the action goes straight out.
  function askManaCounterCost(card: CardView, ability: ManaAbilityView): void {
    if (!hasCounterCost(ability)) {
      sendManaAbility(card, ability, {});
      return;
    }
    const auto = autoCounterChoice(ability);
    if (auto) {
      sendManaAbility(card, ability, counterPaymentParams(ability, auto));
      return;
    }
    manaCounterPrompt = { card, ability };
  }

  function sendManaAbility(
    card: CardView,
    ability: ManaAbilityView,
    counter: CounterPayment,
    sacrificeIDs?: string[],
  ): void {
    guardedSendAction(
      "activate_mana_ability",
      {
        card_id: card.instance_id,
        ability_index: ability.index,
        // ADR 0093: the row this click meant, so a stale one is refused.
        ...manaAbilityRef(card, ability.index),
        ...(sacrificeIDs && sacrificeIDs.length > 0 ? { sacrifice_ids: sacrificeIDs } : {}),
        // #758: absent on ordinary mana abilities, as every optional
        // cost-payment field is.
        ...manaTapPayment(manaTapIDs),
        // #1213: omitted when empty, so every payload a client sent
        // before this field existed is byte-for-byte unchanged.
        ...(manaDiscardIDs.length > 0 ? { discard_ids: manaDiscardIDs } : {}),
        // #1283: the same posture — absent unless the ability exiles.
        ...(manaExileIDs.length > 0 ? { exile_ids: manaExileIDs } : {}),
        // #1600: and the permanents an exile-a-permanent cost names.
        ...manaExilePermanentPayment(manaExilePermanentIDs),
        ...counter,
        // #1443: absent unless the picker named a colour.
        ...manaColorParams(manaColors),
      },
      viewerID ?? undefined,
    );
    resetManaCostPayment();
  }

  function resetManaCostPayment(): void {
    manaDiscardIDs = [];
    manaExileIDs = [];
    manaExilePermanentIDs = [];
    manaTapIDs = [];
    manaColors = [];
  }

  // #170: an ability row picked from the admin context menu. Same two
  // destinations PlayerPanel routes to — the sacrifice picker when the
  // cost needs one, otherwise straight to the action / targeting flow.
  function handleMenuActivate(card: CardView, activate: MenuActivate): void {
    if (activate.kind === "ability") {
      handleActivateAbility(card, activate.index);
      return;
    }
    const ability = (card.mana_abilities ?? []).find((a) => a.index === activate.index);
    if (ability && manaAbilityNeedsPrompt(ability)) {
      handleManaAbilityCost(card, ability, activate.colors);
      return;
    }
    // #1443: the anchored picker's colour rides the activation.
    const params = {
      card_id: card.instance_id,
      ability_index: activate.index,
      ...manaAbilityRef(card, activate.index),
      ...manaColorParams(activate.colors),
    };
    guardedSendAction("activate_mana_ability", params, card.controller);
  }

  function confirmSacrifice(instanceIDs: string[]): void {
    const p = sacrificePrompt;
    sacrificePrompt = null;
    if (!p) return;
    if (p.kind === "mana") {
      // #789: a mana ability could in principle carry both halves, so
      // the counter question is asked after the sacrifice one — the
      // order the engine validates them in.
      if (counterCostNeedsPrompt(p.ability)) {
        manaCounterPrompt = { card: p.card, ability: p.ability, sacrificeIDs: instanceIDs };
        return;
      }
      const auto = autoCounterChoice(p.ability);
      sendManaAbility(
        p.card,
        p.ability,
        counterPaymentParams(p.ability, auto ?? undefined),
        instanceIDs,
      );
      return;
    }
    // #1213: "Sacrifice X Treasures" announces its count AS the X
    // (CR 602.2b), so the payment the player just made is the
    // announcement and the X stepper has nothing left to ask.
    abilitySacrificeX = p.ability.sacrifice_options?.count_from_x ? instanceIDs.length : undefined;
    askCounterCost(p.card, p.ability, instanceIDs, []);
  }

  // continueActivation is the post-cost half: enter targeting for an
  // ability that targets, or fire straight away.
  function continueActivation(
    card: CardView,
    ability: ActivatedAbilityView,
    sacrificeIDs: string[],
    crewIDs: string[] = [],
    xValue?: number,
    counter?: CounterPayment,
    modes?: number[],
    phyrexianLife?: number,
  ): void {
    // CR 602.2b: X is announced with the other choices and before
    // any cost is paid, so the picker opens after the cost picks
    // that name cards and before the targeting step — the same
    // position it holds in a cast's prompt chain.
    if (ability.demands_x && xValue === undefined) {
      // #1213: unless the sacrifice payment already answered it.
      if (abilitySacrificeX !== undefined) {
        xValue = abilitySacrificeX;
      } else if (abilityTapX !== undefined) {
        xValue = abilityTapX;
      } else if (abilityDiscardX !== undefined) {
        xValue = abilityDiscardX;
      } else if (abilityRevealX !== undefined) {
        xValue = abilityRevealX;
      } else {
        xAbilityPrompt = { card, ability, sacrificeIDs, crewIDs, counter };
        return;
      }
    }
    // CR 107.4f / CR 602.2b (#917, #916): the ability's Phyrexian
    // symbols. Same question the cast chain asks, in the same place —
    // after X, before the modes and the targets — and the same
    // stepper asks it, told to price the ABILITY's cost.
    //
    // ADR 0131: the same rule as the cast's. A printed Phyrexian symbol
    // always asks; when every symbol is a granted one (K'rrik's {B}) it
    // asks only if the preview, with nothing claimed, says mana is short.
    if (phyrexianLife === undefined) {
      const symbols = phyrexianSymbolsForAbility(ability);
      const granted = phyrexianGrantedForAbility(ability);
      if (onlyGrantedSymbols(symbols, granted)) {
        if (
          !$settings.gameplay.strictMana ||
          !shouldAskPhyrexianLife(symbols, viewerLife, granted, true)
        ) {
          // Nothing to ask: strict payment is off, or CR 119.4 leaves no
          // symbol to buy. Continue claiming nothing.
          continueActivation(card, ability, sacrificeIDs, crewIDs, xValue, counter, modes, 0);
          return;
        }
        fetchAutoTapPreview(view.id, card.instance_id, { xValue, abilityIndex: ability.index })
          .then((p) => {
            if (p.ok) {
              continueActivation(card, ability, sacrificeIDs, crewIDs, xValue, counter, modes, 0);
            } else {
              phyrexianAbilityPrompt = {
                card,
                ability,
                sacrificeIDs,
                crewIDs,
                xValue,
                counter,
                modes,
                suggest: true,
              };
            }
          })
          .catch(() =>
            continueActivation(card, ability, sacrificeIDs, crewIDs, xValue, counter, modes, 0),
          );
        return;
      }
      if (shouldAskPhyrexianLife(symbols, viewerLife, granted, true)) {
        phyrexianAbilityPrompt = {
          card,
          ability,
          sacrificeIDs,
          crewIDs,
          xValue,
          counter,
          modes,
          suggest: false,
        };
        return;
      }
    }
    // #1310, CR 701.67a: "Waterbend {N}:" — which untapped artifacts
    // and creatures pay part of it. After X, because a Waterbend {X}
    // has no size until X is announced; before the modes and targets,
    // because it is a cost and the cast chain asks its own convoke /
    // waterbend taps in the same place. Skipped when nothing could
    // help: tapping none pays the whole cost with mana.
    if (abilityWaterbendIDs === undefined && shouldAskAbilityWaterbend(ability, xValue)) {
      abilityWaterbendPrompt = {
        card,
        ability,
        sacrificeIDs,
        crewIDs,
        xValue,
        counter,
        phyrexianLife,
      };
      return;
    }
    // #764, CR 602.2b: a modal activated ability chooses its modes
    // with its targets, in the one announcement — so the mode picker
    // sits exactly where a modal cast's does, between the costs and
    // the targeting walk.
    if (ability.modes && modes === undefined) {
      abilityModePrompt = { card, ability, sacrificeIDs, crewIDs, xValue, counter, phyrexianLife };
      return;
    }
    if (ability.legal_targets || ability.clauses?.length || (modes && modes.length > 0)) {
      beginTargetingForAbility(
        card,
        ability,
        sacrificeIDs,
        crewIDs,
        xValue,
        counter,
        modes,
        phyrexianLife,
      );
      const t = $targeting;
      if (t && t.steps.length > 0 && (t.legal || t.steps.length > 1)) return;
      // A modal ability whose chosen bullets take no target falls
      // through to the immediate activation below.
      cancelTargeting();
    }
    const params: Record<string, unknown> = {
      source_card_id: card.instance_id,
      ability_index: ability.index,
      // ADR 0093: the row the activation meant, so a stale one is refused.
      ...activatedAbilityRef(card, ability.index),
      sacrifice_ids: sacrificeIDs,
      crew_ids: crewIDs,
      discard_ids: abilityDiscardIDs,
      // #1213: the return-to-hand picks, made at announce with the
      // rest of the cost and sent in the one activate_ability.
      return_ids: abilityReturnIDs,
      // #759: the station creature, the same way.
      tap_ids: abilityTapIDs,
      ...counter,
    };
    if (xValue !== undefined) params.x_value = xValue;
    // #916: omitted at 0, which is the server default.
    if (phyrexianLife) params.phyrexian_life = phyrexianLife;
    if (modes !== undefined) params.modes = modes;
    // #1310: the waterbend taps, omitted when there are none.
    if (abilityWaterbendIDs && abilityWaterbendIDs.length > 0) {
      params.waterbend_ids = abilityWaterbendIDs;
    }
    // #1297: the exile picks, on their own field, omitted when none.
    if (abilityExileIDs.length > 0) params.exile_ids = abilityExileIDs;
    // #1600: the exiled permanents, likewise.
    if (abilityExilePermanentIDs.length > 0) {
      params.exile_permanent_ids = abilityExilePermanentIDs;
    }
    // ADR 0109 §7: the card put on top of the library, likewise.
    if (abilityTopIDs.length > 0) params.top_ids = abilityTopIDs;
    // #2598: the cards revealed to pay "Reveal X black cards".
    if (abilityRevealIDs.length > 0) params.reveal_ids = abilityRevealIDs;
    abilityWaterbendIDs = undefined;
    abilityRevealIDs = [];
    abilityRevealX = undefined;
    abilityDiscardIDs = [];
    abilityExileIDs = [];
    abilityTopIDs = [];
    abilityReturnIDs = [];
    abilityExilePermanentIDs = [];
    abilityTapIDs = [];
    abilitySacrificeX = undefined;
    abilityDiscardX = undefined;
    abilityTapX = undefined;
    guardedSendAction("activate_ability", params, viewerID ?? undefined);
  }

  // #916: the activation's Phyrexian stepper — the same modal the
  // cast chain opens, priced against the ability's own mana cost.
  let phyrexianAbilityPrompt = $state<{
    card: CardView;
    ability: ActivatedAbilityView;
    sacrificeIDs: string[];
    crewIDs: string[];
    xValue?: number;
    counter?: CounterPayment;
    modes?: number[];
    // ADR 0131: open at the smallest count that makes the mana payable.
    suggest: boolean;
  } | null>(null);

  function confirmAbilityPhyrexianLife(n: number): void {
    const p = phyrexianAbilityPrompt;
    phyrexianAbilityPrompt = null;
    if (!p) return;
    continueActivation(
      p.card,
      p.ability,
      p.sacrificeIDs,
      p.crewIDs,
      p.xValue,
      p.counter,
      p.modes,
      n,
    );
  }

  // #1310: the waterbend picker for an activated ability — the same
  // TapCostModal a spell's convoke / waterbend opens, fed the
  // ability's `waterbend` clause. The answer is held on the side
  // (like the discard and return picks) until the one
  // activate_ability goes out; undefined means "not asked yet".
  let abilityWaterbendPrompt = $state<{
    card: CardView;
    ability: ActivatedAbilityView;
    sacrificeIDs: string[];
    crewIDs: string[];
    xValue?: number;
    counter?: CounterPayment;
    phyrexianLife?: number;
  } | null>(null);
  let abilityWaterbendIDs: string[] | undefined = undefined;

  const abilityWaterbendOptions = $derived.by(() => {
    const p = abilityWaterbendPrompt;
    if (!p?.ability.waterbend) return [];
    const ids = new Set(p.ability.waterbend.options?.cards ?? []);
    return view.battlefield.cards.filter((c) => ids.has(c.instance_id));
  });

  function confirmAbilityWaterbend(ids: string[]): void {
    const p = abilityWaterbendPrompt;
    abilityWaterbendPrompt = null;
    if (!p) return;
    abilityWaterbendIDs = ids;
    continueActivation(
      p.card,
      p.ability,
      p.sacrificeIDs,
      p.crewIDs,
      p.xValue,
      p.counter,
      undefined,
      p.phyrexianLife,
    );
  }

  // #764: the mode picker for a modal ACTIVATED ability. It reuses
  // ModePickerModal by handing it a synthetic card view carrying the
  // ability's ModeSpec — one picker, three owners, the same way the
  // engine has one ModeSpec for three owners.
  let abilityModePrompt = $state<{
    card: CardView;
    ability: ActivatedAbilityView;
    sacrificeIDs: string[];
    crewIDs: string[];
    xValue?: number;
    counter?: CounterPayment;
    // #916: already answered by the time the modes are picked, so it
    // rides through rather than being asked for again.
    phyrexianLife?: number;
  } | null>(null);
  const abilityModeCard = $derived(
    abilityModePrompt
      ? ({ ...abilityModePrompt.card, modes: abilityModePrompt.ability.modes } as CardView)
      : null,
  );
  function confirmAbilityModes(modes: number[]): void {
    const p = abilityModePrompt;
    abilityModePrompt = null;
    if (!p) return;
    continueActivation(
      p.card,
      p.ability,
      p.sacrificeIDs,
      p.crewIDs,
      p.xValue,
      p.counter,
      modes,
      p.phyrexianLife,
    );
  }

  // S20 sub-PR 2: a pick_target pending choice addressed to the
  // viewer drives the same targeting UI a cast does. Enter it when
  // one appears; leave it when it's gone (answered, or resolved
  // elsewhere). A cast prompt already in flight is replaced — the
  // trigger's target is owed first.
  $effect(() => {
    // S27: the legend rule and choose_protector are both answered
    // through this flow — all three are the same question shape (pick
    // one from a server-computed set) and the banner's confirm sends
    // the same payload. The server routes by the choice's kind, so
    // the client needs no second component and no second code path.
    // #1196: the CR 115.7 retarget is the fourth — same question
    // shape, same picker, and the server routes the answer on the
    // kind.
    const mine = (view.pending_choices ?? []).find(
      (c) => answeredOnBoard(c, view, $listFallback) && c.chooser === viewerID,
    );
    const cur = $targeting;
    if (mine) {
      if (cur?.choiceID === mine.id) return;
      const source = findCardAnywhere(mine.source) ?? {
        instance_id: mine.source ?? "",
        name:
          mine.kind === "legend_rule"
            ? "Legend rule"
            : mine.kind === "choose_protector"
              ? "Battle"
              : mine.kind === "retarget"
                ? "Retarget"
                : mine.kind === "untap_choice"
                  ? "Untap step"
                  : mine.kind === "choose_cards"
                    ? "Choose"
                    : "Triggered ability",
        owner: viewerID ?? "",
        controller: viewerID ?? "",
      };
      beginTargetingChoice(mine, source);
    } else if (cur?.choiceID) {
      targeting.set(null);
    }
  });

  function findCardAnywhere(id: string | undefined): CardView | undefined {
    if (!id) return undefined;
    for (const c of view.battlefield.cards) if (c.instance_id === id) return c;
    // #1196: a retarget prompt's source is the spell that is
    // resolving — still on the stack when the prompt opens.
    for (const c of view.stack.cards ?? []) if (c.instance_id === id) return c;
    for (const c of view.exile?.cards ?? []) if (c.instance_id === id) return c;
    for (const s of view.seats) {
      for (const c of s.graveyard?.cards ?? []) if (c.instance_id === id) return c;
      for (const c of s.hand?.cards ?? []) if (c.instance_id === id) return c;
    }
    return undefined;
  }

  function handleDrawCard(): void {
    if (!viewerID) return;
    guardedSendAction("draw_card", undefined, viewerID);
  }

  function handleTargetPlayer(targetPlayerID: string): void {
    const state = $targeting;
    if (!state || !isLegalPlayerTarget(state, targetPlayerID)) return;
    completeTargetedCast("player", targetPlayerID);
  }

  // handleTargetCard is the card-click intercept. Returns true only
  // when a cast-targeting prompt is waiting AND this card is a
  // legal target for that prompt; the caller stops default
  // processing (tap-toggle) in that case.
  function handleTargetCard(card: CardView): boolean {
    const state = $targeting;
    if (!state) return false;
    // S20: with a server legal set, membership decides; free-form
    // cards fall back to the mode heuristic (the caller only routes
    // battlefield / stack / graveyard clicks here).
    if (!isLegalCardTarget(state, card.instance_id)) return false;
    completeTargetedCast("card", card.instance_id);
    // A pick that completed the prompt (single target — the store is
    // cleared once the action fires) closes the zone browser so the
    // table is back in view. Multi-pick keeps it open for more picks.
    if ($targeting === null) closeZoneBrowser();
    return true;
  }

  // ---- Opponent summaries (ADR 0077) --------------------------------
  //
  // Which opponents render as SeatSummary read-outs and which as full
  // PlayerPanels. The decision itself lives in lib/expansion.ts and is
  // unit-tested; Board's job is only to gather the per-seat facts.
  // ADR 0077's in-place pin is gone (ADR 0120 §6): a viewer who wants a
  // seat's board opens it in the expanded overlay below, which does not
  // move the table at all.

  // Seats the viewer may declare an attack against. Computed once per
  // snapshot rather than per seat, because it reads the whole turn.
  const defenderIDs = $derived(legalDefenderIDs(view, viewerID));

  function renderingFor(seat: PlayerView, pos: SeatPosition | null): SeatDecision {
    const controlled = cardsByController.get(seat.id) ?? [];
    return decideSeatRendering(
      {
        isSelf: pos === "self",
        isActiveSeat: seat.id === activeSeatID,
        controlsLegalTarget: seatControlsLegalTarget($targeting, seat.id, controlled),
        hasAttackersOnViewer: seatHasAttackersOn(viewerID, controlled),
        isLegalDefender: defenderIDs.has(seat.id),
      },
      { spectator: isSpectator, combatMode },
      expansionSettingsFor($settings.display),
    );
  }

  // ---- The expanded board (ADR 0120) ----------------------------------
  //
  // One seat's board drawn again, larger, over the table. Hovering its
  // avatar peeks; clicking the avatar, its expand button or the pin pins.
  // Every transition is boardExpand.ts's reducer; Board owns the state,
  // runs the one timer the state asks for, and places the overlay.
  //
  // Per table and never persisted, like ADR 0077's pin before it: a board
  // expanded in one game is not expanded when the next one mounts.
  // `$state.raw` so `expandTimer` below keeps its identity across events
  // that do not touch it, and the timer is not restarted by them.
  let expandState = $state.raw(initialExpandState());
  function dispatchExpand(ev: ExpandEvent): void {
    expandState = reduceExpand(expandState, ev);
  }
  const expandTimer = $derived(expandState.timer);
  $effect(() => {
    const t = expandTimer;
    if (!t) return;
    const h = setTimeout(() => dispatchExpand({ type: "elapsed", seq: t.seq }), t.ms);
    return () => clearTimeout(h);
  });

  // A seat that left the table takes its overlay with it: a derivation,
  // not an effect (ADR 0120 §1), so the state never names a missing seat
  // for even one frame.
  const expanded = $derived(
    liveExpanded(
      expandState.expanded,
      view.seats.map((s) => s.id),
    ),
  );
  const expandedSeat = $derived(
    expanded ? (view.seats.find((s) => s.id === expanded.seatID) ?? null) : null,
  );

  // Phones (under 600px, the dock's own breakpoint) get the pin only, and
  // the overlay fills the board there (ADR 0120 §5).
  function onPhone(): boolean {
    return typeof matchMedia === "function" && matchMedia("(max-width: 599px)").matches;
  }
  // Without a hover-capable pointer nothing peeks (ADR 0120 §1). jsdom
  // has no matchMedia; that counts as a mouse, so a test can hover.
  function canPeek(): boolean {
    if (typeof matchMedia !== "function") return true;
    return matchMedia("(hover: hover)").matches && !onPhone();
  }

  // The avatar's reports (PlayerIdentity.svelte), through context so the
  // table's panels, the summaries and the overlay's own panel all reach
  // it without a prop on each.
  setAvatarExpand({
    enter(seatID, ev) {
      // A touch, a held button (a card dragged from the hand across the
      // avatars) or a pointer that cannot hover opens nothing.
      if (ev.pointerType === "touch" || ev.buttons !== 0 || !canPeek()) return;
      dispatchExpand({
        type: "avatar-enter",
        seatID,
        openDelayMs: peekOpenDelay($settings.display.hoverDelayMs),
      });
    },
    leave(seatID) {
      dispatchExpand({ type: "avatar-leave", seatID });
    },
    press(seatID) {
      dispatchExpand({ type: "avatar-press", seatID });
    },
    click(seatID) {
      returnFocusTo = null;
      dispatchExpand({ type: "avatar-click", seatID });
    },
  });

  // The expand buttons (a full panel's, a summary's ⤢ and its pip click)
  // open the overlay pinned and move focus into it; closing it puts focus
  // back on the button that opened it (ADR 0120 §4).
  let overlayEl = $state<HTMLElement | null>(null);
  let pinButtonEl = $state<HTMLButtonElement | null>(null);
  let returnFocusTo: HTMLElement | null = null;
  async function openExpanded(seatID: string, from: EventTarget | null): Promise<void> {
    returnFocusTo = from instanceof HTMLElement ? from : null;
    dispatchExpand({ type: "expand", seatID });
    await tick();
    pinButtonEl?.focus();
  }
  function restoreFocus(): void {
    const el = returnFocusTo;
    returnFocusTo = null;
    if (el && el.isConnected) el.focus();
  }
  function togglePin(): void {
    const wasPinned = expanded?.pinned === true;
    dispatchExpand({ type: "pin-toggle" });
    if (wasPinned) restoreFocus();
  }

  // Escape closes the overlay only when nothing else owns it (ADR 0120 §4,
  // ADR 0047): not a key something already took, not under a modal, not
  // while a card's popover, the mana picker or a card menu is open, not
  // while a target is being picked, and not when the dock's one key
  // handler would press something with it. That keeps cancelling a pick
  // and closing a menu ahead of closing the board.
  function onExpandKey(e: KeyboardEvent): void {
    if (e.key !== "Escape" || !expanded) return;
    if (e.defaultPrevented || e.isComposing) return;
    if ($foreignModalOpen) return;
    if ($abilityPopover || $manaSourcePicker || $cardMenu || $targeting) return;
    // A popup on the board that closes on its own Escape (the pile's
    // "+N more" list) says so with `data-escape-owner`.
    if (boardEl?.querySelector("[data-escape-owner]")) return;
    if (dockKeyFor(e, currentDockRequest(), { modalOpen: $foreignModalOpen })) return;
    e.preventDefault();
    dispatchExpand({ type: "escape" });
    restoreFocus();
  }

  // Where the overlay sits (ADR 0120 §2): the wider span beside the
  // avatar it belongs to, on the table. Null fills the board, which is
  // the phone layout and the fallback when the avatar cannot be measured.
  let overlaySpan = $state<Span | null>(null);
  function tableAvatar(seatID: string): HTMLElement | null {
    if (!boardEl) return null;
    for (const slot of boardEl.querySelectorAll<HTMLElement>(":scope > .slot")) {
      const el = slot.querySelector<HTMLElement>(seatSelector(seatID));
      if (el) return el;
    }
    return null;
  }
  function measureOverlay(): void {
    const e = expanded;
    if (!e || !boardEl || onPhone()) {
      overlaySpan = null;
      return;
    }
    const av = tableAvatar(e.seatID);
    const b = boardEl.getBoundingClientRect();
    const r = av?.getBoundingClientRect();
    if (!r || (r.width === 0 && r.height === 0) || b.width === 0) {
      overlaySpan = null;
      return;
    }
    const next = placeOverlay(
      { left: r.left - b.left, top: r.top - b.top, width: r.width, height: r.height },
      {
        boardWidth: b.width,
        // ADR 0119's pile publishes its width here; unset is 0.
        leftClear: parsePx(getComputedStyle(boardEl).getPropertyValue("--stack-pile-clear")),
        edge: 6,
      },
    );
    const cur = overlaySpan;
    if (!cur || cur.left !== next.left || cur.width !== next.width || cur.side !== next.side) {
      overlaySpan = next;
    }
  }
  $effect(() => {
    const seatID = expanded?.seatID;
    const board = boardEl;
    if (!seatID || !board) {
      overlaySpan = null;
      return;
    }
    untrack(measureOverlay);
    if (typeof ResizeObserver === "undefined") return;
    const ro = new ResizeObserver(() => measureOverlay());
    ro.observe(board);
    return () => ro.disconnect();
  });

  // The arrows and the fan lane re-measure when the overlay opens,
  // closes, changes seat, moves or resizes (ADR 0120 §3): its copy of a
  // card is the anchor while it is open, and none of that is a snapshot.
  $effect(() => {
    void expanded?.seatID;
    void overlaySpan;
    untrack(bumpBoardExpandLayout);
  });
  $effect(() => {
    const el = overlayEl;
    if (!el || typeof ResizeObserver === "undefined") return;
    const ro = new ResizeObserver(() => bumpBoardExpandLayout());
    ro.observe(el);
    return () => ro.disconnect();
  });

  // The overlay fades in when animations are on and motion is not
  // reduced; otherwise it appears at once. It always leaves at once, so
  // the anchors never prefer a copy that is on its way out. A CSS
  // animation on mount (`.animate`) rather than a Svelte transition, so
  // closing is never delayed by one.
  const overlayMotion = $derived(
    $settings.animations.enabled && !$settings.accessibility.reduceMotion,
  );

  // The overlay's name, and the expand and pin buttons' (ADR 0120 §3, §4).
  // A label contract: none of them contains "<name> board" or
  // "your board", which the e2e suite matches as substrings.
  function expandedRegionName(seat: PlayerView): string {
    return seat.id === viewerID ? "your own board, expanded" : `${seat.name}'s board, expanded`;
  }
  function expandButtonName(seat: PlayerView): string {
    return L.expandBoard(seat.id === viewerID ? null : seat.name);
  }
  function pinButtonName(seat: PlayerView): string {
    return seat.id === viewerID
      ? "Pin your own expanded board"
      : `Pin ${seat.name}'s expanded board`;
  }

  // The four quadrant positions. Iteration order doesn't matter for
  // CSS Grid (each panel sets its own grid-area), but kept stable
  // here so Svelte's keyed each-block reuses DOM across snapshots.
  const positions: SeatPosition[] = ["across_next", "across", "next", "self"];

  // boardEl is the anchor for absolute-positioned overlays (hover
  // zoom, stack, combat arrows). Bound here so children that need
  // board-relative coordinates (CombatArrows reads source/target
  // bounding rects against it) get the same node.
  let boardEl: HTMLDivElement | null = $state(null);

  // ---- The stack's display style (#1467) -----------------------------
  //
  // `compact` is the docked card in the attention strip. Anything else
  // mounts StackLaneHost, which floats over the middle of the table
  // while the stack or pending triggers are live (the host keeps its
  // aria-live announcer mounted in between). While the lane is showing
  // the stack, the docked card is not rendered.
  //
  // ADR 0119 §1: `pile`, the default, falls back to the docked card on
  // a phone and on a board too short for a 200px top card between the
  // strip and the bottom (lib/stackPile.ts). The stored setting is not
  // touched, so the pile comes back when the window grows.
  let pileFallback = $state(false);
  $effect(() => {
    const board = boardEl;
    if (!board) return;
    const mq = typeof matchMedia === "function" ? matchMedia("(max-width: 599px)") : null;
    const read = (): void => {
      pileFallback = pileFallsBack({
        phone: mq?.matches ?? false,
        boardH: board.getBoundingClientRect().height,
        stripBottom: stripContentBottom(board),
      });
    };
    read();
    mq?.addEventListener?.("change", read);
    const ro = typeof ResizeObserver === "undefined" ? null : new ResizeObserver(read);
    ro?.observe(board);
    const strip = attentionStrip(board);
    if (strip) ro?.observe(strip);
    return () => {
      mq?.removeEventListener?.("change", read);
      ro?.disconnect();
    };
  });
  const floatingStackStyle = $derived.by((): StackLaneStyle | null => {
    const stored = $settings.display.stackStyle;
    const s = isStackStyle(stored) ? stored : DEFAULT_STACK_STYLE;
    if (s === "compact") return null;
    if (s === "pile" && pileFallback) return null;
    return s;
  });
  const laneShowsStack = $derived(
    floatingStackStyle !== null && stackLaneLive(view.stack_items, view.pending_triggers),
  );

  // Countering a stack item, from either surface.
  function counterStackItem(item: StackItemView): void {
    const verb = item.kind === "spell" ? "counter_spell" : "counter_ability";
    guardedSendAction(verb, { instance_id: item.id });
  }
</script>

<svelte:window onkeydown={onExpandKey} />

<!-- One seat's PlayerPanel, on the table or (ADR 0120 §3) a second time
     in the expanded overlay, with exactly the same props and handlers, so
     a click in the overlay goes where the same click on the table goes.
     `pos` is null for a spectator's grid. -->
{#snippet seatPanel(seat: PlayerView, pos: SeatPosition | null, expanded: boolean)}
  {#if pos === null}
    <PlayerPanel
      {seat}
      isSelf={false}
      spectator={true}
      {expanded}
      isActive={seat.id === activeSeatID}
      hasPriority={seat.id === prioritySeatID}
      {viewerID}
      {isAdmin}
      sendAction={guardedSendAction}
      isMonarch={seat.id === monarchID}
      isInitiative={seat.id === initiativeID}
      {view}
      controlledCards={cardsByController.get(seat.id) ?? []}
      exile={exileForOwner(seat.id)}
      {combatMode}
      {selectedCombatCardID}
      {onSelectCombatCard}
      {onDeclareAttack}
      {onDeclareBlock}
      onTapToggle={handleTapToggle}
      onPlayCard={handlePlayCard}
      onDrawCard={handleDrawCard}
      onTargetPlayer={handleTargetPlayer}
      onTargetCard={handleTargetCard}
      considering={seat.id === consideringSeatID}
    />
  {:else}
    <PlayerPanel
      {seat}
      isSelf={pos === "self"}
      flipped={opponentsInARow($settings.display.tableLayout) || opponentCount === 2
        ? pos !== "self"
        : pos === "across" || pos === "across_next"}
      {expanded}
      isActive={seat.id === activeSeatID}
      hasPriority={seat.id === prioritySeatID}
      {viewerID}
      {isAdmin}
      sendAction={guardedSendAction}
      isMonarch={seat.id === monarchID}
      isInitiative={seat.id === initiativeID}
      {view}
      controlledCards={cardsByController.get(seat.id) ?? []}
      exile={exileForOwner(seat.id)}
      {combatMode}
      {selectedCombatCardID}
      {onSelectCombatCard}
      {onDeclareAttack}
      {onDeclareBlock}
      onTapToggle={handleTapToggle}
      onPlayCard={handlePlayCard}
      onDrawCard={handleDrawCard}
      onTargetPlayer={handleTargetPlayer}
      onTargetCard={handleTargetCard}
      docked={docked && pos === "self"}
      coached={coached && pos === "self"}
      onActivateAbility={handleActivateAbility}
      onManaAbilityCost={handleManaAbilityCost}
      considering={seat.id === consideringSeatID}
      onDeclareAttackers={pos === "self" && !disabled ? onDeclareAttackers : undefined}
      {legal}
      {legalGate}
    />
  {/if}
{/snippet}

<!-- ADR 0120 §4: the keyboard route to the overlay, and on a touch screen
     the only one. Shown on the slot's hover and on focus. -->
{#snippet expandButton(seat: PlayerView)}
  <button
    class="expand-board"
    type="button"
    aria-label={expandButtonName(seat)}
    title={expandButtonName(seat)}
    onclick={(e) => openExpanded(seat.id, e.currentTarget)}
  >
    ⤢
  </button>
{/snippet}

<div
  class="board"
  class:spectator={isSpectator}
  class:board-disabled={disabled}
  aria-disabled={disabled}
  data-opp-count={opponentCount}
  data-seat-count={view.seats.length}
  bind:this={boardEl}
>
  {#if placements}
    {#each positions as pos (pos)}
      {@const seat = placements[pos]}
      {#if seat}
        {@const decision = renderingFor(seat, pos)}
        <div class="slot" data-pos={pos} style:grid-area={pos}>
          {#if decision.rendering === "summary"}
            <SeatSummary
              {seat}
              {view}
              {viewerID}
              controlledCards={cardsByController.get(seat.id) ?? []}
              isActive={seat.id === activeSeatID}
              hasPriority={seat.id === prioritySeatID}
              isMonarch={seat.id === monarchID}
              isInitiative={seat.id === initiativeID}
              sendAction={guardedSendAction}
              {combatMode}
              {selectedCombatCardID}
              {onDeclareAttack}
              {onDeclareBlock}
              onTargetPlayer={handleTargetPlayer}
              onTargetCard={handleTargetCard}
              onExpand={(from) => openExpanded(seat.id, from)}
              considering={seat.id === consideringSeatID}
              {legalGate}
            />
          {:else}
            {@render expandButton(seat)}
            {@render seatPanel(seat, pos, false)}
          {/if}
        </div>
      {/if}
    {/each}
  {:else}
    <!-- Spectator path: uniform grid, all upright, equal real estate. -->
    {#each view.seats as seat (seat.id)}
      <div class="slot spectator-slot">
        {@render expandButton(seat)}
        {@render seatPanel(seat, null, false)}
      </div>
    {/each}
  {/if}

  {#if expanded && expandedSeat}
    <!-- ADR 0120 §2–§3: the expanded board. Over the table at z 32, beside
         the avatar it belongs to, above the dock's clearance, no scrim.
         Its panel is a second PlayerPanel for the seat, so the table's
         handlers serve its clicks; it is the region, and the panel inside
         is an unnamed group, so "<name> board" stays unique. -->
    <div
      class="board-expanded"
      class:fill={overlaySpan === null}
      class:pinned={expanded.pinned}
      data-board-expanded
      data-seat-id-expanded={expandedSeat.id}
      role="region"
      aria-label={expandedRegionName(expandedSeat)}
      style:left={overlaySpan ? `${overlaySpan.left}px` : null}
      style:width={overlaySpan ? `${overlaySpan.width}px` : null}
      bind:this={overlayEl}
      onpointerenter={() => dispatchExpand({ type: "overlay-enter" })}
      onpointerleave={() => dispatchExpand({ type: "overlay-leave" })}
      class:animate={overlayMotion}
    >
      <button
        class="pin-board"
        type="button"
        aria-label={pinButtonName(expandedSeat)}
        aria-pressed={expanded.pinned}
        title={expanded.pinned ? "Unpin and close" : "Pin open"}
        bind:this={pinButtonEl}
        onclick={togglePin}
      >
        <Icon name="pin" size={13} />
      </button>
      {#key expandedSeat.id}
        {@render seatPanel(
          expandedSeat,
          isSpectator ? null : expandedSeat.id === viewerID ? "self" : "across",
          true,
        )}
      {/key}
    </div>
  {/if}

  <!-- #1467: the fan lane draws its own stack-target arrows, and so
       does the pile (ADR 0119 §1). -->
  <CombatArrows
    {view}
    {boardEl}
    cues={combatCues}
    stackTargets={!(
      laneShowsStack &&
      (floatingStackStyle === "fan" || floatingStackStyle === "pile")
    )}
  />
  <!-- ADR 0134: attackers lunge and hits land. Art-only copies at z 37,
       aria-hidden, no pointer events, on the same beat clock. -->
  <CombatStrikes {view} {boardEl} cues={combatCues} />
  <!-- ADR 0121 §7: every die a card rolls and every coin it flips
       tumbles at the roller's seat, z 41, aria-hidden, no pointer
       events. -->
  <DiceLayer {view} {boardEl} {beatsPrimeKey} />
  <!-- ADR 0119 §4: what the stack targets is ringed in every style,
       compact included; and while the viewer chooses targets, the
       source glows and an arrow runs to each pick and to the pointer. -->
  <StackTargetRings {view} {viewerID} {boardEl} />
  <TargetingArrows {view} {boardEl} />
  <HoverZoomOverlay {view} />
  <!-- Attention strip: one column over the table (the middle
       opponent's hand row in the row layout, the top-left seat's
       hand row otherwise) that stacks every live prompt — the stack
       card first, then whatever Game.svelte renders in `attention`. -->
  <!-- #1467: a floating stack style replaces the docked card while
       it has something to show. The docked card is not rendered at
       all then (rather than hidden), so the stack is never on screen
       twice and nothing in the strip is focusable behind the lane.
       `compact`, the pile's phone and short-board fallback (ADR 0119
       §1) and the rare frame with a stack card but no stack item
       record all keep the docked card. -->
  {#if floatingStackStyle}
    <StackLaneHost
      {view}
      {viewerID}
      {boardEl}
      style={floatingStackStyle}
      considering={consideringSeatID !== null && consideringSeatID === prioritySeatID}
      onCounter={counterStackItem}
      onTargetStackItem={(item) => completeTargetedCast("card", item.id)}
    />
  {/if}
  <!-- ADR 0111 §4 / §10 (PR 4): the attention strip is `region
       "attention"`, the tutorial's step-9 anchor (ADR 0076 §2.4) and an
       e2e contract. It keeps what asks nothing: the stack card, the bot
       feed, reveals, the roll-call and the toasts. -->
  <div class="strip" role="region" aria-label={L.attention}>
    {#if !laneShowsStack}
      <StackOverlay
        stack={view.stack}
        battlefield={view.battlefield}
        exile={view.exile}
        stackItems={view.stack_items ?? []}
        pendingTriggers={view.pending_triggers ?? []}
        seats={view.seats}
        viewerHasPriority={prioritySeatID === viewerID}
        priorityHolderName={view.seats[view.turn.priority_holder]?.name ?? null}
        priorityHolderConsidering={consideringSeatID !== null &&
          consideringSeatID === prioritySeatID}
        splitSecondActive={view.split_second_active === true}
        onCounter={counterStackItem}
        onTargetStackItem={(item) => completeTargetedCast("card", item.id)}
      />
    {/if}
    {@render attention?.()}
  </div>
  <!-- ADR 0119 §3: a card leaving the stack lingers where it was drawn,
       badged resolved / countered / fizzled, then flies to where it
       went. One overlay for every style, aria-hidden and outside the
       labelled regions above. -->
  <StackLinger {view} {viewerID} {boardEl} {beatsPrimeKey} />
  <VotingPanel {view} {viewerID} sendAction={guardedSendAction} {docked} />
  <SacrificeCostModal
    source={sacrificePrompt?.card ?? null}
    label={sacrificePrompt?.ability.sacrifice_label ?? "a permanent"}
    options={sacrificeOptions}
    count={sacrificeBounds.max}
    min={sacrificeBounds.min}
    eachOf={sacrificePrompt?.ability.sacrifice_options?.each_of}
    onConfirm={confirmSacrifice}
    onCancel={() => {
      if (sacrificePrompt?.kind === "mana") resetManaCostPayment();
      sacrificePrompt = null;
    }}
  />
  <CrewCostModal
    card={crewPrompt?.card ?? null}
    ability={crewPrompt?.ability ?? null}
    options={crewOptions}
    onConfirm={confirmCrew}
    onCancel={() => (crewPrompt = null)}
  />
  <!-- #1703: teamwork (CR 702.194a) is crew's picker on a spell. -->
  <CrewCostModal
    card={teamworkPrompt?.card ?? null}
    ability={null}
    threshold={teamworkPrompt?.n}
    keyword="Teamwork"
    rule="CR 702.194"
    options={teamworkOptions}
    onConfirm={confirmTeamwork}
    onCancel={() => (teamworkPrompt = null)}
  />
  <!-- #1703: blight N (CR 701.68a) — one creature you control. -->
  <SacrificeCostModal
    source={blightPrompt?.card ?? null}
    label={blightPrompt?.label ?? "a creature you control"}
    options={blightOptions}
    count={1}
    verb="Choose"
    onConfirm={confirmBlight}
    onCancel={() => (blightPrompt = null)}
  />
  <!-- ADR 0100 amendment: reveal a card from hand / behold — one pick. -->
  <SacrificeCostModal
    source={revealPrompt?.card ?? null}
    label={revealPrompt?.label ?? "a card to show"}
    options={revealOptions}
    count={1}
    verb="Choose"
    onConfirm={confirmReveal}
    onCancel={() => (revealPrompt = null)}
  />
  <CounterCostModal
    card={counterPrompt?.card ?? manaCounterPrompt?.card ?? null}
    ability={counterPrompt?.ability ?? manaCounterPrompt?.ability ?? null}
    board={view.battlefield.cards}
    onConfirm={confirmCounterCost}
    onCancel={() => {
      if (manaCounterPrompt) resetManaCostPayment();
      counterPrompt = null;
      manaCounterPrompt = null;
    }}
  />
  <FacePickerModal
    card={facePromptCard}
    zone={facePromptZone}
    onConfirm={confirmFace}
    onCancel={() => (facePromptCard = null)}
  />
  <AlternativeCostModal
    card={altCostPromptCard}
    seats={view.seats}
    onConfirm={confirmAltCost}
    onCancel={() => {
      altCostPromptCard = null;
      altCostPromptChoices = {};
    }}
  />
  <AltCostPaymentModal
    card={altPayPromptCard}
    offer={altPayOffer ?? null}
    options={altPayOptions}
    onConfirm={confirmAltPay}
    onCancel={() => {
      altPayPromptCard = null;
      altPayPromptChoices = {};
    }}
  />
  <!-- #1727: an alternative cost's sacrifice ("Flashback—Sacrifice
       three creatures") is the sacrifice picker; the picks pay the
       offer, on alt_cost_ids. -->
  <SacrificeCostModal
    source={altSacPromptCard}
    label={altSacPromptLabel}
    options={altSacOptions}
    count={altSacBounds.max}
    min={altSacBounds.min}
    eachOf={altSacPromptClause?.each_of}
    onConfirm={confirmAltSacrifice}
    onCancel={() => {
      altSacPromptCard = null;
      altSacPromptChoices = {};
      altSacPromptClause = undefined;
    }}
  />
  <!-- ADR 0135 §1: an alternative cost's tap ("tap an untapped creature
       you control rather than pay this spell's mana cost") is the tap
       picker; the picks pay the offer, on alt_cost_ids. -->
  <SacrificeCostModal
    source={altTapPromptCard}
    label={altTapPromptLabel}
    options={altTapOptions}
    count={altTapBounds.max}
    min={altTapBounds.min}
    verb="Tap"
    castName={altTapPromptCard?.name}
    onConfirm={confirmAltTap}
    onCancel={() => {
      altTapPromptCard = null;
      altTapPromptChoices = {};
      altTapPromptClause = undefined;
    }}
  />
  <!-- ADR 0100: the count and the label are the cost THIS cast pays —
       the chosen branch of an either/or cost, or the card's own. -->
  <DiscardCostModal
    card={discardPromptCard}
    options={discardCostOptions}
    need={discardPromptCost?.discard_cards}
    label={discardPromptCost?.label}
    onConfirm={confirmDiscardCost}
    onCancel={() => {
      discardPromptCard = null;
      discardPromptChoices = {};
    }}
  />
  <!-- #2598: "Reveal X black cards from your hand" (Martyr of Bones),
       the same picker with the verb changed. Revealing moves nothing, so
       the confirm names how many are shown; the number IS the X. -->
  <DiscardCostModal
    card={abilityRevealPrompt?.card ?? null}
    options={abilityRevealOptions}
    need={abilityRevealPrompt?.ability.reveal_cost_n}
    label={abilityRevealPrompt?.ability.reveal_cost_label}
    variable={abilityRevealPrompt?.ability.reveal_cost_count_from_x}
    verb="Reveal"
    note="cost · CR 602.2b"
    onConfirm={confirmAbilityRevealCost}
    onCancel={() => {
      abilityRevealPrompt = null;
      abilityRevealIDs = [];
      abilityRevealX = undefined;
    }}
  />
  <!-- #660: the same picker, one cost site over — an activated
       ability's "Discard a creature card" (CR 602.2b). -->
  <DiscardCostModal
    card={abilityDiscardPrompt?.card ?? null}
    options={abilityDiscardOptions}
    need={abilityDiscardPrompt?.ability.discard_cost_n}
    label={abilityDiscardPrompt?.ability.discard_cost_label}
    variable={abilityDiscardPrompt?.ability.discard_cost_count_from_x}
    onConfirm={confirmAbilityDiscardCost}
    onCancel={() => {
      abilityDiscardPrompt = null;
      abilityDiscardIDs = [];
    }}
  />
  <!-- #1297: "Exile two cards from your graveyard" (Grim Lavamancer),
       "Exile a card from your hand" (Holistic Wisdom). The discard
       picker with the verb changed and the pile named; the answer
       rides exile_ids, never discard_ids. -->
  <DiscardCostModal
    card={abilityExilePrompt?.card ?? null}
    options={abilityExileOptions}
    need={abilityExilePrompt?.ability.exile_cost_n}
    label={abilityExilePrompt?.ability.exile_cost_label}
    verb="Exile"
    note={abilityExilePrompt ? exileCostNote(abilityExilePrompt.ability) : undefined}
    where={abilityExilePrompt ? exileCostWhere(abilityExilePrompt.ability) : undefined}
    onConfirm={confirmAbilityExileCost}
    onCancel={() => {
      abilityExilePrompt = null;
      abilityExileIDs = [];
      abilityDiscardIDs = [];
    }}
  />
  <!-- ADR 0109 §7 (#1902): "Put a card from your hand on top of your
       library" (Penance, Leashling). The discard picker with its own
       verb and small print; the answer rides top_ids. -->
  <DiscardCostModal
    card={abilityTopPrompt?.card ?? null}
    options={abilityTopOptions}
    need={abilityTopPrompt?.ability.top_cost_n}
    label={abilityTopPrompt?.ability.top_cost_label}
    verb="Put on top"
    note={topCostNote}
    onConfirm={confirmAbilityTopCost}
    onCancel={() => {
      abilityTopPrompt = null;
      abilityTopIDs = [];
      abilityExileIDs = [];
      abilityDiscardIDs = [];
    }}
  />
  <!-- ADR 0109 §7 and owner decision 3: "Discard a card at random" and
       "Exile the top N cards of your library" have nothing to pick, so
       they are confirmed rather than picked. -->
  <CostConfirmModal
    card={abilityCostConfirmPrompt?.card ?? null}
    lines={abilityCostConfirmLines}
    note={abilityCostConfirmPrompt ? costConfirmNote(abilityCostConfirmPrompt.ability) : undefined}
    onConfirm={confirmAbilityCostConfirm}
    onCancel={() => {
      abilityCostConfirmPrompt = null;
      abilityTopIDs = [];
      abilityExileIDs = [];
      abilityDiscardIDs = [];
    }}
  />
  <!-- #1213: the mana-ability half of the discard picker (Skirge
       Familiar). Same modal, same wire field, a different action. -->
  <DiscardCostModal
    card={manaDiscardPrompt?.card ?? null}
    options={manaDiscardOptions}
    need={manaDiscardPrompt?.ability.discard_cost_n}
    label={manaDiscardPrompt?.ability.discard_cost_label}
    onConfirm={confirmManaDiscardCost}
    onCancel={() => {
      manaDiscardPrompt = null;
      manaDiscardIDs = [];
    }}
  />
  <!-- #1283: "Exile a card from your hand" (Cadaverous Bloom). The
       discard picker with the verb changed; the answer rides
       exile_ids, never discard_ids. -->
  <DiscardCostModal
    card={manaExilePrompt?.card ?? null}
    options={manaExileOptions}
    need={manaExilePrompt?.ability.exile_cost_n}
    label={manaExilePrompt?.ability.exile_cost_label}
    verb="Exile"
    note={manaExilePrompt ? exileCostNote(manaExilePrompt.ability) : undefined}
    where={manaExilePrompt ? exileCostWhere(manaExilePrompt.ability) : undefined}
    onConfirm={confirmManaExileCost}
    onCancel={() => {
      manaExilePrompt = null;
      manaExileIDs = [];
      manaDiscardIDs = [];
    }}
  />
  <!-- #1213: "Return a permanent you control to its owner's hand" as
       a cost. The sacrifice picker with a different verb. -->
  <SacrificeCostModal
    source={abilityReturnPrompt?.card ?? null}
    label={abilityReturnPrompt?.ability.return_label ?? "a permanent you control"}
    options={abilityReturnOptions}
    count={abilityReturnPrompt?.ability.return_options?.max ?? 1}
    verb="Return"
    onConfirm={confirmAbilityReturnCost}
    onCancel={() => {
      abilityReturnPrompt = null;
      abilityReturnIDs = [];
    }}
  />
  <!-- #1600: "Exile a creature you control" as a cost (The Soul Stone,
       Altar of Bhaal, City of Shadows). The sacrifice picker with the
       verb "Exile"; the answer rides exile_permanent_ids. -->
  <SacrificeCostModal
    source={abilityExilePermanentPrompt?.card ?? null}
    label={abilityExilePermanentPrompt?.ability.exile_permanent_label ?? "a creature you control"}
    options={abilityExilePermanentOptions}
    count={abilityExilePermanentPrompt?.ability.exile_permanent_options?.max ?? 1}
    verb="Exile"
    onConfirm={confirmAbilityExilePermanentCost}
    onCancel={() => {
      abilityExilePermanentPrompt = null;
      abilityExilePermanentIDs = [];
      abilityReturnIDs = [];
    }}
  />
  <!-- #1600: the same picker for Food Chain's mana ability. -->
  <SacrificeCostModal
    source={manaExilePermanentPrompt?.card ?? null}
    label={manaExilePermanentPrompt?.ability.exile_permanent_label ?? "a creature you control"}
    options={manaExilePermanentOptions}
    count={manaExilePermanentPrompt?.ability.exile_permanent_options?.max ?? 1}
    verb="Exile"
    onConfirm={confirmManaExilePermanentCost}
    onCancel={() => {
      manaExilePermanentPrompt = null;
      resetManaCostPayment();
    }}
  />
  <!-- #759: station's "Tap another untapped creature you control".
       The sacrifice picker once more, with the verb "Tap". -->
  <SacrificeCostModal
    source={abilityTapPrompt?.card ?? null}
    label={abilityTapPrompt?.ability.tap_others_label ?? "another untapped creature you control"}
    options={abilityTapOptions}
    count={abilityTapPrompt?.ability.tap_others_options?.max ?? 1}
    min={abilityTapBounds.min}
    verb="Tap"
    onConfirm={confirmAbilityTapCost}
    onCancel={() => {
      abilityTapPrompt = null;
      abilityTapIDs = [];
      abilityTapX = undefined;
      abilityReturnIDs = [];
    }}
  />
  <!-- #758: Springleaf Drum's mana-ability spelling of the same
       TapOthers component. -->
  <SacrificeCostModal
    source={manaTapPrompt?.card ?? null}
    label={manaTapPrompt?.ability.tap_others_label ?? "an untapped creature you control"}
    options={manaTapOptions}
    count={manaTapPrompt?.ability.tap_others_options?.max ?? 1}
    verb="Tap"
    onConfirm={confirmManaTapCost}
    onCancel={() => {
      manaTapPrompt = null;
      resetManaCostPayment();
    }}
  />
  <SacrificeCostModal
    source={sacrificePromptCard}
    label={sacrificePromptLabel}
    options={castSacrificeOptions}
    count={castSacrificeBounds.max}
    min={castSacrificeBounds.min}
    countIsX={sacrificePromptClause?.count_from_x === true}
    eachOf={sacrificePromptClause?.each_of}
    onConfirm={confirmSacrificeCost}
    onCancel={() => {
      sacrificePromptCard = null;
      sacrificePromptChoices = {};
    }}
  />
  <!-- #1563, CR 601.2d: the shares of a divided step, asked once its
       two or more targets are picked. -->
  <DivideDamageModal
    sourceName={dividePrompt ? dividePrompt.card.name : null}
    targets={divideTargets}
    total={dividePrompt?.divide ?? 0}
    upTo={dividePrompt?.divideUpTo === true}
    onConfirm={confirmDivision}
    onCancel={() => (dividePrompt = null)}
  />
  <XCostModal
    gameID={view.id}
    card={xPromptCard}
    suggestedMax={xPromptCard?.additional_cost?.blight_x
      ? Math.min(suggestedX, xPromptCard.additional_cost.blight_x_max ?? 0)
      : suggestedX}
    costLabel={xPromptCard
      ? alternativeCostByKey(xPromptCard, xPromptChoices.altCost)?.mana_cost
      : undefined}
    castParams={castPreviewParams(xPromptChoices)}
    onConfirm={confirmX}
    onCancel={() => {
      xPromptCard = null;
      xPromptChoices = {};
    }}
  />
  <!-- The same picker for an activated ability's {X} (CR 602.2b),
       priced against the ABILITY's cost rather than the card's. -->
  <XCostModal
    gameID={view.id}
    card={xAbilityPrompt?.card ?? null}
    suggestedMax={suggestedAbilityX}
    abilityIndex={xAbilityPrompt?.ability.index}
    costLabel={xAbilityPrompt?.ability.mana_cost}
    minX={xAbilityPrompt?.ability.min_x ?? 0}
    maxX={xAbilityPrompt ? abilityEnergyMaxX(xAbilityPrompt.ability, viewerEnergy) : undefined}
    confirmVerb="Activate"
    onConfirm={confirmAbilityX}
    onCancel={() => (xAbilityPrompt = null)}
  />
  <!-- CR 107.4f (#916): how many Phyrexian symbols the CAST pays
       with 2 life each. Between the X picker and the convoke picker,
       because it is announced with them and priced after X. -->
  <PhyrexianCostModal
    gameID={view.id}
    card={phyrexianPrompt?.card ?? null}
    symbols={phyrexianPrompt?.symbols ?? 0}
    granted={phyrexianPrompt?.granted ?? 0}
    suggest={phyrexianPrompt?.suggest ?? false}
    life={viewerLife}
    xValue={phyrexianPrompt?.choices.xValue}
    castParams={castPreviewParams(phyrexianPrompt?.choices)}
    onConfirm={confirmPhyrexianLife}
    onCancel={() => (phyrexianPrompt = null)}
  />
  <!-- The same stepper for an activated ability's mana component
       (CR 602.2b), priced against the ABILITY's cost. -->
  <PhyrexianCostModal
    gameID={view.id}
    card={phyrexianAbilityPrompt?.card ?? null}
    symbols={phyrexianAbilityPrompt ? (phyrexianAbilityPrompt.ability.phyrexian_symbols ?? 0) : 0}
    granted={phyrexianAbilityPrompt
      ? phyrexianGrantedForAbility(phyrexianAbilityPrompt.ability)
      : 0}
    suggest={phyrexianAbilityPrompt?.suggest ?? false}
    life={viewerLife}
    xValue={phyrexianAbilityPrompt?.xValue}
    abilityIndex={phyrexianAbilityPrompt?.ability.index}
    costLabel={phyrexianAbilityPrompt?.ability.mana_cost}
    confirmVerb="Activate"
    onConfirm={confirmAbilityPhyrexianLife}
    onCancel={() => (phyrexianAbilityPrompt = null)}
  />
  <TapCostModal
    card={tapPromptCard}
    cost={tapPromptCard ? (tapCostOf(tapPromptCard) ?? null) : null}
    options={tapCostOptions}
    limit={tapCostCap}
    onConfirm={confirmTapCost}
    onCancel={() => {
      tapPromptCard = null;
      tapPromptChoices = {};
    }}
  />
  <!-- ADR 0100: delve's graveyard picker, capped by the server's
       delve_budget for the announcement so far. -->
  <DelveCostModal
    gameID={view.id}
    card={delvePromptCard}
    options={delveCostOptions}
    xValue={delvePromptChoices.xValue ?? 0}
    phyrexianLife={delvePromptChoices.phyrexianLife ?? 0}
    castParams={castPreviewParams(delvePromptChoices)}
    onConfirm={confirmDelve}
    onCancel={() => {
      delvePromptCard = null;
      delvePromptChoices = {};
    }}
  />
  <!-- #1310: the same picker for an activated ability's "Waterbend
       {N}:" cost, fed the ability's own clause. -->
  <TapCostModal
    card={abilityWaterbendPrompt?.card ?? null}
    cost={abilityWaterbendPrompt?.ability.waterbend ?? null}
    options={abilityWaterbendOptions}
    limit={abilityWaterbendPrompt?.ability.waterbend
      ? waterbendLimit(abilityWaterbendPrompt.ability.waterbend, abilityWaterbendPrompt.xValue)
      : 0}
    onConfirm={confirmAbilityWaterbend}
    onCancel={() => {
      abilityWaterbendPrompt = null;
      abilityWaterbendIDs = undefined;
    }}
  />
  <ModePickerModal
    card={modePromptCard}
    onConfirm={confirmModes}
    onCancel={() => {
      modePromptCard = null;
      modePromptChoices = {};
    }}
  />
  <!-- #2126, CR 702.120a: escalate's cards or creatures, named after the
       modes are chosen. The exact-count card picker, in two verbs. -->
  <DiscardCostModal
    card={escalatePrompt?.card ?? null}
    options={escalateOptions}
    need={escalateNeed}
    label={escalatePrompt?.card.modes?.escalate?.label}
    verb={escalatePrompt?.stage === "tap" ? "Tap" : "Discard"}
    note="escalate · CR 702.120a"
    where={escalatePrompt?.stage === "tap" ? "untapped under your control" : "in hand"}
    onConfirm={confirmEscalate}
    onCancel={() => {
      escalatePrompt = null;
    }}
  />
  <ModePickerModal
    card={abilityModeCard}
    onConfirm={confirmAbilityModes}
    onCancel={() => {
      abilityModePrompt = null;
    }}
  />
  {#if $zoneBrowser}
    <ZoneBrowserModal
      {view}
      {viewerID}
      zoneKind={$zoneBrowser.zoneKind}
      ownerSeat={{ id: $zoneBrowser.ownerID, name: $zoneBrowser.ownerName }}
      sendAction={guardedSendAction}
      onClose={closeZoneBrowser}
      onTargetCard={handleTargetCard}
      onCastCard={handlePlayCard}
      onActivateAbility={handleActivateAbility}
      sorcerySpeedBlocked={browsedZoneSorcerySpeedBlocked}
      {legal}
      {legalGate}
    />
  {/if}
  {#if $manaSourcePicker}
    <!-- #1438: the anchored "which mana?" picker a left-click on a
         source with several mana abilities opens. Its pick routes
         exactly as the right-click menu's mana row does. -->
    <ManaSourcePicker
      {view}
      open={$manaSourcePicker}
      onPick={(card, index, colors) => handleMenuActivate(card, { kind: "mana", index, colors })}
      onClose={closeManaSourcePicker}
    />
  {/if}
  {#if castAnywayRequest}
    <DockRequest request={castAnywayRequest} />
  {/if}
  {#if $cardMenu}
    <CardContextMenu
      {view}
      {viewerID}
      {isAdmin}
      open={$cardMenu}
      sendAction={guardedSendAction}
      onActivate={handleMenuActivate}
      onClose={closeCardMenu}
      {legal}
      {legalGate}
    />
  {/if}
</div>

<style>
  .board {
    position: relative;
    width: 100%;
    height: 100%;
    display: grid;
    gap: 8px;
    padding: 6px;
    box-sizing: border-box;
    overflow: hidden;
    /* Table-like depth: a vignette from centre to edges on top of a
       radial "spotlight" in the middle, over a deep navy base. The
       spotlight anchors the eye at the stack/centre area without
       drawing attention from the panels themselves. */
    background:
      radial-gradient(
        60% 55% at 50% 50%,
        color-mix(in srgb, var(--accent) 5%, transparent) 0%,
        rgba(0, 0, 0, 0) 60%
      ),
      radial-gradient(120% 100% at 50% 100%, rgba(0, 0, 0, 0) 0%, rgba(0, 0, 0, 0.45) 80%),
      var(--bg-1);
  }
  .board::before {
    /* Very subtle tiling noise to break up the flat gradients — makes
       the "table" feel material rather than plastic. SVG-data-URI keeps
       it zero-cost on the network. */
    content: "";
    position: absolute;
    inset: 0;
    pointer-events: none;
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='140' height='140'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='2' stitchTiles='stitch'/%3E%3CfeColorMatrix values='0 0 0 0 1 0 0 0 0 1 0 0 0 0 1 0 0 0 0.035 0'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E");
    mix-blend-mode: overlay;
    opacity: 0.5;
  }
  /* #519: the board stays rendered and stays the last-known state
     (see ConnectionBanner.svelte's non-goals) — this only says so.
     Dimming + grayscale is a look, not the guard; guardedSendAction
     above is what actually stops a card action from going anywhere
     while the socket is down. Cards, panels and the stack all dim
     together rather than one at a time, so nothing looks selectively
     broken. */
  .board-disabled {
    filter: grayscale(0.45) brightness(0.82);
    cursor: not-allowed;
  }
  .board-disabled .slot,
  .board-disabled .board-expanded > :global(.panel) {
    pointer-events: none;
  }
  .slot {
    min-height: 0;
    min-width: 0;
    /* Anchors .expand-board. */
    position: relative;
  }

  /* ADR 0120 §4: the expand button, at the slot's top-right corner
     where ADR 0077's collapse button was. Above the panel's own chrome
     (PlayerPanel's rail is z-index 3) and under the attention strip (40)
     and the hover zoom, because a prompt covering this button is strictly
     better than this button covering a prompt. It shows on the slot's
     hover and on focus; without a hovering pointer it always shows,
     because there it is the only way in. */
  .expand-board {
    position: absolute;
    top: 6px;
    right: 6px;
    z-index: 10;
    padding: 3px 6px;
    border-radius: 6px;
    border: 1px solid var(--border);
    background: var(--surface);
    color: var(--fg-dim);
    font-size: 12px;
    line-height: 1;
    cursor: pointer;
    opacity: 0;
    transition: opacity 120ms ease;
  }
  .slot:hover > .expand-board,
  .expand-board:focus-visible {
    opacity: 1;
  }
  .expand-board:hover {
    color: var(--fg);
    border-color: var(--border-strong);
  }
  @media (hover: none) {
    .expand-board {
      opacity: 1;
    }
  }

  /* ADR 0120 §2: the expanded board. z 32: above every panel's chrome
     (≤ 10), below the arrows (35–36), the floating stack (38), the
     attention strip (40), the dock (55), card-local menus (60+), modals
     (200) and the hover zoom (300). Full height, stopping above the dock
     by the zoom's own clearance, so nothing in it is ever under the dock.
     Board.svelte sets `left` and `width` inline to the span beside the
     avatar; `.fill` (no span: a phone, or an avatar with no box) takes
     the whole board. No scrim: the table around it stays clickable. */
  .board-expanded {
    position: absolute;
    top: 10px;
    bottom: calc(10px + var(--dock-zoom-clear, 0px));
    z-index: 32;
    display: flex;
    min-width: 0;
    border-radius: 14px;
    background: var(--surface);
    box-shadow: var(--shadow-lg);
    transform-origin: center;
  }
  .board-expanded.fill {
    left: 6px;
    right: 6px;
  }
  /* §4: a 120 ms fade in, from 0.98, only with animations on and motion
     not reduced. */
  .board-expanded.animate {
    animation: board-expand-in 120ms ease-out;
  }
  @keyframes board-expand-in {
    from {
      opacity: 0;
      transform: scale(0.98);
    }
    to {
      opacity: 1;
      transform: none;
    }
  }
  .board-expanded > :global(.panel) {
    flex: 1 1 auto;
    min-width: 0;
    border-color: var(--border-strong);
  }
  .board-expanded.pinned > :global(.panel) {
    border-color: var(--accent, var(--border-strong));
  }
  /* §5: a phone gets the board's whole area. */
  @media (max-width: 599px) {
    .board-expanded,
    .board-expanded.fill {
      inset: 0;
      width: auto;
    }
  }

  /* §4: the pin toggle, top right of the overlay, over its panel's rail. */
  .pin-board {
    position: absolute;
    top: 6px;
    right: 6px;
    z-index: 20;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 3px 5px;
    border-radius: 6px;
    border: 1px solid var(--border);
    background: var(--surface);
    color: var(--fg-dim);
    line-height: 1;
    cursor: pointer;
  }
  .pin-board:hover {
    color: var(--fg);
    border-color: var(--border-strong);
  }
  .pin-board[aria-pressed="true"] {
    color: var(--accent, var(--fg));
    border-color: var(--accent, var(--border-strong));
  }

  /* The attention strip. Fixed-width column, grows downward, capped
     to one panel's width minus its rail (104px + gaps) so it only
     ever covers an opponent's face-down hand row — never a rail.
     pointer-events pass through the gaps so the cards under it stay
     clickable. */
  .strip {
    position: absolute;
    top: 10px;
    left: 12px;
    width: min(512px, calc(50% - 132px));
    max-height: calc(100% - 20px);
    display: flex;
    flex-direction: column;
    gap: 6px;
    z-index: 40;
    pointer-events: none;
  }
  .strip > :global(*) {
    pointer-events: auto;
    flex: 0 1 auto;
    min-height: 0;
  }
  :global(:root[data-table-layout="row"]) .board[data-opp-count="3"] .strip,
  :global(:root[data-table-layout="focus"]) .board[data-opp-count="3"] .strip {
    left: calc(33.333% + 6px);
    width: min(512px, calc(33.333% - 132px));
  }

  /* Quadrant templates per opponent count. The bottom row (next +
     self) always gets more vertical real estate than the top row
     because the self panel lives there and needs space for the full-
     size hand fan. */
  .board[data-opp-count="0"] {
    grid-template-columns: 1fr;
    grid-template-rows: 1fr;
    grid-template-areas: "self";
  }
  .board[data-opp-count="1"] {
    grid-template-columns: 1fr;
    grid-template-rows: minmax(0, 0.7fr) minmax(0, 1.3fr);
    grid-template-areas:
      "across"
      "self";
  }
  .board[data-opp-count="2"] {
    grid-template-columns: 1fr 1fr;
    grid-template-rows: minmax(0, 0.7fr) minmax(0, 1.3fr);
    /* #956 — with two opponents there is no fourth quadrant to fill,
       so the old "across on top, next beside self" split spent half
       the bottom row on an opponent. A half-width self panel clips
       the hand fan horizontally (--card-h floors at 168px, so the
       cards do not shrink to fit), which is what the reporter hit.
       Both opponents now share the top row and self spans the full
       width below. This makes the quadrant and row layouts identical
       at three players, which is intended: there is no third
       arrangement worth having here. `next` is in the top row so it
       renders flipped — see the markup above. */
    grid-template-areas:
      "next   across"
      "self   self";
  }
  .board[data-opp-count="3"] {
    grid-template-columns: 1fr 1fr;
    grid-template-rows: minmax(0, 0.7fr) minmax(0, 1.3fr);
    /* Turn order runs clockwise around the screen: self (bottom-
       right) → next (bottom-left) → across (top-LEFT, diagonal)
       → across_next (top-right, directly above self). Putting
       "across" top-right instead made the order zig-zag. */
    grid-template-areas:
      "across across_next"
      "next   self";
  }

  /* Row layout (settings.display.tableLayout = "row"; quadrant is the
     default):
     every opponent sits in the top row in turn order, left to right,
     and the self panel takes the full width below. All opponents
     are `flipped` (hand at the top edge). */
  /* grid-template-rows is restated in both rules rather than
     inherited from the quadrant rules above: they match the same
     element, so editing one silently changed the other. */
  :global(:root[data-table-layout="row"]) .board[data-opp-count="2"] {
    grid-template-columns: 1fr 1fr;
    grid-template-rows: minmax(0, 0.7fr) minmax(0, 1.3fr);
    grid-template-areas:
      "next across"
      "self self";
  }
  :global(:root[data-table-layout="row"]) .board[data-opp-count="3"] {
    grid-template-columns: 1fr 1fr 1fr;
    grid-template-rows: minmax(0, 0.7fr) minmax(0, 1.3fr);
    grid-template-areas:
      "next across across_next"
      "self self   self";
  }

  /* Focus layout (#2336): an even split. Your board is the bottom
     half and the opponents share the top half, each a summary
     (expansionSettingsFor in expansion.ts); you hover or click an
     avatar to see a whole board in the expanded overlay. */
  :global(:root[data-table-layout="focus"]) .board[data-opp-count="1"] {
    grid-template-columns: 1fr;
    grid-template-rows: minmax(0, 1fr) minmax(0, 1fr);
    grid-template-areas:
      "across"
      "self";
  }
  :global(:root[data-table-layout="focus"]) .board[data-opp-count="2"] {
    grid-template-columns: 1fr 1fr;
    grid-template-rows: minmax(0, 1fr) minmax(0, 1fr);
    grid-template-areas:
      "next across"
      "self self";
  }
  :global(:root[data-table-layout="focus"]) .board[data-opp-count="3"] {
    grid-template-columns: 1fr 1fr 1fr;
    grid-template-rows: minmax(0, 1fr) minmax(0, 1fr);
    grid-template-areas:
      "next across across_next"
      "self self   self";
  }

  /* Top-row opponents are laid out flipped (hand at the top edge,
     creatures toward the centre) by PlayerPanel's `flipped` prop —
     the same "across the table" reading the old 180° rotation gave,
     with upright text and art. */

  /* Spectator layout: no "around the table" perspective, so every
     seat renders upright and gets equal screen real estate. The grid
     dimensions are chosen by total seat count (set on the .board
     itself via data-seat-count) — 1 = single panel, 2 = stacked, 3-4
     = 2x2. No 180° rotation: the spectator isn't sitting at any one
     seat, so flipping anybody upside-down would just be confusing.

     `grid-template-areas: none` is load-bearing: the quadrant rules
     above key on `.board[data-opp-count="N"]` and set named grid
     areas ("self", "across", …). For spectators, `opponentCount`
     derives to 0, so `[data-opp-count="0"]` matches and drops its
     `grid-template-areas: "self"` onto the same element. Without an
     explicit clear here, the named "self" area leaks in alongside
     the column/row overrides and auto-placement of the spectator
     panels produces a diagonal/quadrant layout instead of a stack. */
  .board.spectator {
    grid-template-columns: 1fr;
    grid-template-rows: 1fr;
    grid-template-areas: none;
  }
  .board.spectator[data-seat-count="2"] {
    grid-template-columns: 1fr;
    grid-template-rows: 1fr 1fr;
  }
  .board.spectator[data-seat-count="3"] {
    grid-template-columns: 1fr 1fr;
    grid-template-rows: 1fr 1fr;
  }
  .board.spectator[data-seat-count="4"] {
    grid-template-columns: 1fr 1fr;
    grid-template-rows: 1fr 1fr;
  }
</style>
