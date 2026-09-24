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

  import type { ActionPayload, ActionType, CardView, PlayerView, ZoneView } from "../../protocol";
  import PileButton from "./PileButton.svelte";
  import CommandZone from "./CommandZone.svelte";
  import { openZoneBrowser } from "../../zoneBrowser";
  import { libraryTopActionLabel, visibleLibraryTop } from "../../libraryTop";
  import type { CastSourceZone } from "../../targeting";

  type ActionSender = (type: ActionType, params?: ActionPayload["params"], player?: string) => void;

  interface Props {
    seat: PlayerView;
    exile: ZoneView;
    isSelf: boolean;
    sendAction: ActionSender;
    onDrawCard?: () => void;
    // #1440: hands a playable library top to the Board's one cast
    // chain, exactly as the graveyard's flashback button and the
    // exile impulse button do — `fromZone: "library"`, so the prompt
    // chain (face, alternative cost, X, modes, targets) runs before
    // anything reaches the wire. Not gated on `isSelf`: a cross-seat
    // grant (Xanathar, Guild Kingpin) puts the button on the GRANTOR's
    // pile for the seat that holds the permission, and `castable_here`
    // is already that viewer's own answer (#1055).
    onPlayCard?: (card: CardView, fromZone?: CastSourceZone, face?: number) => void;
    // #1278: activated abilities that function from the command zone
    // (commander ninjutsu), and the sorcery-speed reason their popover
    // greys with. Passed straight through to CommandZone.
    onActivateAbility?: (card: CardView, abilityIndex: number) => void;
    sorcerySpeedBlocked?: string;
  }

  const {
    seat,
    exile,
    isSelf,
    sendAction,
    onDrawCard,
    onPlayCard,
    onActivateAbility,
    sorcerySpeedBlocked = "",
  }: Props = $props();

  // S42 / CR 401.5: "you may look at the top card of your library any
  // time" and "play with the top card of your library revealed" both
  // reach the client as one readable card at the top of the library
  // zone. When one arrives the pile shows it instead of a back — for
  // its owner under a "look" clause, and for the whole table under a
  // "revealed" one, because the server has already decided which.
  const libraryTop = $derived(visibleLibraryTop(seat.library));

  // #1440: the play/cast affordance on the pile itself. `libraryTop`
  // above only asks whether a card should be shown face up;
  // `libraryTopAction` is the separate, stricter question of whether
  // THIS viewer may play it from here — `castable_here`, which the
  // server stamps per viewer (#1055), so Oracle of Mul Daya reveals a
  // sorcery to the whole table with no button on it anywhere.
  const libraryTopAction = $derived(onPlayCard ? libraryTopActionLabel(seat.library) : null);

  function playLibraryTop(): void {
    if (!onPlayCard || !libraryTop) return;
    onPlayCard(libraryTop, "library");
  }

  // S18.5 — graveyard + exile chips open the browser modal. Always
  // available to every viewer (public-info zones). Library's own
  // click stays owner-only as a "draw the top card" affordance; a
  // future S22 pass adds search / scry / reorder to the library flow.
  // #1440 adds a SEPARATE affordance beside it — playing a visible
  // top card under a permission — which is not owner-gated, because
  // the permission's holder and the pile's owner can be different
  // seats (Xanathar, Guild Kingpin).
  function openGraveyard(): void {
    openZoneBrowser({ zoneKind: "graveyard", ownerID: seat.id, ownerName: seat.name });
  }
  function openExile(): void {
    openZoneBrowser({ zoneKind: "exile", ownerID: seat.id, ownerName: seat.name });
  }
</script>

<div class="pile-bar" aria-label={`${seat.name} piles`}>
  <div class="pile-slot">
    <PileButton
      label="library"
      zone={seat.library}
      faceDown={libraryTop === null}
      disabled={!isSelf || !onDrawCard}
      onClick={isSelf ? onDrawCard : undefined}
    />
    {#if libraryTopAction}
      <!-- #1440: the same always-visible pill treatment as the
           graveyard's flashback button and the exile impulse button —
           a permission is a resource, and a player who has to hover a
           tiny pile thumb to discover it will not discover it. A
           sibling of PileButton's own <button>, never nested inside
           it (button-in-button is invalid markup). -->
      <button
        type="button"
        class="pile-action"
        title={`${libraryTopAction} from the top of your library`}
        aria-label={`${libraryTopAction} ${libraryTop?.name || "card"} from the top of the library`}
        onclick={playLibraryTop}
      >
        {libraryTopAction}
      </button>
    {/if}
  </div>
  <PileButton label="grave" zone={seat.graveyard} onClick={openGraveyard} />
  <PileButton label="exile" zone={exile} onClick={openExile} />
  <CommandZone
    seat={{ id: seat.id, name: seat.name }}
    zone={seat.command}
    {isSelf}
    {sendAction}
    commanderCasts={seat.commander_casts}
    onActivateAbility={isSelf ? onActivateAbility : undefined}
    {sorcerySpeedBlocked}
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
  /* Wraps the library PileButton so the #1440 play/cast pill can sit
     over its bottom edge as a sibling — never nested inside
     PileButton's own <button>, which button-in-button markup forbids. */
  .pile-slot {
    position: relative;
    width: 100%;
  }
  .pile-action {
    position: absolute;
    left: 50%;
    bottom: -3px;
    transform: translate(-50%, 50%);
    height: 15px;
    padding: 0 6px;
    border-radius: 999px;
    border: 1px solid rgba(217, 180, 92, 0.6);
    background: var(--surface-raised);
    color: var(--gold-strong);
    font-family: var(--font-mono);
    font-size: 8.5px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    line-height: 1;
    cursor: pointer;
    white-space: nowrap;
  }
  .pile-action:hover {
    background: var(--surface-hover);
  }
  .pile-action:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 1px;
  }
</style>
