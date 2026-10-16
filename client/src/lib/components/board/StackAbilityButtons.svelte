<script lang="ts">
  // StackAbilityButtons — the activate buttons on a spell that has an
  // ability working on the stack (#1865, ADR 0106 §1 amendment
  // 2026-10-16). Lightning Storm is the one printed card: "Discard a land
  // card: Put two charge counters on Lightning Storm. You may choose a new
  // target for it. Any player may activate this ability but only if
  // Lightning Storm is on the stack."
  //
  // One button per row the server sent this viewer (`item.abilities`,
  // the stack card's `zone_abilities`), which it sends only to a seat
  // that may activate it. Every stack surface (the docked card and the
  // four lane styles) draws this beside its Counter button, so the
  // affordance and its rules live in one place. A click hands the row
  // to Board's announce chain, which asks for the land card to discard
  // as it does for a cycling or a graveyard ability.
  //
  // Live only while the viewer holds priority (CR 117.1b); the engine
  // refuses anything else. A row the server greys (timing, a board-wide
  // ban) stays greyed with its reason.
  import type { StackLaneItem } from "../../stackLane";
  import { L } from "../../labels";

  interface Props {
    item: StackLaneItem;
    /** The viewer holds priority. */
    enabled: boolean;
    /** Undefined: nothing can announce here (a spectator), so no buttons. */
    onActivate?: (item: StackLaneItem, abilityIndex: number) => void;
    small?: boolean;
  }

  const { item, enabled, onActivate, small = false }: Props = $props();

  // The cost half of a "cost: effect" row, for the button's name.
  function costOf(label: string | undefined): string {
    if (!label) return "Activate";
    const at = label.indexOf(":");
    return at > 0 ? label.slice(0, at).trim() : label;
  }

  function blockedReason(row: StackLaneItem["abilities"][number]): string {
    if (row.cant_activate) return row.cant_activate;
    if (row.timing_closed) return "not now";
    if (!enabled) return "you don't hold priority";
    return "";
  }
</script>

{#if onActivate}
  {#each item.abilities as row (row.index)}
    {@const why = blockedReason(row)}
    <button
      type="button"
      class="act ability-btn"
      class:small
      disabled={why !== ""}
      aria-label={L.stackAbility(item.name, costOf(row.label))}
      title={why === "" ? (row.label ?? "") : `${row.label ?? ""} (${why})`}
      onclick={(e) => {
        e.stopPropagation();
        onActivate(item, row.index);
      }}
    >
      {costOf(row.label)}
    </button>
  {/each}
{/if}

<style>
  .ability-btn {
    flex: 0 0 auto;
    max-width: 14em;
    height: 24px;
    padding: 0 9px;
    overflow: hidden;
    font-size: 11px;
    white-space: nowrap;
    text-overflow: ellipsis;
    border-radius: 6px;
    color: var(--fg);
    background: var(--accent-soft);
    border: 1px solid var(--accent-line);
  }
  .ability-btn.small {
    position: relative;
    z-index: 3;
    height: 22px;
    padding: 0 7px;
    font-size: 10.5px;
  }
  .ability-btn:hover:not(:disabled) {
    border-color: var(--accent);
  }
  .ability-btn:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
</style>
