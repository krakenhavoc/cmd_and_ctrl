<script lang="ts">
  // StackLaneRibbon — style C (#1467): one compact, numbered row —
  // resolution order, top item first, left to right — that scrolls
  // horizontally rather than growing tall. StackLaneHost draws the
  // header, summary line and aria-live announcer; this component
  // draws only the items, off the shared model in lib/stackLane.ts.

  import type { StackLaneItem, StackLaneStyle, StackLaneStyleProps } from "../../stackLane";
  import { cardImageURL } from "../../cardImage";
  import { cardArt } from "../../cardArt";
  import Icon from "../Icon.svelte";

  interface Props extends StackLaneStyleProps {
    styleName: StackLaneStyle;
  }

  const { model, controls, styleName }: Props = $props();

  // Resolution order: the stack top-first, then any pending triggers
  // (not on the stack yet) last.
  const items = $derived(model.items);

  const nameByID = $derived(new Map(items.map((it) => [it.id, it.name] as const)));
  function targetedByNames(item: StackLaneItem): string {
    return item.targetedBy.map((id) => nameByID.get(id) ?? "something").join(", ");
  }

  /** The one-line summary: "→ targets" when there are any, else the effect text. */
  function line(item: StackLaneItem): string {
    if (item.targets.length > 0) return `→ ${item.targets.map((t) => t.phrase).join(" / ")}`;
    if (item.effect) return item.effect.text;
    return "";
  }

  function badge(item: StackLaneItem): string {
    if (item.isTop) return `${item.position} · NEXT`;
    if (item.kind === "pending") return "waiting";
    if (item.kind === "triggered") return `${item.position} · trigger`;
    if (item.kind === "activated") return `${item.position} · ability`;
    return `${item.position}`;
  }

  function counterTitle(): string {
    return model.priority.viewerHolds ? "counter this item" : "you don't hold priority";
  }
</script>

