<script lang="ts">
  // CommandStrip — another seat's command zone, beside that seat's hand
  // (#2349). Your own commanders sit in the castable strip beside your
  // hand (ExileStrip, #2202), where you cast or drag them; this is the
  // same place on everyone else's board, so every player's commander is
  // where their hand is. It replaced the command zone tile in the pile
  // rail, the only place an opponent's commander was face up.
  //
  // View only. Each card is the seat's command-zone card, face up, with
  // the hover zoom every Card has and a "+N" commander tax badge (CR
  // 903.8) read off the server's commander_casts. A click opens the
  // command zone browser, as the tile's did. The group carries the
  // tile's accessible name, `<name> command zone, N card(s)`, which the
  // e2e suite and the tutorial's resolver read. Renders nothing for an
  // empty command zone.
  import type { PlayerView } from "../../protocol";
  import { commanderTax } from "../../castStrip";
  import { openZoneBrowser } from "../../zoneBrowser";
  import { L } from "../../labels";
  import Card from "./Card.svelte";

  interface Props {
    seat: Pick<PlayerView, "id" | "name" | "command" | "commander_casts">;
    // A SeatSummary's row: smaller cards, no hand beside them.
    compact?: boolean;
  }

  const { seat, compact = false }: Props = $props();

  const cards = $derived(seat.command?.cards ?? []);

  function browse(): void {
    openZoneBrowser({ zoneKind: "command", ownerID: seat.id, ownerName: seat.name });
  }
</script>

{#if cards.length > 0}
  <div
    class="command-strip"
    class:compact
    role="group"
    aria-label={L.commandZone(seat.name, seat.command?.count ?? cards.length)}
  >
    {#each cards as card (card.instance_id)}
      {@const tax = commanderTax(seat.commander_casts?.[card.instance_id])}
      <div class="slot" title={`${card.name} · command zone`}>
        <Card {card} onClick={browse} />
        {#if tax > 0}
          <span class="tax-badge" title={`commander tax · +${tax} mana`}>+{tax}</span>
        {/if}
      </div>
    {/each}
  </div>
{/if}

<style>
  /* Beside a face-down hand that peeks a third of a card, a commander
     at that size is a sliver, so it is drawn whole at 42% of a creature
     card, and the hover zoom does the rest. PlayerPanel's opponent
     sizing counts that height (H = 1.92h + ~64px), so the creatures
     above still fit. The size is read off the panel's --card-h into
     --strip-h here and handed to the Card one element down: a custom
     property that names itself is invalid. */
  .command-strip {
    --strip-h: calc(var(--card-h, 130px) * 0.42);
    flex: none;
    display: flex;
    gap: 4px;
    align-self: flex-end;
    padding: 0 2px 2px;
  }
  .command-strip.compact {
    --strip-h: 56px;
    align-self: center;
    padding: 0;
  }
  .slot {
    --card-h: var(--strip-h);
    --card-w: calc(var(--strip-h) * 5 / 7);
    position: relative;
    width: var(--card-w);
    height: var(--card-h);
    border-radius: 6px;
    box-shadow: 0 0 0 1px var(--accent-line);
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
</style>
