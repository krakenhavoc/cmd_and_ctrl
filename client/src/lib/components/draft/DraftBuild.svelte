<script lang="ts">
  // DraftBuild — the deckbuilding step of a draft (ADR 0222 §5, #5440).
  //
  // The viewer's pool in two zones, the deck and the sideboard. A card
  // moves with a click or a drag; basic lands come from a stepper for
  // each land type. Beside them: the count, the curve, a land-count
  // suggestion and the one primary, Submit deck.
  //
  // Everything here is the viewer's own: `view.pool` and `view.deck`
  // hold no other seat's cards (CR 905.1c). The server runs the limited
  // deck check (CR 100.2b) on what is submitted; lib/draftBuild.ts says
  // the same thing first.
  import type { Snippet } from "svelte";

  import { cardCount, DRAG_TYPE, type DraftCard, type DraftView } from "../../draft";
  import {
    applySuggestion,
    BASIC_COLOR,
    BASIC_LANDS,
    basicsCount,
    buildProblem,
    curve,
    deckSize,
    initialBuild,
    landCount,
    landSuggestion,
    matchesSuggestion,
    MAX_BASICS,
    moveCard,
    sameDeck,
    setBasic,
    splitPool,
    toggleCard,
    type BasicLand,
  } from "../../draftBuild";
  import { L } from "../../labels";
  import ManaSymbol from "../board/ManaSymbol.svelte";
  import Icon from "../Icon.svelte";
  import DraftPool from "./DraftPool.svelte";

  interface Props {
    view: DraftView;
    /** The socket is open: a deck can be sent. */
    live: boolean;
    /** A deck is on its way to the server. */
    pending: boolean;
    /** The server's sentence for a refused deck. */
    error: string;
    onsubmit: (main: string[], basics: Record<string, number>) => void;
    /** The build timer, drawn by the room. */
    timer: Snippet;
  }
  const { view, live, pending, error, onsubmit, timer }: Props = $props();

  const pool = $derived<DraftCard[]>(view.pool ?? view.picks ?? []);
  const deckMin = $derived(view.build?.deck_min ?? 40);

  // The deck being built is this page's own until it is submitted. It
  // opens on the deck the server already holds for the seat, if any.
  // svelte-ignore state_referenced_locally
  let build = $state(initialBuild(view.pool ?? view.picks ?? [], view.deck));

  const zones = $derived(splitPool(pool, build));
  const size = $derived(deckSize(build));
  const lands = $derived(landCount(zones.deck, build));
  const columns = $derived(curve(zones.deck));
  const tallest = $derived(Math.max(1, ...columns.map((c) => c.count)));
  const suggestion = $derived(landSuggestion(zones.deck, deckMin));
  const problem = $derived(buildProblem(size, deckMin));
  const submitted = $derived(view.build?.submitted === true);
  const unchanged = $derived(submitted && sameDeck(build, view.deck));
  const waiting = $derived(view.players.filter((p) => !p.bot && !p.built && p.seat !== view.you));

  let dropOn = $state<"deck" | "side" | null>(null);

  function drop(ev: DragEvent, to: "deck" | "side"): void {
    ev.preventDefault();
    dropOn = null;
    const id = ev.dataTransfer?.getData(DRAG_TYPE);
    if (id) build = moveCard(build, id, to);
  }
  function over(ev: DragEvent, to: "deck" | "side"): void {
    if (!ev.dataTransfer?.types.includes(DRAG_TYPE)) return;
    ev.preventDefault();
    dropOn = to;
  }
  function step(name: BasicLand, by: number): void {
    build = setBasic(build, name, (build.basics[name] ?? 0) + by);
  }
  function submit(): void {
    if (problem || pending || !live) return;
    onsubmit([...build.main], { ...build.basics });
  }

  const curveSummary = $derived(
    columns.map((c) => `${c.count} at ${c.value === 6 ? "6 or more" : c.value}`).join(", "),
  );
</script>

