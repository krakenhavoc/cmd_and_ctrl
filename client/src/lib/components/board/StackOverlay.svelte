<script lang="ts">
  // StackOverlay is the floating display of the shared stack zone,
  // pinned to the top-left corner of the board. Visually mirrors
  // HoverZoomOverlay (top-right): same clamp width, same
  // image-over-info shape. Each stack item renders as its own panel
  // so players can read the full card art + caster + targets at a
  // glance without hovering.
  //
  // Items stack top-to-bottom with the top of the visual stack
  // (most-recently-cast, resolves first) at the top of the overlay.
  // The container scrolls if a deep stack overflows viewport height.

  import type { CardView, PlayerView, StackItemView, ZoneView } from "../../protocol";
  import { seatColor } from "../../colors";
  import { targeting, isTargetingStack } from "../../targeting";

  interface Props {
    stack: ZoneView;
    stackItems: StackItemView[];
    pendingTriggers: StackItemView[];
    seats: PlayerView[];
    viewerHasPriority: boolean;
    splitSecondActive: boolean;
    onCounter: (item: StackItemView) => void;
    onTargetStackItem?: (item: StackItemView) => void;
  }

  const {
    stack,
    stackItems,
    pendingTriggers,
    seats,
    viewerHasPriority,
    splitSecondActive,
    onCounter,
    onTargetStackItem,
  }: Props = $props();

  const stackTargetable = $derived.by(() => {
    const t = $targeting;
    return t !== null && isTargetingStack(t.mode);
  });

  const cardByID = $derived.by(() => {
    const out = new Map<string, CardView>();
    for (const c of stack.cards) {
      out.set(c.instance_id, c);
    }
    return out;
  });

  const seatBySeatID = $derived.by(() => {
    const out = new Map<string, PlayerView>();
    for (const s of seats) out.set(s.id, s);
    return out;
  });

  function controllerName(item: StackItemView): string {
    return seatBySeatID.get(item.controller)?.name ?? "?";
  }

  function controllerSeatNum(item: StackItemView): number {
    return seatBySeatID.get(item.controller)?.seat ?? 0;
  }

  function targetLabel(item: StackItemView): string {
    if (!item.targets || item.targets.length === 0) return "";
    const parts = item.targets
      .map((t) => {
        if (t.kind === "self") return "self";
        if (t.kind === "none") return "—";
        if (t.kind === "player") {
          return seatBySeatID.get(t.id ?? "")?.name ?? "player";
        }
        const c = cardByID.get(t.id ?? "");
        return c?.name ?? "card";
      })
      .join(" / ");
    return `→ ${parts}`;
  }

  function imgSrcFor(item: StackItemView): string | null {
    const c = cardByID.get(item.id);
    if (!c || !c.scryfall_id) return null;
    return `/cards/${c.scryfall_id}/image?size=normal`;
  }

  function titleFor(item: StackItemView): string {
    const c = cardByID.get(item.id);
    if (c?.name) return c.name;
    if (item.label) return item.label;
    return item.kind === "triggered" ? "trigger" : "ability";
  }

  // Wire delivers stack_items bottom..top (per viewOfStackItemsInStackOrder
  // on the server). The user-facing convention is "top of stack resolves
  // first", so we render in reverse — top entry in the overlay = top of
  // the stack = next to resolve.
  const displayItems = $derived([...(stackItems ?? [])].reverse());
  const triggers = $derived(pendingTriggers ?? []);
  const visible = $derived(displayItems.length > 0 || triggers.length > 0 || stack.count > 0);
</script>

