<script lang="ts">
  // DraftStatsOverlay — the full stats of Main on a screen of their own
  // (#5848 §6): colours, card types, a larger curve and then, laid out in
  // full with no dropdown, the detail: type lines, creatures, what the
  // cards do, and timing and colour. Every row lists the names of the
  // cards it counts, and a row of none is faded, so what the deck lacks
  // can be seen at a glance.
  //
  // It is a modal dialog over the page. The page behind it does not
  // scroll (so there is still one scroll bar), focus goes to Close on
  // open and back to `returnTo` on close, Tab stays on Close, and the
  // key S or Esc closes it. S also opens it, unless the viewer is typing.
  // It opens closed on every page load and is never remembered.
  //
  // The counting is lib/draftDeckStats.ts's; this draws it.
  import { tick, type Snippet } from "svelte";

  import type { DraftCard } from "../../draft";
  import { CREATURES_WANTED, deckStats, namesOf, type Role } from "../../draftDeckStats";
  import { isTypingTarget } from "../../shortcuts";
  import { L } from "../../labels";
  import Icon from "../Icon.svelte";
  import DraftColours from "./DraftColours.svelte";
  import DraftCurve from "./DraftCurve.svelte";

  interface Props {
    open: boolean;
    /** The cards in Main and in the Sideboard. */
    main: DraftCard[];
    side: DraftCard[];
    /** Asks to open or close it (a key, or its Close button). */
    onopen: () => void;
    onclose: () => void;
    /** What gets focus when it closes: the Full stats button. */
    returnTo?: HTMLElement | null;
    /** The pick clock, drawn in the header while drafting. */
    timer?: Snippet;
  }
  const { open, main, side, onopen, onclose, returnTo = null, timer }: Props = $props();

  const stats = $derived(deckStats(main));
  let closeBtn = $state<HTMLButtonElement | null>(null);

  // S toggles, Esc closes; neither while typing or with a modifier held.
  function onkeydown(ev: KeyboardEvent): void {
    if (ev.defaultPrevented || ev.ctrlKey || ev.metaKey || ev.altKey) return;
    if (ev.key === "Escape") {
      if (!open) return;
      ev.preventDefault();
      onclose();
      return;
    }
    if (ev.key !== "s" && ev.key !== "S") return;
    if (ev.repeat || isTypingTarget(ev.target as Element | null)) return;
    ev.preventDefault();
    if (open) onclose();
    else onopen();
  }

  // The only control in the dialog is Close: Tab stays on it.
  function trap(ev: KeyboardEvent): void {
    if (ev.key !== "Tab") return;
    ev.preventDefault();
    closeBtn?.focus();
  }

  // Focus goes in on open and back out on close; the page does not
  // scroll while it is open.
  let wasOpen = false;
  $effect(() => {
    const now = open;
    if (now && !wasOpen) void tick().then(() => closeBtn?.focus());
    if (!now && wasOpen) returnTo?.focus();
    wasOpen = now;
  });
  $effect(() => {
    if (!open) return;
    const root = document.documentElement;
    const was = root.style.overflow;
    root.style.overflow = "hidden";
    return () => {
      root.style.overflow = was;
    };
  });

  const TYPE_NAME = {
    creature: () => L.draftTypeCreature,
    instant: () => L.draftTypeInstant,
    sorcery: () => L.draftTypeSorcery,
    enchantment: () => L.draftTypeEnchantment,
    artifact: () => L.draftTypeArtifact,
    planeswalker: () => L.draftTypePlaneswalker,
    land: () => L.draftGroupLands,
  } as const;
  const ROLE_NAME: Record<Role, () => string> = {
    removal: () => L.draftRoleRemoval,
    fliers: () => L.draftRoleFliers,
    draw: () => L.draftRoleDraw,
    tricks: () => L.draftRoleTricks,
    counters: () => L.draftRoleCounters,
    bounce: () => L.draftRoleBounce,
    tokens: () => L.draftRoleTokens,
    ramp: () => L.draftRoleRamp,
    lifegain: () => L.draftRoleLifegain,
    discard: () => L.draftRoleDiscard,
    answers: () => L.draftRoleAnswers,
  };

  const average = (n: number | null): string => (n === null ? "–" : n.toFixed(1));
</script>

<svelte:window {onkeydown} />

