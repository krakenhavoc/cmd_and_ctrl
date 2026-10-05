<script lang="ts">
  // CommandZone is the first-class commander slot on each PlayerPanel.
  // Replaces the small face-down cmd PileButton with a face-up,
  // card-sized rendering of the commander(s) in the zone, plus a
  // commander-tax badge.
  //
  // Click semantics on the viewer's own zone:
  //   - empty: no-op
  //   - one commander present: cast it from the command zone — the
  //     cast goes through the real pipeline (stack, sorcery-speed +
  //     priority gates, strict-mana with CR 903.8 tax, ETB triggers)
  //     and the server increments CommanderCasts on success. Since
  //     #2202 it starts the Board's one cast chain (onCastCard), the
  //     one the strip beside the hand uses, so an X, a kicker or a
  //     modal double-faced commander is asked its questions first.
  //   - multiple commanders (partner / Background): cycle visible top
  //     and click the visible one to cast it. Partner support on the
  //     deck-import side is currently `ErrUnsupportedMechanic`, so
  //     this path is forward-compatible padding rather than today's
  //     primary flow.
  //
  // The "+N tax" badge reads the server-tracked commander_casts map
  // (S13.1) off the seat's PlayerView — no local bookkeeping, so every
  // viewer (not just the caster's tab) sees the same tax.

  import type { ActionPayload, ActionType, CardView, GameView, ZoneView } from "../../protocol";
  import type { CastSourceZone } from "../../targeting";
  import { commanderTax } from "../../castStrip";
  import Card from "./Card.svelte";
  import { openZoneBrowser } from "../../zoneBrowser";
  import {
    canCastFromHand,
    castAnywayBlocked,
    castAnywayOffered,
    type Legality,
  } from "../../timing";
  import { requestCastAnyway } from "../../castAnyway";
  import { settings } from "../../settings";
  import { NO_LEGAL_ACTIONS, withAvailable, type LegalActions } from "../../legalActions";

  type ActionSender = (type: ActionType, params?: ActionPayload["params"], player?: string) => void;

  interface Props {
    seat: { id: string; name: string };
    zone: ZoneView;
    isSelf: boolean;
    sendAction: ActionSender;
    // Server-side per-commander cast counts (PlayerView.commander_casts),
    // keyed by commander instance UUID. Drives the "+N tax" badge.
    commanderCasts?: Record<string, number>;
    // #2202: the Board's cast chain (handlePlayCard), handed the
    // commander with the command zone as its zone — X, modes, targets,
    // alternative and optional costs. Before #2202 the click sent a bare
    // cast_spell, which skipped every one of those prompts; that bare
    // send is still the fallback for a caller with no chain.
    onCastCard?: (card: CardView, fromZone: CastSourceZone) => void;
    // #1278: a commander can print an activated ability that functions
    // FROM the command zone — commander ninjutsu (CR 702.49c). The
    // server ships those rows on `zone_abilities`, owner-only, exactly
    // as it does for a hand card, so the tile hands them to its Card
    // the way Hand.svelte does: right-click opens the same popover
    // (with the same greying), and the choice goes up to Board's
    // announce chain. Undefined suppresses the popover entirely.
    onActivateAbility?: (card: CardView, abilityIndex: number) => void;
    // Why the CR 307.1 sorcery-speed window is shut, or "" when open —
    // passed through to the popover as the hand passes it.
    sorcerySpeedBlocked?: string;
    // ADR 0105 (#1789): the frame and viewer, for the cast GATE — the
    // same server verdict that greys an uncastable hand card
    // (timing.ts canCastFromHand), so a commander you cannot cast no
    // longer looks castable. Absent (a caller with no frame) leaves
    // the hint live, as it always was.
    view?: GameView | null;
    viewerID?: string | null;
    // ADR 0105: the frame's legal-action lookup, "nothing" while
    // highlights are off or autopass is about to pass. The ready ring
    // on the commander and the hint's accent read it; the gate above
    // does not. The popover's ready-row accent reads it too.
    legal?: LegalActions;
    // ADR 0105 sub-PR 3: the frame's full lookup, for the popover's
    // sorcery-speed gate. The highlight setting never touches it.
    legalGate?: LegalActions;
  }

  const {
    seat,
    zone,
    isSelf,
    sendAction,
    commanderCasts,
    onCastCard,
    onActivateAbility,
    sorcerySpeedBlocked = "",
    view = null,
    viewerID = null,
    legal = NO_LEGAL_ACTIONS,
    legalGate = NO_LEGAL_ACTIONS,
  }: Props = $props();

  // The card slot, so the "ability" hint can open the Card's own
  // popover rather than a second copy of it.
  let cardSlot: HTMLDivElement | undefined = $state();

  let visibleIndex = $state(0);
  const commanders = $derived(zone.cards);
  const visibleCard = $derived<CardView | null>(
    commanders.length > 0 ? (commanders[visibleIndex % commanders.length] ?? null) : null,
  );

  // CR 903.8: each prior cast of this commander adds {2}.
  const visibleTax = $derived(
    visibleCard ? commanderTax(commanderCasts?.[visibleCard.instance_id]) : 0,
  );

  function cycleVisible(): void {
    if (commanders.length <= 1) return;
    visibleIndex = (visibleIndex + 1) % commanders.length;
  }

  // The cast gate for the visible commander. Permissive without a
  // frame, and permissive on a frame with no move list unless the
  // viewer lacks priority — canCastFromHand's own reading, shared with
  // the hand so the two never disagree about what is castable.
  const castGate = $derived.by((): Legality => {
    if (!isSelf || !visibleCard || !view) return { legal: true };
    return canCastFromHand(visibleCard, view, viewerID);
  });
  const castReady = $derived(
    isSelf && !!visibleCard && legal.castableFrom(visibleCard.instance_id, "command"),
  );

  // ADR 0118 §2: with strict payment on, the viewer's own visible
  // commander offers "Cast anyway (don't pay)" in its popover, payable
  // or not. The row opens the dock's confirmation (castAnyway.ts), which
  // starts the Board's cast chain out of the command zone.
  const castAnywayHere = $derived(
    isSelf &&
      !!onCastCard &&
      !!visibleCard &&
      $settings.gameplay.strictMana &&
      castAnywayOffered(visibleCard, "command"),
  );
  const castAnywayReason = $derived(
    castAnywayHere && visibleCard ? castAnywayBlocked(visibleCard, view, viewerID, "command") : "",
  );
  function castAnyway(): void {
    if (visibleCard) requestCastAnyway(visibleCard, "command");
  }

  function castVisible(): void {
    if (!isSelf || !visibleCard) return;
    if (!castGate.legal) return;
    if (onCastCard) {
      onCastCard(visibleCard, "command");
      return;
    }
    // Game.svelte's sendAction shim stamps the strict flag and stashes
    // the payload for the cast-anyway / auto-tap retry paths, which
    // replay it verbatim — so from_zone survives those retries too.
    sendAction(
      "cast_spell",
      { instance_id: visibleCard.instance_id, from_zone: "command" },
      seat.id,
    );
  }

  function handleClick(): void {
    if (commanders.length === 0) return;
    if (commanders.length > 1) {
      // Multi-commander: first click cycles, double-click casts. The
      // double-click affordance is a stretch but prevents an accidental
      // cast when the player is just looking.
      cycleVisible();
      return;
    }
    castVisible();
  }

  function handleDoubleClick(): void {
    castVisible();
  }

  // #1278: the rows the visible commander offers from THIS zone. Only
  // the owner's frame carries any (the server scopes them), and isSelf
  // is checked anyway so a stale frame can never wire an opponent's.
  const visibleAbilities = $derived(isSelf ? (visibleCard?.zone_abilities ?? []) : []);
  const activateVisible = $derived.by(() => {
    const c = visibleCard;
    if (!c || !onActivateAbility || visibleAbilities.length === 0) return undefined;
    return (abilityIndex: number) => onActivateAbility(c, abilityIndex);
  });

  // The hint opens the Card's own popover — the one right-click opens —
  // so there is one menu, one set of greying rules and one path into
  // Board's announce chain.
  function openAbilities(): void {
    const el = cardSlot?.querySelector<HTMLElement>(".card");
    if (!el) return;
    const r = el.getBoundingClientRect();
    el.dispatchEvent(
      new MouseEvent("contextmenu", {
        bubbles: true,
        cancelable: true,
        clientX: r.left + r.width / 2,
        clientY: r.top + r.height / 2,
      }),
    );
  }

  // S18.5 — "browse" affordance opens the ZoneBrowserModal for this
  // player's command zone. Wired on both self (secondary to the
  // cast button) and opponents (whose command zones previously had
  // no click behaviour at all).
  function openBrowser(): void {
    openZoneBrowser({ zoneKind: "command", ownerID: seat.id, ownerName: seat.name });
  }
