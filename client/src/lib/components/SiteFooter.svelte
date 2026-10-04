<script lang="ts">
  // SiteFooter — the unofficial-fan-content and Scryfall notice (#2191).
  // Mounted once by App.svelte for every route outside the table: the
  // game and the practice table never show it, because it must not take
  // table space. The wording deliberately does not claim the Fan Content
  // Policy covers a playable table; do not add that claim here.
</script>

<footer class="site-footer">
  <p class="line">
    Unofficial fan content, not approved or endorsed by Wizards of the Coast. Card data and images
    from <a href="https://scryfall.com" target="_blank" rel="noopener noreferrer">Scryfall</a>.
  </p>
  <details class="legal">
    <summary aria-label="Legal notice">Legal</summary>
    <p>
      cmd_and_ctrl is unofficial fan content. It is not approved or endorsed by Wizards of the
      Coast. Portions of the materials used are property of Wizards of the Coast. ©Wizards of the
      Coast LLC. Magic: The Gathering is a trademark of Wizards of the Coast LLC. Card data and card
      images are provided by Scryfall, which is not affiliated with this site.
    </p>
    <p>
      <a
        href="https://company.wizards.com/en/legal/fancontentpolicy"
        target="_blank"
        rel="noopener noreferrer">Wizards Fan Content Policy</a
      >
    </p>
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
  .site-footer p {
    margin: 0 0 0.5rem;
  }
  .site-footer a {
    color: var(--fg-muted);
    text-decoration: underline;
  }
  .site-footer a:hover,
  .site-footer a:focus-visible {
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
