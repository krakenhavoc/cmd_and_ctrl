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

  import type {
    ActionPayload,
    ActionType,
    ActivatedAbilityView,
    CardView,
    GameView,
    ZoneView,
  } from "../../protocol";
  import { seatPlacements, type SeatPosition } from "../../cardTypes";
  import PlayerPanel from "./PlayerPanel.svelte";
  import HoverZoomOverlay from "./HoverZoomOverlay.svelte";
  // CommanderDamageTooltip was folded into HoverZoomOverlay — the
  // damage readout now lives inside the card preview panel instead
  // of the lower-right corner.
  import StackOverlay from "./StackOverlay.svelte";
  import CombatArrows from "./CombatArrows.svelte";
  import VotingPanel from "./VotingPanel.svelte";
  import ZoneBrowserModal from "./ZoneBrowserModal.svelte";
  import { zoneBrowser, closeZoneBrowser } from "../../zoneBrowser";
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
    isLegalCardTarget,
    isLegalPlayerTarget,
    isMultiPick,
    togglePick,
    canConfirm,
    setConfirmHandler,
    type TargetingMode,
    type TargetingState,
    type TargetRef,
  } from "../../targeting";
  import XCostModal from "./XCostModal.svelte";
  import SacrificeCostModal from "./SacrificeCostModal.svelte";
  import ModePickerModal from "./ModePickerModal.svelte";
  import DiscardCostModal from "./DiscardCostModal.svelte";

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

  // S20 sub-PR 3: an {X} spell asks for X first. The modal's
  // confirm continues into targeting / cast with the chosen value.
  let xPromptCard = $state<CardView | null>(null);
  let xPromptDiscardIDs: string[] | undefined;
  function confirmX(x: number): void {
    const card = xPromptCard;
    const discardIDs = xPromptDiscardIDs;
    xPromptCard = null;
    xPromptDiscardIDs = undefined;
    if (!card) return;
    continueCast(card, x, discardIDs);
  }

  // S21 sub-PR 5: a spell with an additional cost ("As an
  // additional cost to cast this spell, discard a card") asks for
  // the payment first. The picked IDs thread through the rest of
  // the cast flow the way xValue does and ride cast_spell as
  // discard_ids; the server validates them at announce and pays
  // them once the spell is on the stack.
  let discardPromptCard = $state<CardView | null>(null);

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
    discardPromptCard = null;
    if (!card) return;
    if (hasXCost(card)) {
      xPromptDiscardIDs = ids;
      xPromptCard = card;
      return;
    }
    continueCast(card, undefined, ids);
  }

  function handlePlayCard(card: CardView): void {
    if (discardCostOf(card) > 0) {
      discardPromptCard = card;
      return;
    }
    if (hasXCost(card)) {
      xPromptCard = card;
      return;
    }
    continueCast(card, undefined);
  }

  // S20 sub-PR 4: a modal spell asks for its mode(s) after X and
  // before targeting. The picker's confirm continues with the
  // chosen indexes; a targeted option enters targeting with that
  // option's legal set, otherwise the cast fires straight away.
  let modePromptCard = $state<CardView | null>(null);
  let modePromptX: number | undefined;
  let modePromptDiscardIDs: string[] | undefined;
  function confirmModes(modes: number[]): void {
    const card = modePromptCard;
    const xValue = modePromptX;
    const discardIDs = modePromptDiscardIDs;
    modePromptCard = null;
    modePromptX = undefined;
    modePromptDiscardIDs = undefined;
    if (!card) return;
    const targeted = modes.find((i) => card.modes?.options[i]?.legal_targets !== undefined);
    if (targeted !== undefined) {
      const option = card.modes!.options[targeted];
      beginTargetingForMode(card, option, modes, xValue, discardIDs);
      return;
    }
    const params: Record<string, unknown> = { instance_id: card.instance_id, modes };
    if (xValue !== undefined) params.x_value = xValue;
    if (discardIDs !== undefined) params.discard_ids = discardIDs;
    sendAction("cast_spell", params, viewerID ?? undefined);
  }

  // continueCast is the post-X half of the cast flow: pick modes for
  // a modal card, enter targeting for a targeted card, or fire
  // cast_spell straight away.
  function continueCast(card: CardView, xValue: number | undefined, discardIDs?: string[]): void {
    if (isModal(card)) {
      modePromptX = xValue;
      modePromptDiscardIDs = discardIDs;
      modePromptCard = card;
      return;
    }
    // S14: if the card declares a target_mode (catalog cards with
    // a target slot — Lightning Bolt, Counterspell), enter the
    // targeting flow and wait for a second click on a legal target.
    // Otherwise fire cast_spell immediately (lands, sorceries with
    // no targets, vanilla permanents).
    const mode = card.target_mode as TargetingMode | undefined;
    if (
      mode === "any" ||
      mode === "player" ||
      mode === "creature" ||
      mode === "permanent" ||
      mode === "stack_spell" ||
      mode === "card_in_graveyard"
    ) {
      beginTargeting(card, mode, xValue, discardIDs);
      return;
    }
    const params: Record<string, unknown> = { instance_id: card.instance_id };
    if (xValue !== undefined) params.x_value = xValue;
    if (discardIDs !== undefined) params.discard_ids = discardIDs;
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
    if (state.xValue !== undefined) params.x_value = state.xValue;
    if (state.modes !== undefined) params.modes = state.modes;
    if (state.discardIDs !== undefined) params.discard_ids = state.discardIDs;
    sendAction("cast_spell", params, viewerID ?? undefined);
    cancelTargeting();
  }

  // --- S21 sub-PR 2: activated abilities ------------------------
  //
  // Order of operations mirrors CR 601.2 / 602.2: choose the mode
  // (n/a today), pay the costs, then choose targets. So a sacrifice
  // cost is picked BEFORE targeting, and both are sent together in
  // one activate_ability — the server pays and announces atomically.
  let sacrificePrompt = $state<{
    card: CardView;
    ability: ActivatedAbilityView;
  } | null>(null);

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
      sacrificePrompt = { card, ability };
      return;
    }
    continueActivation(card, ability, []);
  }

  function confirmSacrifice(instanceID: string): void {
    const p = sacrificePrompt;
    sacrificePrompt = null;
    if (!p) return;
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
    const mine = (view.pending_choices ?? []).find(
      (c) => c.kind === "pick_target" && c.chooser === viewerID,
    );
    const cur = $targeting;
    if (mine) {
      if (cur?.choiceID === mine.id) return;
      const source = findCardAnywhere(mine.source) ?? {
        instance_id: mine.source ?? "",
        name: "Triggered ability",
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
  <StackOverlay
    stack={view.stack}
    battlefield={view.battlefield}
    exile={view.exile}
    stackItems={view.stack_items ?? []}
    pendingTriggers={view.pending_triggers ?? []}
    seats={view.seats}
    viewerHasPriority={prioritySeatID === viewerID}
    splitSecondActive={view.split_second_active === true}
    onCounter={(item) => {
      const verb = item.kind === "spell" ? "counter_spell" : "counter_ability";
      sendAction(verb, { instance_id: item.id });
    }}
    onTargetStackItem={(item) => completeTargetedCast("card", item.id)}
  />
  <VotingPanel {view} {viewerID} {sendAction} />
  <SacrificeCostModal
    source={sacrificePrompt?.card ?? null}
    label={sacrificePrompt?.ability.sacrifice_label ?? "a permanent"}
    options={sacrificeOptions}
    onConfirm={confirmSacrifice}
    onCancel={() => (sacrificePrompt = null)}
  />
  <DiscardCostModal
    card={discardPromptCard}
    options={discardCostOptions}
    onConfirm={confirmDiscardCost}
    onCancel={() => (discardPromptCard = null)}
  />
  <XCostModal
    gameID={view.id}
    card={xPromptCard}
    suggestedMax={suggestedX}
    onConfirm={confirmX}
    onCancel={() => (xPromptCard = null)}
  />
  <ModePickerModal
    card={modePromptCard}
    onConfirm={confirmModes}
    onCancel={() => {
      modePromptCard = null;
      modePromptX = undefined;
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
      radial-gradient(60% 55% at 50% 50%, rgba(80, 120, 220, 0.08) 0%, rgba(6, 10, 20, 0) 60%),
      radial-gradient(120% 100% at 50% 100%, rgba(10, 15, 35, 0), rgba(2, 4, 10, 0.6) 70%),
      radial-gradient(120% 100% at 50% 0%, rgba(10, 15, 35, 0), rgba(2, 4, 10, 0.6) 70%),
      linear-gradient(180deg, #070c18 0%, #0a1122 60%, #050811 100%);
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
    grid-template-areas:
      "across_next across"
      "next        self";
  }

  /* Top-row opponents are flipped so the layout reads "across the
     table" — their hand strip ends up at the top of the screen
     (closest to them) and their header strip ends up at the bottom
     of the slot (closest to the centre, and the viewer). Bottom-row
     opponents stay upright; the user's mental model is "this player
     sits next to me, facing the same direction I do." */
  .slot[data-pos="across"] :global(.panel),
  .slot[data-pos="across_next"] :global(.panel) {
    transform: rotate(180deg);
    transform-origin: center;
  }

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
