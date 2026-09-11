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
    ManaAbilityView,
    ZoneView,
  } from "../../protocol";
  import { seatPlacements, type SeatPosition } from "../../cardTypes";
  import PlayerPanel from "./PlayerPanel.svelte";
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
  import CardContextMenu from "./CardContextMenu.svelte";
  import { cardMenu, closeCardMenu } from "../../contextMenu";
  import type { MenuActivate } from "../../contextMenu.logic";
  import {
    targeting,
    begin as beginTargeting,
    cancel as cancelTargeting,
    beginChoice as beginTargetingChoice,
    beginForAbility as beginTargetingForAbility,
    beginForMode as beginTargetingForMode,
    hasXCost,
    isModal,
    discardCostOf,
    sacrificeCostOptions,
    tapCostOf,
    tapCostLimit,
    alternativeCostsOf,
    alternativeCostByKey,
    applyCastChoices,
    isLegalCardTarget,
    isLegalPlayerTarget,
    isMultiPick,
    togglePick,
    canConfirm,
    setConfirmHandler,
    type CastChoices,
    type CastSourceZone,
    type TargetingMode,
    type TargetingState,
    type TargetRef,
  } from "../../targeting";
  import XCostModal from "./XCostModal.svelte";
  import SacrificeCostModal from "./SacrificeCostModal.svelte";
  import ModePickerModal from "./ModePickerModal.svelte";
  import DiscardCostModal from "./DiscardCostModal.svelte";
  import AlternativeCostModal from "./AlternativeCostModal.svelte";
  import FacePickerModal from "./FacePickerModal.svelte";
  import { cardAsFace, needsFacePicker } from "../../faces";
  import TapCostModal from "./TapCostModal.svelte";

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
    // Priority controls forwarded to the self-panel's PhaseDisplay.
    autopassEnabled?: boolean;
    onPassPriority?: () => void;
    onToggleAutopass?: () => void;
    // Game.svelte's live prompts (targeting, combat hint, mulligan
    // roll-call, toasts, game end) render inside the attention strip
    // under the stack card so every "look here" surface shares one
    // anchor over the table.
    attention?: Snippet;
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
    autopassEnabled,
    onPassPriority,
    onToggleAutopass,
    attention,
  }: Props = $props();

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

  const activeSeatID = $derived(view.seats[view.turn.active_seat]?.id ?? null);
  const prioritySeatID = $derived(view.seats[view.turn.priority_holder]?.id ?? null);
  const monarchID = $derived(view.monarch ?? null);
  const initiativeID = $derived(view.initiative ?? null);

  function handleTapToggle(card: CardView): void {
    sendAction(card.tapped ? "untap" : "tap", { instance_id: card.instance_id });
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
  let xPromptChoices: CastChoices = {};
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
  function confirmAltCost(key: string | undefined): void {
    const card = altCostPromptCard;
    const choices = altCostPromptChoices;
    altCostPromptCard = null;
    altCostPromptChoices = {};
    if (!card) return;
    afterAltCost(card, key === undefined ? choices : { ...choices, altCost: key });
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

  const castSacrificeOptions = $derived.by(() => {
    const card = sacrificePromptCard;
    if (!card) return [];
    const ids = new Set(sacrificeCostOptions(card) ?? []);
    return view.battlefield.cards.filter((c) => ids.has(c.instance_id));
  });

  function confirmSacrificeCost(instanceID: string): void {
    const card = sacrificePromptCard;
    const choices = sacrificePromptChoices;
    sacrificePromptCard = null;
    sacrificePromptChoices = {};
    if (!card) return;
    afterCastCosts(card, { ...choices, sacrificeIDs: [instanceID] });
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

  // afterAltCost / afterDiscardCost / afterCastCosts are the seams
  // between the cost prompts and the rest of the cast flow, so adding
  // a cost kind doesn't mean editing every earlier prompt's confirm.
  //
  // An alternative cost REPLACES the mana cost but not the additional
  // costs (CR 601.2f is evaluated independently), so the chain
  // continues through the discard / sacrifice prompts rather than
  // short-circuiting past them.
  function afterAltCost(card: CardView, choices: CastChoices): void {
    if (discardCostOf(card) > 0) {
      discardPromptChoices = choices;
      discardPromptCard = card;
      return;
    }
    afterDiscardCost(card, choices);
  }

  function afterDiscardCost(card: CardView, choices: CastChoices): void {
    if (sacrificeCostOptions(card) !== undefined) {
      sacrificePromptChoices = choices;
      sacrificePromptCard = card;
      return;
    }
    afterCastCosts(card, choices);
  }

  function afterCastCosts(card: CardView, choices: CastChoices): void {
    if (hasXCost(card)) {
      xPromptChoices = choices;
      xPromptCard = card;
      return;
    }
    afterXCost(card, choices);
  }

  // afterXCost is the seam between the X prompt and targeting, and
  // the only reason the tap picker isn't in afterCastCosts with the
  // others: a waterbend {X} cost can't size its picker until X is
  // known, so this step has to come after the X prompt rather than
  // before it.
  function afterXCost(card: CardView, choices: CastChoices): void {
    if (tapCostOf(card)) {
      tapPromptChoices = choices;
      tapPromptCard = card;
      return;
    }
    continueCast(card, choices);
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
    if (face <= 0) {
      afterFace(card, fromZone ? { fromZone } : {});
      return;
    }
    // Run the rest of the chain against the CHOSEN face, so the
    // prompts and the cast-timing checks see its type line and cost
    // rather than the front's. cardAsFace drops the announce-prompt
    // fields, which describe face 0's catalog spec and would be
    // wrong here — see the note on cardAsFace.
    afterFace(cardAsFace(card, face), fromZone ? { face, fromZone } : { face });
  }

  function afterFace(card: CardView, choices: CastChoices): void {
    if (alternativeCostsOf(card).length > 0) {
      altCostPromptChoices = choices;
      altCostPromptCard = card;
      return;
    }
    afterAltCost(card, choices);
  }

  // handlePlayCard is the head of the chain. `fromZone` is undefined
  // for the hand, which is every cast the board's own surfaces fire;
  // S29's zone browser passes "graveyard" so a flashback cast walks
  // the identical prompt chain and lands the zone on the payload via
  // CastChoices.
  //
  // The face picker's confirm re-enters at afterFace with its own
  // choices object, so the zone has to be seeded here rather than at
  // the end — otherwise a modal DFC cast out of the graveyard would
  // lose it.
  function handlePlayCard(card: CardView, fromZone?: CastSourceZone): void {
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
  function confirmModes(modes: number[]): void {
    const card = modePromptCard;
    const choices = modePromptChoices;
    modePromptCard = null;
    modePromptChoices = {};
    if (!card) return;
    const targeted = modes.find((i) => card.modes?.options[i]?.legal_targets !== undefined);
    if (targeted !== undefined) {
      const option = card.modes!.options[targeted];
      beginTargetingForMode(card, option, modes, choices);
      return;
    }
    const params: Record<string, unknown> = { instance_id: card.instance_id, modes };
    applyCastChoices(params, choices);
    sendAction("cast_spell", params, viewerID ?? undefined);
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
    const mode = (alt ? alt.target_mode : card.target_mode) as TargetingMode | undefined;
    if (
      mode === "any" ||
      mode === "player" ||
      mode === "creature" ||
      mode === "permanent" ||
      mode === "stack_spell" ||
      mode === "card_in_graveyard"
    ) {
      beginTargeting(card, mode, choices, alt);
      return;
    }
    const params: Record<string, unknown> = { instance_id: card.instance_id };
    applyCastChoices(params, choices);
    sendAction("cast_spell", params, viewerID ?? undefined);
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
    fireTargets(state, [ref]);
  }

  function confirmTargets(): void {
    const state = $targeting;
    if (!state || !canConfirm(state)) return;
    fireTargets(state, state.picked);
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
      sendAction("resolve_choice", { choice_id: state.choiceID, targets }, viewerID ?? undefined);
      targeting.set(null);
      return;
    }
    // S21 sub-PR 2: the prompt belongs to an activated ability, not
    // a cast — its cost was already paid at announce.
    if (state.ability) {
      sendAction(
        "activate_ability",
        {
          source_card_id: state.card.instance_id,
          ability_index: state.ability.index,
          sacrifice_ids: state.ability.sacrificeIDs,
          targets,
        },
        viewerID ?? undefined,
      );
      targeting.set(null);
      return;
    }
    const params: Record<string, unknown> = { instance_id: state.card.instance_id, targets };
    if (state.modes !== undefined) params.modes = state.modes;
    applyCastChoices(params, state.choices);
    sendAction("cast_spell", params, viewerID ?? undefined);
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
    const ids = new Set(p.ability.sacrifice_options?.cards ?? []);
    return view.battlefield.cards.filter((c) => ids.has(c.instance_id));
  });

  function handleActivateAbility(card: CardView, index: number): void {
    const ability = (card.activated_abilities ?? []).find((a) => a.index === index);
    if (!ability) return;
    if (ability.sacrifice_options) {
      sacrificePrompt = { kind: "ability", card, ability };
      return;
    }
    continueActivation(card, ability, []);
  }

  // S21: a mana ability with a sacrifice-another cost, handed up by
  // PlayerPanel because the picker is board-wide.
  function handleManaSacrificeCost(card: CardView, ability: ManaAbilityView): void {
    sacrificePrompt = { kind: "mana", card, ability };
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
    if (ability?.sacrifice_options) {
      handleManaSacrificeCost(card, ability);
      return;
    }
    const params = { card_id: card.instance_id, ability_index: activate.index };
    sendAction("activate_mana_ability", params, card.controller);
  }

  function confirmSacrifice(instanceID: string): void {
    const p = sacrificePrompt;
    sacrificePrompt = null;
    if (!p) return;
    if (p.kind === "mana") {
      sendAction(
        "activate_mana_ability",
        {
          card_id: p.card.instance_id,
          ability_index: p.ability.index,
          sacrifice_ids: [instanceID],
        },
        viewerID ?? undefined,
      );
      return;
    }
    continueActivation(p.card, p.ability, [instanceID]);
  }

  // continueActivation is the post-cost half: enter targeting for an
  // ability that targets, or fire straight away.
  function continueActivation(
    card: CardView,
    ability: ActivatedAbilityView,
    sacrificeIDs: string[],
  ): void {
    if (ability.legal_targets) {
      beginTargetingForAbility(card, ability, sacrificeIDs);
      return;
    }
    sendAction(
      "activate_ability",
      {
        source_card_id: card.instance_id,
        ability_index: ability.index,
        sacrifice_ids: sacrificeIDs,
      },
      viewerID ?? undefined,
    );
  }

  // S20 sub-PR 2: a pick_target pending choice addressed to the
  // viewer drives the same targeting UI a cast does. Enter it when
  // one appears; leave it when it's gone (answered, or resolved
  // elsewhere). A cast prompt already in flight is replaced — the
  // trigger's target is owed first.
  $effect(() => {
    // S27: the legend rule is answered through the same flow — it is
    // the same question shape (pick one from a server-computed set)
    // and the banner's confirm sends the same payload. The server
    // routes by the choice's kind, so the client needs no second
    // component and no second code path.
    const mine = (view.pending_choices ?? []).find(
      (c) => (c.kind === "pick_target" || c.kind === "legend_rule") && c.chooser === viewerID,
    );
    const cur = $targeting;
    if (mine) {
      if (cur?.choiceID === mine.id) return;
      const source = findCardAnywhere(mine.source) ?? {
        instance_id: mine.source ?? "",
        name: mine.kind === "legend_rule" ? "Legend rule" : "Triggered ability",
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
    for (const c of view.exile?.cards ?? []) if (c.instance_id === id) return c;
    for (const s of view.seats) {
      for (const c of s.graveyard?.cards ?? []) if (c.instance_id === id) return c;
      for (const c of s.hand?.cards ?? []) if (c.instance_id === id) return c;
    }
    return undefined;
  }

  function handleDrawCard(): void {
    if (!viewerID) return;
    sendAction("draw_card", undefined, viewerID);
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
  data-opp-count={opponentCount}
  data-seat-count={view.seats.length}
  bind:this={boardEl}
>
  {#if placements}
    {#each positions as pos (pos)}
      {@const seat = placements[pos]}
      {#if seat}
        <div class="slot" data-pos={pos} style:grid-area={pos}>
          <PlayerPanel
            {seat}
            isSelf={pos === "self"}
            flipped={$settings.display.tableLayout === "row"
              ? pos !== "self"
              : pos === "across" || pos === "across_next"}
            isActive={seat.id === activeSeatID}
            hasPriority={seat.id === prioritySeatID}
            {viewerID}
            {isAdmin}
            {sendAction}
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
            {onPassPriority}
            {onToggleAutopass}
            onActivateAbility={handleActivateAbility}
            onManaSacrificeCost={handleManaSacrificeCost}
          />
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
          isActive={seat.id === activeSeatID}
          hasPriority={seat.id === prioritySeatID}
          {viewerID}
          {isAdmin}
          {sendAction}
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

  <CombatArrows {view} {boardEl} />
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
        sendAction(verb, { instance_id: item.id });
      }}
      onTargetStackItem={(item) => completeTargetedCast("card", item.id)}
      onPass={onPassPriority}
    />
    {@render attention?.()}
  </div>
  <VotingPanel {view} {viewerID} {sendAction} />
  <SacrificeCostModal
    source={sacrificePrompt?.card ?? null}
    label={sacrificePrompt?.ability.sacrifice_label ?? "a permanent"}
    options={sacrificeOptions}
    onConfirm={confirmSacrifice}
    onCancel={() => (sacrificePrompt = null)}
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
  <DiscardCostModal
    card={discardPromptCard}
    options={discardCostOptions}
    onConfirm={confirmDiscardCost}
    onCancel={() => {
      discardPromptCard = null;
      discardPromptChoices = {};
    }}
  />
  <SacrificeCostModal
    source={sacrificePromptCard}
    label={sacrificePromptCard?.additional_cost?.label ?? "a permanent"}
    options={castSacrificeOptions}
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
    onConfirm={confirmX}
    onCancel={() => {
      xPromptCard = null;
      xPromptChoices = {};
    }}
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
  {#if $zoneBrowser}
    <ZoneBrowserModal
      {view}
      {viewerID}
      zoneKind={$zoneBrowser.zoneKind}
      ownerSeat={{ id: $zoneBrowser.ownerID, name: $zoneBrowser.ownerName }}
      {sendAction}
      onClose={closeZoneBrowser}
      onTargetCard={handleTargetCard}
      onCastCard={handlePlayCard}
    />
  {/if}
  {#if $cardMenu}
    <CardContextMenu
      {view}
      {viewerID}
      {isAdmin}
      open={$cardMenu}
      {sendAction}
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
  .slot {
    min-height: 0;
    min-width: 0;
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
    /* "across" spans the full top row so the across player feels
       primary the same way they do in a 2-player game. The bottom
       splits between next (left) and self (right). */
    grid-template-areas:
      "across across"
      "next   self";
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
  :global(:root[data-table-layout="row"]) .board[data-opp-count="2"] {
    grid-template-columns: 1fr 1fr;
    grid-template-areas:
      "next across"
      "self self";
  }
  :global(:root[data-table-layout="row"]) .board[data-opp-count="3"] {
    grid-template-columns: 1fr 1fr 1fr;
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
