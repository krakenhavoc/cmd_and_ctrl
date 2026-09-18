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

  import type { CardView, PlayerView, StackItemView, ZoneView } from "../../protocol";
  import { seatColor } from "../../colors";
  import { cardImageURL } from "../../cardImage";
  import { cardArt } from "../../cardArt";
  import Icon from "../Icon.svelte";
  import { targeting, isLegalCardTarget } from "../../targeting";
  import { hoveredCard } from "../../cardTypes";
  import { settings } from "../../settings";
  import { holdPriority, toggleHoldPriority } from "../../holdPriority";
  import { doubledTriggerLabel } from "../../triggerDoubling";
  import { previewableCard } from "../../stackPreview";

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
    return cardImageURL(artCardFor(item), "small");
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

  // ---- #322 hover preview -------------------------------------
  // Deliberately plain `let`, not `$state`: the cleanup effect below
  // wants to react to the stack changing under the cursor, not to
  // its own bookkeeping writes.
  let hoverTimer: ReturnType<typeof setTimeout> | null = null;
  let hoveredItemID: string | null = null;
  let previewedInstanceID: string | null = null;

  function cancelHoverTimer(): void {
    if (hoverTimer !== null) {
      clearTimeout(hoverTimer);
      hoverTimer = null;
    }
  }

  // previewCardFor is the card the overlay should show for an item:
  // the spell itself, or an ability's source permanent. Returns null
  // for anything the viewer isn't allowed to read — the same rule
  // Card.svelte applies before writing the store, shared through
  // stackPreview.ts so the two cannot disagree.
  //
  // #697: this used to test `known_by_you === false`, which the server
  // never sends (the field is omitempty), so the guard never fired and
  // an unreadable card opened a blank zoom panel.
  function previewCardFor(item: StackItemView): CardView | null {
    return previewableCard(artCardFor(item));
  }

  // clearPreview drops our own write to the shared store, never
  // anyone else's — a battlefield card hovered after us owns the
  // slot and must survive.
  function clearPreview(): void {
    const inst = previewedInstanceID;
    hoveredItemID = null;
    previewedInstanceID = null;
    if (!inst) return;
    hoveredCard.update((c) => (c?.instance_id === inst ? null : c));
  }

  function handleItemEnter(item: StackItemView): void {
    const c = previewCardFor(item);
    if (!c) return;
    cancelHoverTimer();
    hoveredItemID = item.id;
    const delay = $settings.display.hoverDelayMs;
    if (delay <= 0) {
      previewedInstanceID = c.instance_id;
      hoveredCard.set(c);
      return;
    }
    hoverTimer = setTimeout(() => {
      hoverTimer = null;
      previewedInstanceID = c.instance_id;
      hoveredCard.set(c);
    }, delay);
  }

  function handleItemLeave(item: StackItemView): void {
    cancelHoverTimer();
    if (hoveredItemID !== item.id) return;
    clearPreview();
  }

  // A stack item resolves out from under the cursor without ever
  // firing pointerleave — the row is simply removed. Without this the
  // preview of a resolved spell would hang around until the user
  // hovered something else.
  $effect(() => {
    const items = displayItems;
    if (hoveredItemID === null) return;
    if (items.some((it) => it.id === hoveredItemID)) return;
    cancelHoverTimer();
    clearPreview();
  });

  // Unmount (last item resolved, game over, seat switch) is the other
  // way the row can vanish mid-hover.
  $effect(() => {
    return () => {
      cancelHoverTimer();
      clearPreview();
    };
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
        <span class="hint">{priorityHolderName} holds priority</span>
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
        {@const seatNum = controllerSeatNum(item)}
        {@const src = imgSrcFor(item)}
        {@const stackTargetable = itemTargetable(item)}
        {@const previewable = previewCardFor(item) !== null}
        {@const doubledLabel = doubledTriggerLabel(item.doubled_by, item.doubled_by_name)}
        <div class="line">
          <!-- svelte-ignore a11y_no_noninteractive_tabindex -->
          <div
            class="item"
            class:top={i === 0}
            class:cast-targetable={stackTargetable}
            class:previewable
            style:--seat-color={seatColor(seatNum)}
            data-stack-item-id={item.id}
            title={previewable ? `${titleFor(item)} — hover to preview` : titleFor(item)}
            onpointerenter={() => handleItemEnter(item)}
            onpointerleave={() => handleItemLeave(item)}
            onfocusin={() => handleItemEnter(item)}
            onfocusout={() => handleItemLeave(item)}
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
                {#if doubledLabel}
                  <span class="chip flag">{doubledLabel}</span>
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
                <!-- #764: the chosen bullets, in announce order (CR
                     700.2c) and with repeats (CR 700.2d). This used to
                     print the raw indexes ("modes: 0, 2"), which nobody
                     at the table could read: the caster's hand card is
                     gone once the spell is on the stack, so the labels
                     have to travel with the item. -->
                {#if item.mode_labels && item.mode_labels.length > 0}
                  {#each item.mode_labels as modeLabel, mi (mi)}
                    <span class="chip">{modeLabel}</span>
                  {/each}
                {:else if item.modes && item.modes.length > 0}
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
