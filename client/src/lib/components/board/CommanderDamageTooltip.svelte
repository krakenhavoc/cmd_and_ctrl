<script lang="ts">
  // CommanderDamageTooltip is the hover-activated replacement for the
  // retired CommanderDamageGrid matrix. It reads the shared
  // `hoveredCard` store and, when the hovered card IS a commander,
  // renders a compact list of "<opponent-name>: <N>" rows showing how
  // much commander damage THIS specific commander instance has dealt
  // to each other player.
  //
  // Why instance-scoped:
  //   PlayerView.commander_damage is keyed by the commander's
  //   instance_id, so a partner-deck commander pair shows two
  //   independent totals (one per card) — the old NxN matrix
  //   flattened those into a single row per player and couldn't
  //   disambiguate.
  //
  // Data shape:
  //   PlayerView.commander_damage holds damage TAKEN by that player,
  //   keyed by the commander instance_id that dealt it. So to compute
  //   "how much has commander X dealt to player P", we read
  //   seatP.commander_damage[X.instance_id].
  //
  // Visual language matches PhaseDisplay / priority pills (--surface
  // bg, --border outline, rounded, tight padding) so the tooltip
  // reads as part of the existing chrome rather than a modal.

  import type { GameView, PlayerView } from "../../protocol";
  import { hoveredCard } from "../../cardTypes";
  import { playerColor } from "../../avatarColor";
  import { seatColor } from "../../colors";

  interface Props {
    view: GameView;
  }

  const { view }: Props = $props();

  const card = $derived($hoveredCard);
  const active = $derived(card !== null && card.is_commander === true);

  // Per-seat accent colour map, mirroring PhaseDisplay's pattern:
  // seeded from seatColor() and flipped async to the avatar-derived
  // hex once extraction finishes. Keeps colouring consistent with
  // priority pills / phase display.
  const playerColors = $state<Record<string, string>>({});
  $effect(() => {
    for (const s of view.seats) {
      const id = s.id;
      playerColors[id] = playerColor(s, (c) => {
        playerColors[id] = c;
      });
    }
  });
  const colorFor = (seat: PlayerView): string => playerColors[seat.id] ?? seatColor(seat.seat);

  // Rows: every seat OTHER than the commander's controller, with the
  // damage this specific instance has dealt to them (0 when none).
  // Controller is excluded on CR grounds — a player can't deal
  // commander damage to themselves — and because showing a self-row
  // would just muddy the readout.
  const rows = $derived.by(() => {
    if (!card) return [];
    const instanceID = card.instance_id;
    const controllerID = card.controller;
    return view.seats
      .filter((s) => s.id !== controllerID)
      .map((s) => ({
        seat: s,
        amount: s.commander_damage?.[instanceID] ?? 0,
      }));
  });

  const hasAnyDamage = $derived(rows.some((r) => r.amount > 0));
</script>

{#if active && card}
  <div class="tooltip" role="tooltip" aria-label="commander damage dealt">
    <header class="head">
      <span class="dot" aria-hidden="true"></span>
      <span class="label">cmdr dmg dealt</span>
    </header>
    {#if hasAnyDamage}
      <ul class="rows">
        {#each rows as row (row.seat.id)}
          <li class="row" class:zero={row.amount === 0}>
            <span class="name" style:--seat-color={colorFor(row.seat)} title={row.seat.name}>
              <span class="name-dot" aria-hidden="true"></span>
              <span class="name-text">{row.seat.name}</span>
            </span>
            <span class="amount" class:lethal={row.amount >= 21}>
              {row.amount}
            </span>
          </li>
        {/each}
      </ul>
    {:else}
      <div class="empty">no commander damage dealt yet</div>
    {/if}
  </div>
{/if}

<style>
  .tooltip {
    position: absolute;
    right: 12px;
    bottom: 12px;
    min-width: 160px;
    max-width: 240px;
    padding: 8px 10px;
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 0.04) 0%, rgba(0, 0, 0, 0.2) 100%), var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    box-shadow: var(--shadow-sm);
    color: var(--fg);
    font-size: 11px;
    z-index: 35;
    pointer-events: none;
  }
  .head {
    display: flex;
    align-items: center;
    gap: 6px;
    padding-bottom: 6px;
    border-bottom: 1px solid var(--border);
    margin-bottom: 6px;
  }
  .dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--gold);
    display: inline-block;
  }
  .label {
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 0.14em;
    color: var(--fg-dim);
    font-weight: 700;
  }
  .rows {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 3px;
  }
  .row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    padding: 1px 0;
  }
  .row.zero {
    opacity: 0.55;
  }
  .name {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    min-width: 0;
    color: var(--fg);
  }
  .name-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--seat-color, #888);
    flex: 0 0 auto;
  }
  .name-text {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 14ch;
  }
  .amount {
    font-variant-numeric: tabular-nums;
    font-weight: 700;
    color: var(--gold);
    min-width: 2ch;
    text-align: right;
  }
  .amount.lethal {
    color: var(--danger);
    text-shadow: 0 0 6px rgba(255, 122, 122, 0.55);
  }
  .row.zero .amount {
    color: var(--fg-dim);
    font-weight: 500;
  }
  .empty {
    font-size: 10px;
    color: var(--fg-dim);
    font-style: italic;
    text-align: center;
    padding: 2px 0;
  }
</style>
