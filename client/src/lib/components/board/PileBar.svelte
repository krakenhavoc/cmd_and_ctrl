<script lang="ts">
  // PileBar renders the three pile controls in the player rail:
  // LIBRARY / GRAVEYARD on top, EXILE below. LIBRARY shows a card back.
  // #2962: clicking your own library opens a small menu (Draw a card,
  // Shuffle library) rather than drawing at once, because a stray click
  // while reaching for the pile drew a card nobody meant to draw.
  //
  // #2349: the command zone tile that was the fourth is gone. Every
  // seat's commander sits beside its hand instead: yours in the
  // castable strip (ExileStrip), everyone else's in CommandStrip.
  //
  // Exile is conceptually a shared zone in the wire protocol
  // (GameView.exile), so callers pass a pre-filtered ZoneView containing
  // just this player's owned exiled cards.

  import { tick } from "svelte";
  import type { ActionPayload, ActionType, CardView, PlayerView, ZoneView } from "../../protocol";
  import PileButton from "./PileButton.svelte";
  import { openZoneBrowser } from "../../zoneBrowser";
  import {
    libraryTopActionLabel,
    libraryTopSpecialActions,
    visibleLibraryTop,
  } from "../../libraryTop";
  import type { CastSourceZone } from "../../targeting";
  import { heldPlayableCount } from "../../exileGrants";
  import { NO_LEGAL_ACTIONS, withAvailable, type LegalActions } from "../../legalActions";

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
    // ADR 0105 (#1789): the legal-action lookup ("nothing" while
    // highlights are off) for the piles' "N ready" counts and the
    // library-top pill's accent.
    legal?: LegalActions;
  }

  const {
    seat,
    exile,
    isSelf,
    sendAction,
    onDrawCard,
    onPlayCard,
    legal = NO_LEGAL_ACTIONS,
  }: Props = $props();

  // "N ready" per pile. The exile pile is this seat's own slice, so it
  // counts by owner — an impulse-exiled card the viewer may cast sits
  // on its owner's pile, and that is where it is counted.
  const graveReady = $derived(legal.readyCount("graveyard", seat.id));
  const exileReady = $derived(legal.readyCount("exile", seat.id));
  const libraryReady = $derived(legal.readyCount("library", seat.id));
  // #2559: on another seat's pile, the exiled cards THAT seat may play
  // (Memory Vessel, Rocco). The viewer's own are the "ready" count.
  const exileHeld = $derived(isSelf ? 0 : heldPlayableCount(exile.cards, seat.id));

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
  const libraryTopReady = $derived(
    libraryTop !== null && legal.castableFrom(libraryTop.instance_id, "library"),
  );

  function playLibraryTop(): void {
    if (!onPlayCard || !libraryTop) return;
    onPlayCard(libraryTop, "library");
  }

  // #1391: special actions on the top card — Fblthp, Lost on the
  // Range's plot. A pill beside the cast pill, fired straight from the
  // row like the hand menu's special-action rows (no targets, no
  // picker). Only on the viewer's OWN pile: the action is the library
  // owner's, and the server already sends the rows to nobody else.
  const libraryTopSpecial = $derived(isSelf ? libraryTopSpecialActions(seat.library) : []);

  // S18.5 — graveyard + exile chips open the browser modal. Always
  // available to every viewer (public-info zones). Library's own
  // click stays owner-only as a "draw the top card" affordance; a
  // future S22 pass adds search / scry / reorder to the library flow.
  // #1440 adds a SEPARATE affordance beside it — playing a visible
  // top card under a permission — which is not owner-gated, because
  // the permission's holder and the pile's owner can be different
  // seats (Xanathar, Guild Kingpin).
  // #2962: the library's own click opens this menu; nothing is sent
  // until a row is chosen. The draw shortcut and the game menu still
  // draw in one step. Position is fixed from the pile's rect so no
  // ancestor's overflow clips it.
  let slotEl: HTMLDivElement | undefined = $state();
  let menuEl: HTMLDivElement | undefined = $state();
  let menuOpen = $state(false);
  let menuPos = $state({ left: 0, top: 0 });

  async function openLibraryMenu(): Promise<void> {
    if (menuOpen) {
      menuOpen = false;
      return;
    }
    const r = slotEl?.getBoundingClientRect();
    if (r) menuPos = { left: r.left, top: r.bottom + 4 };
    menuOpen = true;
    await tick();
    menuEl?.querySelector<HTMLElement>("[role=menuitem]")?.focus();
  }
  function closeLibraryMenu(): void {
    menuOpen = false;
  }
  function chooseDraw(): void {
    menuOpen = false;
    onDrawCard?.();
  }
  function chooseShuffle(): void {
    menuOpen = false;
    sendAction("shuffle_library", undefined, seat.id);
  }
  function onWindowPointer(e: Event): void {
    if (!menuOpen) return;
    const t = e.target as Node | null;
    if (t && (menuEl?.contains(t) || slotEl?.contains(t))) return;
    menuOpen = false;
  }
  function onMenuKey(e: KeyboardEvent): void {
    if (e.key === "Escape") {
      e.stopPropagation();
      menuOpen = false;
      slotEl?.querySelector<HTMLElement>("button.pile")?.focus();
    }
  }

  function openGraveyard(): void {
    openZoneBrowser({ zoneKind: "graveyard", ownerID: seat.id, ownerName: seat.name });
  }
  function openExile(): void {
    openZoneBrowser({ zoneKind: "exile", ownerID: seat.id, ownerName: seat.name });
  }
