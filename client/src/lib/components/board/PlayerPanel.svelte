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
    <Hand hand={seat.hand} {isSelf} onPlayCard={isSelf ? onPlayCard : undefined} />
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
    /* Card sizes — easy to dial in further by tweaking just these
       vars; everything downstream (Card.svelte, PileButton.svelte,
       Hand.svelte) reads them. Aspect ratio held at 0.714 (the real
       63mm × 88mm magic-card proportion) so face art stays
       undistorted.

       --card-scale is set on :root by the S11.5 settings panel
       (Display → card size). Multiplying here rather than letting
       :root override --card-w directly avoids a specificity fight
       with .panel.opponent's smaller baseline. */
    --card-w: calc(120px * var(--card-scale, 1));
    --card-h: calc(168px * var(--card-scale, 1));
    --pile-w: calc(94px * var(--card-scale, 1));
    --thumb-w: calc(75px * var(--card-scale, 1));
    --thumb-h: calc(105px * var(--card-scale, 1));
    display: grid;
    grid-template-columns: minmax(0, 1.4fr) minmax(0, 3fr) minmax(0, 1.3fr);
    /* Header is fixed; piles/hand row is content-sized (the largest
       child is the hand fan, naturally about --card-h tall). The
       creatures and lands rows split the remaining space so the
       battlefield zones grow to fill the panel instead of leaving a
       blank gap above the piles row.

       The right column (artifacts / enchantments) spans every row
       below the header so it can grow vertically alongside the
       creatures row instead of overflowing off-screen at the bottom
       when many permanents are in play. */
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
    padding: 4px;
    background: #0d1424;
    border-radius: 6px;
    overflow: hidden;
  }
  .panel.opponent {
    background: #0a1020;
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
