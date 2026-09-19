<script lang="ts">
  // KeywordBadgeRow renders the S18 combat keyword badges along the
  // bottom edge of a card. Reads `abilities` from the wire
  // (`CardView.abilities`) which is the effective list — includes
  // printed keywords via `Spec.PrintedKeywords` AND granted
  // keywords via Layer 6 static abilities (Lord of Atlantis →
  // islandwalk on other Merfolk).
  //
  // Known keywords (the S18 set of 12 plus the S23 targeting pair,
  // hexproof and shroud) render as flat currentColor
  // SVG icons from keywordIcons.ts. Unknown tokens fall back to a
  // 3-letter text badge so a Lord of Atlantis "islandwalk" grant
  // stays readable as "ISL" until S26 tribal landswalk handling
  // lands. The tooltip carries the full keyword name in both cases.
  //
  // #781 adds the chosen-value chips to the same row. A chosen colour
  // and a named creature type are not keywords, but they read the same
  // way — a short public fact about this permanent that a player has
  // to be able to see at a glance — and putting them anywhere else on
  // the tile would mean a second row competing for the same two
  // pixels. They render first, in words rather than icons, because
  // "Elf" cannot be abbreviated into a glyph anyone would recognise.
  //
  // PROTECTION (#662) is the third kind of chip here, and the one
  // keyword whose token carries a PARAMETER. It does not ride
  // `abilities` with the rest: the badge abbreviates the QUALITY
  // ("DEM" for Demons, "ALL" for everything) and the tooltip names it
  // in full. The parse comes from the server on `card.protection` —
  // the engine has exactly one protection grammar and the client is
  // not a second copy of it. The raw tokens are filtered out of
  // `abilities` so the same ability is not badged twice.
  //
  // Order on the row: chosen values, then protections, then the
  // ordinary keyword icons. Words before glyphs, and the two facts a
  // player most often has to check before pointing a spell at the
  // permanent come first.

  import { chosenValueChips } from "../../chosenValues";
  import { KEYWORD_ICONS } from "../../keywordIcons";
  import type { ProtectionView } from "../../protocol";

  interface Props {
    abilities?: string[];
    // #781, CR 105.4 / CR 614.12. Passed through rather than read off
    // a CardView so this component keeps taking only what it renders.
    chosenColor?: string;
    namedTribe?: string;
    // #662, CR 702.16. The qualities the SERVER parsed; the client
    // owns no protection grammar.
    protection?: ProtectionView[];
  }

  const { abilities = [], chosenColor, namedTribe, protection = [] }: Props = $props();

  const chosen = $derived(chosenValueChips({ chosen_color: chosenColor, named_tribe: namedTribe }));

  // CR 702.16m: a permanent can have the same protection twice (two
  // Swords of Fire and Ice on one creature). It is ONE quality for
  // every rules check, and two identical badges would also be a
  // duplicate `{#each}` key, which Svelte 5 throws on. Deduped here
  // rather than server-side, because the ability list is the
  // engine's truth and the row is a presentation of it.
  const plain = $derived([
    ...new Set((abilities ?? []).filter((kw) => !kw.toLowerCase().startsWith("protection from "))),
  ]);
  const protections = $derived([
    ...new Map((protection ?? []).map((p) => [p.printed.toLowerCase(), p])).values(),
  ]);
  const anyBadges = $derived(plain.length > 0 || protections.length > 0 || chosen.length > 0);

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
    hexproof: "Hexproof",
    shroud: "Shroud",
  };

  function labelFor(kw: string): string {
    return KEYWORD_LONG[kw] ?? kw;
  }

  function fallbackShort(kw: string): string {
    return kw.slice(0, 3).toUpperCase();
  }

  // "Protection from Demons" — the tooltip a player reads. The
  // quality keeps the card's own spelling and plural, which is why
  // the server's token does too.
  function protectionLabel(p: ProtectionView): string {
    return `Protection from ${p.printed}`;
  }

  // The badge face. Three-letter abbreviations of the printed quality,
  // except where the first three letters say nothing: "everything"
  // reads better as ALL than as EVE, and "the chosen player" (CR
  // 702.16k, #980) would abbreviate to THE. Both are the qualities that
  // are not characteristics of the source at all, which is why they are
  // the two that need naming rather than truncating.
  function protectionShort(p: ProtectionView): string {
    if (p.kind === "everything") return "ALL";
    if (p.kind === "player") return "PLR";
    return fallbackShort(p.printed);
  }
</script>

{#if anyBadges}
  <div class="keyword-row" aria-label="keywords">
    {#each chosen as chip (chip.kind)}
      <span
        class="kw-badge kw-text kw-chosen kw-chosen-{chip.kind}"
        title={chip.title}
        aria-label={chip.title}
      >
        {chip.label}
      </span>
    {/each}
    {#each protections as p (p.printed)}
      {@const long = protectionLabel(p)}
      <span class="kw-badge kw-text kw-protection" title={long} aria-label={long}>
        {protectionShort(p)}
      </span>
    {/each}
    {#each plain as kw (kw)}
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
  .kw-protection {
    /* Protection is the only badge that is a shield rather than an
       ability the creature uses, so it reads as a different thing. */
    background: rgba(24, 48, 92, 0.85);
    border-color: rgba(160, 200, 255, 0.45);
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
  /* #781: a chosen value is a full word, not a three-letter keyword
     abbreviation, so it gets a lighter border and its own tint to
     read as a different kind of fact from the badges beside it. */
  .kw-chosen {
    letter-spacing: 0;
    text-transform: none;
    border-color: rgba(255, 255, 255, 0.45);
  }
  .kw-chosen-color {
    background: rgba(32, 58, 40, 0.82);
  }
  .kw-chosen-tribe {
    background: rgba(28, 40, 66, 0.82);
  }
</style>
