<script lang="ts">
  // PlayerPanel is one player's full board, laid out in a CSS Grid
  // matching the wireframe:
  //
  //   ┌────────────────────────────────────────┐
  //   │ header                                 │
  //   ├────────────────────────────────────────┤
  //   │ creatures (full width)                 │
  //   ├────────────────────────────────┬───────┤
  //   │ lands                          │       │
  //   ├──────────┬─────────────────────┤ right │
  //   │ piles    │ hand                │ col   │
  //   └──────────┴─────────────────────┴───────┘
  //
  // The same component is reused for self + opponents; Board.svelte
  // wraps opponent instances in a transform container that scales
  // and rotates them around the table. The panel itself is layout-
  // only and doesn't know whether it's rotated.
  //
  // Click routing for battlefield cards lives here so combat vs.
  // tap-toggle logic stays in one place. The router mirrors the old
  // Pixi wireTapClick: combat select on your own creature, declare-
  // block on an incoming attacker, otherwise tap/untap.

  import type { ActionPayload, CardView, GameView, PlayerView, ZoneView } from "../../protocol";
  import { bucketForBattlefield, isCreature } from "../../cardTypes";
  import BattlefieldRow from "./BattlefieldRow.svelte";
  import BattlefieldColumn from "./BattlefieldColumn.svelte";
  import PileBar from "./PileBar.svelte";
  import Hand from "./Hand.svelte";
  import PlayerHeader from "./PlayerHeader.svelte";
  import PromisesRow from "./PromisesRow.svelte";

  type ActionSender = (type: string, params?: ActionPayload["params"], player?: string) => void;

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
  }: Props = $props();

  const buckets = $derived.by(() => {
    const out = { creature: [] as CardView[], land: [] as CardView[], right: [] as CardView[] };
    for (const c of controlledCards) {
      out[bucketForBattlefield(c)].push(c);
    }
    return out;
  });

  const attackTargetable = $derived(
    !isSelf && !seat.eliminated && combatMode === "attack" && !!selectedCombatCardID,
  );

  function handleCardClick(card: CardView): void {
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
    // Default: tap/untap if the viewer controls it (or is admin).
    if (card.controller !== viewerID && !isAdmin) return;
    onTapToggle(card);
  }
</script>

