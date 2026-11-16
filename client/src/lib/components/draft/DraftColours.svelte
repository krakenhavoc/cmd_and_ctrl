<script lang="ts">
  // DraftColours — the colours of Main by coloured mana symbols (#5848
  // §6): a bar split by the symbols in the cards' costs, a row for each
  // colour, and the basic lands a 17-land deck would play by them. The
  // compact look is the one under the picks and across the top of Deck
  // view; the full look is the overlay's.
  //
  // A bar never stands alone: every colour in it has its number in a row.
  import type { DraftCard } from "../../draft";
  import { BASIC_OF, colourStats } from "../../draftDeckStats";
  import { L } from "../../labels";

  interface Props {
    main: DraftCard[];
    /** The strip's look: rows run together, and no card counts. */
    compact?: boolean;
  }
  const { main, compact = false }: Props = $props();

  const stats = $derived(colourStats(main));

  const NAME = {
    W: () => L.draftGroupWhite,
    U: () => L.draftGroupBlue,
    B: () => L.draftGroupBlack,
    R: () => L.draftGroupRed,
    G: () => L.draftGroupGreen,
  } as const;
  // The colour-identity tokens (DESIGN.md §2), which no skin changes.
  const INK = {
    W: "var(--identity-w)",
    U: "var(--identity-u)",
    B: "var(--identity-b)",
    R: "var(--identity-r)",
    G: "var(--identity-g)",
  } as const;
</script>

<div class="colours" class:compact data-testid="draft-colours">
  {#if stats.rows.length === 0}
    <p class="hint">{L.draftStatsNoColour}</p>
  {:else}
    <div class="bar" aria-hidden="true">
      {#each stats.rows as r (r.colour)}
        <i style:width={`${(r.pips / stats.totalPips) * 100}%`} style:background={INK[r.colour]}
        ></i>
      {/each}
    </div>
    <ul class="rows">
      {#each stats.rows as r (r.colour)}
        <li data-colour={r.colour}>
          <span class="dot" style:--dot={INK[r.colour]} aria-hidden="true"></span>
          <span class="l">{NAME[r.colour]()}</span>
          <b data-pips={r.colour}>{r.pips}</b>
          {#if !compact}<span class="dim">{L.draftStatsCards(r.cards)}</span>{/if}
        </li>
      {/each}
    </ul>
    <p class="hint" data-testid="draft-basics">
      {L.draftStatsBasics(stats.basics.map(([c, n]) => `${n} ${BASIC_OF[c]}`).join(" · "))}
    </p>
  {/if}
</div>

<style>
  .colours {
    min-width: 0;
  }
  .bar {
    display: flex;
    height: 10px;
    margin-bottom: 8px;
    border-radius: 5px;
    overflow: hidden;
    background: var(--surface-sunken);
    box-shadow: inset 0 0 0 1px var(--border);
  }
  .bar i {
    display: block;
    height: 100%;
  }
  .rows {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 3px;
    font-size: 12.5px;
  }
  .compact .rows {
    flex-direction: row;
    flex-wrap: wrap;
    gap: 2px 14px;
  }
  li {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .l {
    flex: 1;
  }
  .compact .l {
    /* A colour's name is not what a glance wants: the dot says it. */
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip-path: inset(50%);
    white-space: nowrap;
    flex: none;
  }
  b {
    font-family: var(--font-mono);
    font-variant-numeric: tabular-nums;
    font-weight: 600;
  }
  .dim {
    min-width: 54px;
    text-align: right;
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--fg-muted);
  }
  .dot {
    flex: none;
    width: 10px;
    height: 10px;
    border-radius: 999px;
    background: var(--dot);
    border: 1px solid var(--border-strong);
    box-sizing: border-box;
  }
  .hint {
    margin: 6px 0 0;
    font-size: 12px;
    color: var(--fg-muted);
  }
</style>
