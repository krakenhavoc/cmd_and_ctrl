<script lang="ts">
  // MyDrafts — a signed-in person's finished drafts (ADR 0222 §9, #5440).
  //
  // #/drafts is the list; #/drafts/<id> is one draft, read-only: the
  // deck that was built, the rest of the pool, and each round. Both
  // read GET /me/drafts, which answers for the caller's own records
  // only. A guest has none: a guest's draft is kept only while it runs.
  import { onMount } from "svelte";

  import DraftPool from "../lib/components/draft/DraftPool.svelte";
  import Page from "../lib/components/Page.svelte";
  import { DRAFT_NEW_HASH, pickModeName } from "../lib/draft";
  import { L } from "../lib/labels";
  import {
    basicsText,
    draftDate,
    draftRecord,
    fetchMyDraft,
    fetchMyDrafts,
    MY_DRAFTS_HASH,
    myDraftHash,
    roundText,
    splitPicks,
    type MyDraft,
  } from "../lib/myDrafts";

  interface Props {
    /** One draft's id; absent for the list. */
    id?: string;
  }
  const { id }: Props = $props();

  let drafts = $state<MyDraft[] | null>(null);
  let draft = $state<MyDraft | null>(null);
  let loading = $state(true);
  let failed = $state("");

  onMount(() => {
    const ctrl = new AbortController();
    const load = id
      ? fetchMyDraft(id, ctrl.signal).then((d) => (draft = d))
      : fetchMyDrafts(ctrl.signal).then((d) => (drafts = d));
    load
      .catch(() => {
        if (ctrl.signal.aborted) return;
        failed = id
          ? "That draft isn’t here. It may belong to another account, or have been deleted with one."
          : "Couldn’t load your drafts. Try again in a moment.";
      })
      .finally(() => (loading = false));
    return () => ctrl.abort();
  });

  const split = $derived(draft ? splitPicks(draft) : { deck: [], side: [] });
  const basics = $derived(basicsText(draft?.basics));
  const basicsTotal = $derived(Object.values(draft?.basics ?? {}).reduce((n, k) => n + k, 0));
</script>

