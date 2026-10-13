<script lang="ts">
  // Page — the one frame every non-game page sits in (ADR 0156 §3).
  //
  // The site header, then a centred column with the page's head: an
  // optional eyebrow, a title (sentence case, the display face, one size
  // everywhere), an optional lede, and an actions slot on the right. The
  // page's first-use tip is drawn under the lede (PageTipSlot), never
  // over the page's controls. The body follows.
  //
  // Two widths: "wide" (1200px) for tables and grids, "prose" (72ch) for
  // reading. One breakpoint: the design target is a tablet (820px and
  // up); below it the head wraps, and nothing gets a layout of its own.
  //
  // The footer is not here: App.svelte draws it once, below every page.

  import type { Snippet } from "svelte";
  import SiteHeader from "./SiteHeader.svelte";
  import PageTipSlot from "./hints/PageTipSlot.svelte";

  interface Props {
    /** The small line above the title. */
    eyebrow?: string;
    /** The h1: text, or a snippet for a title with markup of its own. */
    title: string | Snippet;
    /** The h1's accessible name, when it must differ from its text (a contract label). */
    titleLabel?: string;
    /** The line under the title: text, or a snippet of inline content. */
    lede?: string | Snippet;
    /** Anything else that belongs in the head, under the lede (chips, a contact line). */
    meta?: Snippet;
    /** Buttons on the head's right. */
    actions?: Snippet;
    width?: "wide" | "prose";
    children: Snippet;
  }

  const {
    eyebrow,
    title,
    titleLabel,
    lede,
    meta,
    actions,
    width = "wide",
    children,
  }: Props = $props();
</script>

<SiteHeader />

<main class="site-page {width}">
  <header class="page-head">
    <div class="page-head-main">
      {#if eyebrow}<p class="page-eyebrow">{eyebrow}</p>{/if}
      <!-- prettier-ignore -->
      <h1 class="page-title" aria-label={titleLabel}>{#if typeof title === "string"}{title}{:else}{@render title()}{/if}</h1>
      {#if lede}
        <!-- prettier-ignore -->
        <p class="page-lede">{#if typeof lede === "string"}{lede}{:else}{@render lede()}{/if}</p>
      {/if}
      {@render meta?.()}
      <PageTipSlot />
    </div>
    {#if actions}
      <div class="page-actions">{@render actions()}</div>
    {/if}
  </header>

  {@render children()}
</main>

<style>
  .site-page {
    flex: 1 0 auto;
    width: 100%;
    margin: 0 auto;
    padding: 40px var(--page-gutter) 56px;
    box-sizing: border-box;
    min-width: 0;
  }
  .site-page.wide {
    max-width: calc(1200px + 2 * var(--page-gutter));
  }
  .site-page.prose {
    max-width: calc(72ch + 2 * var(--page-gutter));
  }

  .page-head {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-end;
    justify-content: space-between;
    gap: 16px 24px;
    margin-bottom: 28px;
  }
  .page-head-main {
    flex: 1 1 420px;
    min-width: 0;
  }
  .page-eyebrow {
    margin: 0 0 8px;
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--accent);
  }
  .page-title {
    margin: 0;
    font-family: var(--font-display);
    font-size: 34px;
    font-weight: 700;
    line-height: 1.12;
    letter-spacing: -0.02em;
    text-transform: none;
    color: var(--fg);
    overflow-wrap: anywhere;
  }
  .page-lede {
    margin: 10px 0 0;
    max-width: 68ch;
    color: var(--fg-muted);
    font-size: 15px;
    line-height: 1.55;
  }
  .page-actions {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 10px;
  }
</style>