{#snippet row(label: string, cards: DraftCard[], key: string)}
  <li class:zero={cards.length === 0} data-row={key}>
    <span class="l">{label}</span>
    <b>{cards.length}</b>
    {#if cards.length > 0}<span class="names">{namesOf(cards).join(", ")}</span>{/if}
  </li>
{/snippet}

{#if open}
  <div
    class="overlay"
    role="dialog"
    aria-modal="true"
    aria-labelledby="draft-stats-title"
    tabindex="-1"
    data-testid="draft-stats-overlay"
    onkeydown={trap}
  >
    <div class="in">
      <header class="head">
        <h2 id="draft-stats-title">{L.deckStats}</h2>
        <span class="count" data-testid="draft-stats-counts"
          >{L.draftStatsMainSide(main.length, side.length)}</span
        >
        {#if timer}<span class="clock">{@render timer()}</span>{/if}
        <button
          type="button"
          class="btn close"
          bind:this={closeBtn}
          aria-keyshortcuts="S Escape"
          data-testid="draft-stats-close"
          onclick={onclose}
        >
          <Icon name="x" size={14} />
          {L.draftStatsClose}
        </button>
      </header>

      <div class="body">
        <div class="panel grid">
          <section aria-labelledby="ds-colours">
            <h3 id="ds-colours">
              {L.draftStatColours} <span class="sub">{L.draftStatsPipsNote}</span>
            </h3>
            <DraftColours {main} />
          </section>

          <section aria-labelledby="ds-types">
            <h3 id="ds-types">{L.draftStatTypes}</h3>
            <ul class="srows">
              {#each stats.types.rows as t (t.type)}
                {@render row(TYPE_NAME[t.type](), t.cards, t.type)}
              {/each}
              {@render row(L.draftStatsOtherSpells, stats.types.otherSpells, "other")}
            </ul>
            {#if stats.types.creatures < CREATURES_WANTED}
              <p class="hint" data-testid="draft-creature-note">
                {L.draftStatsCreatureNote(stats.types.creatures)}
              </p>
            {/if}
          </section>

          <section aria-labelledby="ds-curve">
            <h3 id="ds-curve">{L.draftStatCurve}</h3>
            <DraftCurve {main} large />
          </section>
        </div>

        <h3 class="sub-h">{L.draftStatsMore}</h3>
        {#if main.length === 0}
          <p class="hint">{L.draftStatsNoCards}</p>
        {:else}
          <div class="panel grid more" data-testid="draft-more-detail">
            <section aria-labelledby="ds-lines">
              <h3 id="ds-lines">{L.draftStatsTypeLines}</h3>
              <ul class="srows">
                {#each stats.typeLines as r (r.key)}
                  {@render row(r.key, r.cards, `line-${r.key}`)}
                {/each}
              </ul>
            </section>

            <section aria-labelledby="ds-creatures">
              <h3 id="ds-creatures">
                {L.draftTypeCreature}
                {#if stats.creatures.avgPower !== null}
                  <span class="sub"
                    >{L.draftStatsAverageSize(
                      average(stats.creatures.avgPower),
                      average(stats.creatures.avgToughness),
                    )}</span
                  >
                {/if}
              </h3>
              <ul class="srows">
                {@render row(L.draftStatsSmall, stats.creatures.small, "small")}
                {@render row(L.draftStatsMedium, stats.creatures.medium, "medium")}
                {@render row(L.draftStatsBig, stats.creatures.big, "big")}
                {@render row(L.draftStatsEvasive, stats.creatures.evasive, "evasive")}
              </ul>
              {#if stats.creatureTypes.length > 0}
                <h3 class="again">{L.draftStatsCreatureTypes}</h3>
                <ul class="srows">
                  {#each stats.creatureTypes as r (r.key)}
                    {@render row(r.key, r.cards, `kind-${r.key}`)}
                  {/each}
                </ul>
              {/if}
            </section>

            <section aria-labelledby="ds-do">
              <h3 id="ds-do">{L.draftStatsDo}</h3>
              <ul class="srows">
                {#each stats.roles as r (r.role)}
                  {@render row(ROLE_NAME[r.role](), r.cards, r.role)}
                {/each}
              </ul>
            </section>

            <section aria-labelledby="ds-timing">
              <h3 id="ds-timing">{L.draftStatsTiming}</h3>
              <ul class="srows">
                {@render row(L.draftStatsInstant, stats.instant, "instant-speed")}
                {@render row(L.draftStatsSorcery, stats.sorcery, "sorcery-speed")}
                {@render row(L.draftStatsTwoColours, stats.multicolour, "multicolour")}
                {@render row(L.draftGroupColourless, stats.colourless, "colourless")}
              </ul>
            </section>
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  .overlay {
    position: fixed;
    inset: 0;
    z-index: 80;
    overflow-y: auto;
    background: color-mix(in srgb, var(--bg) 94%, transparent);
    -webkit-backdrop-filter: blur(8px);
    backdrop-filter: blur(8px);
    padding-top: env(safe-area-inset-top, 0px);
    padding-bottom: 64px;
    padding-left: max(16px, env(safe-area-inset-left, 0px));
    padding-right: max(16px, env(safe-area-inset-right, 0px));
  }
  .in {
    max-width: 1240px;
    margin: 0 auto;
  }
  .head {
    position: sticky;
    top: 0;
    z-index: 1;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px 16px;
    padding: 14px 0;
    margin-bottom: 8px;
    background: var(--bg);
    border-bottom: 1px solid var(--border);
  }
  h2 {
    margin: 0;
    font-family: var(--font-display);
    font-size: 22px;
    font-weight: 800;
    letter-spacing: -0.005em;
    text-transform: none;
    color: var(--fg);
  }
  .count {
    margin-right: auto;
    font-size: 13px;
    color: var(--fg-muted);
  }
  .close {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    border-radius: 999px;
  }
  .body {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .panel {
    padding: 16px;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--surface);
  }
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 22px 32px;
  }
  .grid > section {
    min-width: 0;
  }
  .more {
    grid-template-columns: repeat(auto-fit, minmax(230px, 1fr));
  }
  h3 {
    margin: 0 0 8px;
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 2px 8px;
    font-family: var(--font-display);
    font-size: 14px;
    font-weight: 800;
    letter-spacing: 0;
    text-transform: none;
    color: var(--fg);
  }
  h3.again {
    margin-top: 14px;
  }
  .sub-h {
    margin: 18px 0 4px;
    font-size: 17px;
  }
  .sub {
    font-family: var(--font-body);
    font-size: 11.5px;
    font-weight: 400;
    color: var(--fg-muted);
  }
  .srows {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 3px;
    font-size: 13.5px;
  }
  .srows li {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 0 8px;
  }
  .l {
    flex: 1;
  }
  b {
    font-family: var(--font-mono);
    font-variant-numeric: tabular-nums;
    font-weight: 600;
  }
  .names {
    width: 100%;
    font-size: 11px;
    line-height: 1.35;
    color: var(--fg-muted);
  }
  /* A row of none is an answer, so it stays, fainter. */
  .zero {
    opacity: 0.45;
  }
  .hint {
    margin: 6px 0 0;
    font-size: 12px;
    color: var(--fg-muted);
  }
</style>