<div class="cols">
  <div class="zones">
    <section
      class="card zone"
      class:drop={dropOn === "deck"}
      aria-labelledby="draft-deck-h"
      data-testid="draft-deck"
      ondragover={(ev) => over(ev, "deck")}
      ondragleave={() => (dropOn = null)}
      ondrop={(ev) => drop(ev, "deck")}
    >
      <h2 id="draft-deck-h">
        {L.draftDeckZone} <span class="n">{zones.deck.length}</span>
      </h2>
      <DraftPool
        cards={zones.deck}
        images
        verb="Move to sideboard"
        empty="No cards yet. Click a card in your sideboard, or drag it here, to play it."
        onpick={(card) => (build = toggleCard(build, card.id))}
      />
    </section>

    <section
      class="card zone"
      class:drop={dropOn === "side"}
      aria-labelledby="draft-side-h"
      data-testid="draft-side"
      ondragover={(ev) => over(ev, "side")}
      ondragleave={() => (dropOn = null)}
      ondrop={(ev) => drop(ev, "side")}
    >
      <h2 id="draft-side-h">
        {L.draftSideboardZone} <span class="n">{zones.side.length}</span>
      </h2>
      <DraftPool
        cards={zones.side}
        images
        verb="Add to deck"
        empty="Every card you drafted is in your deck."
        onpick={(card) => (build = toggleCard(build, card.id))}
      />
    </section>
  </div>

  <section class="card panel" aria-labelledby="draft-build-h" data-testid="draft-build">
    <div class="head">
      <h2 id="draft-build-h">{L.draftBuildDeck}</h2>
      {@render timer()}
    </div>
    <p class="help">
      Play at least {deckMin} cards: the ones you drafted, plus as many basic lands as you like. The rest
      of what you drafted is your sideboard. When the timer ends, a deck is built for anyone who has not
      submitted one.
    </p>

    <dl class="counts" data-testid="draft-deck-count">
      <div>
        <dt>Deck</dt>
        <dd class:short={problem !== ""}>{size}<span class="of">/{deckMin}</span></dd>
      </div>
      <div>
        <dt>Lands</dt>
        <dd>{lands}</dd>
      </div>
      <div>
        <dt>Other cards</dt>
        <dd>{size - lands}</dd>
      </div>
    </dl>

    <div class="curve" role="img" aria-label={`${L.draftCurve}: ${curveSummary}`}>
      {#each columns as col (col.value)}
        <div class="col" aria-hidden="true">
          <span class="count">{col.count}</span>
          <span class="bar" style:height={`${Math.round((col.count / tallest) * 100)}%`}></span>
          <span class="mv">{col.label}</span>
        </div>
      {/each}
    </div>

    <div class="lands">
      <h3>{L.draftBasicLands} <span class="n">{basicsCount(build)}</span></h3>
      <ul>
        {#each BASIC_LANDS as name (name)}
          {@const n = build.basics[name] ?? 0}
          <li role="group" aria-label={L.draftBasicStepper(name)}>
            <span class="land">
              <ManaSymbol symbol={BASIC_COLOR[name]} size={16} />
              {name}
            </span>
            <span class="stepper">
              <button
                type="button"
                class="btn ghost sm"
                aria-label={`one fewer ${name}`}
                disabled={n === 0}
                onclick={() => step(name, -1)}>−</button
              >
              <output class="num" aria-live="off">{n}</output>
              <button
                type="button"
                class="btn ghost sm"
                aria-label={`one more ${name}`}
                disabled={n >= MAX_BASICS}
                onclick={() => step(name, 1)}>+</button
              >
            </span>
          </li>
        {/each}
      </ul>
      {#if suggestion.lands > 0}
        <p class="help suggest" data-testid="draft-land-suggestion">
          Suggested: {suggestion.lands} lands for {cardCount(
            zones.deck.length - (lands - basicsCount(build)),
          )}
          that are not lands{#if Object.keys(suggestion.basics).length > 0}{": "}{BASIC_LANDS.filter(
              (b) => suggestion.basics[b],
            )
              .map((b) => `${suggestion.basics[b]} ${b}`)
              .join(", ")}{/if}.
        </p>
        {#if suggestion.add > 0 && Object.keys(suggestion.basics).length > 0 && !matchesSuggestion(build, suggestion)}
          <button
            type="button"
            class="btn ghost sm"
            onclick={() => (build = applySuggestion(build, suggestion))}
          >
            {L.draftUseSuggestion}
          </button>
        {/if}
      {:else}
        <p class="help suggest">A 40-card deck usually plays about 17 lands.</p>
      {/if}
    </div>

    {#if error}
      <p class="notice err" role="alert">
        <Icon name="x" size={14} />
        <span>{error}</span>
      </p>
    {/if}

    <div class="submit">
      <button
        type="button"
        class="btn primary"
        disabled={problem !== "" || pending || !live || unchanged}
        onclick={submit}
      >
        {pending ? "Submitting…" : L.draftSubmitDeck}
      </button>
      <p class="help" role="status" data-testid="draft-build-status">
        {#if problem}
          {problem}
        {:else if !live}
          Reconnecting… Your deck is kept on this page.
        {:else if unchanged}
          Deck submitted.
          {#if waiting.length > 0}
            Waiting for {waiting.map((p) => p.name).join(", ")}. You can still change your deck.
          {/if}
        {:else if submitted}
          You changed your deck since you submitted it. Submit again to play this one.
        {:else}
          {cardCount(size)}. Ready to submit.
        {/if}
      </p>
    </div>
  </section>
</div>

<style>
  .cols {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(280px, 360px);
    gap: 22px;
    align-items: start;
  }
  @media (max-width: 819px) {
    .cols {
      grid-template-columns: minmax(0, 1fr);
    }
    .panel {
      order: -1;
    }
  }
  .zones {
    display: flex;
    flex-direction: column;
    gap: 22px;
    min-width: 0;
  }
  /* A zone a card is being dragged over: the hairline thickens to the
     accent, and nothing moves. */
  .zone.drop {
    border-color: var(--accent);
    box-shadow: 0 0 0 1px var(--accent);
  }
  .panel {
    position: sticky;
    top: 16px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  .head {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 8px 12px;
  }
  h2 {
    margin: 0 0 12px;
    font-family: var(--font-display);
    font-size: 19px;
    font-weight: 800;
    color: var(--fg);
  }
  .head h2 {
    margin: 0;
  }
  h3 {
    margin: 0 0 8px;
    font-family: var(--font-display);
    font-size: 15px;
    font-weight: 700;
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
    font-size: 0.8125rem;
    line-height: 1.6;
    color: var(--fg-muted);
  }
  .counts {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 8px;
    margin: 0;
  }
  .counts div {
    padding: 8px 10px;
    border-radius: var(--radius);
    background: var(--surface-sunken);
    border: 1px solid var(--border);
  }
  dt {
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--fg-muted);
  }
  dd {
    margin: 2px 0 0;
    font-family: var(--font-mono);
    font-size: 20px;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
    color: var(--fg);
  }
  /* Below the minimum: the count is underlined as well as tinted, and
     the line by the button says how many more. */
  dd.short {
    color: var(--danger);
    text-decoration: underline;
    text-underline-offset: 3px;
  }
  .of {
    font-size: 12px;
    color: var(--fg-muted);
  }
  .curve {
    display: grid;
    grid-template-columns: repeat(7, minmax(0, 1fr));
    gap: 6px;
    height: 96px;
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
  .lands ul {
    list-style: none;
    margin: 0 0 10px;
    padding: 0;
  }
  .lands li {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    padding: 4px 0;
  }
  .lands li + li {
    border-top: 1px solid var(--border);
  }
  .land {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-size: 0.875rem;
    color: var(--fg);
  }
  .stepper {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
  .stepper .btn {
    min-width: 32px;
    padding: 0;
    justify-content: center;
    font-family: var(--font-mono);
    font-size: 16px;
  }
  .num {
    min-width: 3ch;
    text-align: center;
    font-family: var(--font-mono);
    font-size: 14px;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
    color: var(--fg);
  }
  .suggest {
    margin-bottom: 8px;
  }
  .submit {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
    padding-top: 14px;
    border-top: 1px solid var(--border);
  }
  .notice {
    display: flex;
    align-items: center;
    gap: 10px;
    margin: 0;
    padding: 10px 14px;
    border-radius: var(--radius);
    font-size: 0.8125rem;
    line-height: 1.5;
    color: var(--fg);
  }
  .notice.err {
    background: color-mix(in srgb, var(--danger) 8%, transparent);
    border: 1px solid color-mix(in srgb, var(--danger) 35%, transparent);
  }
</style>
