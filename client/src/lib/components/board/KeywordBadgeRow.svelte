<script lang="ts">
  // KeywordBadgeRow renders the S18 combat keyword badges along the
  // bottom edge of a card. Reads `abilities` from the wire
  // (`CardView.abilities`) which is the effective list — includes
  // printed keywords via `Spec.PrintedKeywords` AND granted
  // keywords via Layer 6 static abilities (Lord of Atlantis →
  // islandwalk on other Merfolk).
  //
  // Canonical lowercase tokens come from the server; this
  // component maps them to short labels + readable titles so the
  // tiny battlefield thumbnails stay legible without sacrificing
  // the full keyword name in the hover tooltip.
  //
  // Keywords not in KEYWORD_META are rendered with their raw token
  // truncated to 3 characters — this handles unknown tokens
  // gracefully (e.g. "islandwalk" grants from Lord of Atlantis
  // show as "ISL" until S26 tribal gives them proper handling).

  interface Props {
    abilities?: string[];
  }

  const { abilities = [] }: Props = $props();

  const KEYWORD_META: Record<string, { short: string; long: string }> = {
    flying: { short: "FLY", long: "Flying" },
    reach: { short: "RCH", long: "Reach" },
    "first strike": { short: "FS", long: "First strike" },
    "double strike": { short: "DS", long: "Double strike" },
    deathtouch: { short: "DT", long: "Deathtouch" },
    lifelink: { short: "LL", long: "Lifelink" },
    trample: { short: "TR", long: "Trample" },
    vigilance: { short: "VIG", long: "Vigilance" },
    menace: { short: "MEN", long: "Menace" },
    defender: { short: "DEF", long: "Defender" },
    haste: { short: "HST", long: "Haste" },
    flash: { short: "FLS", long: "Flash" },
  };

  function labelFor(kw: string): { short: string; long: string } {
    const meta = KEYWORD_META[kw];
    if (meta) return meta;
    // Unknown keyword — take first 3 chars uppercased for the short
    // label, full token for the tooltip. Handles Lord of Atlantis's
    // "islandwalk" grant until S26 tribal landswalk handling.
    return { short: kw.slice(0, 3).toUpperCase(), long: kw };
  }
</script>

{#if abilities.length > 0}
  <div class="keyword-row" aria-label="keywords">
    {#each abilities as kw (kw)}
      {@const meta = labelFor(kw)}
      <span class="kw-badge" title={meta.long} aria-label={meta.long}>
        {meta.short}
      </span>
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
    padding: 1px 4px;
    font-size: 9px;
    font-weight: 700;
    letter-spacing: 0.04em;
    color: #fff;
    background: rgba(0, 0, 0, 0.72);
    border: 1px solid rgba(255, 255, 255, 0.25);
    border-radius: 4px;
    backdrop-filter: blur(6px);
    -webkit-backdrop-filter: blur(6px);
    text-shadow: 0 1px 0 rgba(0, 0, 0, 0.6);
    line-height: 1;
    min-width: 16px;
    text-align: center;
  }
</style>