<div class="panel" class:self={isSelf} class:opponent={!isSelf}>
  <div class="grid-header">
    <PlayerHeader
      {seat}
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
    {#if !isSelf}
      <PromisesRow {view} {viewerID} opponentID={seat.id} {sendAction} />
    {/if}
  </div>
  <div class="grid-creatures">
    <BattlefieldRow
      label="creatures"
      cards={buckets.creature}
      {viewerID}
      {selectedCombatCardID}
      onCardClick={handleCardClick}
    />
  </div>
  <div class="grid-lands">
    <BattlefieldRow
      label="lands"
      cards={buckets.land}
      {viewerID}
      {selectedCombatCardID}
      onCardClick={handleCardClick}
    />
  </div>
  <div class="grid-piles">
    <PileBar {seat} {exile} {isSelf} {sendAction} onDrawCard={isSelf ? onDrawCard : undefined} />
  </div>
  <div class="grid-hand">
    <Hand
      hand={seat.hand}
      {isSelf}
      onPlayCard={isSelf ? onPlayCard : undefined}
      snap={view}
      {viewerID}
    />
  </div>
  <div class="grid-rightcol">
    <BattlefieldColumn
      label="enchant / artifact"
      cards={buckets.right}
      {viewerID}
      {selectedCombatCardID}
      onCardClick={handleCardClick}
    />
  </div>
</div>

<style>
  .panel {
    /* Wireframe layout. The right column spans the lower two rows
       (lands + piles/hand). Header is a fixed 32px strip; the
       remaining vertical space is split with creatures getting the
       largest share. min-height: 0 on every grid item lets cards
       shrink to fit instead of forcing the panel to grow past its
       container.

       --card-w / --card-h cascade into every nested Card so the
       opponent panels can shrink the whole board with one rule
       (see .panel.opponent below). --pile-w / --thumb-w /
       --thumb-h play the same role for the PileBar. */
    --card-w: calc(120px * var(--card-scale, 1));
    --card-h: calc(168px * var(--card-scale, 1));
    --pile-w: calc(94px * var(--card-scale, 1));
    --thumb-w: calc(75px * var(--card-scale, 1));
    --thumb-h: calc(105px * var(--card-scale, 1));
    display: grid;
    grid-template-columns: minmax(0, 1.4fr) minmax(0, 3fr) minmax(0, 1.3fr);
    grid-template-rows: 32px minmax(0, 1.4fr) minmax(0, 1fr) auto;
    grid-template-areas:
      "header     header     header"
      "creatures  creatures  rightcol"
      "lands      lands      rightcol"
      "piles      hand       rightcol";
    gap: 6px;
    width: 100%;
    height: 100%;
    box-sizing: border-box;
    padding: 8px;
    /* Layered surface: subtle inner highlight at the top for a sheen,
       gentle gradient from raised→sunken so the panel reads like a
       felt-topped playmat rather than a flat rectangle. */
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 0.025) 0%, rgba(255, 255, 255, 0) 20%),
      linear-gradient(180deg, #131c34 0%, #0b1325 100%);
    border: 1px solid rgba(122, 167, 255, 0.1);
    border-radius: var(--radius-lg);
    box-shadow:
      0 12px 30px rgba(0, 0, 0, 0.4),
      inset 0 1px 0 rgba(255, 255, 255, 0.04);
    overflow: hidden;
    position: relative;
  }
  /* Self panel carries a subtle amber accent along the bottom to
     remind the user which row is theirs without drawing too much
     attention. */
  .panel.self {
    border-color: rgba(255, 208, 122, 0.18);
    box-shadow:
      0 14px 34px rgba(0, 0, 0, 0.45),
      inset 0 -1px 0 rgba(255, 208, 122, 0.18),
      inset 0 1px 0 rgba(255, 255, 255, 0.05);
  }
  .panel.opponent {
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 0.02) 0%, rgba(255, 255, 255, 0) 20%),
      linear-gradient(180deg, #0f1628 0%, #080d1b 100%);
    /* Opponent baseline is shrunk vs. pre-S11.5 (78×110 → 66×92)
       to leave headroom for the rotated-180° hand fan's bounding
       box, which adds ~20px above the card height. Scales by
       --card-scale-opponent (defined on :root by the S11.5
       settings panel), which uses a gentler curve than self
       (0.85 / 1 / 1.1) so a "large" setting doesn't overflow the
       fixed 0.7fr opponent row in the Board grid. */
    --card-w: calc(66px * var(--card-scale-opponent, 1));
    --card-h: calc(92px * var(--card-scale-opponent, 1));
    --pile-w: calc(56px * var(--card-scale-opponent, 1));
    --thumb-w: calc(44px * var(--card-scale-opponent, 1));
    --thumb-h: calc(62px * var(--card-scale-opponent, 1));
  }
  .grid-header {
    grid-area: header;
    min-width: 0;
  }
  .grid-creatures {
    grid-area: creatures;
    min-height: 0;
    min-width: 0;
  }
  .grid-lands {
    grid-area: lands;
    min-height: 0;
    min-width: 0;
  }
  .grid-piles {
    grid-area: piles;
    min-height: 0;
    min-width: 0;
    display: flex;
    /* Bottom-align so the pile chips sit on the same baseline as the
       hand fan in the adjacent cell, instead of floating at the top
       of an oversized row. */
    align-items: flex-end;
  }
  .grid-hand {
    grid-area: hand;
    min-height: 0;
    min-width: 0;
  }
  .grid-rightcol {
    grid-area: rightcol;
    min-height: 0;
    min-width: 0;
  }
</style>
