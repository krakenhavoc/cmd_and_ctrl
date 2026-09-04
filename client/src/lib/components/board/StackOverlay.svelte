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

  // Every card the overlay might need to name or draw: stack (spell
  // items), battlefield (ability sources + most targets), exile and
  // graveyards (dies-trigger sources, exiled targets).
  const anyCardByID = $derived.by(() => {
    const out = new Map<string, CardView>(cardByID);
    for (const c of battlefield?.cards ?? []) out.set(c.instance_id, c);
    for (const c of exile?.cards ?? []) out.set(c.instance_id, c);
    for (const s of seats) {
      for (const c of s.graveyard?.cards ?? []) out.set(c.instance_id, c);
    }
    return out;
  });

  // The card whose art represents an item: the spell itself, or an
  // ability's source permanent.
  function artCardFor(item: StackItemView): CardView | undefined {
    if (item.kind === "spell") return cardByID.get(item.id);
    return anyCardByID.get(item.source_card_id);
  }

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
        const c = anyCardByID.get(t.id ?? "");
        return c?.name ?? "card";
      })
      .join(" / ");
    return `→ ${parts}`;
  }

  function imgSrcFor(item: StackItemView): string | null {
    const c = artCardFor(item);
    if (!c || !c.scryfall_id) return null;
    return `/cards/${c.scryfall_id}/image?size=normal`;
  }

  function titleFor(item: StackItemView): string {
    if (item.kind === "spell") {
      const c = cardByID.get(item.id);
      if (c?.name) return c.name;
    }
    if (item.label) return item.label;
    const src = anyCardByID.get(item.source_card_id);
    if (src?.name) return src.name;
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
          data-stack-item-id={item.id}
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
    background: linear-gradient(180deg, rgba(19, 26, 44, 0.92) 0%, rgba(8, 12, 24, 0.92) 100%);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border: 1px solid rgba(122, 167, 255, 0.22);
    border-radius: var(--radius-lg);
    box-shadow:
      0 12px 40px rgba(0, 0, 0, 0.6),
      0 0 0 1px rgba(0, 0, 0, 0.4),
      inset 0 1px 0 rgba(255, 255, 255, 0.06);
    z-index: 40;
    display: flex;
    flex-direction: column;
    color: var(--fg);
    font-size: 12px;
    overflow: hidden;
    pointer-events: auto;
  }
  .split-second {
    background: linear-gradient(180deg, #5a1a1a 0%, #3a1010 100%);
    color: #ffd6d6;
    border-bottom: 1px solid rgba(197, 72, 72, 0.6);
    padding: 6px 12px;
    font-size: 11px;
    text-align: center;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    font-weight: 700;
    text-shadow: 0 1px 0 rgba(0, 0, 0, 0.4);
  }
  .stack-head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    padding: 10px 14px 8px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    background: linear-gradient(180deg, rgba(255, 255, 255, 0.03) 0%, rgba(0, 0, 0, 0) 100%);
  }
  .label {
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 0.16em;
    color: var(--fg-dim);
    font-weight: 700;
  }
  .count {
    font-size: 12px;
    font-weight: 800;
    color: var(--fg);
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
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-left: 3px solid var(--seat-color);
    border-radius: var(--radius);
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 0.03) 0%, rgba(0, 0, 0, 0.25) 100%), #0f162a;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    box-shadow:
      0 4px 12px rgba(0, 0, 0, 0.3),
      inset 0 1px 0 rgba(255, 255, 255, 0.03);
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
    padding: 5px 12px;
    border-radius: 999px;
    background: rgba(20, 26, 44, 0.85);
    color: var(--fg);
    border: 1px solid rgba(255, 255, 255, 0.18);
    cursor: pointer;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    font-weight: 700;
    backdrop-filter: blur(6px);
    box-shadow: 0 4px 10px rgba(0, 0, 0, 0.4);
    transition:
      background 120ms var(--ease),
      border-color 120ms var(--ease);
  }
  .counter-btn:hover:not(:disabled) {
    background: rgba(255, 122, 122, 0.25);
    border-color: rgba(255, 122, 122, 0.6);
    color: var(--danger);
  }
  .counter-btn:disabled {
    opacity: 0.4;
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
