<script lang="ts">
  // SiteFooter — the unofficial-fan-content and Scryfall notice (#2191).
  // Mounted once by App.svelte for every route outside the table: the
  // game and the practice table never show it, because it must not take
  // table space. They reach the same words from the ⋯ menu's Help group
  // instead (#2982). Since #2982 it carries the Fan Content Policy's own
  // sentence, "unofficial Fan Content permitted under the Fan Content
  // Policy"; the sentence is the owner's call and was added on purpose.
  import LegalNotice from "./LegalNotice.svelte";
  import { FAN_CONTENT_POLICY_URL } from "../legalNotice";
  import { PRIVACY_HASH } from "../privacy";
</script>

<footer class="site-footer">
  <p class="line">
    cmd_and_ctrl is unofficial Fan Content permitted under the
    <a href={FAN_CONTENT_POLICY_URL} target="_blank" rel="noopener noreferrer">Fan Content Policy</a
    >. Not approved or endorsed by Wizards of the Coast. Card data and images from
    <a href="https://scryfall.com" target="_blank" rel="noopener noreferrer">Scryfall</a>.
  </p>
  <!-- #2983: the privacy notice is one click from every page outside the
       table; the table's own copy is in the ⋯ menu's Legal notice. -->
  <p class="line"><a href={PRIVACY_HASH}>Privacy and your data</a></p>
  <details class="legal">
    <summary aria-label="Legal notice">Legal</summary>
    <LegalNotice />
  </details>
</footer>

<style>
  .site-footer {
    margin-top: 2rem;
    padding-top: 1rem;
    border-top: 1px solid var(--border);
    color: var(--fg-muted);
    font-size: 0.75rem;
    line-height: 1.5;
    overflow-wrap: anywhere;
  }
  .site-footer :global(p) {
    margin: 0 0 0.5rem;
  }
  .site-footer :global(a) {
    color: var(--fg-muted);
    text-decoration: underline;
  }
  .site-footer :global(a:hover),
  .site-footer :global(a:focus-visible) {
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
