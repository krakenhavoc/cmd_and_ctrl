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
</script>

<div
  class="cmd-zone"
  class:self={isSelf}
  class:empty={commanders.length === 0}
  aria-label={`${seat.name} command zone, ${zone.count} card${zone.count === 1 ? "" : "s"}`}
>
  <div class="card-slot">
    {#if visibleCard}
      <Card card={visibleCard} onClick={isSelf ? handleClick : undefined} />
      {#if visibleTax > 0}
        <span class="tax-badge" title={`commander tax · +${visibleTax} mana`}>+{visibleTax}</span>
      {/if}
    {:else}
      <div class="empty-slot" aria-hidden="true"></div>
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
</div>

<style>
  .cmd-zone {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
    padding: 4px;
    background: #111a2b;
    border: 1px solid #4a3a1a;
    border-radius: 6px;
    box-sizing: border-box;
    /* Width matches the regular pile column so the bar still aligns,
       but the card slot inside is the full --card-w/--card-h so the
       commander reads as a real card rather than a thumbnail. */
    width: var(--pile-w, 64px);
    /* Self gets the brighter "this is your commander" border. */
  }
  .cmd-zone.self {
    border-color: #ffd07a;
    box-shadow: 0 0 8px rgba(255, 208, 122, 0.18);
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
  }
  .tax-badge {
    position: absolute;
    top: 2px;
    right: 2px;
    background: #4a3a1a;
    color: #ffd07a;
    font-size: 9px;
    font-weight: 800;
    padding: 1px 4px;
    border-radius: 3px;
    pointer-events: none;
    text-shadow: 0 1px 0 rgba(0, 0, 0, 0.6);
  }
  .meta {
    display: flex;
    flex-direction: column;
    align-items: center;
    line-height: 1.1;
  }
  .label {
    font-size: 8px;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: #c8a86a;
  }
  .count {
    font-size: 13px;
    font-weight: 700;
    color: #e0e6f5;
  }
  .cycle,
  .cast-hint {
    margin-top: 2px;
    background: transparent;
    color: #c8a86a;
    border: 1px solid #4a3a1a;
    border-radius: 3px;
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    padding: 1px 4px;
    cursor: pointer;
    font-family: inherit;
  }
  .cycle:hover,
  .cast-hint:hover {
    background: #4a3a1a;
    color: #ffd07a;
  }
  .cast-hint:focus-visible,
  .cycle:focus-visible {
    outline: 1px solid #ffd07a;
    outline-offset: 1px;
  }
</style>
