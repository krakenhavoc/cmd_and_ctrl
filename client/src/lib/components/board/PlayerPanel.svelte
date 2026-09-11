<script lang="ts">
  // PlayerPanel is one player's full board, laid out in a CSS Grid
  // matching the S16.5 avatar-centric redesign:
  //
  //   ┌─────────────────────────────────────────────┐
  //   │  creatures (full width)                     │
  //   ├──────────────────────┬──────────────────────┤
  //   │  lands               │  enchant / artifact  │
  //   ├─────────────────────────────────────────────┤
  //   │  ⭕ piles  hand (peek) ……………… phases (self)│
  //   └─────────────────────────────────────────────┘
  //
  // The bottombar is a single flex row: avatar anchors bottom-left,
  // piles immediately beside it, the hand peeks up from the same base-
  // line (top ~55% of each card visible, bottom-clipped at the pile
  // bottom), and PhaseDisplay floats bottom-right on the viewer's own
  // panel. Hovering the self hand lifts the whole fan up over the board
  // to reveal full cards.
  //
  // The same component is reused for self + opponents; Board.svelte
  // wraps opponent instances in a transform container that rotates
  // them 180° across the top so they read "across the table". The
  // panel itself is layout-only and doesn't know whether it's rotated.
  //
  // Click routing for battlefield cards lives here so combat vs.
  // tap-toggle logic stays in one place. The router mirrors the old
  // Pixi wireTapClick: combat select on your own creature, declare-
  // block on an incoming attacker, otherwise tap/untap — except for
  // planeswalkers, whose click opens the card menu (#329). The
  // decision itself is battlefieldClickIntent, in contextMenu.logic,
  // so it is testable without rendering Svelte.

  import type {
    ActionPayload,
    ActionType,
    CardView,
    GameView,
    ManaAbilityView,
    PlayerView,
    ZoneView,
  } from "../../protocol";
  import { bucketForBattlefield, isCreature } from "../../cardTypes";
  import { battlefieldClickIntent } from "../../contextMenu.logic";
  import { sorcerySpeedWindowOpen } from "../../timing";
  import { openCardMenu } from "../../contextMenu";
  import BattlefieldRow from "./BattlefieldRow.svelte";
  import PileBar from "./PileBar.svelte";
  import Hand from "./Hand.svelte";
  import PlayerIdentity from "./PlayerIdentity.svelte";
  import PhaseDisplay from "./PhaseDisplay.svelte";
  import PromisesRow from "./PromisesRow.svelte";

  type ActionSender = (type: ActionType, params?: ActionPayload["params"], player?: string) => void;

  interface Props {
    seat: PlayerView;
    isSelf: boolean;
    isActive: boolean;
    hasPriority: boolean;
    viewerID: string | null;
    isAdmin: boolean;
    sendAction: ActionSender;
    isMonarch: boolean;
    isInitiative: boolean;
    view: GameView;
    // Battlefield slice already filtered to cards with controller === seat.id
    controlledCards: CardView[];
    // Player-owned slice of the shared exile zone (filtered by Board)
    exile: ZoneView;
    combatMode: "idle" | "attack" | "block";
    selectedCombatCardID: string | null;
    onSelectCombatCard: (cardID: string) => void;
    onDeclareAttack: (targetPlayerID: string) => void;
    onDeclareBlock: (attackerCardID: string) => void;
    onTapToggle: (card: CardView) => void;
    onPlayCard: (card: CardView) => void;
    onDrawCard: () => void;
    onTargetPlayer?: (targetPlayerID: string) => void;
    // onTargetCard returns true when a cast-targeting prompt
    // consumed the click (so the caller stops propagating into
    // the tap-toggle / combat default). Returns false when no
    // prompt is active or the card isn't a legal target.
    onTargetCard?: (card: CardView) => boolean;
    // S21 sub-PR 2: a CR 602 activated ability on one of this
    // seat's permanents was chosen from the card menu. Board owns
    // the follow-up (sacrifice pick, targeting) because those are
    // board-wide modals. Only wired for the viewer's own panel.
    onActivateAbility?: (card: CardView, abilityIndex: number) => void;
    // S21: a mana ability on one of this seat's permanents needs a
    // sacrifice chosen before it can be activated. Board owns that
    // modal, so the panel forwards the click instead of sending the
    // action. Only wired for the viewer's own panel.
    onManaSacrificeCost?: (card: CardView, ability: ManaAbilityView) => void;
    // Priority controls forwarded to PhaseDisplay — only the
    // self panel mounts the widget, so these only matter when
    // isSelf=true but they're plumbed uniformly for prop typing.
    autopassEnabled?: boolean;
    onPassPriority?: () => void;
    onToggleAutopass?: () => void;
    // flipped — top-row opponents. The panel keeps its zones in the
    // same grid but reverses the row order (hand at the top edge,
    // creatures toward the table centre) instead of rotating 180°,
    // so text and card art stay upright.
    flipped?: boolean;
  }

  const {
    seat,
    isSelf,
    isActive,
    hasPriority,
    viewerID,
    isAdmin,
    sendAction,
    isMonarch,
    isInitiative,
    view,
    controlledCards,
    exile,
    combatMode,
    selectedCombatCardID,
    onSelectCombatCard,
    onDeclareAttack,
    onDeclareBlock,
    onTapToggle,
    onPlayCard,
    onDrawCard,
    onTargetPlayer,
    onTargetCard,
    autopassEnabled = false,
    onPassPriority,
    onToggleAutopass,
    onActivateAbility,
    onManaSacrificeCost,
    flipped = false,
  }: Props = $props();

  // The seat's commander, wherever it is right now: the command zone
  // first, then the shared zones (battlefield, stack, exile) and its
  // graveyard. Feeds the art-crop avatar fallback in PlayerIdentity.
  const commanderScryfallID = $derived.by((): string | null => {
    const inCommand = seat.command?.cards?.find((c) => c.scryfall_id);
    if (inCommand?.scryfall_id) return inCommand.scryfall_id;
    const zones = [view.battlefield, view.stack, view.exile, seat.graveyard];
    for (const z of zones) {
      const hit = z?.cards?.find(
        (c) => c.is_commander && (c.owner === seat.id || c.controller === seat.id) && c.scryfall_id,
      );
      if (hit?.scryfall_id) return hit.scryfall_id;
    }
    return null;
  });

  // S31: the CR 307.1 sorcery-speed window, derived once per panel
  // and handed down to every Card so the ability popover can grey an
  // "activate only as a sorcery" row. The flag has ridden the wire as
  // ActivatedAbilityView.sorcery_speed since S21 and nothing read it,
  // so those abilities stayed clickable through combat and an
  // opponent's turn and came back rejected — the live example
  // ADR 0033 §1 cites for why the client stopped re-deriving timing.
  //
  // Empty string means "open, no opinion"; opponents' panels are
  // never gated on the VIEWER's window, so they get "" too.
  const sorcerySpeedBlocked = $derived(
    isSelf ? (sorcerySpeedWindowOpen(view, viewerID).reason ?? "") : "",
  );

  const buckets = $derived.by(() => {
    const out = { creature: [] as CardView[], land: [] as CardView[], right: [] as CardView[] };
    for (const c of controlledCards) {
      out[bucketForBattlefield(c)].push(c);
    }
    return out;
  });

  // S15: mana-ability activation handler for battlefield permanents
  // the viewer controls. Only installed on the viewer's own panel;
  // opponent panels pass undefined down so the context menu stays
  // closed on cards they don't control.
  //
  // S21: a mana ability whose cost sacrifices ANOTHER permanent
  // (Ashnod's Altar) needs a choice first, and that modal is
  // board-wide — so the click is handed to Board, which owns the
  // same SacrificeCostModal the CR 602 abilities use and sends the
  // action itself once a card is picked.
  const activateManaAbility = $derived(
    isSelf
      ? (card: CardView, abilityIndex: number) => {
          const ability = (card.mana_abilities ?? []).find((a) => a.index === abilityIndex);
          if (ability?.sacrifice_options && onManaSacrificeCost) {
            onManaSacrificeCost(card, ability);
            return;
          }
          sendAction(
            "activate_mana_ability",
            { card_id: card.instance_id, ability_index: abilityIndex },
            seat.id,
          );
        }
      : undefined,
  );

  const attackTargetable = $derived(
    !isSelf && !seat.eliminated && combatMode === "attack" && !!selectedCombatCardID,
  );

  function handleCardClick(card: CardView, ev?: MouseEvent): void {
    // The S17 sub-PR 5 Shift+click / Shift+Alt+click +1/+1 debug
    // chord used to live here. #170 retired it: the right-click
    // override menu offers both directions on every counter type
    // (and every other manual override) from a surface the player
    // can find without being told it exists. Enable it under
    // Settings → Gameplay → "Enable admin overrides".
    //
    // S14: targeting intercept. If a cast-targeting prompt is live
    // and this card is a legal target (battlefield creature for
    // "any" / "creature" modes), route through onTargetCard. Board
    // clears the targeting state when cast_spell fires.
    if (onTargetCard) {
      // onTargetCard itself checks the targeting store; only call
      // when a prompt is waiting. Board wires this to the
      // targeting-aware completion.
      // To avoid double-handling, we check the targeting store here
      // via a light dynamic import — but easier: the callback
      // returns a boolean "handled" signal. Lacking that, just
      // call unconditionally: Board's handler is a no-op when no
      // prompt is active.
      if (onTargetCard(card)) return;
    }
    // Combat select on viewer's own creature wins ahead of tap/untap.
    if (
      (combatMode === "attack" || combatMode === "block") &&
      card.controller === viewerID &&
      isCreature(card)
    ) {
      onSelectCombatCard(card.instance_id);
      return;
    }
    // Block-mode click on an incoming attacker commits the block.
    if (combatMode === "block" && card.attacking_target === viewerID) {
      onDeclareBlock(card.instance_id);
      return;
    }
    // #329: a planeswalker's click means "which loyalty ability?",
    // not "tap it". The card menu is the surface that already
    // renders activated abilities (ADR 0028), and it carries the
    // manual loyalty rows for the planeswalkers with no catalog
    // entry — which is still most of them.
    switch (battlefieldClickIntent(card, viewerID, isAdmin)) {
      case "none":
        return;
      case "abilities":
        openCardMenu({ card, x: ev?.clientX ?? 0, y: ev?.clientY ?? 0 });
        return;
      case "tap":
        onTapToggle(card);
    }
  }
