<script lang="ts">
  // PileBar renders the four bottom-left pile controls in the order
  // shown on the wireframe: EXILE / GRAVEYARD / LIBRARY / CMD ZONE.
  //
  // The CMD slot is no longer a face-down PileButton; it's the
  // first-class CommandZone component (face-up commander, tax badge,
  // cast affordance). The other three stay as PileButtons; LIBRARY
  // shows a card back and is the only one wired to an action at v1
  // (draw_card on the viewer's own seat).
  //
  // Exile is conceptually a shared zone in the wire protocol
  // (GameView.exile), so callers pass a pre-filtered ZoneView containing
  // just this player's owned exiled cards.

  import type { ActionPayload, PlayerView, ZoneView } from "../../protocol";
  import PileButton from "./PileButton.svelte";
  import CommandZone from "./CommandZone.svelte";

  type ActionSender = (type: string, params?: ActionPayload["params"], player?: string) => void;

  interface Props {
    seat: PlayerView;
    exile: ZoneView;
    isSelf: boolean;
    sendAction: ActionSender;
    onDrawCard?: () => void;
  }

  const { seat, exile, isSelf, sendAction, onDrawCard }: Props = $props();
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
  <CommandZone seat={{ id: seat.id, name: seat.name }} zone={seat.command} {isSelf} {sendAction} />
</div>

<style>
  .pile-bar {
    display: flex;
    gap: 6px;
    align-items: stretch;
  }
</style>
