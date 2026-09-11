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

  import type { CardView, PlayerView, StackItemView, ZoneView } from "../../protocol";
  import { seatColor } from "../../colors";
  import Icon from "../Icon.svelte";
  import { targeting, isLegalCardTarget } from "../../targeting";

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
    onPass,
  }: Props = $props();

  // S20: per-item legality — with a server legal set only the
  // matching spells light up (Negate can't point at a creature
  // spell); free-form prompts fall back to "any stack item".
  function itemTargetable(item: StackItemView): boolean {
    const t = $targeting;
    return t !== null && isLegalCardTarget(t, item.id);
  }

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
    return `/cards/${c.scryfall_id}/image?size=small`;
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
    <header class="stack-head">
      <span class="label">stack <b class="count">{displayItems.length}</b></span>
      {#if splitSecondActive}
        <span class="hint split-second" title="no responses allowed (split second is active)">
          <Icon name="bolt" size={11} /> split second — no responses
        </span>
      {:else if viewerHasPriority}
        <span class="hint">top resolves when everyone passes · you hold priority</span>
      {:else if priorityHolderName}
        <span class="hint">{priorityHolderName} holds priority</span>
      {/if}
    </header>
    <div class="items">
      {#each displayItems as item, i (item.id)}
        {@const seatNum = controllerSeatNum(item)}
        {@const src = imgSrcFor(item)}
        {@const stackTargetable = itemTargetable(item)}
        <div class="line">
          <!-- svelte-ignore a11y_no_noninteractive_tabindex -->
          <div
            class="item"
            class:top={i === 0}
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
            <div class="thumb">
              {#if src}
                <img {src} alt="" loading="lazy" decoding="async" />
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
              <div class="title">{titleFor(item)}</div>
              <div class="sub">
                <span class="seat-dot"></span>
                <span class="caster-name">{controllerName(item)}</span>
                {#if item.targets && item.targets.length > 0}
                  <span class="target-text">{targetLabel(item)}</span>
                {/if}
                {#if item.kind !== "spell"}
                  <span class="chip">{item.kind}</span>
                {/if}
                {#if cardByID.get(item.id)?.auto}
                  <span
                    class="chip flag"
                    title="this card auto-resolves — effect fires when priority passes to empty stack"
                  >
                    auto
                  </span>
                {/if}
                <!-- The moment the expectation forms. The spell is on
                     the stack, everyone is looking at it, and in a
                     second it will resolve and appear to do nothing.
                     Saying so here is what reports #321 / #324 /
                     #325 / #332 / #333 each needed and none of them
                     got. Transient by construction — the chip leaves
                     with the stack item, so it never becomes board
                     furniture. -->
                {#if cardByID.get(item.id)?.unimplemented}
                  <span
                    class="chip manual"
                    title="this card's rules aren't implemented yet — it resolves with no effect, so resolve it by hand"
                  >
                    manual
                  </span>
                {/if}
                {#if item.x_value}
                  <span class="chip">X = {item.x_value}</span>
                {/if}
                <!-- S22: an overloaded Rift wipes and a hard-cast one bounces one thing. -->
                {#if item.alt_cost}
                  <span class="chip flag">{item.alt_cost}</span>
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
          <button
            type="button"
            class="act counter-btn"
            disabled={!viewerHasPriority}
            onclick={(e) => {
              e.stopPropagation();
              onCounter(item);
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
  .thumb {
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
