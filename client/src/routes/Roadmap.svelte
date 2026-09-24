<script lang="ts">
  // Roadmap — stub for #/roadmap (#1386, PR 3 of the site-portal plan).
  //
  // This route and its SiteHeader placement land here so the nav has
  // somewhere to point; the real page — GET /roadmap, implemented
  // mechanics with example cards, partial ones with what's missing,
  // missing seams with the cards waiting on them, and an "up next"
  // strip — is PR 2 of the same plan and replaces this file wholesale.
  // Public on purpose, same as the page it will become: no card art,
  // no session.
  import SiteHeader from "../lib/components/SiteHeader.svelte";
</script>

<SiteHeader />

<section class="stub">
  <div class="head">
    <p class="eyebrow">What the engine plays, and what's next</p>
    <h1>roadmap</h1>
    <p class="lede">The roadmap is on its way.</p>
  </div>
</section>

<style>
  .stub {
    width: min(720px, 100%);
    margin: 40px auto 0;
    text-align: center;
  }
  .eyebrow {
    margin: 0 0 6px;
    font-family: var(--font-mono);
    font-size: 10.5px;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--gold-strong);
  }
  h1 {
    margin: 0;
    font-family: var(--font-display);
    font-weight: 800;
    letter-spacing: -0.01em;
  }
  .lede {
    margin: 10px 0 0;
    color: var(--fg-muted);
    font-size: 14px;
    line-height: 1.55;
  }
</style>
