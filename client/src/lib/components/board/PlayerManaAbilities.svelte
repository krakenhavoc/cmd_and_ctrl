<script lang="ts">
  // PlayerManaAbilities — the mana abilities the viewer has as a PLAYER,
  // bound to no card (#4885): Channel's "until end of turn, any time you
  // could activate a mana ability, you may pay 1 life. If you do, add
  // {C}."
  //
  // The server decides everything: `seat.mana_abilities` lists each live
  // one with its price, its mana and whether the life can be paid now
  // (CR 119.4). One press is one activation:
  //   activate_mana_ability { player_ability: <source> }
  // It uses no stack, so it works wherever a card's mana ability does:
  // with priority, or while a "pay {N}" prompt is open. The mana lands in
  // the pool beside this button and pays the next cost from there. The
  // auto-tapper never spends the life for you; this button is the only
  // way it is spent.
  //
  // Renders nothing when the seat has none, which is almost always.
  import type { ActionPayload, ActionType, PlayerManaAbilityView } from "../../protocol";
  import { L } from "../../labels";
  import ManaCost from "./ManaCost.svelte";

  type ActionSender = (type: ActionType, params?: ActionPayload["params"], player?: string) => void;

  interface Props {
    seatID: string;
    abilities: PlayerManaAbilityView[] | undefined;
    sendAction: ActionSender;
  }

  const { seatID, abilities, sendAction }: Props = $props();

  function activate(row: PlayerManaAbilityView): void {
    if (!row.usable) return;
    sendAction("activate_mana_ability", { player_ability: row.source }, seatID);
  }
</script>

{#if abilities && abilities.length > 0}
  <div class="player-mana" role="group" aria-label="Your mana abilities">
    {#each abilities as row (row.source)}
      <button
        type="button"
        class="player-mana-ability"
        aria-label={L.playerManaAbility(row.label, row.life_cost, row.produced)}
        title={row.usable
          ? `${row.label}: pay ${row.life_cost} life to add ${row.produced}, until end of turn`
          : `You can't pay ${row.life_cost} life`}
        disabled={!row.usable}
        onclick={() => activate(row)}
      >
        <span class="name">{row.label}</span>
        <span class="price">{row.life_cost} life →</span>
        <ManaCost cost={row.produced} size={12} label={`adds ${row.produced}`} />
      </button>
    {/each}
  </div>
{/if}

<style>
  .player-mana {
    flex: none;
    display: flex;
    flex-direction: column;
    gap: 3px;
    align-self: flex-end;
    padding: 0 2px 2px;
  }
  .player-mana-ability {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-family: var(--font-mono);
    font-size: 10px;
    padding: 2px 8px;
    border-radius: 999px;
    border: 1px solid var(--accent-strong);
    background: rgba(8, 7, 6, 0.85);
    color: var(--accent-strong);
    cursor: pointer;
    white-space: nowrap;
  }
  .player-mana-ability:hover:not(:disabled),
  .player-mana-ability:focus-visible {
    box-shadow: 0 0 0 1px var(--accent-strong);
  }
  .player-mana-ability:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
  .name {
    font-weight: 700;
  }
</style>
