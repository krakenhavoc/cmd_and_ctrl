<script lang="ts">
  // DraftArchetypes — a set's ten colour pairs with a short label and
  // the gold uncommons that name each (ADR 0222 §7, #5703): what Arena
  // shows when you choose a draft, so a drafter knows what a pair of
  // colours is for in this set.
  //
  // Each pair is a row: its two mana pips, the written label, and the
  // signpost cards by name. A row with signposts opens to show the cards
  // themselves; their art is not asked for until it is opened. A pair
  // with no label and no signpost says so: the data supports a pair, not
  // a name, in a set whose gold cards are not one per pair.
  //
  // The component fetches its own list (public GET
  // /draft/sets/{code}/archetypes), once per set per page, and draws
  // nothing but a short line while it loads or if it cannot.
  import { onMount } from "svelte";

  import {
    fetchArchetypes,
    pairName,
    pairTitle,
    signpostCard,
    type ArchetypePair,
    type Archetypes,
  } from "../../draftArchetypes";
  import ManaCost from "../board/ManaCost.svelte";
  import DraftCardArt from "./DraftCardArt.svelte";

  interface Props {
    /** The set code to show the pairs of. */
    set: string;
  }
  const { set }: Props = $props();

  let data = $state<Archetypes | null>(null);
  let failed = $state(false);
  // The pairs whose cards the viewer has opened.
  let opened = $state<Record<string, boolean>>({});

  onMount(() => {
    let live = true;
    fetchArchetypes(set)
      .then((a) => {
        if (live) data = a;
      })
      .catch(() => {
        if (live) failed = true;
      });
    return () => {
      live = false;
    };
  });

  const pips = (p: ArchetypePair): string => [...p.colors].map((c) => `{${c}}`).join("");
</script>

{#snippet head(p: ArchetypePair)}
  <span class="pips"><ManaCost cost={pips(p)} size={14} label={pairName(p.colors)} /></span>
  <span class="pair-name">{pairTitle(p.colors)}</span>
  {#if p.label}<span class="label">{p.label}</span>{/if}
  {#if p.signposts.length > 0}
    <span class="names">{p.signposts.map((s) => s.name).join(", ")}</span>
  {:else}
    <span class="names none">No gold uncommon in this set</span>
  {/if}
{/snippet}

<div class="arch" data-testid="archetypes" data-set={set}>
  {#if failed}
    <p class="note" role="status">Couldn’t load this set’s colour pairs.</p>
  {:else if !data}
    <p class="note" aria-live="polite">Loading colour pairs…</p>
  {:else}
    <ul class="pairs">
      {#each data.pairs as p (p.colors)}
        <li class="pair" data-pair={p.colors}>
          {#if p.signposts.length > 0}
            <details ontoggle={(e) => (opened[p.colors] = e.currentTarget.open)}>
              <summary>{@render head(p)}</summary>
              {#if opened[p.colors]}
                <ul class="cards">
                  {#each p.signposts as s (s.id)}
                    <li><DraftCardArt card={signpostCard(s)} alt={s.name} /></li>
                  {/each}
                </ul>
              {/if}
            </details>
          {:else}
            <div class="row">{@render head(p)}</div>
          {/if}
        </li>
      {/each}
    </ul>
  {/if}
</div>

<style>
  .arch {
    min-width: 0;
  }
  .note {
    margin: 0;
    color: var(--fg-muted);
    font-size: 13px;
  }
  .pairs {
    list-style: none;
    margin: 0;
    padding: 0;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--surface);
  }
  .pair + .pair {
    border-top: 1px solid var(--border);
  }
  .row,
  summary {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 2px 10px;
    padding: 8px 12px;
    min-width: 0;
  }
  summary {
    cursor: pointer;
    list-style: none;
    border-radius: var(--radius);
  }
  summary::-webkit-details-marker {
    display: none;
  }
  summary:focus-visible {
    outline: 2px solid var(--accent-strong);
    outline-offset: -2px;
  }
  .pips {
    display: inline-flex;
  }
  /* The colour pair is named in words beside its pips, never pips alone. */
  .pair-name {
    font-size: 13px;
    font-weight: 700;
    color: var(--fg);
    overflow-wrap: anywhere;
  }
  .label {
    margin-left: auto;
    font-size: 13px;
    font-weight: 600;
    color: var(--fg);
    overflow-wrap: anywhere;
    text-align: right;
  }
  .names {
    flex: 1 0 100%;
    font-size: 12px;
    line-height: 1.5;
    color: var(--fg-muted);
    overflow-wrap: anywhere;
  }
  .names.none {
    font-style: italic;
  }
  .cards {
    list-style: none;
    margin: 0;
    padding: 0 12px 12px;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(110px, 1fr));
    gap: 8px;
  }
</style>
