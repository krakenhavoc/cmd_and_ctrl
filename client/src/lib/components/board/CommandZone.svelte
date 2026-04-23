<script lang="ts">
  // CommandZone is the first-class commander slot on each PlayerPanel.
  // Replaces the small face-down cmd PileButton with a face-up,
  // card-sized rendering of the commander(s) in the zone, plus a
  // commander-tax badge.
  //
  // Click semantics on the viewer's own zone:
  //   - empty: no-op
  //   - one commander present: send move_card from command → battlefield
  //   - multiple commanders (partner / Background): cycle visible top
  //     and click the visible one to cast it. Partner support on the
  //     deck-import side is currently `ErrUnsupportedMechanic`, so
  //     this path is forward-compatible padding rather than today's
  //     primary flow.
  //
  // Casts increment a local tax counter shown in the UI. The tax is
  // not yet tracked server-side (a future S13+ rules graft item once
  // the cast pipeline exists); for now it's a viewer hint that mirrors
  // what a casting player would track on paper.

  import type { ActionPayload, CardView, ZoneView } from "../../protocol";
  import Card from "./Card.svelte";
  import { openZoneBrowser } from "../../zoneBrowser";

  type ActionSender = (type: string, params?: ActionPayload["params"], player?: string) => void;

  interface Props {
    seat: { id: string; name: string };
    zone: ZoneView;
    isSelf: boolean;
    sendAction: ActionSender;
  }

  const { seat, zone, isSelf, sendAction }: Props = $props();

  let visibleIndex = $state(0);
  const commanders = $derived(zone.cards);
  const visibleCard = $derived<CardView | null>(
    commanders.length > 0 ? (commanders[visibleIndex % commanders.length] ?? null) : null,
  );

  // Local-only tax counter (++ per cast). Reset when the underlying
  // commander instance changes (the same commander returning to the
  // command zone keeps the same instance ID, so its tax persists; a
  // freshly-imported deck or a new game gets a fresh count).
  let castCounts = $state<Record<string, number>>({});
  const visibleTax = $derived(visibleCard ? (castCounts[visibleCard.instance_id] ?? 0) * 2 : 0);

  function cycleVisible(): void {
    if (commanders.length <= 1) return;
    visibleIndex = (visibleIndex + 1) % commanders.length;
  }

  function castVisible(): void {
    if (!isSelf || !visibleCard) return;
    const id = visibleCard.instance_id;
    sendAction(
      "move_card",
      {
        src: { kind: "command", owner: seat.id },
        dst: { kind: "battlefield" },
        instance_id: id,
      },
      seat.id,
    );
    castCounts = { ...castCounts, [id]: (castCounts[id] ?? 0) + 1 };
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

<style>
  .cmd-zone {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    padding: 6px 4px 4px;
    background:
      linear-gradient(180deg, rgba(255, 208, 122, 0.08) 0%, rgba(0, 0, 0, 0.25) 100%), #0f1324;
    border: 1px solid #5a4520;
    border-radius: var(--radius);
    box-sizing: border-box;
    width: var(--pile-w, 64px);
  }
  .cmd-zone.self {
    border-color: var(--gold);
    box-shadow:
      0 0 12px rgba(255, 208, 122, 0.28),
      inset 0 1px 0 rgba(255, 208, 122, 0.12);
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
    border: 1px dashed #2e3a55;
    background: transparent;
    box-sizing: border-box;
    padding: 0;
  }
  .empty-slot-btn {
    cursor: pointer;
  }
  .empty-slot-btn:hover {
    border-color: var(--accent, #6a8dff);
  }
  .tax-badge {
    position: absolute;
    top: 3px;
    right: 3px;
    background: linear-gradient(180deg, #8a6a2e 0%, #5a4520 100%);
    color: var(--gold);
    font-size: 10px;
    font-weight: 800;
    padding: 2px 6px;
    border-radius: 999px;
    border: 1px solid rgba(255, 208, 122, 0.55);
    box-shadow:
      0 2px 6px rgba(0, 0, 0, 0.45),
      inset 0 1px 0 rgba(255, 255, 255, 0.18);
    pointer-events: none;
    text-shadow: 0 1px 0 rgba(0, 0, 0, 0.6);
  }
  .meta {
    display: flex;
    flex-direction: column;
    align-items: center;
    line-height: 1.2;
    gap: 2px;
  }
  .label {
    font-size: 8px;
    text-transform: uppercase;
    letter-spacing: 0.12em;
    color: #c8a86a;
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
    margin-top: 2px;
    background: transparent;
    color: #c8a86a;
    border: 1px solid #5a4520;
    border-radius: 999px;
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    padding: 2px 8px;
    cursor: pointer;
    font-family: inherit;
    font-weight: 700;
    box-shadow: none;
    transition:
      background 120ms var(--ease),
      color 120ms var(--ease),
      border-color 120ms var(--ease);
  }
  .browse-hint {
    /* Dim "browse" so it reads as the secondary affordance next to
       the primary cast / cycle button. */
    color: var(--fg-muted, #8a93a8);
    border-color: rgba(255, 255, 255, 0.12);
  }
  .cycle:hover,
  .cast-hint:hover {
    background: rgba(255, 208, 122, 0.15);
    color: var(--gold);
    border-color: var(--gold);
  }
  .browse-hint:hover {
    background: rgba(122, 167, 255, 0.12);
    color: var(--accent);
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