{#snippet meta()}
  {#if draft}
    <div class="chips">
      <span class="chip">{draft.set_name || draft.set.toUpperCase()}</span>
      <span class="chip">{draftDate(draft.drafted_at)}</span>
      <span class="chip">{pickModeName(draft.mode)}</span>
      <span class="chip">{draft.seats} seats</span>
      <span class="chip">{draftRecord(draft)}</span>
    </div>
  {/if}
{/snippet}

{#if id}
  <Page eyebrow={L.myDrafts} title={draft?.name ?? "A draft"} {meta} width="wide">
    <div class="stack">
      <a class="link-arrow back" href={MY_DRAFTS_HASH}>← {L.myDrafts}</a>
      {#if loading}
        <p class="help" role="status">Loading the draft…</p>
      {:else if failed || !draft}
        <div class="card"><p class="help" role="alert">{failed}</p></div>
      {:else}
        <section class="card" aria-labelledby="md-rounds-h" data-testid="my-draft-rounds">
          <h2 id="md-rounds-h">Rounds</h2>
          {#if draft.matches.length === 0}
            <p class="help">You left this draft before it played a round.</p>
          {:else}
            <ol class="rounds">
              {#each draft.matches as m (m.round)}
                <li><span class="r">Round {m.round}</span> {roundText(m)}</li>
              {/each}
            </ol>
          {/if}
        </section>

        <section aria-labelledby="md-deck-h" data-testid="my-draft-deck">
          <h2 id="md-deck-h">
            Deck <span class="n">{split.deck.length + basicsTotal}</span>
          </h2>
          {#if split.deck.length === 0 && basicsTotal === 0}
            <p class="help">No deck was built in this draft.</p>
          {:else}
            {#if basics}<p class="help lands">Basic lands added: {basics}.</p>{/if}
            <DraftPool cards={split.deck} images />
          {/if}
        </section>

        <section aria-labelledby="md-side-h" data-testid="my-draft-side">
          <h2 id="md-side-h">
            Sideboard <span class="n">{split.side.length}</span>
          </h2>
          <DraftPool cards={split.side} images empty="Every card you drafted was in the deck." />
        </section>
      {/if}
    </div>
  </Page>
{:else}
  <Page
    eyebrow="Draft"
    title={L.myDrafts}
    lede="The drafts you have finished while signed in: what you picked, what you built, and how it went. Only you can see them."
    width="wide"
  >
    <div class="stack">
      {#if loading}
        <p class="help" role="status">Loading your drafts…</p>
      {:else if failed}
        <div class="card"><p class="help" role="alert">{failed}</p></div>
      {:else if !drafts || drafts.length === 0}
        <div class="card empty">
          <p class="help strong">No drafts yet.</p>
          <p class="help">
            A draft is kept here once its picks are done. Drafting without signing in keeps nothing.
          </p>
          <a class="link-arrow" href={DRAFT_NEW_HASH}>{L.draftNewTitle} →</a>
        </div>
      {:else}
        <ul class="list" data-testid="my-drafts">
          {#each drafts as d (d.id)}
            <li class="card">
              <a class="title" href={myDraftHash(d.id)}>{d.name}</a>
              <span class="what">
                {d.set_name || d.set.toUpperCase()} · {draftDate(d.drafted_at)} · {pickModeName(
                  d.mode,
                )} · {d.seats} seats
              </span>
              <span class="rec">
                {#if d.matches.length > 0}{draftRecord(d)}{:else}no rounds{/if}
              </span>
            </li>
          {/each}
        </ul>
        <a class="link-arrow" href={DRAFT_NEW_HASH}>{L.draftNewTitle} →</a>
      {/if}
    </div>
  </Page>
{/if}

<style>
  .stack {
    display: flex;
    flex-direction: column;
    gap: 22px;
    min-width: 0;
  }
  .back {
    align-self: flex-start;
  }
  h2 {
    margin: 0 0 12px;
    font-family: var(--font-display);
    font-size: 19px;
    font-weight: 800;
    color: var(--fg);
  }
  .n {
    margin-left: 6px;
    font-family: var(--font-mono);
    font-size: 13px;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    color: var(--fg-muted);
  }
  .help {
    margin: 0;
    font-size: 0.875rem;
    line-height: 1.6;
    color: var(--fg-muted);
  }
  .help.strong {
    font-weight: 700;
    color: var(--fg);
  }
  .help.lands {
    margin-bottom: 14px;
  }
  .empty {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
  .chips {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px 8px;
    margin-top: 12px;
  }
  .chip {
    display: inline-flex;
    align-items: center;
    padding: 3px 10px;
    border-radius: 999px;
    border: 1px solid var(--border-strong);
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 600;
    letter-spacing: 0.04em;
    color: var(--fg-muted);
  }
  .list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }
  .list li {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 2px 16px;
    align-items: baseline;
    padding: 14px 18px;
  }
  .title {
    font-family: var(--font-display);
    font-size: 17px;
    font-weight: 700;
    color: var(--fg);
    text-decoration: none;
    overflow-wrap: anywhere;
  }
  .title:hover {
    text-decoration: underline;
  }
  .what {
    grid-column: 1;
    font-size: 0.8125rem;
    color: var(--fg-muted);
  }
  .rec {
    grid-column: 2;
    grid-row: 1 / span 2;
    align-self: center;
    font-family: var(--font-mono);
    font-size: 15px;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
    color: var(--fg);
  }
  .rounds {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .rounds li {
    padding: 6px 0;
    font-size: 0.875rem;
    color: var(--fg);
  }
  .rounds li + li {
    border-top: 1px solid var(--border);
  }
  .r {
    display: inline-block;
    min-width: 9ch;
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--fg-muted);
  }
</style>
