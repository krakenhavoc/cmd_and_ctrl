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

  import type { ActionPayload, CardView, GameView, ZoneView } from "../../protocol";
  import { seatPlacements, type SeatPosition } from "../../cardTypes";
  import PlayerPanel from "./PlayerPanel.svelte";
  import HoverZoomOverlay from "./HoverZoomOverlay.svelte";
  import StackOverlay from "./StackOverlay.svelte";
  import CombatArrows from "./CombatArrows.svelte";

  type ActionSender = (type: string, params?: ActionPayload["params"], player?: string) => void;

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
  }: Props = $props();

  const placements = $derived(seatPlacements(view.seats, viewerID));
  const opponentCount = $derived(
    (["next", "across", "across_next"] as SeatPosition[]).filter((p) => placements[p]).length,
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

  const activeSeatID = $derived(view.seats[view.turn.active_seat]?.id ?? null);
  const prioritySeatID = $derived(view.seats[view.turn.priority_holder]?.id ?? null);

  function handleTapToggle(card: CardView): void {
    sendAction(card.tapped ? "untap" : "tap", { instance_id: card.instance_id });
  }

  function handlePlayCard(card: CardView): void {
    sendAction("play_card", { instance_id: card.instance_id }, viewerID ?? undefined);
  }

  function handleDrawCard(): void {
    if (!viewerID) return;
    sendAction("draw_card", undefined, viewerID);
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

<div class="board" data-opp-count={opponentCount} bind:this={boardEl}>
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
        />
      </div>
    {/if}
  {/each}

  <CombatArrows {view} {boardEl} />
  <HoverZoomOverlay />
  <StackOverlay stack={view.stack} />
</div>

<style>
  .board {
    position: relative;
    width: 100%;
    height: 100%;
    display: grid;
    gap: 6px;
    box-sizing: border-box;
    overflow: hidden;
    background: #0b1220;
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
</style>
