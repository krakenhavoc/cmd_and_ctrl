<script lang="ts">
  // DraftSets — which sets can be drafted (ADR 0222 §4, #5437).
  //
  // A read-only list from the public GET /draft/sets. A set is ready
  // only when every card its packs can contain is fully built; every
  // other set names the cards it is waiting on. Public on purpose
  // (lib/signedInHome.ts), like the catalogue: it is card data.
  //
  // A ready set links to the create form with that set chosen (#5438).
  // No pack is opened here.
  import { onMount } from "svelte";

  import Icon from "../lib/components/Icon.svelte";
  import Page from "../lib/components/Page.svelte";
  import DraftArchetypes from "../lib/components/draft/DraftArchetypes.svelte";
  import {
    blockedSets,
    boosterName,
    cardsToGo,
    DRAFT_SETS_TITLE,
    fetchDraftSets,
    PACKS_ARE_APPROXIMATIONS,
    readySets,
    releaseYear,
    type DraftSet,
  } from "../lib/draftSets";
  import { draftNewHash } from "../lib/draft";
  import { L } from "../lib/labels";

  let sets = $state<DraftSet[] | null>(null);
  let loading = $state(true);
  let error = $state("");
  // The sets whose colour pairs the viewer has opened; each is fetched
  // when first opened, not for the whole list.
  let pairsOpen = $state<Record<string, boolean>>({});

  const ready = $derived(readySets(sets ?? []));
  const blocked = $derived(blockedSets(sets ?? []));

  onMount(() => {
    const ctrl = new AbortController();
    fetchDraftSets(ctrl.signal)
      .then((r) => {
        sets = r.sets;
      })
      .catch((err: unknown) => {
        if (ctrl.signal.aborted) return;
        error = err instanceof Error ? err.message : "couldn’t load the sets";
      })
      .finally(() => {
        loading = false;
      });
    return () => ctrl.abort();
  });
</script>

