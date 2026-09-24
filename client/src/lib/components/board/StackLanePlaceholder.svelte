<script lang="ts">
  // StackLanePlaceholder — the stand-in body for the three floating
  // stack styles (#1467) until their own designs land: spotlight and
  // ribbon in the second PR, fan in the third. Each of those replaces
  // its entry in StackLaneHost's STYLE_BODIES and takes exactly these
  // props, so swapping one in touches nothing else.
  //
  // It is deliberately plain — a numbered, top-first list read off the
  // shared model — so the setting works end to end today and every
  // wired control (counter, target, hover preview) can be exercised in
  // a real game before any design is built on top of it.

  import type { StackLaneStyle, StackLaneStyleProps } from "../../stackLane";
  import { cardImageURL } from "../../cardImage";
  import { cardArt } from "../../cardArt";
  import Icon from "../Icon.svelte";

  interface Props extends StackLaneStyleProps {
    styleName: StackLaneStyle;
  }

  const { model, controls, styleName }: Props = $props();
</script>

<div class="body" data-stack-body={styleName}>
  {#if model.stackItems.length > 0}
    <ol class="list">
      {#each model.stackItems as item (item.id)}
        {@const src = cardImageURL(item.artCard, "small")}
        {@const targetable = controls.targetable(item)}
        <li class="line">
          <!-- svelte-ignore a11y_no_noninteractive_tabindex -->
          <div
            class="item"
            class:top={item.isTop}
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
            <span class="pos" aria-hidden="true">{item.position}</span>
            <div class="thumb">
              {#if src}
                <img {src} alt="" loading="lazy" decoding="async" use:cardArt={src} />
              {:else}
                <Icon name={item.kind === "triggered" ? "bolt" : "spark"} size={16} />
              {/if}
            </div>
            <div class="info">
              <div class="title">
                {#if item.isTop}<span class="next">next</span>{/if}
                <span class="name">{item.name}</span>
              </div>
              <div class="sub">
                <span class="seat-dot"></span>
                <span class="caster">{item.casterIsViewer ? "You" : item.casterName}</span>
                {#if item.effect && item.kind === "spell"}
                  <span class="effect">· {item.effect.text}</span>
                {/if}
                {#if item.targets.length > 0}
                  <span class="targets">→ {item.targets.map((t) => t.phrase).join(" / ")}</span>
                {/if}
                {#each item.chips as chip, ci (ci)}
                  <span
                    class="chip"
                    class:flag={chip.tone === "flag"}
                    class:manual={chip.tone === "manual"}
                    title={chip.title}>{chip.label}</span
                  >
                {/each}
              </div>
            </div>
          </div>
          <button
            type="button"
            class="act counter-btn"
            disabled={!model.priority.viewerHolds}
            onclick={(e) => {
              e.stopPropagation();
              controls.counter(item);
            }}
            title={model.priority.viewerHolds ? "counter this item" : "you don't hold priority"}
          >
            Counter
          </button>
        </li>
      {/each}
    </ol>
  {/if}
  {#if model.pendingTriggers.length > 0}
    <div class="pending">
      <span class="label">waiting to go on the stack</span>
      <ul>
        {#each model.pendingTriggers as t (t.id)}
          <li style:--seat-color={t.casterColor} title={`from ${t.casterName}`}>
            <span class="seat-dot"></span>{t.name}
          </li>
        {/each}
      </ul>
    </div>
  {/if}
</div>

<style>
  .body {
    display: flex;
    flex-direction: column;
    gap: 8px;
    min-height: 0;
    overflow-y: auto;
  }
  .list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
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
    padding: 5px 10px 5px 6px;
    border-radius: 10px;
    background: var(--surface-raised);
    border: 1px solid var(--border);
    transition:
      border-color 120ms var(--ease),
      box-shadow 120ms var(--ease);
  }
  .item.top {
    border-color: var(--gold);
    box-shadow: var(--ring-gold);
  }
  .item.previewable {
    cursor: zoom-in;
  }
  .item.previewable:hover {
    border-color: var(--border-strong);
    background: var(--surface-hover);
  }
  .item.cast-targetable {
    cursor: pointer;
    box-shadow: var(--ring-gold);
    border-color: var(--gold);
  }
  .pos {
    width: 16px;
    flex: 0 0 auto;
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--fg-dim);
    text-align: center;
    font-variant-numeric: tabular-nums;
  }
  .item.top .pos {
    color: var(--gold);
  }
  .thumb {
    position: relative;
    --art-error-top: 2px;
    --art-error-right: 2px;
    width: 46px;
    height: 64px;
    border-radius: 4px;
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
    gap: 3px;
  }
  .title {
    display: flex;
    align-items: baseline;
    gap: 8px;
    min-width: 0;
  }
  .name {
    font-weight: 700;
    font-size: 14px;
    color: var(--fg);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .next {
    font-family: var(--font-mono);
    font-size: 9px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    font-weight: 700;
    color: var(--gold);
    flex: 0 0 auto;
  }
  .sub {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 5px;
    font-size: 12px;
    color: var(--fg-muted);
    min-width: 0;
  }
  .seat-dot {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--seat-color);
    flex: 0 0 auto;
    margin-right: 5px;
  }
  .sub .seat-dot {
    margin-right: 0;
  }
  .caster {
    color: var(--fg);
    font-weight: 600;
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
  }
  .chip.flag {
    color: var(--gold-strong);
    border-color: color-mix(in srgb, var(--gold) 45%, transparent);
  }
  .chip.manual {
    border-style: dashed;
    border-color: var(--border-strong);
  }
  .act {
    flex: 0 0 auto;
    height: 30px;
    padding: 0 10px;
    font-size: 12px;
    border-radius: 7px;
  }
  .act:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
  .counter-btn:hover:not(:disabled) {
    color: var(--danger);
    border-color: color-mix(in srgb, var(--danger) 50%, transparent);
  }
  .pending {
    padding-top: 6px;
    border-top: 1px solid var(--border);
  }
  .pending .label {
    font-family: var(--font-mono);
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.14em;
    color: var(--fg-dim);
    font-weight: 700;
  }
  .pending ul {
    list-style: none;
    margin: 6px 0 0;
    padding: 0;
    display: flex;
    flex-wrap: wrap;
    gap: 4px 14px;
    font-size: 12px;
    color: var(--fg-muted);
  }
</style>