{#if visible}
  <div class="overlay" aria-label={`stack: ${displayItems.length} on the stack`}>
    {#if splitSecondActive}
      <div class="split-second" title="no responses allowed (split second is active)">
        ⚡ split second
      </div>
    {/if}
    <header class="stack-head">
      <span class="label">stack</span>
      <span class="count">{displayItems.length}</span>
    </header>
    <div class="items">
      {#each displayItems as item (item.id)}
        {@const seatNum = controllerSeatNum(item)}
        {@const src = imgSrcFor(item)}
        <!-- svelte-ignore a11y_no_noninteractive_tabindex -->
        <div
          class="item"
          class:cast-targetable={stackTargetable}
          style:--seat-color={seatColor(seatNum)}
          onclick={stackTargetable ? () => onTargetStackItem?.(item) : undefined}
          onkeydown={(e) => {
            if (stackTargetable && (e.key === "Enter" || e.key === " ")) {
              e.preventDefault();
              onTargetStackItem?.(item);
            }
          }}
          role={stackTargetable ? "button" : undefined}
          tabindex={stackTargetable ? 0 : undefined}
        >
          <div class="image-wrap">
            {#if src}
              <img {src} alt={titleFor(item)} loading="lazy" decoding="async" />
            {:else}
              <div class="image-fallback">
                <span class="glyph">{item.kind === "triggered" ? "⚡" : "✦"}</span>
                <span class="fb-name">{titleFor(item)}</span>
              </div>
            {/if}
            <button
              type="button"
              class="counter-btn"
              disabled={!viewerHasPriority}
              onclick={(e) => {
                e.stopPropagation();
                onCounter(item);
              }}
              title={viewerHasPriority ? "counter this item" : "you don't hold priority"}
            >
              counter
            </button>
          </div>
          <div class="info">
            <div class="title">{titleFor(item)}</div>
            <div class="caster">
              <span class="seat-dot"></span>
              <span class="caster-name">{controllerName(item)}</span>
              {#if item.kind !== "spell"}
                <span class="kind-chip">{item.kind}</span>
              {/if}
              {#if cardByID.get(item.id)?.auto}
                <span
                  class="auto-chip"
                  title="this card auto-resolves — effect fires when priority passes to empty stack"
                >
                  auto
                </span>
              {/if}
            </div>
            {#if item.targets && item.targets.length > 0}
              <div class="target-text">{targetLabel(item)}</div>
            {/if}
            <div class="meta-flags">
              {#if item.x_value}
                <span class="chip">X = {item.x_value}</span>
              {/if}
              {#if item.modes && item.modes.length > 0}
                <span class="chip">modes: {item.modes.join(", ")}</span>
              {/if}
              {#if item.split_second}
                <span class="chip flag">split-second</span>
              {/if}
              {#if item.hold_priority}
                <span class="chip flag">held priority</span>
              {/if}
            </div>
          </div>
        </div>
      {/each}
    </div>
    {#if triggers.length > 0}
      <div class="pending-section">
        <span class="label">pending triggers · {triggers.length}</span>
        <ul class="trigger-list">
          {#each triggers as t (t.id)}
            <li
              class="trigger-row"
              style:--seat-color={seatColor(controllerSeatNum(t))}
              title={`from ${controllerName(t)}`}
            >
              <span class="seat-dot"></span>
              <span class="trigger-label">{t.label || "trigger"}</span>
            </li>
          {/each}
        </ul>
      </div>
    {/if}
  </div>
{/if}

<style>
  .overlay {
    position: absolute;
    left: 12px;
    top: 12px;
    width: clamp(220px, 26vw, 380px);
    max-height: calc(100vh - 24px);
    background: #0b1220;
    border: 1px solid #4a5270;
    border-radius: 12px;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.6);
    z-index: 40;
    display: flex;
    flex-direction: column;
    color: #e0e6f5;
    font-size: 12px;
    overflow: hidden;
    pointer-events: auto;
  }
  .split-second {
    background: #4a1a1a;
    color: #ffd6d6;
    border-bottom: 1px solid #c54848;
    padding: 4px 10px;
    font-size: 11px;
    text-align: center;
    letter-spacing: 0.05em;
  }
  .stack-head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    padding: 8px 10px 6px;
    border-bottom: 1px solid #1a2335;
  }
  .label {
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 0.12em;
    color: #6c7a99;
  }
  .count {
    font-size: 11px;
    font-weight: 700;
    color: #cfd6ee;
    font-variant-numeric: tabular-nums;
  }
  .items {
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 10px;
    overflow-y: auto;
    min-height: 0;
  }
  .item {
    border: 1px solid #2a3148;
    border-left: 3px solid var(--seat-color);
    border-radius: 8px;
    background: #131a2c;
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }
  .item.cast-targetable {
    cursor: pointer;
    outline: 2px solid #ffd07a;
    outline-offset: 1px;
    box-shadow: 0 0 12px rgba(255, 208, 122, 0.45);
  }
  .item.cast-targetable:hover {
    background: color-mix(in srgb, #ffd07a 18%, #131a2c);
  }
  .image-wrap {
    position: relative;
    background: #0b1220;
  }
  .image-wrap img {
    width: 100%;
    aspect-ratio: 63 / 88;
    object-fit: contain;
    display: block;
  }
  .image-fallback {
    width: 100%;
    aspect-ratio: 63 / 88;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 10px;
    background: #1a2335;
    color: #cfd6ee;
    text-align: center;
    padding: 12px;
    box-sizing: border-box;
  }
  .image-fallback .glyph {
    font-size: 36px;
    color: var(--seat-color);
  }
  .image-fallback .fb-name {
    font-size: 13px;
  }
  .counter-btn {
    position: absolute;
    right: 8px;
    bottom: 8px;
    font-size: 10px;
    padding: 4px 10px;
    border-radius: 4px;
    background: rgba(20, 26, 44, 0.92);
    color: #cfd6ee;
    border: 1px solid #4a5270;
    cursor: pointer;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    backdrop-filter: blur(2px);
  }
  .counter-btn:hover:not(:disabled) {
    background: #3a4263;
    border-color: #6c7a99;
  }
  .counter-btn:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
  .info {
    padding: 6px 10px 8px;
    display: flex;
    flex-direction: column;
    gap: 3px;
    border-top: 1px solid #1a2335;
  }
  .title {
    font-weight: 700;
    font-size: 12px;
    color: #e0e6f5;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .caster {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 11px;
  }
  .seat-dot {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--seat-color);
    flex: 0 0 auto;
  }
  .caster-name {
    color: var(--seat-color);
    font-weight: 600;
  }
  .kind-chip {
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    padding: 1px 5px;
    border-radius: 3px;
    background: #1a2335;
    color: #9aa5cd;
    border: 1px solid #2e3a55;
  }
  .auto-chip {
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    padding: 1px 5px;
    border-radius: 3px;
    background: rgba(80, 60, 0, 0.7);
    color: #ffd07a;
    border: 1px solid #c8a86a;
    font-weight: 700;
  }
  .target-text {
    font-size: 11px;
    color: #9aa5cd;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .meta-flags {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
  }
  .chip {
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    padding: 1px 5px;
    border-radius: 3px;
    background: #1a2335;
    color: #c8a86a;
    border: 1px solid #2e3a55;
  }
  .chip.flag {
    color: #ffd07a;
    border-color: #5a4a1a;
    background: #2a2010;
  }
  .pending-section {
    padding: 8px 10px;
    border-top: 1px solid #2a3148;
    background: #0e1424;
  }
  .trigger-list {
    list-style: none;
    margin: 4px 0 0;
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
    color: #cfd6ee;
  }
  .trigger-row .seat-dot {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--seat-color);
  }
  .trigger-label {
    color: #9aa5cd;
  }
</style>
