<script lang="ts">
  // PileBar renders the four pile controls as a 2×2 grid in the
  // player rail: LIBRARY / GRAVEYARD on top, EXILE / CMD ZONE below.
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

  import type { ActionPayload, ActionType, PlayerView, ZoneView } from "../../protocol";
  import PileButton from "./PileButton.svelte";
  import CommandZone from "./CommandZone.svelte";
  import { openZoneBrowser } from "../../zoneBrowser";
  import { visibleLibraryTop } from "../../libraryTop";

  type ActionSender = (type: ActionType, params?: ActionPayload["params"], player?: string) => void;

  interface Props {
    seat: PlayerView;
    exile: ZoneView;
    isSelf: boolean;
    sendAction: ActionSender;
    onDrawCard?: () => void;
  }

  const { seat, exile, isSelf, sendAction, onDrawCard }: Props = $props();

  // S42 / CR 401.5: "you may look at the top card of your library any
  // time" and "play with the top card of your library revealed" both
  // reach the client as one readable card at the top of the library
  // zone. When one arrives the pile shows it instead of a back — for
  // its owner under a "look" clause, and for the whole table under a
  // "revealed" one, because the server has already decided which.
  const libraryTop = $derived(visibleLibraryTop(seat.library));

  // S18.5 — graveyard + exile chips open the browser modal. Always
  // available to every viewer (public-info zones). Library stays
  // owner-only as a "draw the top card" affordance; a future S22
  // pass adds search / scry / reorder to the library flow.
  function openGraveyard(): void {
    openZoneBrowser({ zoneKind: "graveyard", ownerID: seat.id, ownerName: seat.name });
  }
  function openExile(): void {
    openZoneBrowser({ zoneKind: "exile", ownerID: seat.id, ownerName: seat.name });
  }
</script>

<div class="pile-bar" aria-label={`${seat.name} piles`}>
  <PileButton
    label="library"
    zone={seat.library}
    faceDown={libraryTop === null}
    disabled={!isSelf || !onDrawCard}
    onClick={isSelf ? onDrawCard : undefined}
  />
  <PileButton label="grave" zone={seat.graveyard} onClick={openGraveyard} />
  <PileButton label="exile" zone={exile} onClick={openExile} />
  <CommandZone
    seat={{ id: seat.id, name: seat.name }}
    zone={seat.command}
    {isSelf}
    {sendAction}
    commanderCasts={seat.commander_casts}
  />
</div>

<style>
  /* 2×2 in the player rail: library / grave on top, exile / command
     below. Tiles size to the rail; the thumb inside sizes from
     --thumb-w / --thumb-h set by PlayerPanel. */
  .pile-bar {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 4px;
    width: 100%;
  }
</style>