</script>

<svelte:window onpointerdown={onWindowPointer} onresize={closeLibraryMenu} />

<div class="pile-bar" aria-label={`${seat.name} piles`}>
  <div class="pile-slot" bind:this={slotEl}>
    <PileButton
      label="library"
      pile="library"
      owner={seat.id}
      zone={seat.library}
      faceDown={libraryTop === null}
      disabled={!isSelf || !onDrawCard}
      onClick={isSelf ? openLibraryMenu : undefined}
      haspopup={isSelf && !!onDrawCard}
      expanded={menuOpen}
      readyCount={libraryReady}
    />
    {#if menuOpen}
      <div
        class="lib-menu"
        role="menu"
        aria-label="library actions"
        tabindex="-1"
        bind:this={menuEl}
        style:left="{menuPos.left}px"
        style:top="{menuPos.top}px"
        onkeydown={onMenuKey}
      >
        <button type="button" class="mi" role="menuitem" onclick={chooseDraw}>Draw a card</button>
        <button type="button" class="mi" role="menuitem" onclick={chooseShuffle}>
          Shuffle library
        </button>
      </div>
    {/if}
    {#if libraryTopAction || libraryTopSpecial.length > 0}
      <div class="pile-actions">
        {#if libraryTopAction}
          <!-- #1440: the same always-visible pill treatment as the
               graveyard's flashback button and the exile impulse
               button — a permission is a resource, and a player who
               has to hover a tiny pile thumb to discover it will not
               discover it. A sibling of PileButton's own <button>,
               never nested inside it (button-in-button is invalid
               markup). -->
          <button
            type="button"
            class="pile-action"
            class:ready={libraryTopReady}
            title={`${libraryTopAction} from the top of your library`}
            aria-label={withAvailable(
              `${libraryTopAction} ${libraryTop?.name || "card"} from the top of the library`,
              libraryTopReady,
            )}
            onclick={playLibraryTop}
          >
            {libraryTopAction}
          </button>
        {/if}
        {#each libraryTopSpecial as sa (sa.key)}
          <!-- #1391: a special action on the top card (Fblthp's plot),
               sent as `special_action` exactly as the hand menu's row
               sends it. -->
          <button
            type="button"
            class="pile-action special"
            title={`${sa.label} from the top of your library`}
            aria-label={`${sa.label}: ${libraryTop?.name || "card"} from the top of the library`}
            onclick={() => sendAction("special_action", sa.params, seat.id)}
          >
            {sa.text}
          </button>
        {/each}
      </div>
    {/if}
  </div>
  <PileButton
    label="grave"
    pile="graveyard"
    owner={seat.id}
    zone={seat.graveyard}
    onClick={openGraveyard}
    readyCount={graveReady}
  />
  <PileButton
    label="exile"
    pile="exile"
    owner={seat.id}
    zone={exile}
    onClick={openExile}
    readyCount={exileReady}
    note={exileHeld > 0 ? `${exileHeld} theirs to play` : undefined}
  />
</div>

<style>
  /* Two across in the player rail: library / grave on top, exile
     below. Tiles size to the rail; the thumb inside sizes from
     --thumb-w / --thumb-h set by PlayerPanel. */
  .pile-bar {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 4px;
    width: 100%;
  }
  /* Wraps the library PileButton so the #1440 play/cast pill and the
     #1391 special-action pills can sit over its bottom edge as
     siblings — never nested inside
     PileButton's own <button>, which button-in-button markup forbids. */
  .pile-slot {
    position: relative;
    width: 100%;
  }
  /* The row of pills over the library's bottom edge: the #1440
     play/cast pill and the #1391 special-action pills side by side. */
  .pile-actions {
    position: absolute;
    left: 50%;
    bottom: -3px;
    transform: translate(-50%, 50%);
    display: flex;
    gap: 3px;
  }
  .lib-menu {
    position: fixed;
    z-index: 60;
    display: flex;
    flex-direction: column;
    min-width: 140px;
    padding: 4px;
    background: var(--surface-raised);
    border: 1px solid color-mix(in srgb, var(--accent) 45%, transparent);
    border-radius: var(--radius);
    box-shadow: 0 6px 18px rgb(0 0 0 / 0.4);
  }
  .lib-menu .mi {
    padding: 7px 10px;
    text-align: left;
    background: none;
    border: 0;
    border-radius: 4px;
    color: inherit;
    font: inherit;
    cursor: pointer;
  }
  .lib-menu .mi:hover,
  .lib-menu .mi:focus-visible {
    background: var(--surface-hover);
    outline: none;
  }
  .pile-action {
    height: 15px;
    padding: 0 6px;
    border-radius: 999px;
    border: 1px solid color-mix(in srgb, var(--accent) 60%, transparent);
    background: var(--surface-raised);
    color: var(--accent-strong);
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
  /* ADR 0105: the server's move list says this top card can be played
     right now, mana included — the ready accent. */
  .pile-action.ready {
    color: var(--ready);
    border-color: var(--ready);
    background: var(--ready-soft);
  }
  .pile-action:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 1px;
  }
</style>
