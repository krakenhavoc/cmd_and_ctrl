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
    CardView,
    GameView,
    LegalTargetsView,
    ManaAbilityView,
    PlayerView,
    ZoneView,
  } from "../../protocol";
  import { seatPlacements, type SeatPosition } from "../../cardTypes";
  import PlayerPanel from "./PlayerPanel.svelte";
  import SeatSummary from "./SeatSummary.svelte";
  import {
    decideSeatRendering,
    legalDefenderIDs,
    nextPinnedSeat,
    seatControlsLegalTarget,
    seatHasAttackersOn,
    type SeatDecision,
  } from "../../expansion";
  import { settings } from "../../settings";
  import HoverZoomOverlay from "./HoverZoomOverlay.svelte";
  // CommanderDamageTooltip was folded into HoverZoomOverlay — the
  // damage readout now lives inside the card preview panel instead
  // of the lower-right corner.
  import StackOverlay from "./StackOverlay.svelte";
  import CombatArrows from "./CombatArrows.svelte";
  import VotingPanel from "./VotingPanel.svelte";
  import ZoneBrowserModal from "./ZoneBrowserModal.svelte";
  import { zoneBrowser, closeZoneBrowser } from "../../zoneBrowser";
  import { phasedOutCards } from "../../phasedOut";
  import { canActivateSorcerySpeedAbility } from "../../timing";
  import CardContextMenu from "./CardContextMenu.svelte";
  import { cardMenu, closeCardMenu } from "../../contextMenu";
  import type { MenuActivate } from "../../contextMenu.logic";
  import {
    targeting,
    begin as beginTargeting,
    cancel as cancelTargeting,
    beginChoice as beginTargetingChoice,
    beginForAbility as beginTargetingForAbility,
    beginForModes as beginTargetingForModes,
    advance,
    allPicks,
    hasXCost,
    castLocksXAtZero,
    isModal,
    discardCostOf,
    castSacrificeClause,
    castSacrificeLabel,
    optionalCostsOf,
    tapCostOf,
    tapCostLimit,
    alternativeCostsOf,
    alternativeCostByKey,
    altCostPayOptions,
    applyCastChoices,
    isLegalCardTarget,
    isLegalPlayerTarget,
    isMultiPick,
    opensTargetPicker,
    togglePick,
    canConfirm,
    setConfirmHandler,
    type CastChoices,
    type CastSourceZone,
    type TargetingState,
    type TargetRef,
  } from "../../targeting";
  import { suggestedAbilityX as suggestedAbilityXFor } from "../../abilityX";
  import { castPreviewParams } from "../../castPreview";
  import { orderSacrificeOptions, sacrificeCount, sacrificeRange } from "../../sacrificeCost";
  import XCostModal from "./XCostModal.svelte";
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
  import { manaAbilityNeedsPrompt } from "../../manaAbilityCost";
  import AlternativeCostModal from "./AlternativeCostModal.svelte";
  import FacePickerModal from "./FacePickerModal.svelte";
  import { cardAsFace, needsFacePicker } from "../../faces";
  import TapCostModal from "./TapCostModal.svelte";
  import PhyrexianCostModal from "./PhyrexianCostModal.svelte";
  import {
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
    // #519: connectionBanner.ts's actionsDisabled(status), computed by
    // Game.svelte and handed down rather than recomputed here — Board
    // has no socket of its own to ask. Dims the table and turns every
    // card affordance into a no-op instead of a click that silently
    // goes nowhere while the connection is down.
    disabled?: boolean;
    // Priority controls forwarded to the self-panel's PhaseDisplay.
    autopassEnabled?: boolean;
    // #628: the CR 726 loop-breaker banner line, empty when quiet.
    loopNotice?: string;
    onPassPriority?: () => void;
    onToggleAutopass?: () => void;
    // Game.svelte's live prompts (targeting, combat hint, mulligan
    // roll-call, toasts, game end) render inside the attention strip
    // under the stack card so every "look here" surface shares one
    // anchor over the table.
    attention?: Snippet;
    // #187 / ADR 0053: changes whenever the combat damage beats should
    // prime again instead of cueing what they missed (reconnect, replay
    // toggle). Passed straight to CombatArrows.
    beatsPrimeKey?: string;
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
    disabled = false,
    autopassEnabled,
    loopNotice = "",
    onPassPriority,
    onToggleAutopass,
    attention,
    beatsPrimeKey,
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
    xAbilityPrompt ? suggestedAbilityXFor(xAbilityPrompt.ability, suggestedX) : 0,
  );

  // #916: the viewer's life total, which is CR 119.4's cap on a
  // Phyrexian life payment. Read off the live snapshot so a life loss
  // while a prompt is open shrinks its ceiling.
  const viewerLife = $derived(view.seats.find((s) => s.id === viewerID)?.life ?? 0);

  const activeSeatID = $derived(view.seats[view.turn.active_seat]?.id ?? null);
  const prioritySeatID = $derived(view.seats[view.turn.priority_holder]?.id ?? null);
  const monarchID = $derived(view.monarch ?? null);
  const initiativeID = $derived(view.initiative ?? null);

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
  function confirmAltCost(key: string | undefined, optional: number[]): void {
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

  function confirmSacrificeCost(instanceIDs: string[]): void {
    const card = sacrificePromptCard;
    const choices = sacrificePromptChoices;
    sacrificePromptCard = null;
    sacrificePromptChoices = {};
    if (!card) return;
    afterCastCosts(card, { ...choices, sacrificeIDs: instanceIDs });
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
    continueCast(card, { ...choices, tapIDs: ids });
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
    if (altCostPayOptions(alternativeCostByKey(card, choices.altCost)) !== undefined) {
      altPayPromptChoices = choices;
      altPayPromptCard = card;
      return;
    }
    afterAltCostPayment(card, choices);
  }

  function afterAltCostPayment(card: CardView, choices: CastChoices): void {
    if (discardCostOf(card) > 0) {
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
    if (clause !== undefined) {
      sacrificePromptClause = clause;
      sacrificePromptLabel = castSacrificeLabel(card, choices);
      sacrificePromptChoices = choices;
      sacrificePromptCard = card;
      return;
    }
    afterCastCosts(card, choices);
  }

  function afterCastCosts(card: CardView, choices: CastChoices): void {
    // CR 107.3b (#831): a free cast of an {X} spell has one legal X
    // and it is 0, so the picker is skipped and nothing is sent. The
    // server refuses a non-zero X on such a cast, which is what makes
    // this a prompt decision rather than a rule the client enforces.
    if (hasXCost(card) && !castLocksXAtZero(card, choices.altCost)) {
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
    // CR 107.4c/f (#916): "{U/P} can be paid with either {U} or 2
    // life", and CR 601.2b makes which one part of announcing the
    // spell. Asked after X for the same reason the tap picker is:
    // the stepper's live readout prices the mana half, and an {X}
    // cost has no size until X is announced. Skipped when the cost
    // prints no Phyrexian symbol, and when CR 119.4 leaves the
    // caster unable to buy even one — a prompt whose only answer is
    // 0 is a click, not a choice.
    const symbols = phyrexianSymbolsForCast(card, choices.altCost);
    if (shouldAskPhyrexianLife(symbols, viewerLife)) {
      phyrexianPrompt = { card, symbols, choices };
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
    continueCast(card, choices);
  }

  // #916: the cast's Phyrexian stepper. One reactive object rather
  // than the card / choices pair the older stages use, because the
  // modal reads the stashed choices (the announced X, the symbol
  // count the chosen cost prints) as well as the card — the same
  // shape xAbilityPrompt has, for the same reason.
  let phyrexianPrompt = $state<{
    card: CardView;
    symbols: number;
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
  let facePromptZone: CastSourceZone | undefined;
  function confirmFace(face: number): void {
    const card = facePromptCard;
    const fromZone = facePromptZone;
    facePromptCard = null;
    facePromptZone = undefined;
    if (!card) return;
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
    afterFace(cardAsFace(card, face), fromZone ? { face, fromZone } : { face });
  }

  function afterFace(card: CardView, choices: CastChoices): void {
    // ADR 0073: a card with kicker and no alternative cost opens the
    // same picker with only the add-ons showing — one prompt for one
    // question (CR 601.2b), rather than a second modal asking the
    // other half of it.
    if (alternativeCostsOf(card).length > 0 || optionalCostsOf(card).length > 0) {
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
  function handlePlayCard(card: CardView, fromZone?: CastSourceZone, face?: number): void {
    if (face !== undefined) {
      afterFace(cardAsFace(card, face), fromZone ? { face, fromZone } : { face });
      return;
    }
    if (needsFacePicker(card)) {
      facePromptZone = fromZone;
      facePromptCard = card;
      return;
    }
    afterFace(card, fromZone ? { fromZone } : {});
  }

  // S20 sub-PR 4: a modal spell asks for its mode(s) after X and
  // before targeting. The picker's confirm continues with the
  // chosen indexes; a targeted option enters targeting with that
  // option's legal set, otherwise the cast fires straight away.
  let modePromptCard = $state<CardView | null>(null);
  let modePromptChoices: CastChoices = {};
  // #764: EVERY chosen bullet contributes its clauses to the walk,
  // in the order they were chosen (CR 700.2c), and a repeated bullet
  // (CR 700.2d) contributes them once per occurrence. The old shape
  // took the FIRST targeted option and dropped the rest, which is
  // why Kolaghan's Command could not be cast.
  function confirmModes(modes: number[]): void {
    const card = modePromptCard;
    const choices = modePromptChoices;
    modePromptCard = null;
    modePromptChoices = {};
    if (!card) return;
    if (beginTargetingForModes(card, modes, choices)) return;
    const params: Record<string, unknown> = { instance_id: card.instance_id, modes };
    applyCastChoices(params, choices);
    guardedSendAction("cast_spell", params, viewerID ?? undefined);
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
      modePromptCard = card;
      return;
    }
    // S14: if the card declares a target_mode (catalog cards with
    // a target slot — Lightning Bolt, Counterspell), enter the
    // targeting flow and wait for a second click on a legal target.
    // Otherwise fire cast_spell immediately (lands, sorceries with
    // no targets, vanilla permanents).
    const alt = alternativeCostByKey(card, choices.altCost);
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
  function stepOrFire(state: TargetingState): void {
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
      guardedSendAction(
        "resolve_choice",
        { choice_id: state.choiceID, targets },
        viewerID ?? undefined,
      );
      targeting.set(null);
      return;
    }
    // S21 sub-PR 2: the prompt belongs to an activated ability, not
    // a cast — its cost was already paid at announce.
    if (state.ability) {
      const params: Record<string, unknown> = {
        source_card_id: state.card.instance_id,
        ability_index: state.ability.index,
        sacrifice_ids: state.ability.sacrificeIDs,
        crew_ids: state.ability.crewIDs,
        // #660: the discard picks were made at announce, before the
        // targeting step, and ride the one activate_ability with the
        // rest of the cost.
        discard_ids: abilityDiscardIDs,
        // #1213: same announcement, same message.
        return_ids: abilityReturnIDs,
        ...state.ability.counter,
        targets,
      };
      if (state.ability.xValue !== undefined) params.x_value = state.ability.xValue;
      // #916, CR 107.4f: announced with the rest of the cost, before
      // these targets, and sent in the same message.
      if (state.ability.phyrexianLife) params.phyrexian_life = state.ability.phyrexianLife;
      if (state.modes !== undefined) params.modes = state.modes;
      abilityDiscardIDs = [];
      abilityReturnIDs = [];
      abilitySacrificeX = undefined;
      guardedSendAction("activate_ability", params, viewerID ?? undefined);
      targeting.set(null);
      return;
    }
    const params: Record<string, unknown> = { instance_id: state.card.instance_id, targets };
    if (state.modes !== undefined) params.modes = state.modes;
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

  // #1213: the X a "Sacrifice X Treasures" clause announces. It is
  // the SIZE of the payment rather than a number the player types, so
  // the X stepper is skipped for such an ability — asking twice could
  // only produce an announcement the server refuses.
  let abilitySacrificeX: number | undefined;

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
    // #660: the discard payment is asked FIRST, as the cast flow asks
    // its own — it is the cost most likely to make a player back out.
    // Skipped when the hand holds exactly the cards the clause
    // demands: a modal with one possible answer is a worse version of
    // no modal.
    if (ability.discard_cost_n) {
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
    if (p.ability.sacrifice_options) {
      sacrificePrompt = { kind: "ability", card: p.card, ability: p.ability };
      return;
    }
    if (p.ability.crew_cost) {
      crewPrompt = { card: p.card, ability: p.ability };
      return;
    }
    askCounterCost(p.card, p.ability, [], []);
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
  function handleManaAbilityCost(card: CardView, ability: ManaAbilityView): void {
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
  // answered: the sacrifice picker, then the counter cost.
  function afterManaCardCosts(card: CardView, ability: ManaAbilityView): void {
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
    const ids = new Set(p.ability.exile_cost_options ?? []);
    const me = view.seats.find((s) => s.id === viewerID);
    return (me?.hand.cards ?? []).filter((c) => ids.has(c.instance_id));
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
        ...(sacrificeIDs && sacrificeIDs.length > 0 ? { sacrifice_ids: sacrificeIDs } : {}),
        // #1213: omitted when empty, so every payload a client sent
        // before this field existed is byte-for-byte unchanged.
        ...(manaDiscardIDs.length > 0 ? { discard_ids: manaDiscardIDs } : {}),
        // #1283: the same posture — absent unless the ability exiles.
        ...(manaExileIDs.length > 0 ? { exile_ids: manaExileIDs } : {}),
        ...counter,
      },
      viewerID ?? undefined,
    );
    manaDiscardIDs = [];
    manaExileIDs = [];
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
      handleManaAbilityCost(card, ability);
      return;
    }
    const params = { card_id: card.instance_id, ability_index: activate.index };
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
      } else {
        xAbilityPrompt = { card, ability, sacrificeIDs, crewIDs, counter };
        return;
      }
    }
    // CR 107.4f / CR 602.2b (#917, #916): the ability's Phyrexian
    // symbols. Same question the cast chain asks, in the same place —
    // after X, before the modes and the targets — and the same
    // stepper asks it, told to price the ABILITY's cost.
    if (
      phyrexianLife === undefined &&
      shouldAskPhyrexianLife(phyrexianSymbolsForAbility(ability), viewerLife)
    ) {
      phyrexianAbilityPrompt = { card, ability, sacrificeIDs, crewIDs, xValue, counter, modes };
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
      sacrifice_ids: sacrificeIDs,
      crew_ids: crewIDs,
      discard_ids: abilityDiscardIDs,
      // #1213: the return-to-hand picks, made at announce with the
      // rest of the cost and sent in the one activate_ability.
      return_ids: abilityReturnIDs,
      ...counter,
    };
    if (xValue !== undefined) params.x_value = xValue;
    // #916: omitted at 0, which is the server default.
    if (phyrexianLife) params.phyrexian_life = phyrexianLife;
    if (modes !== undefined) params.modes = modes;
    abilityDiscardIDs = [];
    abilityReturnIDs = [];
    abilitySacrificeX = undefined;
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
      (c) =>
        (c.kind === "pick_target" ||
          c.kind === "legend_rule" ||
          c.kind === "choose_protector" ||
          c.kind === "retarget") &&
        c.chooser === viewerID,
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
  // unit-tested; Board's job is only to gather the per-seat facts and
  // to own the one piece of state the decision cannot derive — the pin.
  //
  // The pin lives HERE rather than in a store because it is per-table
  // view state with no reason to outlive the component: a summary
  // pinned open in one game should not still be pinned when the next
  // game mounts.
  let pinnedSeatID = $state<string | null>(null);

  // Seats the viewer may declare an attack against. Computed once per
  // snapshot rather than per seat, because it reads the whole turn.
  const defenderIDs = $derived(legalDefenderIDs(view, viewerID));

  // A pin naming a seat that is no longer at the table (conceded,
  // eliminated and removed, or a different game entirely) would hold a
  // panel open for a player who is not there — and, worse, would be
  // unclearable, because the control that clears it lives on the
  // panel. Dropping it here keeps the invariant "a pin always has a
  // seat" without needing an effect to watch for departures.
  const pinned = $derived(
    pinnedSeatID && view.seats.some((s) => s.id === pinnedSeatID) ? pinnedSeatID : null,
  );

  function renderingFor(seat: PlayerView, pos: SeatPosition | null): SeatDecision {
    const controlled = cardsByController.get(seat.id) ?? [];
    return decideSeatRendering(
      {
        isSelf: pos === "self",
        isPinned: pinned === seat.id,
        isActiveSeat: seat.id === activeSeatID,
        controlsLegalTarget: seatControlsLegalTarget($targeting, seat.id, controlled),
        hasAttackersOnViewer: seatHasAttackersOn(viewerID, controlled),
        isLegalDefender: defenderIDs.has(seat.id),
      },
      { spectator: isSpectator, combatMode },
      {
        opponentDetail: $settings.display.opponentDetail,
        expandActivePlayer: $settings.display.expandActivePlayer,
      },
    );
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
</script>

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
              onExpand={() => (pinnedSeatID = nextPinnedSeat(pinned, seat.id))}
            />
          {:else}
            {#if decision.reason === "pinned"}
              <!-- The only way back. A pin is the one expansion the
                   viewer has to undo by hand — every other reason
                   clears itself when the prompt closes or the turn
                   moves on — and PlayerPanel has nowhere to put the
                   control, so it lives on the slot instead. -->
              <button
                class="unpin"
                type="button"
                aria-label={`Collapse ${seat.name}'s board back to a summary`}
                onclick={() => (pinnedSeatID = null)}
              >
                ⤡
              </button>
            {/if}
            <PlayerPanel
              {seat}
              isSelf={pos === "self"}
              flipped={$settings.display.tableLayout === "row" || opponentCount === 2
                ? pos !== "self"
                : pos === "across" || pos === "across_next"}
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
              {autopassEnabled}
              {loopNotice}
              {onPassPriority}
              {onToggleAutopass}
              onActivateAbility={handleActivateAbility}
              onManaAbilityCost={handleManaAbilityCost}
            />
          {/if}
        </div>
      {/if}
    {/each}
  {:else}
    <!-- Spectator path: uniform grid, all upright, equal real estate. -->
    {#each view.seats as seat (seat.id)}
      <div class="slot spectator-slot">
        <PlayerPanel
          {seat}
          isSelf={false}
          spectator={true}
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
        />
      </div>
    {/each}
  {/if}

  <CombatArrows {view} {boardEl} {beatsPrimeKey} />
  <HoverZoomOverlay {view} />
  <!-- Attention strip: one column over the table (the middle
       opponent's hand row in the row layout, the top-left seat's
       hand row otherwise) that stacks every live prompt — the stack
       card first, then whatever Game.svelte renders in `attention`. -->
  <div class="strip">
    <StackOverlay
      stack={view.stack}
      battlefield={view.battlefield}
      exile={view.exile}
      stackItems={view.stack_items ?? []}
      pendingTriggers={view.pending_triggers ?? []}
      seats={view.seats}
      viewerHasPriority={prioritySeatID === viewerID}
      priorityHolderName={view.seats[view.turn.priority_holder]?.name ?? null}
      splitSecondActive={view.split_second_active === true}
      onCounter={(item) => {
        const verb = item.kind === "spell" ? "counter_spell" : "counter_ability";
        guardedSendAction(verb, { instance_id: item.id });
      }}
      onTargetStackItem={(item) => completeTargetedCast("card", item.id)}
      onPass={onPassPriority}
    />
    {@render attention?.()}
  </div>
  <VotingPanel {view} {viewerID} sendAction={guardedSendAction} />
  <SacrificeCostModal
    source={sacrificePrompt?.card ?? null}
    label={sacrificePrompt?.ability.sacrifice_label ?? "a permanent"}
    options={sacrificeOptions}
    count={sacrificeBounds.max}
    min={sacrificeBounds.min}
    onConfirm={confirmSacrifice}
    onCancel={() => (sacrificePrompt = null)}
  />
  <CrewCostModal
    card={crewPrompt?.card ?? null}
    ability={crewPrompt?.ability ?? null}
    options={crewOptions}
    onConfirm={confirmCrew}
    onCancel={() => (crewPrompt = null)}
  />
  <CounterCostModal
    card={counterPrompt?.card ?? manaCounterPrompt?.card ?? null}
    ability={counterPrompt?.ability ?? manaCounterPrompt?.ability ?? null}
    board={view.battlefield.cards}
    onConfirm={confirmCounterCost}
    onCancel={() => {
      counterPrompt = null;
      manaCounterPrompt = null;
    }}
  />
  <FacePickerModal
    card={facePromptCard}
    onConfirm={confirmFace}
    onCancel={() => (facePromptCard = null)}
  />
  <AlternativeCostModal
    card={altCostPromptCard}
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
  <DiscardCostModal
    card={discardPromptCard}
    options={discardCostOptions}
    onConfirm={confirmDiscardCost}
    onCancel={() => {
      discardPromptCard = null;
      discardPromptChoices = {};
    }}
  />
  <!-- #660: the same picker, one cost site over — an activated
       ability's "Discard a creature card" (CR 602.2b). -->
  <DiscardCostModal
    card={abilityDiscardPrompt?.card ?? null}
    options={abilityDiscardOptions}
    need={abilityDiscardPrompt?.ability.discard_cost_n}
    label={abilityDiscardPrompt?.ability.discard_cost_label}
    onConfirm={confirmAbilityDiscardCost}
    onCancel={() => {
      abilityDiscardPrompt = null;
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
    note="exiled from your hand · not a discard"
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
  <SacrificeCostModal
    source={sacrificePromptCard}
    label={sacrificePromptLabel}
    options={castSacrificeOptions}
    count={sacrificeCount(sacrificePromptClause)}
    onConfirm={confirmSacrificeCost}
    onCancel={() => {
      sacrificePromptCard = null;
      sacrificePromptChoices = {};
    }}
  />
  <XCostModal
    gameID={view.id}
    card={xPromptCard}
    suggestedMax={suggestedX}
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
    confirmVerb="Activate"
    onConfirm={confirmAbilityX}
    onCancel={() => (xAbilityPrompt = null)}
  />
  <!-- CR 107.4c/f (#916): how many Phyrexian symbols the CAST pays
       with 2 life each. Between the X picker and the convoke picker,
       because it is announced with them and priced after X. -->
  <PhyrexianCostModal
    gameID={view.id}
    card={phyrexianPrompt?.card ?? null}
    symbols={phyrexianPrompt?.symbols ?? 0}
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
  <ModePickerModal
    card={modePromptCard}
    onConfirm={confirmModes}
    onCancel={() => {
      modePromptCard = null;
      modePromptChoices = {};
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
    />
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
      radial-gradient(60% 55% at 50% 50%, rgba(217, 180, 92, 0.05) 0%, rgba(0, 0, 0, 0) 60%),
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
  .board-disabled .slot {
    pointer-events: none;
  }
  .slot {
    min-height: 0;
    min-width: 0;
    /* Anchors .unpin. Nothing else in a slot is positioned, so this
       costs nothing until a seat is pinned. */
    position: relative;
  }

  /* The collapse control on a pinned panel. Sits above the panel's own
     chrome (PlayerPanel's rail is z-index 3) but under the attention
     strip (40) and the hover zoom, because a prompt covering this
     button is strictly better than this button covering a prompt. */
  .unpin {
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
  }
  .unpin:hover {
    color: var(--fg);
    border-color: var(--border-strong);
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
  :global(:root[data-table-layout="row"]) .board[data-opp-count="3"] .strip {
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
