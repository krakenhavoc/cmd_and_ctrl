<script lang="ts">
  // DraftSets — which sets can be drafted (ADR 0222 §4, #5437).
  //
  // A read-only list from the public GET /draft/sets. A set is ready
  // only when every card its packs can contain is fully built; every
  // other set names the cards it is waiting on. Public on purpose
  // (lib/signedInHome.ts), like the catalogue: it is card data.
  //
  // There is no draft and no pack to open here. That is #5438.
  import { onMount } from "svelte";

  import Icon from "../lib/components/Icon.svelte";
  import Page from "../lib/components/Page.svelte";
  import {
    blockedSets,
    cardsToGo,
    DRAFT_SETS_TITLE,
    fetchDraftSets,
    PACKS_ARE_APPROXIMATIONS,
    readySets,
    releaseYear,
    type DraftSet,
  } from "../lib/draftSets";

  let sets = $state<DraftSet[] | null>(null);
  let loading = $state(true);
  let error = $state("");

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
    <span>{s.cards} cards</span>
  </span>
{/snippet}

<Page
  eyebrow="Draft"
  title={DRAFT_SETS_TITLE}
  lede="A set can be drafted once every card its packs can contain is fully built, so a draft never hands you a card that doesn’t work. Drafting itself is not open yet; this page shows which sets are ready for it."
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
              <li class="set">{@render setHead(s)}</li>
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
    gap: 10px;
    font-family: var(--font-mono);
    font-size: 12px;
    font-variant-numeric: tabular-nums;
    color: var(--fg-muted);
  }
  .code {
    text-transform: uppercase;
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
