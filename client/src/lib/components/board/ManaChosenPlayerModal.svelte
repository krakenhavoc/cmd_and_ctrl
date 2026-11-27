<script lang="ts">
  // ManaChosenPlayerModal — #6101: "{T}: Choose a player. That player
  // adds one mana of any color they choose." (Spectral Searchlight).
  //
  // Opens in the mana-ability cost chain before the activation is sent.
  // One button per player still in the game, the activator first as
  // "You". Choosing yourself is an ordinary colour pick after the tap;
  // choosing anyone else taps the source and asks THEM the colour, and
  // the mana goes to their pool.
  //
  // A sheet in the action dock, like the other cost pickers: the confirm
  // and Cancel are the dock's action bar.
  import { onDestroy } from "svelte";
  import type { CardView } from "../../protocol";
  import type { ChosenPlayerOption } from "../../manaAbilityCost";
  import { cancelAction, confirmAction } from "../../dock";
  import { L } from "../../labels";
  import DockSheet from "./DockSheet.svelte";

  interface Props {
    // The source; null closes the sheet.
    card: CardView | null;
    players: ChosenPlayerOption[];
    onConfirm: (playerID: string) => void;
    onCancel: () => void;
  }

  const { card, players, onConfirm, onCancel }: Props = $props();

  let chosen = $state(0);
  let lastCardID: string | null = null;
  $effect(() => {
    const id = card?.instance_id ?? null;
    if (id !== lastCardID) {
      lastCardID = id;
      chosen = 0;
    }
  });

  const picked = $derived(players[chosen]);
  const verb = $derived(
    !picked ? "Choose" : picked.you ? "Add mana yourself" : `Give the mana to ${picked.name}`,
  );

  function confirm(): void {
    if (picked) onConfirm(picked.id);
  }

  // The arrows and 1-9 pick a player while the sheet is open; Enter and
  // Escape are the dock's.
  function handleKey(e: KeyboardEvent): void {
    if (!card) return;
    if (e.key === "ArrowLeft" || e.key === "ArrowUp") {
      e.preventDefault();
      chosen = Math.max(0, chosen - 1);
    } else if (e.key === "ArrowRight" || e.key === "ArrowDown") {
      e.preventDefault();
      chosen = Math.min(players.length - 1, chosen + 1);
    } else if (e.key >= "1" && e.key <= "9") {
      const n = Number(e.key) - 1;
      if (n < players.length) {
        e.preventDefault();
        chosen = n;
      }
    }
  }
  $effect(() => {
    if (!card) return;
    document.addEventListener("keydown", handleKey);
    return () => document.removeEventListener("keydown", handleKey);
  });
  onDestroy(() => document.removeEventListener("keydown", handleKey));
</script>

{#if card && players.length > 0}
  <DockSheet
    label={L.manaChosenPlayer}
    title={card.name}
    src="choose a player"
    width={420}
    sheetKey={`mana-chosen-player:${card.instance_id}`}
    primary={confirmAction(verb, confirm)}
    secondary={[cancelAction(onCancel)]}
  >
    <p class="prompt-hint">
      That player adds one mana of any color they choose. If you choose someone else, they pick the
      color and the mana goes to their pool.
    </p>
    <ul class="player-options">
      {#each players as p, i (p.id)}
        <li>
          <button
            type="button"
            class="player-opt"
            class:on={chosen === i}
            aria-pressed={chosen === i}
            onclick={() => (chosen = i)}
            ondblclick={() => {
              chosen = i;
              confirm();
            }}
          >
            <span class="player-name">{p.you ? "You" : p.name}</span>
          </button>
        </li>
      {/each}
    </ul>
  </DockSheet>
{/if}

<style>
  .prompt-hint {
    margin: 0 0 10px;
    font-size: 13px;
    opacity: 0.85;
  }

  .player-options {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    list-style: none;
    margin: 0;
    padding: 0;
  }

  .player-opt {
    min-width: 120px;
    min-height: 44px;
    padding: 8px 12px;
    border-radius: 10px;
    border: 2px solid transparent;
    background: var(--surface-sunken);
    color: inherit;
    cursor: pointer;
    text-align: left;
  }

  .player-opt:hover {
    border-color: var(--border-strong, rgba(255, 255, 255, 0.25));
  }

  .player-opt.on {
    border-color: var(--accent, #d9a441);
  }

  .player-opt:focus-visible {
    outline: 2px solid var(--accent, #d9a441);
    outline-offset: 2px;
  }

  .player-name {
    font-weight: 600;
  }
</style>
