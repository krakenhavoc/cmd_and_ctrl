<script lang="ts">
  // CommandZone is the first-class commander slot on each PlayerPanel.
  // Replaces the small face-down cmd PileButton with a face-up,
  // card-sized rendering of the commander(s) in the zone, plus a
  // commander-tax badge.
  //
  // Click semantics on the viewer's own zone:
  //   - empty: no-op
  //   - one commander present: cast_spell from the command zone — the
  //     cast goes through the real pipeline (stack, sorcery-speed +
  //     priority gates, strict-mana with CR 903.8 tax, ETB triggers)
  //     and the server increments CommanderCasts on success.
  //   - multiple commanders (partner / Background): cycle visible top
  //     and click the visible one to cast it. Partner support on the
  //     deck-import side is currently `ErrUnsupportedMechanic`, so
  //     this path is forward-compatible padding rather than today's
  //     primary flow.
  //
  // The "+N tax" badge reads the server-tracked commander_casts map
  // (S13.1) off the seat's PlayerView — no local bookkeeping, so every
  // viewer (not just the caster's tab) sees the same tax.

  import type { ActionPayload, ActionType, CardView, ZoneView } from "../../protocol";
  import Card from "./Card.svelte";
  import { openZoneBrowser } from "../../zoneBrowser";

  type ActionSender = (type: ActionType, params?: ActionPayload["params"], player?: string) => void;

  interface Props {
    seat: { id: string; name: string };
    zone: ZoneView;
    isSelf: boolean;
    sendAction: ActionSender;
    // Server-side per-commander cast counts (PlayerView.commander_casts),
    // keyed by commander instance UUID. Drives the "+N tax" badge.
    commanderCasts?: Record<string, number>;
  }

  const { seat, zone, isSelf, sendAction, commanderCasts }: Props = $props();

  let visibleIndex = $state(0);
  const commanders = $derived(zone.cards);
  const visibleCard = $derived<CardView | null>(
    commanders.length > 0 ? (commanders[visibleIndex % commanders.length] ?? null) : null,
  );

  // CR 903.8: each prior cast of this commander adds {2}.
  const visibleTax = $derived(
    visibleCard ? (commanderCasts?.[visibleCard.instance_id] ?? 0) * 2 : 0,
  );

  function cycleVisible(): void {
    if (commanders.length <= 1) return;
    visibleIndex = (visibleIndex + 1) % commanders.length;
  }

  function castVisible(): void {
    if (!isSelf || !visibleCard) return;
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
  aria-label={`${seat.name} command zone, ${zone.count} card${zone.count === 1 ? "" : "s"}`}
>
  <div class="card-slot">
    {#if visibleCard}
      <Card card={visibleCard} onClick={isSelf ? handleClick : openBrowser} />
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
        onclick={castVisible}
        ondblclick={handleDoubleClick}
        title="cast commander"
        aria-label="cast commander"
      >
        cast
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
    border: 1px solid rgba(217, 180, 92, 0.25);
    border-radius: var(--radius);
    box-sizing: border-box;
    width: 100%;
  }
  .cmd-zone.self {
    border-color: rgba(217, 180, 92, 0.5);
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
    color: var(--gold-strong);
    font-family: var(--font-mono);
    font-size: 9px;
    font-weight: 700;
    padding: 1px 5px;
    border-radius: 999px;
    border: 1px solid rgba(217, 180, 92, 0.5);
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
    color: var(--gold-strong);
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
  .browse-hint {
    background: transparent;
    color: var(--gold-strong);
    border: 1px solid rgba(217, 180, 92, 0.35);
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
  .cast-hint:hover {
    background: var(--accent-soft);
    color: var(--gold-strong);
    border-color: var(--gold);
  }
  .browse-hint:hover {
    background: var(--accent-soft);
    color: var(--accent-strong);
    border-color: var(--accent);
  }
  .cast-hint:focus-visible,
  .cycle:focus-visible {
    outline: 1px solid var(--gold);
    outline-offset: 1px;
  }
  .browse-hint:focus-visible {
    outline: 1px solid var(--accent);
    outline-offset: 1px;
  }
</style>