<div class="ribbon" data-stack-body={styleName}>
  <div class="row">
    {#each items as item, i (item.id)}
      {#if i > 0}
        <svg class="chevron" width="20" height="16" viewBox="0 0 28 20" aria-hidden="true">
          <path
            d="M2 10 H22 M16 4 L23 10 L16 16"
            fill="none"
            stroke="var(--fg-dim)"
            stroke-width="2"
          />
        </svg>
      {/if}
      {@const src = cardImageURL(item.artCard, "small")}
      {@const targetable = controls.targetable(item)}
      <div class="r-line">
        <!-- svelte-ignore a11y_no_noninteractive_tabindex -->
        <div
          class="r-item"
          class:top={item.isTop}
          class:waiting={item.kind === "pending"}
          class:targeted={item.targetedBy.length > 0}
          class:cast-targetable={targetable}
          class:previewable={item.previewCard !== null}
          style:--seat-color={item.casterColor}
          data-stack-item-id={item.id}
          title={item.previewCard ? `${item.name} — hover to preview` : item.name}
          onpointerenter={() => controls.hoverEnter(item)}
          onpointerleave={() => controls.hoverLeave(item)}
          onfocusin={() => controls.hoverEnter(item)}
          onfocusout={() => controls.hoverLeave(item)}
          onclick={targetable ? () => controls.target(item) : undefined}
          onkeydown={(e) => {
            if (targetable && (e.key === "Enter" || e.key === " ")) {
              e.preventDefault();
              controls.target(item);
            }
          }}
          role={targetable ? "button" : undefined}
          tabindex={targetable ? 0 : undefined}
        >
          <div class="r-thumb">
            {#if src}
              <img {src} alt="" loading="lazy" decoding="async" use:cardArt={src} />
            {:else}
              <Icon name={item.kind === "triggered" ? "bolt" : "spark"} size={16} />
            {/if}
          </div>
          <div class="r-info">
            <span class="r-badge">{badge(item)}</span>
            <span class="r-name">{item.name}</span>
            <span class="r-caster">
              <span class="seat-dot"></span>{item.casterIsViewer ? "You" : item.casterName}
            </span>
            {#if line(item)}<span class="r-sub">{line(item)}</span>{/if}
            {#if item.targetedBy.length > 0}
              <span class="r-targeted-by">targeted by {targetedByNames(item)}</span>
            {/if}
          </div>
        </div>
        <button
          type="button"
          class="act counter-btn"
          disabled={!model.priority.viewerHolds}
          aria-label={`Counter ${item.name}`}
          onclick={(e) => {
            e.stopPropagation();
            controls.counter(item);
          }}
          title={counterTitle()}
        >
          Counter
        </button>
      </div>
    {/each}
  </div>
</div>

<style>
  .ribbon {
    min-width: 0;
  }
  .row {
    display: flex;
    align-items: center;
    gap: 8px;
    overflow-x: auto;
    overflow-y: hidden;
    padding-bottom: 4px;
  }
  .chevron {
    flex: 0 0 auto;
  }
  .r-line {
    flex: 0 0 auto;
    display: flex;
    flex-direction: column;
    align-items: stretch;
    gap: 4px;
  }
  .r-item {
    flex: 0 0 auto;
    width: 230px;
    box-sizing: border-box;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px;
    border-radius: 12px;
    background: var(--surface-raised);
    border: 1px solid var(--border);
    transition:
      border-color 120ms var(--ease),
      box-shadow 120ms var(--ease);
  }
  .r-item.top {
    border-color: var(--gold);
    box-shadow: 0 0 0 2px var(--gold);
    animation: ribbon-pulse 2.2s ease-in-out infinite;
  }
  @keyframes ribbon-pulse {
    0%,
    100% {
      box-shadow: 0 0 0 2px var(--gold);
    }
    50% {
      box-shadow: 0 0 0 2px color-mix(in srgb, var(--gold) 55%, transparent);
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .r-item.top {
      animation: none;
    }
  }
  .r-item.waiting {
    border-style: dashed;
  }
  .r-item.targeted {
    border-color: color-mix(in srgb, var(--rose) 50%, transparent);
    box-shadow: 0 0 0 1px color-mix(in srgb, var(--rose) 45%, transparent);
  }
  .r-item.previewable {
    cursor: zoom-in;
  }
  .r-item.previewable:hover {
    border-color: var(--border-strong);
    background: var(--surface-hover);
  }
  .r-item.cast-targetable {
    cursor: pointer;
    box-shadow: var(--ring-gold);
    border-color: var(--gold);
  }
  .r-thumb {
    position: relative;
    --art-error-top: 2px;
    --art-error-right: 2px;
    width: 96px;
    height: 134px;
    border-radius: 8px;
    overflow: hidden;
    flex: 0 0 auto;
    background: var(--surface-sunken);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--seat-color);
  }
  .r-thumb img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
  .r-info {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .r-badge {
    font-family: var(--font-mono);
    font-size: 10.5px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--fg-dim);
  }
  .r-item.top .r-badge {
    color: var(--gold);
    font-weight: 700;
  }
  .r-name {
    font-family: var(--font-display);
    font-weight: 700;
    font-size: 14px;
    color: var(--fg);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 128px;
  }
  .r-caster {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    color: var(--fg-muted);
  }
  .seat-dot {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--seat-color);
    flex: 0 0 auto;
  }
  .r-sub {
    font-size: 11.5px;
    color: var(--fg-muted);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 128px;
  }
  .r-targeted-by {
    font-size: 11px;
    color: var(--rose);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 128px;
  }
  .act {
    flex: 0 0 auto;
    height: 26px;
    padding: 0 10px;
    font-size: 11.5px;
    border-radius: 7px;
    align-self: center;
  }
  .act:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
  .counter-btn:hover:not(:disabled) {
    color: var(--danger);
    border-color: color-mix(in srgb, var(--danger) 50%, transparent);
  }
</style>
