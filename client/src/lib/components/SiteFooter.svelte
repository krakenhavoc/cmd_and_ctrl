<script lang="ts">
  // SiteFooter — one row under every page outside the table (ADR 0156 §3
  // item 4): Discord, About, Privacy, Terms (#3358) and GitHub, then the Fan Content
  // line (#2191). Mounted once by App.svelte; the game and the practice
  // table never show it, because it must not take table space. They
  // reach the same words from the ⋯ menu's Help group instead (#2982).
  //
  // The Fan Content sentence and the Legal disclosure (ADR 0149) are the
  // owner's call and are unchanged. The Discord link comes first, and
  // only when the server names an invite (GET /community).
  import LegalNotice from "./LegalNotice.svelte";
  import Icon from "./Icon.svelte";
  import { FAN_CONTENT_POLICY_URL } from "../legalNotice";
  import { PRIVACY_HASH } from "../privacy";
  import { TERMS_HASH } from "../terms";
  import { FUNDING_HASH } from "../funding";
  import { ABOUT_HASH, GUIDE_HASH, REPO_URL } from "../about";
  import { community } from "../community";
  import Wordmark from "./Wordmark.svelte";
  import { seatBandStyle } from "../colors";

  const SEAT_BAND = seatBandStyle();
</script>

<footer class="site-footer" style={SEAT_BAND}>
  <div class="close">
    <a class="close-mark" href="#/" aria-label="CMD & CTRL home"><Wordmark size="bar" /></a>
    <p class="close-line" aria-hidden="true">Commander, over the wire.</p>
  </div>
  <div class="row">
    <nav class="links" aria-label="Site footer">
      {#if $community.invite_url}
        <a class="discord" href={$community.invite_url} target="_blank" rel="noopener noreferrer"
          ><Icon name="discord" size={14} /> Discord</a
        >
      {/if}
      <a href={GUIDE_HASH}>Guide</a>
      <a href={ABOUT_HASH}>About</a>
      <!-- #2983: the privacy notice is one click from every page outside
           the table; the table's own copy is in the ⋯ menu's Legal notice. -->
      <a href={PRIVACY_HASH}>Privacy</a>
      <a href={TERMS_HASH}>Terms</a>
      <a href={FUNDING_HASH}>Funding</a>
      <a href={REPO_URL} target="_blank" rel="noopener noreferrer">GitHub</a>
    </nav>
    <div class="fine">
      <p class="line">
        cmd_and_ctrl is unofficial Fan Content permitted under the
        <a href={FAN_CONTENT_POLICY_URL} target="_blank" rel="noopener noreferrer"
          >Fan Content Policy</a
        >. Not approved or endorsed by Wizards of the Coast. Card data and images from
        <a href="https://scryfall.com" target="_blank" rel="noopener noreferrer">Scryfall</a>.
      </p>
      <details class="legal">
        <summary aria-label="Legal notice">Legal</summary>
        <LegalNotice />
      </details>
    </div>
  </div>
</footer>

<style>
  .site-footer {
    margin-top: auto;
    color: var(--fg-dim);
    font-size: 12.5px;
    line-height: 1.6;
    background: var(--bg-1);
  }
  /* The four seats again, quieter than the header's: the page closes
     where it opened. */
  .site-footer::before {
    content: "";
    display: block;
    height: 2px;
    opacity: 0.7;
    background: linear-gradient(
      90deg,
      var(--s0) 0 25%,
      var(--s1) 25% 50%,
      var(--s2) 50% 75%,
      var(--s3) 75% 100%
    );
  }
  /* The close: the mark, then the name's promise set large and quiet. */
  .close {
    max-width: 1200px;
    margin: 0 auto;
    padding: 36px var(--page-gutter) 0;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 12px 32px;
    box-sizing: content-box;
  }
  .close-mark {
    display: inline-flex;
    font-size: 20px;
    text-decoration: none;
    border-radius: var(--radius);
  }
  .close-line {
    margin: 0;
    font-family: var(--font-display);
    font-size: clamp(26px, 3.4vw, 44px);
    font-weight: 800;
    line-height: 1;
    letter-spacing: -0.035em;
    color: color-mix(in srgb, var(--fg) 22%, transparent);
  }
  .row {
    max-width: 1200px;
    margin: 0 auto;
    padding: 22px var(--page-gutter) 30px;
    display: flex;
    flex-wrap: wrap;
    align-items: flex-start;
    gap: 10px 28px;
    box-sizing: content-box;
  }
  .links {
    display: flex;
    flex-wrap: wrap;
    gap: 4px 18px;
    white-space: nowrap;
  }
  .links a {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    color: var(--fg-muted);
    font-weight: 600;
    text-decoration: none;
  }
  .links a:hover,
  .links a:focus-visible {
    color: var(--fg);
  }
  .links a.discord :global(svg) {
    color: #aab3ff;
  }
  .fine {
    flex: 1 1 420px;
    min-width: 0;
    overflow-wrap: anywhere;
  }
  .fine :global(p) {
    margin: 0 0 0.4rem;
  }
  .fine :global(a) {
    color: var(--fg-muted);
    text-decoration: underline;
  }
  .fine :global(a:hover),
  .fine :global(a:focus-visible) {
    color: var(--fg);
  }
  summary {
    cursor: pointer;
    width: fit-content;
    text-decoration: underline;
  }
  summary:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
  .legal[open] summary {
    margin-bottom: 0.5rem;
  }
</style>
