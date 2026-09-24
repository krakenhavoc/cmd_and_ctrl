<script lang="ts">
  // StackLaneSpotlight — style B (#1467): the top stack item shown
  // large, with the rest of the stack and any pending triggers queued
  // beside it under "THEN". StackLaneHost already draws the header
  // (count, priority, split second, hold, Pass), the summary line and
  // the aria-live announcer — this component draws only the items,
  // reading everything (names, targets, casters, effect text) off the
  // shared model in lib/stackLane.ts so it can never say something the
  // model didn't derive honestly.

  import type { StackLaneItem, StackLaneStyle, StackLaneStyleProps } from "../../stackLane";
  import { cardImageURL } from "../../cardImage";
  import { cardArt } from "../../cardArt";
  import Icon from "../Icon.svelte";

  interface Props extends StackLaneStyleProps {
    styleName: StackLaneStyle;
  }

  const { model, controls, styleName }: Props = $props();

  const top = $derived(model.stackItems[0] ?? null);
  // Everything else, in order: the rest of the stack, then the
  // pending triggers last — they aren't on the stack yet.
  const queue = $derived([...model.stackItems.slice(1), ...model.pendingTriggers]);

  const nameByID = $derived(new Map(model.items.map((it) => [it.id, it.name] as const)));
  function targetedByNames(item: StackLaneItem): string {
    return item.targetedBy.map((id) => nameByID.get(id) ?? "something").join(", ");
  }

  /** The queue line: "→ targets" when there are any, else the effect text. */
  function queueLine(item: StackLaneItem): { text: string; kind: "targets" | "effect" | "none" } {
    if (item.targets.length > 0) {
      return { text: item.targets.map((t) => t.phrase).join(" / "), kind: "targets" };
    }
    if (item.effect) return { text: item.effect.text, kind: "effect" };
    return { text: "", kind: "none" };
  }

  function kindLabel(item: StackLaneItem): string | null {
    if (item.kind === "pending") return "waiting";
    if (item.kind === "triggered") return "trigger";
    if (item.kind === "activated") return "ability";
    return null;
  }

  function counterTitle(): string {
    return model.priority.viewerHolds ? "counter this item" : "you don't hold priority";
  }
</script>