{#snippet setHead(s: DraftSet)}
  <span class="set-name">{s.name}</span>
  <span class="set-meta">
    <span class="code">{s.code}</span>
    {#if releaseYear(s.released)}<span>{releaseYear(s.released)}</span>{/if}
    {#if boosterName(s.booster)}<span class="booster">{boosterName(s.booster)}</span>{/if}
    <span>{s.cards} cards</span>
  </span>
{/snippet}

<Page
  eyebrow="Draft"
  title={DRAFT_SETS_TITLE}
  lede="A set can be drafted once every card its packs can contain is fully built, so a draft never hands you a card that doesn’t work. This page shows which sets are ready, and which cards the others are waiting on."
  width="prose"
>
  <div class="stack">
    <p class="approx" data-testid="packs-approximate">{PACKS_ARE_APPROXIMATIONS}</p>

    {#if !loading && error}
      <p class="notice err" role="alert">
        <Icon name="x" size={14} />
        <span>Couldn’t load the sets ({error}). Reload the page to try again.</span>
      </p>
    {/if}

    {#if loading}
      <p class="help" aria-live="polite">Loading the sets…</p>
    {/if}

    {#if sets}
      <section aria-labelledby="ready-title">
        <h2 id="ready-title">Ready to draft <span class="n">{ready.length}</span></h2>
        {#if ready.length === 0}
          <p class="help">
            No set is ready yet. The closest ones are listed below with the cards they are waiting
            on.
          </p>
        {:else}
          <ul class="sets" data-testid="ready-sets">
            {#each ready as s (s.code)}
              <li class="set">
                {@render setHead(s)}
                <a class="link-arrow go" href={draftNewHash(s.code)}
                  >{L.draftThisSet}<span class="sr-only">: {s.name}</span> →</a
                >
                <details
                  class="pairs"
                  data-testid="pairs-{s.code}"
                  ontoggle={(e) => (pairsOpen[s.code] = e.currentTarget.open)}
                >
                  <summary>{L.draftArchetypesTitle}<span class="sr-only">: {s.name}</span></summary>
                  {#if pairsOpen[s.code]}
                    <div class="pairs-body"><DraftArchetypes set={s.code} /></div>
                  {/if}
                </details>
              </li>
            {/each}
          </ul>
        {/if}
      </section>

      {#if blocked.length > 0}
        <section aria-labelledby="blocked-title">
          <h2 id="blocked-title">Not ready yet <span class="n">{blocked.length}</span></h2>
          <p class="help">
            Closest first. Open a set to see the cards that are not fully built. They all still play
            at a Commander table, where you can resolve a card by hand.
          </p>
          <ul class="sets" data-testid="blocked-sets">
            {#each blocked as s (s.code)}
              <li class="set">
                <details>
                  <summary>
                    {@render setHead(s)}
                    <span class="togo">{cardsToGo(s.blockers.length)}</span>
                  </summary>
                  <ul class="blockers">
                    {#each s.blockers as name (name)}
                      <li>{name}</li>
                    {/each}
                  </ul>
                </details>
              </li>
            {/each}
          </ul>
        </section>
      {/if}
    {/if}
  </div>
</Page>

<style>
  .stack {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 24px;
  }
  h2 {
    margin: 0 0 8px;
    font-family: var(--font-display);
    font-size: 19px;
    font-weight: 800;
    letter-spacing: -0.005em;
    text-transform: none;
    color: var(--fg);
  }
  h2 .n {
    margin-left: 6px;
    font-family: var(--font-mono);
    font-size: 13px;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    color: var(--fg-muted);
  }
  .approx {
    margin: 0;
    padding: 12px 14px;
    border-radius: var(--radius);
    background: var(--surface);
    border: 1px solid var(--border-strong);
    color: var(--fg);
    font-size: 14px;
    line-height: 1.6;
  }
  .help {
    margin: 0 0 8px;
    color: var(--fg-muted);
    font-size: 13px;
    line-height: 1.6;
  }
  .notice {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 0;
    padding: 10px 12px;
    border-radius: var(--radius);
    background: var(--surface);
    border: 1px solid var(--border-strong);
    font-size: 13px;
  }
  .notice.err {
    color: var(--danger);
  }
  .sets {
    list-style: none;
    margin: 0;
    padding: 0;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--surface);
  }
  .set + .set {
    border-top: 1px solid var(--border);
  }
  .set,
  summary {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 4px 12px;
    min-width: 0;
  }
  .set {
    padding: 10px 14px;
  }
  .set:has(details) {
    display: block;
    padding: 0;
  }
  summary {
    padding: 10px 14px;
    cursor: pointer;
    border-radius: var(--radius);
  }
  /* The flex summary loses the browser's marker; this is its stand-in. */
  summary {
    list-style: none;
  }
  summary::-webkit-details-marker {
    display: none;
  }
  summary::before {
    content: "";
    align-self: center;
    width: 0;
    height: 0;
    border-left: 5px solid var(--fg-muted);
    border-top: 4px solid transparent;
    border-bottom: 4px solid transparent;
  }
  details[open] > summary::before {
    transform: rotate(90deg);
  }
  summary:focus-visible {
    outline: 2px solid var(--accent-strong);
    outline-offset: -2px;
  }
  .set-name {
    font-size: 14px;
    font-weight: 700;
    color: var(--fg);
    overflow-wrap: anywhere;
  }
  .set-meta {
    display: inline-flex;
    flex-wrap: wrap;
    gap: 2px 10px;
    font-family: var(--font-mono);
    font-size: 12px;
    font-variant-numeric: tabular-nums;
    color: var(--fg-muted);
  }
  .code {
    text-transform: uppercase;
  }
  .go {
    margin-left: auto;
    font-size: 13px;
  }
  /* The colour pairs of a ready set: a plain disclosure under its row. */
  .pairs {
    flex: 1 0 100%;
    min-width: 0;
  }
  .pairs summary {
    display: inline-flex;
    align-items: center;
    padding: 2px 0;
    font-size: 13px;
    font-weight: 600;
    color: var(--fg-muted);
    cursor: pointer;
  }
  .pairs summary:focus-visible {
    outline: 2px solid var(--accent-strong);
    outline-offset: 2px;
  }
  .pairs-body {
    margin-top: 8px;
  }
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }
  .togo {
    margin-left: auto;
    font-family: var(--font-mono);
    font-size: 12px;
    font-variant-numeric: tabular-nums;
    color: var(--fg-muted);
  }
  .blockers {
    margin: 0;
    padding: 0 14px 12px 34px;
    columns: 2;
    column-gap: 24px;
    color: var(--fg-muted);
    font-size: 13px;
    line-height: 1.6;
  }
  .blockers li {
    break-inside: avoid;
    overflow-wrap: anywhere;
  }
  @media (max-width: 819px) {
    .blockers {
      columns: 1;
    }
  }
</style>
