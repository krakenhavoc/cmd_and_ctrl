<script lang="ts">
  // StackLaneHost — the floating stack lane (#1467).
  //
  // When a player picks a stack style other than "compact", the stack
  // leaves the docked card in the attention strip and floats across
  // the middle of the table while anything is on it. It FLOATS: it is
  // absolutely positioned over the board, so the grid never reflows
  // when a spell is cast or the stack empties — a panel jumping under
  // the cursor mid-click is the one thing it must not cause.
  //
  // This component owns everything the three designs share, so a
  // design is presentation only:
  //
  //   - the model (lib/stackLane.ts), including the oracle text the
  //     effect summary may read, fetched through the /cards cache;
  //   - where the lane sits (see `measure` below);
  //   - the header: count, whose priority it is, split second, the
  //     #323 hold toggle and Pass — wired exactly as StackOverlay's;
  //   - the controls handed to the design: counter, target a stack
  //     item, and the #322 hover preview (lib/stackHover.ts);
  //   - an aria-live region that reads out the top item's line.
  //
  // The design itself is STYLE_BODIES[style]. Today all three are the
  // placeholder list; the follow-up PRs replace their entries.

  import type { Component } from "svelte";
  import { untrack } from "svelte";
  import type { CardView, GameView, StackItemView } from "../../protocol";
  import {
    buildStackLane,
    type StackLaneControls,
    type StackLaneItem,
    type StackLaneStyle,
    type StackLaneStyleProps,
  } from "../../stackLane";
  import { createStackHover } from "../../stackHover";
  import { hoveredCard } from "../../cardTypes";
  import { settings } from "../../settings";
  import { targeting, isLegalCardTarget } from "../../targeting";
  import { holdPriority, toggleHoldPriority } from "../../holdPriority";
  import { metaFor } from "../../cardMetaCache";
  import Icon from "../Icon.svelte";
  import StackLanePlaceholder from "./StackLanePlaceholder.svelte";

  interface Props {
    view: GameView;
    viewerID: string | null;
    /** The .board element the lane is positioned against. */
    boardEl: HTMLElement | null;
    style: StackLaneStyle;
    /** #1307: Board's "the holder is considering a response" read. */
    considering?: boolean;
    onCounter: (item: StackItemView) => void;
    onTargetStackItem?: (item: StackItemView) => void;
    onPass?: () => void;
  }

  const {
    view,
    viewerID,
    boardEl,
    style,
    considering = false,
    onCounter,
    onTargetStackItem,
    onPass,
  }: Props = $props();

  // A design is a component taking StackLaneStyleProps. The follow-up
  // PRs swap their entry; nothing else here changes.
  type StyleBody = Component<StackLaneStyleProps & { styleName: StackLaneStyle }>;
  const STYLE_BODIES: Record<StackLaneStyle, StyleBody> = {
    fan: StackLanePlaceholder,
    spotlight: StackLanePlaceholder,
    ribbon: StackLanePlaceholder,
  };
  const Body = $derived(STYLE_BODIES[style]);

  // ---- oracle text for the effect summary --------------------------
  // Only the stack's own spell cards, only ones the viewer may read.
  // The key is a string so the effect below re-subscribes when the
  // set of printings changes, not on every snapshot.
  let oracle = $state<Record<string, string>>({});
  const printingKey = $derived(
    [
      ...new Set(
        view.stack.cards
          .filter((c) => c.known_by_you === true && c.scryfall_id)
          .map((c) => c.scryfall_id as string),
      ),
    ]
      .sort()
      .join(","),
  );

  // #740: metaFor is called from an effect, never a derived, and the
  // subscriber only writes this component's own $state.
  $effect(() => {
    const ids = printingKey === "" ? [] : printingKey.split(",");
    const unsubs = ids.map((id) =>
      metaFor(id).subscribe((meta) => {
        const text = meta?.oracle_text;
        if (!text) return;
        untrack(() => {
          if (oracle[id] !== text) oracle = { ...oracle, [id]: text };
        });
      }),
    );
    return () => unsubs.forEach((u) => u());
  });

  function oracleTextFor(card: CardView): string | undefined {
    return card.scryfall_id ? oracle[card.scryfall_id] : undefined;
  }

  const prioritySeatID = $derived(view.seats[view.turn.priority_holder]?.id ?? null);

  const model = $derived(
    buildStackLane({
      stack: view.stack,
      stackItems: view.stack_items,
      pendingTriggers: view.pending_triggers,
      seats: view.seats,
      battlefield: view.battlefield,
      exile: view.exile,
      viewerID,
      priorityHolder: prioritySeatID,
      splitSecondActive: view.split_second_active === true,
      considering,
      oracleTextFor,
    }),
  );

  // ---- controls ------------------------------------------------------
  const hover = createStackHover(hoveredCard);

  $effect(() => {
    hover.sync(model.items.map((it) => it.id));
  });
  $effect(() => {
    return () => hover.destroy();
  });

  const controls: StackLaneControls = $derived({
    counter: (item: StackLaneItem) => onCounter(item.raw),
    pass: model.priority.viewerHolds && onPass ? onPass : undefined,
    // Only a stack item can be a target; a pending trigger is not on
    // the stack yet.
    targetable: (item: StackLaneItem) =>
      item.kind !== "pending" && $targeting !== null && isLegalCardTarget($targeting, item.id),
    target: (item: StackLaneItem) => onTargetStackItem?.(item.raw),
    hoverEnter: (item: StackLaneItem) =>
      hover.enter(item.id, item.previewCard, $settings.display.hoverDelayMs),
    hoverLeave: (item: StackLaneItem) => hover.leave(item.id),
  });

  // ---- position --------------------------------------------------------
  // The lane centres on the seam between the opponents' row and the
  // viewer's row. Every seated layout — 2, 3 or 4 players, quadrant or
  // row — puts the viewer's panel in the bottom grid row and every
  // opponent above it, so the seam is the top edge of the "self" slot
  // whatever the seat count, and measuring it means no per-layout
  // table to keep in step with Board's grid templates. With no seam
  // (a spectator's uniform grid, a table of one) it sits at the middle.
  //
  // Re-measured when the board or the self slot resizes, which covers
  // a window resize, a table-layout switch and an opponent expanding;
  // and clamped so a tall lane never hangs off the board's edge.
  let laneEl: HTMLElement | null = $state(null);
  let seamY = $state<number | null>(null);
  let boardH = $state(0);
  let laneH = $state(0);

  function measure(): void {
    if (!boardEl) return;
    const b = boardEl.getBoundingClientRect();
    boardH = b.height;
    const self = boardEl.querySelector<HTMLElement>(':scope > .slot[data-pos="self"]');
    const opponents = boardEl.querySelectorAll(':scope > .slot:not([data-pos="self"])');
    if (!self || opponents.length === 0) {
      seamY = null;
    } else {
      // Half the grid's 8px gap above the slot is the seam itself.
      seamY = self.getBoundingClientRect().top - b.top - 4;
    }
    laneH = laneEl?.offsetHeight ?? 0;
  }

  $effect(() => {
    if (!boardEl) return;
    // Re-measure on every snapshot too: a seat expanding or collapsing
    // (ADR 0077) moves the seam without resizing the board.
    void view;
    void model.items.length;
    measure();
    if (typeof ResizeObserver === "undefined") return;
    const ro = new ResizeObserver(() => measure());
    ro.observe(boardEl);
    const self = boardEl.querySelector<HTMLElement>(':scope > .slot[data-pos="self"]');
    if (self) ro.observe(self);
    if (laneEl) ro.observe(laneEl);
    return () => ro.disconnect();
  });

  // In pixels from the board's top edge, or null for the CSS fallback.
  const laneTop = $derived.by((): number | null => {
    if (boardH <= 0) return null;
    const anchor = seamY ?? boardH / 2;
    const margin = 8;
    const top = anchor - laneH / 2;
    return Math.max(margin, Math.min(top, boardH - laneH - margin));
  });

  const stackCount = $derived(model.stackItems.length);
