<script lang="ts">
  // KeywordBadgeRow renders the S18 combat keyword badges along the
  // bottom edge of a card. Reads `abilities` from the wire
  // (`CardView.abilities`) which is the effective list — includes
  // printed keywords via `Spec.PrintedKeywords` AND granted
  // keywords via Layer 6 static abilities (Lord of Atlantis →
  // islandwalk on other Merfolk).
  //
  // Known keywords (the S18 set of 12) render as flat currentColor
  // SVG icons from keywordIcons.ts. Unknown tokens fall back to a
  // 3-letter text badge so a Lord of Atlantis "islandwalk" grant
  // stays readable as "ISL" until S26 tribal landswalk handling
  // lands. The tooltip carries the full keyword name in both cases.

  import { KEYWORD_ICONS } from "../../keywordIcons";

  interface Props {
    abilities?: string[];
  }

  const { abilities = [] }: Props = $props();

  const KEYWORD_LONG: Record<string, string> = {
    flying: "Flying",
    reach: "Reach",
    "first strike": "First strike",
    "double strike": "Double strike",
    deathtouch: "Deathtouch",
    lifelink: "Lifelink",
    trample: "Trample",
    vigilance: "Vigilance",
    menace: "Menace",
    defender: "Defender",
    haste: "Haste",
    flash: "Flash",
  };

  function labelFor(kw: string): string {
    return KEYWORD_LONG[kw] ?? kw;
  }

  function fallbackShort(kw: string): string {
    return kw.slice(0, 3).toUpperCase();
  }
</script>

{#if abilities.length > 0}
  <div class="keyword-row" aria-label="keywords">
    {#each abilities as kw (kw)}
      {@const long = labelFor(kw)}
      {@const icon = KEYWORD_ICONS[kw]}
      {#if icon}
        <span class="kw-badge kw-icon" title={long} aria-label={long}>
          <!-- eslint-disable-next-line svelte/no-at-html-tags -->
          {@html icon}
        </span>
      {:else}
        <span class="kw-badge kw-text" title={long} aria-label={long}>
          {fallbackShort(kw)}
        </span>
      {/if}
    {/each}
  </div>
{/if}

<style>
  .keyword-row {
    position: absolute;
    bottom: 2px;
    left: 2px;
    right: 2px;
    display: flex;
    flex-wrap: wrap;
    gap: 2px;
    justify-content: flex-start;
    pointer-events: none;
    z-index: 3;
  }
  .kw-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    background: rgba(0, 0, 0, 0.72);
    border: 1px solid rgba(255, 255, 255, 0.25);
    border-radius: 3px;
    backdrop-filter: blur(6px);
    -webkit-backdrop-filter: blur(6px);
    line-height: 1;
  }
  .kw-icon {
    width: 14px;
    height: 14px;
    padding: 1px;
  }
  .kw-icon :global(svg) {
    width: 100%;
    height: 100%;
    display: block;
  }
  .kw-text {
    padding: 1px 3px;
    font-size: 9px;
    font-weight: 700;
    letter-spacing: 0.04em;
    text-shadow: 0 1px 0 rgba(0, 0, 0, 0.6);
    min-width: 14px;
    text-align: center;
  }
</style>