</script>

<div
  class="cmd-zone"
  class:self={isSelf}
  class:empty={commanders.length === 0}
  data-pile="command"
  data-pile-owner={seat.id}
  aria-label={`${seat.name} command zone, ${zone.count} card${zone.count === 1 ? "" : "s"}`}
>
  <div class="card-slot" bind:this={cardSlot}>
    {#if visibleCard}
      <Card
        card={visibleCard}
        ready={castReady}
        readyZone="command"
        onClick={isSelf ? handleClick : openBrowser}
        onActivateAbility={activateVisible}
        onCastAnyway={castAnywayHere ? castAnyway : undefined}
        castAnywayBlocked={castAnywayReason}
        {sorcerySpeedBlocked}
        legal={isSelf ? legal : undefined}
        legalGate={isSelf ? legalGate : undefined}
      />
      {#if visibleTax > 0}
        <span class="tax-badge" title={`commander tax · +${visibleTax} mana`}>+{visibleTax}</span>
      {/if}
    {:else if isSelf}
      <div class="empty-slot" aria-hidden="true"></div>
    {:else}
      <button
        type="button"
        class="empty-slot empty-slot-btn"
        onclick={openBrowser}
        aria-label={`browse ${seat.name}'s command zone`}
      ></button>
    {/if}
  </div>
  <div class="meta">
    <span class="label">cmd</span>
    <span class="count">{zone.count}</span>
    {#if commanders.length > 1}
      <button
        type="button"
        class="cycle"
        onclick={cycleVisible}
        title="cycle commanders"
        aria-label="cycle commanders"
      >
        {visibleIndex + 1}/{commanders.length}
      </button>
    {/if}
  </div>
  <div class="hints">
    {#if isSelf && commanders.length === 1}
      <button
        type="button"
        class="cast-hint"
        class:ready={castReady}
        disabled={!castGate.legal}
        onclick={castVisible}
        ondblclick={handleDoubleClick}
        title={castGate.legal ? "cast commander" : (castGate.reason ?? "Can't cast right now")}
        aria-label={withAvailable("cast commander", castReady)}
      >
        cast
      </button>
    {/if}
    {#if activateVisible}
      <button
        type="button"
        class="ability-hint"
        onclick={openAbilities}
        title="abilities this commander can use from the command zone"
        aria-label="commander abilities"
      >
        ability
      </button>
    {/if}
    <button
      type="button"
      class="browse-hint"
      onclick={openBrowser}
      title={`browse ${seat.name}'s command zone`}
      aria-label={`browse ${seat.name}'s command zone`}
    >
      browse
    </button>
  </div>
</div>

<style>
  .cmd-zone {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    padding: 5px 3px 4px;
    background: var(--surface-raised);
    border: 1px solid color-mix(in srgb, var(--accent) 25%, transparent);
    border-radius: var(--radius);
    box-sizing: border-box;
    width: 100%;
  }
  .cmd-zone.self {
    border-color: var(--accent-line);
  }
  .cmd-zone.empty {
    border-style: dashed;
    opacity: 0.5;
  }
  .card-slot {
    /* Render the commander at the same size as the pile thumb so it
       fits beside EXILE / GRAVE / LIBRARY without towering over them.
       The face-up rendering and amber border are what carry "this
       zone is special"; size parity keeps the row visually coherent. */
    width: var(--thumb-w, 50px);
    height: var(--thumb-h, 70px);
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    --card-w: var(--thumb-w, 50px);
    --card-h: var(--thumb-h, 70px);
  }
  .empty-slot {
    width: 100%;
    height: 100%;
    border-radius: 4px;
    border: 1px dashed var(--border-strong);
    background: transparent;
    box-sizing: border-box;
    padding: 0;
  }
  .empty-slot-btn {
    cursor: pointer;
  }
  .empty-slot-btn:hover {
    border-color: var(--accent);
  }
  .tax-badge {
    position: absolute;
    top: 3px;
    right: 3px;
    background: rgba(8, 7, 6, 0.9);
    color: var(--accent-strong);
    font-family: var(--font-mono);
    font-size: 9px;
    font-weight: 700;
    padding: 1px 5px;
    border-radius: 999px;
    border: 1px solid var(--accent-line);
    pointer-events: none;
  }
  .meta {
    display: flex;
    flex-direction: column;
    align-items: center;
    line-height: 1.2;
    gap: 2px;
  }
  .hints {
    display: flex;
    gap: 3px;
    flex-wrap: wrap;
    justify-content: center;
  }
  .label {
    font-family: var(--font-mono);
    font-size: 8px;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--accent-strong);
    font-weight: 600;
  }
  .count {
    font-size: 14px;
    font-weight: 800;
    color: var(--fg);
    font-variant-numeric: tabular-nums;
    letter-spacing: -0.01em;
  }
  .cycle,
  .cast-hint,
  .ability-hint,
  .browse-hint {
    background: transparent;
    color: var(--accent-strong);
    border: 1px solid color-mix(in srgb, var(--accent) 35%, transparent);
    border-radius: 999px;
    font-family: var(--font-mono);
    font-size: 8px;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    padding: 1px 6px;
    cursor: pointer;
    font-weight: 700;
    box-shadow: none;
    line-height: 1.4;
    transition:
      background 120ms var(--ease),
      color 120ms var(--ease),
      border-color 120ms var(--ease);
  }
  .browse-hint {
    /* Dim "browse" so it reads as the secondary affordance next to
       the primary cast / cycle button. */
    color: var(--fg-dim);
    border-color: var(--border);
  }
  .cycle:hover,
  .cast-hint:hover,
  .ability-hint:hover {
    background: var(--accent-soft);
    color: var(--accent-strong);
    border-color: var(--accent);
  }
  /* ADR 0105: the hint is a gate. With no cast move it greys and
     withholds the click, like an uncastable hand card; with one, and
     highlights live, it takes the ready accent. */
  .cast-hint:disabled {
    cursor: not-allowed;
    color: var(--fg-dim);
    border-color: var(--border);
    opacity: 0.6;
  }
  .cast-hint:disabled:hover {
    background: transparent;
    color: var(--fg-dim);
    border-color: var(--border);
  }
  .cast-hint.ready:not(:disabled) {
    color: var(--ready);
    border-color: var(--ready);
    background: var(--ready-soft);
  }
  .browse-hint:hover {
    background: var(--accent-soft);
    color: var(--accent-strong);
    border-color: var(--accent);
  }
  .cast-hint:focus-visible,
  .ability-hint:focus-visible,
  .cycle:focus-visible {
    outline: 1px solid var(--accent);
    outline-offset: 1px;
  }
  .browse-hint:focus-visible {
    outline: 1px solid var(--accent);
    outline-offset: 1px;
  }
</style>