</script>

{#if model.live}
  <div
    class="lane-host"
    class:measured={laneTop !== null}
    style:top={laneTop !== null ? `${laneTop}px` : null}
    data-stack-style={style}
  >
    <section class="stack-lane" bind:this={laneEl} aria-label={`stack: ${stackCount} on the stack`}>
      <header class="head">
        <span class="label">the stack <b class="count">{stackCount}</b></span>
        {#if model.splitSecond}
          <span class="status split-second" title="no responses allowed (split second is active)">
            <Icon name="bolt" size={11} /> split second — no responses
          </span>
        {:else if model.priority.viewerHolds}
          <span class="status mine">you hold priority</span>
        {:else if model.priority.holderName}
          <span class="status">
            {#if model.priority.considering}
              {model.priority.holderName} is considering a response…
            {:else}
              {model.priority.holderName} holds priority
            {/if}
          </span>
        {/if}
        <span class="spacer"></span>
        <!-- #323: the same session hold as the phase widget and the
             docked card. Engaged is --magenta, never gold: gold means
             priority everywhere on the table. -->
        <button
          type="button"
          class="hold-toggle"
          class:on={$holdPriority}
          aria-pressed={$holdPriority}
          onclick={toggleHoldPriority}
          title={$holdPriority
            ? "hold ON — your own spells and triggers keep the cursor so you can respond to them; click to release"
            : "hold OFF — your own spells and triggers resolve without asking. Click before you cast to keep priority and respond to them"}
        >
          {$holdPriority ? "hold ✓" : "hold"}
        </button>
        {#if controls.pass && stackCount > 0}
          <button
            type="button"
            class="act primary pass-btn"
            onclick={controls.pass}
            title="pass priority — the top item resolves once everyone passes"
          >
            Pass <Icon name="chevronRight" size={12} />
          </button>
        {/if}
      </header>
      {#if model.summary}
        <p class="summary">{model.summary}</p>
      {/if}
      <Body {model} {controls} styleName={style} />
    </section>
  </div>
{/if}
<!-- Outside the {#if}: a live region has to exist before its text
     changes for a screen reader to announce the change, so it stays
     mounted for as long as a floating style is chosen. -->
<p class="sr-only" aria-live="polite" data-stack-lane-announcer>{model.summary}</p>

<style>
  /* The host spans the lane's width and nothing else, and takes no
     pointer events itself: only the lane's own content does, so the
     table beside and under it stays clickable. z-index sits above the
     combat / target arrows (35–36) and below the attention strip (40),
     so a prompt is never covered by the stack. */
  .lane-host {
    position: absolute;
    left: 50%;
    top: 40%;
    transform: translate(-50%, -50%);
    width: min(1040px, calc(100% - 32px));
    z-index: 38;
    pointer-events: none;
    display: flex;
    justify-content: center;
  }
  .lane-host.measured {
    transform: translateX(-50%);
  }
  .stack-lane {
    pointer-events: auto;
    width: 100%;
    max-height: min(420px, 48vh);
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 10px 14px 12px;
    box-sizing: border-box;
    background: color-mix(in srgb, var(--surface) 94%, transparent);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border: 1px solid color-mix(in srgb, var(--gold) 35%, transparent);
    border-radius: var(--radius-xl);
    box-shadow:
      0 0 0 1px color-mix(in srgb, var(--gold) 18%, transparent),
      var(--shadow-lg);
    color: var(--fg);
    font-size: 12px;
  }
  .head {
    display: flex;
    align-items: center;
    gap: 10px;
    min-width: 0;
  }
  .label {
    font-family: var(--font-mono);
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.14em;
    color: var(--gold);
    font-weight: 700;
    flex: 0 0 auto;
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
  .count {
    color: var(--fg);
    font-variant-numeric: tabular-nums;
  }
  .status {
    font-size: 12px;
    color: var(--fg-dim);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    min-width: 0;
  }
  .status.mine {
    color: var(--fg);
  }
  .status :global(svg) {
    vertical-align: -2px;
    margin-right: 4px;
  }
  .status.split-second {
    color: var(--danger);
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    font-weight: 700;
  }
  .spacer {
    flex: 1 1 auto;
  }
  .hold-toggle {
    flex: 0 0 auto;
    font-family: var(--font-mono);
    font-size: 9px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--fg-dim);
    background: transparent;
    border: 1px solid var(--border);
    border-radius: 999px;
    padding: 2px 8px;
    cursor: pointer;
  }
  .hold-toggle:hover {
    color: var(--fg);
    border-color: var(--border-strong);
  }
  .hold-toggle.on {
    color: var(--magenta);
    border-color: color-mix(in srgb, var(--magenta) 55%, transparent);
    background: color-mix(in srgb, var(--magenta) 16%, transparent);
    font-weight: 700;
  }
  .act {
    flex: 0 0 auto;
    height: 30px;
    padding: 0 12px;
    font-size: 12px;
    border-radius: 7px;
    display: inline-flex;
    align-items: center;
    gap: 3px;
  }
  .summary {
    margin: 0;
    font-size: 13.5px;
    line-height: 1.35;
    color: var(--fg);
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
</style>
