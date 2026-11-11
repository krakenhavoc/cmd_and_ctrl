<script lang="ts">
  // DraftStats — what the picks are made of (#5702): creatures against
  // the rest, a count for each card type, the mana curve and the
  // colours, for Main and for the whole pool. It is on the pick screen
  // and on the deckbuilding screen, and follows the cards as they move
  // between Main and the Sideboard.
  //
  // The tallies are lib/draftStats.ts's; this draws them. Every figure
  // is a number first: a bar or a colour dot never stands alone.
  import type { DraftCard } from "../../draft";
  import { CARD_TYPES, curveSummary, STAT_COLOURS, tally, type CardType } from "../../draftStats";
  import { L } from "../../labels";

  interface Props {
    /** The cards in Main. */
    main: DraftCard[];
    /** Every card picked: Main and the Sideboard. */
    pool: DraftCard[];
  }
  const { main, pool }: Props = $props();

  const scopes = $derived([
    { key: "main", name: L.draftDeckZone, curveName: L.draftCurve, stats: tally(main) },
    { key: "pool", name: L.draftStatsPool, curveName: L.draftCurvePool, stats: tally(pool) },
  ]);

  const typeName = (t: CardType): string => {
    switch (t) {
      case "creature":
        return L.draftTypeCreature;
      case "instant":
        return L.draftTypeInstant;
      case "sorcery":
        return L.draftTypeSorcery;
      case "artifact":
        return L.draftTypeArtifact;
      case "enchantment":
        return L.draftTypeEnchantment;
      case "planeswalker":
        return L.draftTypePlaneswalker;
      case "land":
        return L.draftGroupLands;
    }
  };

  const COLOUR_NAME = {
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
    multi: "var(--fg)",
    colourless: "var(--identity-c)",
  } as const;
</script>

<section class="card stats" aria-labelledby="draft-stats-h" data-testid="draft-stats">
  <h2 id="draft-stats-h">{L.draftStats}</h2>
  <div class="scopes">
    {#each scopes as s (s.key)}
      {@const st = s.stats}
      {@const tallest = Math.max(1, ...st.curve.map((c) => c.count))}
      <section class="scope" data-scope={s.key} aria-labelledby={`draft-stats-${s.key}-h`}>
        <h3 id={`draft-stats-${s.key}-h`}>
          {s.name} <span class="n">{st.total}</span>
        </h3>
        {#if st.total === 0}
          <p class="none">No cards yet.</p>
        {:else}
          <dl class="pair">
            <div>
              <dt>{L.draftTypeCreature}</dt>
              <dd data-stat="creatures">{st.creatures}</dd>
            </div>
            <div>
              <dt>{L.draftStatNonCreatures}</dt>
              <dd data-stat="non-creatures">{st.nonCreatures}</dd>
            </div>
          </dl>

          <h4>{L.draftStatTypes}</h4>
          <dl class="types">
            {#each CARD_TYPES as t (t)}
              <div class:zero={st.types[t] === 0}>
                <dt>{typeName(t)}</dt>
                <dd data-type={t}>{st.types[t]}</dd>
              </div>
            {/each}
          </dl>

          <h4>{L.draftStatCurve}</h4>
          <div class="curve" role="img" aria-label={`${s.curveName}: ${curveSummary(st.curve)}`}>
            {#each st.curve as col (col.value)}
              <div class="col" aria-hidden="true">
                <span class="count">{col.count}</span>
                <span class="bar" style:height={`${Math.round((col.count / tallest) * 100)}%`}
                ></span>
                <span class="mv">{col.label}</span>
              </div>
            {/each}
          </div>

          <h4>{L.draftStatColours}</h4>
          <ul class="colours">
            {#each STAT_COLOURS as c (c)}
              <li class:zero={st.colours.byColour[c] === 0}>
                <span class="dot" style:--dot={INK[c]} aria-hidden="true"></span>
                <span class="cname">{COLOUR_NAME[c]()}</span>
                <span class="num" data-colour={c}>{st.colours.byColour[c]}</span>
              </li>
            {/each}
            <li class:zero={st.colours.multi === 0}>
              <span class="dot" style:--dot={INK.multi} aria-hidden="true"></span>
              <span class="cname">{L.draftGroupMulticolour}</span>
              <span class="num" data-colour="multi">{st.colours.multi}</span>
            </li>
            <li class:zero={st.colours.colourless === 0}>
              <span class="dot" style:--dot={INK.colourless} aria-hidden="true"></span>
              <span class="cname">{L.draftGroupColourless}</span>
              <span class="num" data-colour="colourless">{st.colours.colourless}</span>
            </li>
          </ul>
        {/if}
      </section>
    {/each}
  </div>
</section>

<style>
  .stats {
    min-width: 0;
  }
  h2 {
    margin: 0 0 12px;
    font-family: var(--font-display);
    font-size: 19px;
    font-weight: 800;
    /* a .card's h2 is a mono eyebrow by default; this is a heading */
    letter-spacing: -0.005em;
    text-transform: none;
    color: var(--fg);
  }
  .scopes {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
    gap: 20px 28px;
  }
  .scope {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  h3 {
    margin: 0;
    font-family: var(--font-display);
    font-size: 15px;
    font-weight: 700;
    color: var(--fg);
  }
  h4 {
    margin: 6px 0 0;
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--fg-muted);
  }
  .n {
    margin-left: 6px;
    font-family: var(--font-mono);
    font-size: 13px;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    color: var(--fg-muted);
  }
  .none {
    margin: 0;
    font-size: 0.8125rem;
    color: var(--fg-muted);
  }
  dl {
    margin: 0;
  }
  dt,
  .cname {
    font-size: 0.8125rem;
    color: var(--fg);
  }
  dd,
  .num {
    margin: 0;
    font-family: var(--font-mono);
    font-size: 0.8125rem;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
    color: var(--fg);
  }
  .pair {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 8px;
  }
  .pair div {
    padding: 8px 10px;
    border-radius: var(--radius);
    background: var(--surface-sunken);
    border: 1px solid var(--border);
  }
  .pair dt {
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--fg-muted);
  }
  .pair dd {
    margin-top: 2px;
    font-size: 20px;
  }
  .types div,
  .colours li {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    padding: 2px 0;
  }
  .types div + div,
  .colours li + li {
    border-top: 1px solid var(--border);
  }
  /* A count of nothing is quieter, never hidden: it is still an answer. */
  .zero dt,
  .zero dd,
  .zero .cname,
  .zero .num {
    color: var(--fg-muted);
    font-weight: 500;
  }
  .colours {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .colours li {
    justify-content: flex-start;
  }
  .colours .num {
    margin-left: auto;
  }
  .dot {
    flex: none;
    width: 9px;
    height: 9px;
    border-radius: 999px;
    background: var(--dot);
    border: 1px solid var(--border-strong);
    box-sizing: border-box;
  }
  .curve {
    display: grid;
    grid-template-columns: repeat(7, minmax(0, 1fr));
    gap: 6px;
    height: 84px;
  }
  .col {
    display: grid;
    grid-template-rows: auto minmax(0, 1fr) auto;
    justify-items: center;
    gap: 3px;
    min-height: 0;
  }
  .col .count,
  .col .mv {
    font-family: var(--font-mono);
    font-size: 10px;
    font-variant-numeric: tabular-nums;
    color: var(--fg-muted);
  }
  .col .bar {
    align-self: end;
    width: 100%;
    min-height: 2px;
    border-radius: var(--radius-sm) var(--radius-sm) 0 0;
    background: var(--accent);
  }
</style>