<div class="spotlight" data-stack-body={styleName}>
  {#if top}
    {@const src = cardImageURL(top.artCard, "normal")}
    {@const targetable = controls.targetable(top)}
    <div class="hero">
      <!-- svelte-ignore a11y_no_noninteractive_tabindex -->
      <div
        class="art"
        class:cast-targetable={targetable}
        class:previewable={top.previewCard !== null}
        style:--seat-color={top.casterColor}
        data-stack-item-id={top.id}
        title={top.previewCard ? `${top.name} — hover to preview` : top.name}
        onpointerenter={() => controls.hoverEnter(top)}
        onpointerleave={() => controls.hoverLeave(top)}
        onfocusin={() => controls.hoverEnter(top)}
        onfocusout={() => controls.hoverLeave(top)}
        onclick={targetable ? () => controls.target(top) : undefined}
        onkeydown={(e) => {
          if (targetable && (e.key === "Enter" || e.key === " ")) {
            e.preventDefault();
            controls.target(top);
          }
        }}
        role={targetable ? "button" : undefined}
        tabindex={targetable ? 0 : undefined}
      >
        {#if src}
          <img {src} alt="" loading="lazy" decoding="async" use:cardArt={src} />
        {:else}
          <Icon name={top.kind === "triggered" ? "bolt" : "spark"} size={28} />
        {/if}
      </div>
      <div class="hero-info">
        <span class="next-chip">Resolves next</span>
        <h3 class="name">{top.name}</h3>
        <div class="caster" style:--seat-color={top.casterColor}>
          <span class="seat-dot"></span>
          <span class="caster-name">{top.casterIsViewer ? "You" : top.casterName}</span>
        </div>
        {#if top.effect}
          <p class="effect">{top.effect.text}</p>
        {/if}
        {#if top.targets.length > 0}
          <div class="target-chips">
            {#each top.targets as t (t.id)}
              <span class="target-chip"
                >{t.phrase}{#if t.stackPosition !== null}<span class="queue-num">
                    · queue #{t.stackPosition}</span
                  >{/if}</span
              >
            {/each}
          </div>
        {/if}
        <button
          type="button"
          class="act counter-btn"
          disabled={!model.priority.viewerHolds}
          aria-label={`Counter ${top.name}`}
          onclick={(e) => {
            e.stopPropagation();
            controls.counter(top);
          }}
          title={counterTitle()}
        >
          Counter
        </button>
      </div>
    </div>
  {/if}
  {#if queue.length > 0}
    <div class="queue">
      <span class="queue-label">then</span>
      <ol>
        {#each queue as item (item.id)}
          {@const src = cardImageURL(item.artCard, "small")}
          {@const targetable = controls.targetable(item)}
          {@const line = queueLine(item)}
          {@const kl = kindLabel(item)}
          <li class="q-line">
            <!-- svelte-ignore a11y_no_noninteractive_tabindex -->
            <div
              class="q-item"
              class:waiting={item.kind === "pending"}
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
              <span class="q-pos" aria-hidden="true">{item.position ?? "•"}</span>
              <div class="q-thumb">
                {#if src}
                  <img {src} alt="" loading="lazy" decoding="async" use:cardArt={src} />
                {:else}
                  <Icon name={item.kind === "triggered" ? "bolt" : "spark"} size={14} />
                {/if}
              </div>
              <div class="q-info">
                <div class="q-title">
                  <span class="q-name">{item.name}</span>
                  {#if kl}<span class="q-kind">{kl}</span>{/if}
                </div>
                <div class="q-sub">
                  <span class="seat-dot"></span>
                  <span class="caster-name">{item.casterIsViewer ? "You" : item.casterName}</span>
                  {#if line.kind === "targets"}
                    <span class="q-targets">→ {line.text}</span>
                  {:else if line.kind === "effect"}
                    <span class="q-effect">{line.text}</span>
                  {/if}
                </div>
                {#if item.targetedBy.length > 0}
                  <span class="targeted-by">targeted by {targetedByNames(item)}</span>
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
          </li>
        {/each}
      </ol>
    </div>
  {/if}
</div>

<style>
  .spotlight {
    display: flex;
    gap: 22px;
    align-items: flex-start;
    min-width: 0;
    min-height: 0;
    overflow-y: auto;
  }
  .hero {
    display: flex;
    gap: 18px;
    flex: 1 1 auto;
    min-width: 0;
  }
  .art {
    position: relative;
    --art-error-top: 4px;
    --art-error-right: 4px;
    flex: 0 0 auto;
    width: 150px;
    height: 210px;
    border-radius: var(--radius-lg);
    overflow: hidden;
    background: var(--surface-sunken);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--seat-color);
    box-shadow:
      0 0 0 3px var(--gold),
      var(--shadow-lg);
    transition:
      box-shadow 120ms var(--ease),
      transform 120ms var(--ease);
  }
  .art img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
  .art.previewable {
    cursor: zoom-in;
  }
  .art.cast-targetable {
    cursor: pointer;
    box-shadow:
      0 0 0 3px var(--gold),
      var(--ring-gold),
      var(--shadow-lg);
  }
  .art.cast-targetable:hover {
    transform: translateY(-2px);
  }
  .hero-info {
    flex: 1 1 auto;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .next-chip {
    align-self: flex-start;
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--accent-fg);
    background: var(--gold);
    border-radius: 999px;
    padding: 3px 10px;
    animation: spotlight-pulse 2.2s ease-in-out infinite;
  }
  @keyframes spotlight-pulse {
    0%,
    100% {
      opacity: 1;
    }
    50% {
      opacity: 0.72;
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .next-chip {
      animation: none;
    }
  }
  .name {
    margin: 0;
    font-family: var(--font-display);
    font-weight: 700;
    font-size: 21px;
    color: var(--fg);
    line-height: 1.15;
  }
  .caster {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
  }
  .caster-name {
    color: var(--fg);
    font-weight: 600;
  }
  .seat-dot {
    display: inline-block;
    width: 9px;
    height: 9px;
    border-radius: 50%;
    background: var(--seat-color);
    flex: 0 0 auto;
  }
  .effect {
    margin: 0;
    font-size: 13.5px;
    line-height: 1.35;
    color: var(--fg-muted);
  }
  .target-chips {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }
  .target-chip {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 5px 9px;
    border-radius: 8px;
    font-size: 12.5px;
    color: var(--fg);
    background: color-mix(in srgb, var(--rose) 14%, transparent);
    border: 1px solid color-mix(in srgb, var(--rose) 45%, transparent);
  }
  .queue-num {
    color: var(--gold-strong);
    font-family: var(--font-mono);
    font-size: 11px;
  }
  .act {
    flex: 0 0 auto;
    height: 32px;
    padding: 0 14px;
    font-size: 12px;
    border-radius: 7px;
    align-self: flex-start;
  }
  .act:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
  .counter-btn:hover:not(:disabled) {
    color: var(--danger);
    border-color: color-mix(in srgb, var(--danger) 50%, transparent);
  }

  .queue {
    flex: 1 1 220px;
    min-width: 200px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .queue-label {
    font-family: var(--font-mono);
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.14em;
    color: var(--fg-dim);
    font-weight: 700;
  }
  .queue ol {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .q-line {
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
  }
  .q-item {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 6px 10px 6px 6px;
    border-radius: 10px;
    background: var(--surface-raised);
    border: 1px solid var(--border);
    transition:
      border-color 120ms var(--ease),
      box-shadow 120ms var(--ease);
  }
  .q-item.waiting {
    border-style: dashed;
  }
  .q-item.previewable {
    cursor: zoom-in;
  }
  .q-item.previewable:hover {
    border-color: var(--border-strong);
    background: var(--surface-hover);
  }
  .q-item.cast-targetable {
    cursor: pointer;
    box-shadow: var(--ring-gold);
    border-color: var(--gold);
  }
  .q-pos {
    width: 16px;
    flex: 0 0 auto;
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--fg-dim);
    text-align: center;
    font-variant-numeric: tabular-nums;
  }
  .q-thumb {
    position: relative;
    --art-error-top: 2px;
    --art-error-right: 2px;
    width: 56px;
    height: 78px;
    border-radius: 5px;
    overflow: hidden;
    flex: 0 0 auto;
    background: var(--surface-sunken);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--seat-color);
  }
  .q-thumb img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
  .q-info {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 3px;
  }
  .q-title {
    display: flex;
    align-items: baseline;
    gap: 6px;
    min-width: 0;
  }
  .q-name {
    font-family: var(--font-display);
    font-weight: 700;
    font-size: 13.5px;
    color: var(--fg);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .q-kind {
    font-family: var(--font-mono);
    font-size: 9px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--fg-dim);
    flex: 0 0 auto;
  }
  .q-sub {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 5px;
    font-size: 11.5px;
    color: var(--fg-muted);
    min-width: 0;
  }
  .q-sub .seat-dot {
    width: 7px;
    height: 7px;
  }
  .targeted-by {
    font-size: 11px;
    color: var(--rose);
  }
</style>
