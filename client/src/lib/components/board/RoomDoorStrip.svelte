<script lang="ts">
  // RoomDoorStrip — a Room's two doors on its card tile (CR 709.5, ADR
  // 0103). One segment per half: its name and whether it is locked.
  // A locked door whose unlock the server offers becomes a button,
  // "Unlock {cost}", for the Room's controller (owner decision 4); the
  // same unlock is also a row in the card's right-click menu.
  //
  // The button asks through the roomDoors store and the Board sends
  // the special_action, so this component sends nothing itself. A row
  // the server says is not available right now (not your main phase,
  // the stack is not empty) is drawn greyed with the reason, never
  // hidden: a player has to be able to see the door is there.
  import type { CardView } from "../../protocol";
  import { doorRows, requestUnlock, unlockPrice } from "../../roomDoors";

  interface Props {
    card: CardView;
    // True when the viewer controls this Room — the only seat whose
    // unlock the server accepts (CR 709.5e).
    canUnlock: boolean;
  }

  const { card, canUnlock }: Props = $props();

  const rows = $derived(doorRows(card));
</script>

{#if rows.length === 2}
  <div class="doors" role="group" aria-label="Room doors">
    {#each rows as row (row.door)}
      <div class="door" class:unlocked={row.unlocked} data-door={row.door}>
        <span class="door-name" title={row.name}>
          <span class="lock" aria-hidden="true">{row.unlocked ? "🔓" : "🔒"}</span>
          {row.name}
        </span>
        {#if !row.unlocked && canUnlock && row.unlock}
          <button
            type="button"
            class="unlock"
            disabled={!row.unlock.available}
            title={row.unlock.available ? `Unlock ${row.name}` : "not right now"}
            aria-label={`Unlock ${row.name} for ${unlockPrice(row)}`}
            onclick={(e) => {
              e.stopPropagation();
              requestUnlock(card.instance_id, row.door);
            }}
          >
            Unlock {unlockPrice(row)}
          </button>
        {/if}
      </div>
    {/each}
  </div>
{/if}

<style>
  .doors {
    position: absolute;
    left: 2px;
    right: 2px;
    bottom: 2px;
    display: flex;
    gap: 2px;
    z-index: 3;
    pointer-events: none;
  }

  .door {
    flex: 1 1 0;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 1px;
    padding: 1px 2px;
    border-radius: 4px;
    background: rgba(10, 14, 26, 0.82);
    border: 1px solid rgba(255, 255, 255, 0.12);
    color: rgba(255, 255, 255, 0.6);
    font-size: 8px;
    line-height: 1.2;
  }

  .door.unlocked {
    color: #fff;
    border-color: var(--accent, #d9a441);
  }

  .door-name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .lock {
    font-size: 7px;
  }

  .unlock {
    pointer-events: auto;
    font: inherit;
    font-weight: 600;
    padding: 0 2px;
    border-radius: 3px;
    border: 1px solid var(--accent, #d9a441);
    background: rgba(217, 164, 65, 0.2);
    color: #fff;
    cursor: pointer;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .unlock:disabled {
    opacity: 0.5;
    cursor: default;
  }
</style>
