<script lang="ts">
  // StackOverlay is the stack card of the board's attention strip
  // (Board.svelte's .strip column). Each item is one compact row —
  // 40×56 thumb, title, caster → targets — so a deep stack still
  // reads top-down without covering the table. The top item (next
  // to resolve) carries a gold ring; when the viewer holds priority
  // it also carries Counter + Pass, the one deliberate duplication
  // of the phase widget's "next".
  //
  // Items stack top-to-bottom with the top of the visual stack
  // (most-recently-cast, resolves first) at the top of the overlay.
  // The container scrolls if a deep stack overflows the strip.
  //
  // #322 — hover preview. The rows are compact by design, which made
  // the stack the one zone you could not actually read: it renders
  // its own <img> thumbs rather than mounting Card.svelte, and
  // Card.svelte was the only writer of the shared `hoveredCard`
  // store, so HoverZoomOverlay never heard about a stack row. Fixed
  // by writing the store from here on pointerenter, honouring the
  // same settings.display.hoverDelayMs dwell as the table. The
  // overlay pins top-right and the strip pins top-left, so the big
  // preview lands beside the stack rather than over it — which is
  // why this is the fix instead of permanently growing the rows and
  // paying board space every turn for a read that only happens while
  // the stack is live.
  //
  // For an ability item the previewed card is the source permanent
  // (abilities have no card of their own on the stack) — the same
  // card artCardFor() already draws the thumb from.

  // #1467: every derivation this card used to make for itself — the
  // title, the art card, the caster, the "→ target" line, the chips,
  // the order — now comes from lib/stackLane.ts, the model the
  // floating lane styles read too, so the docked card and the lane
  // cannot say different things about the same stack. The #322 hover
  // bookkeeping is lib/stackHover.ts for the same reason.

  import type { PlayerView, StackItemView, ZoneView } from "../../protocol";
  import { cardImageURL } from "../../cardImage";
  import { cardArt } from "../../cardArt";
  import Icon from "../Icon.svelte";
  import { targeting, isLegalCardTarget } from "../../targeting";
  import { hoveredCard } from "../../cardTypes";
  import { settings } from "../../settings";
  import { holdPriority, toggleHoldPriority } from "../../holdPriority";
  import { buildStackLane, type StackLaneItem } from "../../stackLane";
  import { createStackHover } from "../../stackHover";

  interface Props {
    stack: ZoneView;
    // S19 sub-PR 8: the zones an ability item's SOURCE card (and a
    // trigger's target) can live in. Abilities have no card in the
    // stack zone, so their art + target names resolve against
    // these. Optional — the overlay degrades to the label glyph.
    battlefield?: ZoneView;
    exile?: ZoneView;
    stackItems: StackItemView[];
    pendingTriggers: StackItemView[];
    seats: PlayerView[];
    viewerHasPriority: boolean;
    splitSecondActive: boolean;
    onCounter: (item: StackItemView) => void;
    onTargetStackItem?: (item: StackItemView) => void;
    // Whose window it is when the viewer doesn't hold priority — the
    // header names them instead of showing actions.
    priorityHolderName?: string | null;
    // #1307: true when the named priority holder is the seat Board
    // has decided is "considering a response" — a public-timing
    // read, never a hint about whether they actually have one.
    priorityHolderConsidering?: boolean;
    // Pass priority from the strip (same handler as the phase widget).
    onPass?: () => void;
  }

  const {
    stack,
    battlefield,
    exile,
    stackItems,
    pendingTriggers,
    seats,
    viewerHasPriority,
    splitSecondActive,
    onCounter,
    onTargetStackItem,
    priorityHolderName = null,
    priorityHolderConsidering = false,
    onPass,
  }: Props = $props();

  // The docked card never says "your", so it needs no viewer; the
  // priority header keeps reading its own props, which Board computes
  // from the same snapshot the lane's priority block is built from.
  const model = $derived(
    buildStackLane({
      stack,
      stackItems,
      pendingTriggers,
      seats,
      battlefield,
      exile,
      viewerID: null,
      priorityHolder: null,
      splitSecondActive,
    }),
  );

  // S20: per-item legality — with a server legal set only the
  // matching spells light up (Negate can't point at a creature
  // spell); free-form prompts fall back to "any stack item".
  function itemTargetable(item: StackLaneItem): boolean {
    const t = $targeting;
    return t !== null && isLegalCardTarget(t, item.id);
  }

  const displayItems = $derived(model.stackItems);
  const triggers = $derived(model.pendingTriggers);
  const visible = $derived(displayItems.length > 0 || triggers.length > 0 || stack.count > 0);

  // ---- #322 hover preview -------------------------------------
  // The overlay pins top-right and the strip pins top-left, so the
  // big preview lands beside the stack rather than over it.
  const hover = createStackHover(hoveredCard);

  function handleItemEnter(item: StackLaneItem): void {
    hover.enter(item.id, item.previewCard, $settings.display.hoverDelayMs);
  }

  function handleItemLeave(item: StackLaneItem): void {
    hover.leave(item.id);
  }

  // A stack item resolves out from under the cursor without ever
  // firing pointerleave — the row is simply removed.
  $effect(() => {
    hover.sync(displayItems.map((it) => it.id));
  });

  // Unmount (last item resolved, game over, seat switch) is the other
  // way the row can vanish mid-hover.
  $effect(() => {
    return () => hover.destroy();
  });