</script>

<div
  class="panel"
  class:self={isSelf}
  class:opponent={!isSelf}
  class:flipped
  role="region"
  aria-label={isSelf ? "your board" : `${seat.name} board`}
>
  <div class="grid-creatures">
    <BattlefieldRow
      label="creatures"
      cards={buckets.creature}
      {viewerID}
      {selectedCombatCardID}
      onCardClick={handleCardClick}
      onActivateManaAbility={activateManaAbility}
      {onActivateAbility}
      {sorcerySpeedBlocked}
    />
  </div>
  <div class="grid-middle">
    <BattlefieldRow
      label="enchant / artifact"
      cards={buckets.right}
      compact
      {viewerID}
      {selectedCombatCardID}
      onCardClick={handleCardClick}
      onActivateManaAbility={activateManaAbility}
      {onActivateAbility}
      {sorcerySpeedBlocked}
    />
    <BattlefieldRow
      label="lands"
      cards={buckets.land}
      compact
      strip
      {viewerID}
      {selectedCombatCardID}
      onCardClick={handleCardClick}
      onActivateManaAbility={activateManaAbility}
      {onActivateAbility}
      {sorcerySpeedBlocked}
    />
  </div>
  <div class="grid-bottom">
    <!-- Hand sits inline with the phase widget; its clipped-bottom
         line coincides with the panel edge. flex: 1 lets it absorb
         the width the rail freed up. -->
    <div class="hand-zone">
      <Hand
        hand={seat.hand}
        {isSelf}
        onPlayCard={isSelf ? onPlayCard : undefined}
        snap={view}
        {viewerID}
      />
    </div>
    {#if !isSelf}
      <PromisesRow {view} {viewerID} opponentID={seat.id} {sendAction} />
    {/if}
    {#if isSelf && view.turn}
      <PhaseDisplay
        turn={view.turn}
        seats={view.seats}
        mulligansOpen={view.mulligans_open === true}
        viewerHasPriority={hasPriority}
        {autopassEnabled}
        onPassPriority={onPassPriority ?? (() => {})}
        onToggleAutopass={onToggleAutopass ?? (() => {})}
      />
    {/if}
  </div>
  <!-- The rail is the player card: identity, floating mana and
       markers on top, the four piles below. It frees the bottom row
       for the hand and the phase widget. -->
  <div class="rail">
    <PlayerIdentity
      {seat}
      {commanderScryfallID}
      {isSelf}
      {isActive}
      {hasPriority}
      {attackTargetable}
      {isMonarch}
      {isInitiative}
      {sendAction}
      {onDeclareAttack}
      {onTargetPlayer}
    />
    <div class="rail-gap"></div>
    <PileBar {seat} {exile} {isSelf} {sendAction} onDrawCard={isSelf ? onDrawCard : undefined} />
  </div>
</div>

<style>
  .panel {
    /* Sept 2026 redesign: zones on the left, a player rail on the
       right. Creatures on top at full size; the middle band splits
       artifacts/enchantments and the land strip, both one size down;
       the bottom row is the hand plus (self only) the phase widget.
       Top-row opponents set `flipped`, which reverses the row order
       instead of rotating the panel.

       --card-w / --card-h cascade into every nested Card; the
       compact rows override them with --card-w-sm / --card-h-sm.
       --thumb-w / --thumb-h size the pile thumbnails in the rail. */
    --card-w: calc(120px * var(--card-scale, 1));
    --card-h: calc(168px * var(--card-scale, 1));
    --card-w-sm: calc(88px * var(--card-scale, 1));
    --card-h-sm: calc(123px * var(--card-scale, 1));
    --thumb-w: calc(40px * var(--card-scale, 1));
    --thumb-h: calc(56px * var(--card-scale, 1));
    --avatar-size-base: 84px;
    --rail-w: 112px;
    display: grid;
    grid-template-columns: minmax(0, 1fr) var(--rail-w);
    grid-template-rows: minmax(0, 1fr) auto auto;
    grid-template-areas:
      "creatures rail"
      "middle    rail"
      "bottom    rail";
    gap: 6px;
    width: 100%;
    height: 100%;
    box-sizing: border-box;
    padding: 8px;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 14px;
    overflow: hidden;
    position: relative;
  }
  /* Self panel carries a gold hairline so the row that's yours reads
     without drawing attention from the cards. */
  .panel.self {
    border-color: rgba(217, 180, 92, 0.28);
    box-shadow: inset 0 0 0 1px rgba(217, 180, 92, 0.08);
  }
  .panel.opponent {
    /* Upright opponent (the "next" seat) gets the medium scale — it
       has the tall row to itself. Scales by --card-scale-opponent
       (settings) with a gentler curve than self. */
    --card-w: calc(88px * var(--card-scale-opponent, 1));
    --card-h: calc(123px * var(--card-scale-opponent, 1));
    --card-w-sm: calc(64px * var(--card-scale-opponent, 1));
    --card-h-sm: calc(90px * var(--card-scale-opponent, 1));
    --thumb-w: calc(36px * var(--card-scale-opponent, 1));
    --thumb-h: calc(50px * var(--card-scale-opponent, 1));
    --avatar-size-base: 72px;
    --rail-w: 104px;
  }
  .panel.opponent.flipped {
    /* Across-table seats: one more size down (the top row is the
       short one), rows reversed so the hand hugs the top edge and
       creatures face the centre of the table. */
    --card-w: calc(64px * var(--card-scale-opponent, 1));
    --card-h: calc(90px * var(--card-scale-opponent, 1));
    --card-w-sm: calc(48px * var(--card-scale-opponent, 1));
    --card-h-sm: calc(67px * var(--card-scale-opponent, 1));
    --thumb-w: calc(32px * var(--card-scale-opponent, 1));
    --thumb-h: calc(45px * var(--card-scale-opponent, 1));
    --avatar-size-base: 60px;
    grid-template-rows: auto auto minmax(0, 1fr);
    grid-template-areas:
      "bottom    rail"
      "middle    rail"
      "creatures rail";
  }
  .grid-creatures {
    grid-area: creatures;
    min-height: 0;
    min-width: 0;
  }
  .grid-middle {
    grid-area: middle;
    min-height: 0;
    min-width: 0;
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(0, 1.15fr);
    gap: 6px;
  }
  .grid-bottom {
    grid-area: bottom;
    min-height: 0;
    min-width: 0;
    display: flex;
    align-items: flex-end;
    gap: 8px;
  }
  .flipped .grid-bottom {
    align-items: flex-start;
  }
  .hand-zone {
    flex: 1 1 0;
    min-width: 0;
    align-self: flex-end;
    /* Fixed rest-state height matches the Hand's clipped peek. Keeping
       it explicit means the Hand's hover lift (max-height: none plus a
       translateY transform) doesn't grow this wrapper and push the
       row taller — the lift stays purely visual. */
    height: calc(var(--card-h, 168px) * 0.55);
    overflow: visible;
    position: relative;
  }
  .flipped .hand-zone {
    align-self: flex-start;
  }
  .rail {
    grid-area: rail;
    min-height: 0;
    min-width: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
    padding: 6px 0 2px 8px;
    border-left: 1px solid var(--border);
  }
  .rail-gap {
    flex: 1 1 0;
  }
  .flipped .rail-gap {
    flex: 0 0 4px;
  }
  /* Short panels (the top row at 900px tall) must never clip the
     piles: the rail scrolls before it hides anything, and the
     across-table rails drop the pile labels (the tile's title and
     aria-label still carry them) and the command-zone hints. */
  .rail {
    overflow-y: auto;
    overflow-x: hidden;
  }
  .flipped .rail :global(.pile .label),
  .flipped .rail :global(.cmd-zone .label),
  .flipped .rail :global(.cmd-zone .hints) {
    display: none;
  }
</style>
