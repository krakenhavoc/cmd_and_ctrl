<script lang="ts">
  // ManaSymbol — one mana symbol, drawn inline (#1438).
  //
  // The five colours and colourless get the familiar disc-and-glyph
  // shape (sun, drop, skull, flame, tree, the hollow colourless
  // diamond) in the conventional disc colours; anything else — a
  // generic number, X, a hybrid or snow symbol — is a grey disc with
  // its text on it, which is what a printed card does too.
  //
  // Hand-drawn SVG rather than Scryfall's symbol files: the server's
  // image proxy caches card art by card ID and has no route for
  // symbology, and hot-linking svgs.scryfall.io from the client would
  // be the only third-party request on the table. The glyphs are a few
  // simple shapes each, not art (AGENTS.md §8).
  //
  // Decorative by default (aria-hidden): every caller already says in
  // words what the symbol means — the pool pip's "2 green mana", the
  // picker button's "add green mana". Pass `label` to make the symbol
  // itself speak.
  import { MANA_SYMBOL_META, manaSymbolMeta } from "../../manaSymbol";

  interface Props {
    /** The symbol with its braces stripped: "G", "C", "2", "X". */
    symbol: string;
    /** Rendered diameter in CSS px. */
    size?: number;
    /** Accessible name. Omit to leave the symbol decorative. */
    label?: string;
  }

  const { symbol, size = 16, label }: Props = $props();

  const key = $derived(symbol.toUpperCase());
  const meta = $derived(manaSymbolMeta(key));
  const known = $derived(key in MANA_SYMBOL_META);
  // A long generic ("10", "W/U") shrinks to stay inside the disc.
  const textSize = $derived(key.length >= 3 ? 10 : key.length === 2 ? 14 : 19);
</script>

<svg
  class="mana-symbol"
  data-symbol={key}
  width={size}
  height={size}
  viewBox="0 0 32 32"
  role={label ? "img" : undefined}
  aria-label={label}
  aria-hidden={label ? undefined : "true"}
  focusable="false"
>
  <!-- The offset shadow disc under the face, as on a printed symbol. -->
  <circle cx="16.6" cy="17.2" r="15" fill="#0d0f0f" opacity="0.55" />
  <circle cx="16" cy="16" r="15" fill={meta.fill} />
  {#if known}
    <path d={meta.glyph} fill="#0d0f0f" fill-rule="evenodd" />
  {:else}
    <text
      x="16"
      y="16.5"
      text-anchor="middle"
      dominant-baseline="central"
      font-size={textSize}
      font-weight="800"
      font-family="ui-sans-serif, system-ui, sans-serif"
      fill="#0d0f0f">{key}</text
    >
  {/if}
</svg>

<style>
  .mana-symbol {
    display: inline-block;
    flex: 0 0 auto;
    vertical-align: middle;
    overflow: visible;
  }
</style>