</script>

{#if visible}
  <div class="overlay" aria-label={`stack: ${displayItems.length} on the stack`}>
    <header class="stack-head">
      <span class="label">stack <b class="count">{displayItems.length}</b></span>
      {#if splitSecondActive}
        <span class="hint split-second" title="no responses allowed (split second is active)">
          <Icon name="bolt" size={11} /> split second — no responses
        </span>
      {:else if viewerHasPriority}
        <span class="hint">top resolves when everyone passes · you hold priority</span>
      {:else if priorityHolderName}
        <span class="hint">
          {#if priorityHolderConsidering}
            {priorityHolderName} is considering a response…
          {:else}
            {priorityHolderName} holds priority
          {/if}
        </span>
      {/if}
      <!-- #323: the same session hold as the phase widget's button,
           mirrored here because this card is on screen exactly when
           the stack is live — the moment the reporter is in when
           they want to change their mind. Armed, your own spells and
           triggers keep the cursor instead of auto-passing; it has
           to be armed before the cast, since the pass fires on the
           next snapshot. -->
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
    </header>
    <div class="items">
      {#each displayItems as item, i (item.id)}
        {@const src = cardImageURL(item.artCard, "small")}
        {@const stackTargetable = itemTargetable(item)}
        {@const previewable = item.previewCard !== null}
        <div class="line">
          <!-- svelte-ignore a11y_no_noninteractive_tabindex -->
          <div
            class="item"
            class:top={i === 0}
            class:cast-targetable={stackTargetable}
            class:previewable
            style:--seat-color={item.casterColor}
            data-stack-item-id={item.id}
            title={previewable ? `${item.name} — hover to preview` : item.name}
            onpointerenter={() => handleItemEnter(item)}
            onpointerleave={() => handleItemLeave(item)}
            onfocusin={() => handleItemEnter(item)}
            onfocusout={() => handleItemLeave(item)}
            onclick={stackTargetable ? () => onTargetStackItem?.(item.raw) : undefined}
            onkeydown={(e) => {
              if (stackTargetable && (e.key === "Enter" || e.key === " ")) {
                e.preventDefault();
                onTargetStackItem?.(item.raw);
              }
            }}
            role={stackTargetable ? "button" : undefined}
            tabindex={stackTargetable ? 0 : undefined}
          >
            <div class="thumb">
              {#if src}
                <img {src} alt="" loading="lazy" decoding="async" use:cardArt={src} />
              {:else}
                <span class="glyph">
                  {#if item.kind === "triggered"}<Icon name="bolt" size={16} />{:else}<Icon
                      name="spark"
                      size={16}
                    />{/if}
                </span>
              {/if}
            </div>
            <div class="info">
              <div class="title">{item.name}</div>
              <div class="sub">
                <span class="seat-dot"></span>
                <span class="caster-name">{item.casterName}</span>
                {#if item.targetText}
                  <span class="target-text">{item.targetText}</span>
                {/if}
                <!-- The chips — kind, doubled trigger, auto / manual,
                     X, alt cost, gift, chosen modes, split second, held
                     priority — and why each exists are in
                     lib/stackLane.ts (chipsFor). -->
                {#each item.chips as chip, ci (ci)}
                  <span
                    class="chip"
                    class:flag={chip.tone === "flag"}
                    class:manual={chip.tone === "manual"}
                    title={chip.title}
                  >
                    {chip.label}
                  </span>
                {/each}
              </div>
            </div>
          </div>
          <button
            type="button"
            class="act counter-btn"
            disabled={!viewerHasPriority}
            onclick={(e) => {
              e.stopPropagation();
              onCounter(item.raw);
            }}
            title={viewerHasPriority ? "counter this item" : "you don't hold priority"}
          >
            Counter
          </button>
          {#if i === 0 && viewerHasPriority && onPass}
            <button
              type="button"
              class="act primary pass-btn"
              onclick={onPass}
              title="pass priority — the top item resolves once everyone passes"
            >
              Pass <Icon name="chevronRight" size={12} />
            </button>
          {/if}
        </div>
      {/each}
    </div>
    {#if triggers.length > 0}
      <div class="pending-section">
        <span class="label">pending triggers <b class="count">{triggers.length}</b></span>
        <ul class="trigger-list">
          {#each triggers as t (t.id)}
            <li
              class="trigger-row"
              style:--seat-color={t.casterColor}
              title={`from ${t.casterName}`}
            >
              <span class="seat-dot"></span>
              <span class="trigger-label">{t.name}</span>
            </li>
          {/each}
        </ul>
      </div>
    {/if}
  </div>
{/if}

<style>
  .overlay {
    width: 100%;
    max-height: 60vh;
    background: color-mix(in srgb, var(--surface) 94%, transparent);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-lg);
    display: flex;
    flex-direction: column;
    color: var(--fg);
    font-size: 12px;
    overflow: hidden;
    box-sizing: border-box;
  }
  .stack-head {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 12px 6px 14px;
    min-width: 0;
  }
  .label {
    font-family: var(--font-mono);
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.14em;
    color: var(--fg-dim);
    font-weight: 700;
    flex: 0 0 auto;
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
  .count {
    color: var(--fg);
    font-weight: 700;
    font-variant-numeric: tabular-nums;
  }
  .hint {
    font-size: 11px;
    color: var(--fg-dim);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    min-width: 0;
    flex: 1 1 auto;
  }
  .hint :global(svg) {
    vertical-align: -2px;
    margin-right: 4px;
  }
  /* #323 hold toggle — the mono micro-chip vocabulary the rest of
     this card already uses for `.chip`, sized to sit on the header
     baseline. Engaged uses --magenta (a standing user instruction),
     never gold: gold means priority everywhere on the table. */
  .hold-toggle {
    flex: 0 0 auto;
    margin-left: auto;
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
    transition:
      color 120ms var(--ease),
      border-color 120ms var(--ease),
      background 120ms var(--ease);
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
  .hint.split-second {
    color: var(--danger);
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    font-weight: 700;
  }
  .items {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 0 10px 10px 12px;
    overflow-y: auto;
    min-height: 0;
  }
  .line {
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
  }
  .item {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 4px 10px 4px 4px;
    border-radius: 9px;
    background: var(--surface-raised);
    border: 1px solid var(--border);
    transition:
      border-color 120ms var(--ease),
      box-shadow 120ms var(--ease);
  }
  .item.top {
    border-color: rgba(217, 180, 92, 0.5);
    box-shadow: 0 0 0 1px rgba(217, 180, 92, 0.15);
  }
  .item.cast-targetable {
    cursor: pointer;
    box-shadow: var(--ring-gold);
    border-color: var(--gold);
  }
  .item.cast-targetable:hover {
    background: var(--surface-hover);
  }
  /* #322: the row is the hover-preview trigger, so it needs to look
     like one. zoom-in reads as "there is more of this to see"; the
     border lift is the same 120ms transition the row already
     declares. Targeting wins the cursor when it's live — a click
     there means "point my spell at this", not "show me the art". */
  .item.previewable {
    cursor: zoom-in;
  }
  .item.previewable:hover {
    border-color: var(--border-strong);
    background: var(--surface-hover);
  }
  .item.cast-targetable.previewable {
    cursor: pointer;
  }
  .thumb {
    /* positioned for the failed-art pip (#33) */
    position: relative;
    --art-error-top: 2px;
    --art-error-right: 2px;
    width: 40px;
    height: 56px;
    border-radius: 3px;
    overflow: hidden;
    flex: 0 0 auto;
    background: var(--surface-sunken);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--seat-color);
  }
  .thumb img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
  .info {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .title {
    font-weight: 700;
    font-size: 12.5px;
    color: var(--fg);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .sub {
    display: flex;
    align-items: center;
    gap: 5px;
    font-size: 11px;
    color: var(--fg-muted);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    min-width: 0;
  }
  .seat-dot {
    display: inline-block;
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--seat-color);
    flex: 0 0 auto;
  }
  .caster-name {
    color: var(--fg);
    font-weight: 600;
  }
  .target-text {
    color: var(--fg-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    min-width: 0;
  }
  .chip {
    font-family: var(--font-mono);
    font-size: 9px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--fg-dim);
    border: 1px solid var(--border);
    border-radius: 999px;
    padding: 1px 6px;
    flex: 0 0 auto;
  }
  .chip.flag {
    color: var(--gold-strong);
    border-color: rgba(217, 180, 92, 0.45);
  }
  /* Dashed rather than coloured. The gold `flag` chips mark things
     the engine is doing; this one marks the absence of one, and an
     outline with a gap in it says that without competing with them
     for attention. */
  .chip.manual {
    border-style: dashed;
    border-color: var(--border-strong);
  }
  .act {
    flex: 0 0 auto;
    height: 28px;
    padding: 0 10px;
    font-size: 11.5px;
    border-radius: 7px;
    display: inline-flex;
    align-items: center;
    gap: 3px;
  }
  .act:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
  .counter-btn:hover:not(:disabled) {
    color: var(--danger);
    border-color: rgba(255, 107, 107, 0.5);
  }
  .pending-section {
    padding: 8px 12px 10px 14px;
    border-top: 1px solid var(--border);
    background: var(--surface-sunken);
  }
  .trigger-list {
    list-style: none;
    margin: 6px 0 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .trigger-row {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 11px;
    color: var(--fg-muted);
  }
  .trigger-label {
    color: var(--fg-muted);
  }
</style>
