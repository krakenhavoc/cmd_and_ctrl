<script lang="ts">
  // BotFeed renders what the bot seats have said — improvisation
  // disclosures always, per-move reasoning behind the "show bot
  // reasoning" setting. S31 sub-PR 8 / ADR 0033 §8.
  //
  // Why this exists as its own surface rather than in a chat panel:
  // S08.5 removed the chat UI (players coordinate on Discord), so the
  // transport is live but nothing renders it. An announcement nobody
  // can see is not an announcement, and the improvisation feature
  // rests on the table not being deceived — so the disclosure gets a
  // minimal read-only feed of its own in the board's attention strip.
  // #427 owns the rest of the bot-seat client surface (BOT chips on
  // PlayerHeader, the lobby picker); when a full chat panel returns,
  // this collapses into it.
  import { visibleBotLines, dismissable, type BotLine } from "../botChat";
  import type { ChatMessage } from "../ws";
  import { settings } from "../settings";
  import Icon from "./Icon.svelte";

  interface Props {
    chat: readonly ChatMessage[];
  }
  const { chat }: Props = $props();

  // Lines the viewer has closed. Only reasoning is closable — an
  // improvisation ages out on the feed's cap instead, because a
  // disclosure you can click away before reading is half a
  // disclosure.
  let dismissed = $state<Set<string>>(new Set());

  const lines = $derived(
    visibleBotLines(chat, $settings.gameplay.showBotReasoning).filter((l) => !dismissed.has(l.id)),
  );

  function dismiss(line: BotLine) {
    if (!dismissable(line)) return;
    const next = new Set(dismissed);
    next.add(line.id);
    dismissed = next;
  }
</script>

{#each lines as line (line.id)}
  <div
    class="bot-line"
    class:improvised={line.improvised}
    role={line.improvised ? "alert" : "status"}
    aria-live="polite"
  >
    <span class="label" class:gold={line.improvised}>
      <Icon name={line.improvised ? "spark" : "dot"} size={12} />
      {line.improvised ? "improvised" : "bot"}
    </span>
    <span class="text">
      <strong>{line.authorName}</strong>
      {line.text}
      {#if line.reason}
        <span class="muted">· {line.reason}</span>
      {/if}
    </span>
    {#if line.improvised}
      <span class="muted undo-hint">
        <Icon name="undo" size={12} /> any player can undo this
      </span>
    {:else}
      <button type="button" class="ghost close" onclick={() => dismiss(line)} aria-label="dismiss">
        <Icon name="x" size={12} />
      </button>
    {/if}
  </div>
{/each}

<style>
  /* Styled to match Game.svelte's attention strip (.att) without
     borrowing its class — Svelte scopes styles per component, so a
     shared class name would render unstyled here. Same tokens, same
     shape; if the strip is ever extracted into a component of its
     own, this collapses into it. */
  .bot-line {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 9px 10px 9px 14px;
    background: color-mix(in srgb, var(--surface) 94%, transparent);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-lg);
    color: var(--fg-muted);
    font-size: 13px;
    line-height: 1.35;
    box-sizing: border-box;
  }
  .bot-line.improvised {
    border-color: rgba(217, 180, 92, 0.45);
  }
  .label {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    font-weight: 700;
    color: var(--fg-dim);
    display: inline-flex;
    align-items: center;
    gap: 6px;
    flex: 0 0 auto;
    white-space: nowrap;
  }
  .label.gold {
    color: var(--gold-strong);
  }
  .text {
    flex: 1;
    min-width: 0;
  }
  .text strong {
    color: var(--fg);
    font-weight: 700;
  }
  .muted {
    color: var(--fg-dim);
  }
  .undo-hint {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    flex: 0 0 auto;
    white-space: nowrap;
    font-size: 11px;
  }
  .close {
    flex: 0 0 auto;
    width: 26px;
    height: 26px;
    padding: 0;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 7px;
  }
</style>
