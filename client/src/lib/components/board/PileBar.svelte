<script lang="ts">
  // PileBar renders the four bottom-left pile controls in the order
  // shown on the wireframe: EXILE / GRAVEYARD / DECK / CMD ZONE.
  //
  // Wires the only v1 action (DECK click on the viewer's own seat →
  // draw_card); the others stay passive until a pile-browser modal
  // lands. Exile is conceptually a shared zone in the wire protocol
  // (GameView.exile), so callers pass a pre-filtered ZoneView containing
  // just this player's owned exiled cards.

  import type { PlayerView, ZoneView } from "../../protocol";
  import PileButton from "./PileButton.svelte";

  interface Props {
    seat: PlayerView;
    exile: ZoneView;
    isSelf: boolean;
    onDrawCard?: () => void;
  }

  const { seat, exile, isSelf, onDrawCard }: Props = $props();
</script>

<div class="pile-bar" aria-label={`${seat.name} piles`}>
  <PileButton label="exile" zone={exile} />
  <PileButton label="grave" zone={seat.graveyard} />
  <PileButton
    label="library"
    zone={seat.library}
    faceDown
    disabled={!isSelf || !onDrawCard}
    onClick={isSelf ? onDrawCard : undefined}
  />
  <PileButton label="cmd" zone={seat.command} />
</div>

<style>
  .pile-bar {
    display: flex;
    gap: 6px;
    align-items: stretch;
  }
</style>
