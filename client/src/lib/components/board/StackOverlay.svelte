<script lang="ts">
  // StackOverlay is the floating display of the shared stack zone,
  // anchored at (0.2, 0.2) of the board area — top-left quadrant.
  // Mirrors the placement convention of HoverZoomOverlay (top-right
  // at 0.8, 0.2) so both overlays share visual hierarchy without
  // colliding. Visible only when the stack has metadata items
  // (StackMeta is the source of truth post-S13.1; the legacy
  // ZoneView fallback covers admin direct-drops that bypass the
  // cast verb).
  //
  // Each entry is a small Card stacked from bottom to top with a
  // fixed offset, decorated with the caster's seat colour and a
  // counter button visible to the priority holder.

  import type { CardView, PlayerView, StackItemView, ZoneView } from "../../protocol";
  import { seatColor } from "../../colors";
  import Card from "./Card.svelte";

  interface Props {
    stack: ZoneView;
    stackItems: StackItemView[];
    pendingTriggers: StackItemView[];
    seats: PlayerView[];
    viewerHasPriority: boolean;
    splitSecondActive: boolean;
    onCounter: (item: StackItemView) => void;
  }

  const {
    stack,
    stackItems,
    pendingTriggers,
    seats,
    viewerHasPriority,
    splitSecondActive,
    onCounter,
  }: Props = $props();

  // Map instance_id → CardView for rendering spell items. Ability
  // items don't have a card on the stack; their source card stays
  // in its origin zone, so for them we render a label-only chip.
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
        // Card target: try to find a meaningful name.
        const c = cardByID.get(t.id ?? "");
        return c?.name ?? "card";
      })
      .join(" / ");
    return `→ ${parts}`;
  }

  // Items render bottom-to-top so the most-recently-cast (top of
  // the stack) sits on top of the pile. The wire delivers
  // stack_items in stack order (bottom..top) per the server's
  // viewOfStackItemsInStackOrder helper.
  const items = $derived(stackItems ?? []);
  const triggers = $derived(pendingTriggers ?? []);
  const visible = $derived(items.length > 0 || triggers.length > 0 || stack.count > 0);
</script>

{#if visible}
  <div class="overlay" aria-label={`stack: ${items.length} on the stack`}>
    {#if splitSecondActive}
      <div class="split-second" title="No responses allowed (split second is active)">
        ⚡ split second
      </div>
    {/if}
    <span class="label">stack · {items.length}</span>
    <div class="items">
      {#each items as item (item.id)}
        <div class="item-row" style:--seat-color={seatColor(controllerSeatNum(item))}>
          {#if cardByID.has(item.id)}
            <div class="thumb">
              <Card card={cardByID.get(item.id) as CardView} />
            </div>
          {:else}
            <div class="ability-thumb" title="ability">
              {item.kind === "triggered" ? "⚡" : "✦"}
            </div>
          {/if}
          <div class="meta">
            <div class="line caster">{controllerName(item)}</div>
            {#if cardByID.has(item.id) && cardByID.get(item.id)?.name}
              <div class="line spell-name">{cardByID.get(item.id)?.name}</div>
            {/if}
            {#if item.label}
              <div class="line label-text">{item.label}</div>
            {/if}
            {#if item.targets && item.targets.length > 0}
              <div class="line target-text">{targetLabel(item)}</div>
            {/if}
            {#if item.x_value}
              <div class="line">X = {item.x_value}</div>
            {/if}
            {#if item.modes && item.modes.length > 0}
              <div class="line">modes: {item.modes.join(", ")}</div>
            {/if}
            {#if item.split_second}
              <div class="line flag">split-second</div>
            {/if}
            {#if item.hold_priority}
              <div class="line flag">held priority</div>
            {/if}
          </div>
          <button
            type="button"
            class="counter-btn"
            disabled={!viewerHasPriority}
            onclick={() => onCounter(item)}
            title={viewerHasPriority ? "counter this item" : "you don't hold priority"}
          >
            counter
          </button>
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
    left: 20%;
    top: 20%;
    transform: translate(-50%, -50%);
    background: #0b1220;
    border: 1px solid #4a5270;
    border-radius: 12px;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.6);
    z-index: 40;
    padding: 10px 10px 14px;
    box-sizing: border-box;
    min-width: 240px;
    max-width: 320px;
    pointer-events: auto;
  }
  .label {
    display: block;
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: #6c7a99;
    margin-bottom: 6px;
  }
  .split-second {
    background: #4a1a1a;
    color: #ffd6d6;
    border: 1px solid #c54848;
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 11px;
    margin-bottom: 8px;
    text-align: center;
  }
  .items {
    position: relative;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .item-row {
    display: grid;
    grid-template-columns: 56px 1fr auto;
    align-items: center;
    gap: 8px;
    padding: 4px 6px;
    border-left: 3px solid var(--seat-color);
    background: #131a2c;
    border-radius: 4px;
  }
  .thumb {
    width: 56px;
    height: 78px;
    /* Cascade card size to the child Card via the CSS vars it reads
       for --card-w / --card-h, so the image fills the thumb rather
       than rendering at its 80x112 default and getting cropped. */
    --card-w: 56px;
    --card-h: 78px;
    overflow: hidden;
    border-radius: 3px;
  }
  .ability-thumb {
    width: 40px;
    height: 40px;
    background: var(--seat-color);
    color: #0c1426;
    font-size: 22px;
    font-weight: 700;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 4px;
  }
  .meta {
    font-size: 11px;
    color: #cfd6ee;
    line-height: 1.3;
  }
  .meta .line {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .meta .caster {
    color: var(--seat-color);
    font-weight: 600;
  }
  .meta .spell-name {
    color: #e0e6f5;
    font-weight: 600;
  }
  .meta .target-text {
    color: #9aa5cd;
  }
  .meta .flag {
    color: #c5a648;
    font-size: 10px;
    text-transform: uppercase;
  }
  .counter-btn {
    font-size: 10px;
    padding: 4px 8px;
    border-radius: 4px;
    background: #2a3148;
    color: #cfd6ee;
    border: 1px solid #4a5270;
    cursor: pointer;
  }
  .counter-btn:hover:not(:disabled) {
    background: #3a4263;
  }
  .counter-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
  .pending-section {
    margin-top: 10px;
    padding-top: 8px;
    border-top: 1px solid #2a3148;
  }
  .trigger-list {
    list-style: none;
    margin: 0;
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
