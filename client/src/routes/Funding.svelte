<script lang="ts">
  // Funding — "How CMD & CTRL is funded" (#3558): free to play, what
  // running it costs in plain words (no numbers), the ways to help, that
  // support unlocks nothing, and who runs it.
  //
  // Public on purpose (lib/signedInHome.ts), like Terms and Privacy: a
  // stranger can read it before signing in. A voluntary thank-you jar,
  // nothing more: keep it that way, and promise nothing about the future.
  // The operator and contact come from lib/legalNotice.ts. The ways to
  // help are GitHub Sponsors, the optional Ko-fi and Patreon links the
  // server configures (lib/supportLinks.ts, GET /config "support"), and
  // the shop links of ADR 0179 with their commission note.
  import Page from "../lib/components/Page.svelte";
  import { appConfig } from "../lib/env";
  import { FUNDING_TITLE } from "../lib/funding";
  import { LEGAL_CONTACT_EMAIL, LEGAL_CONTACT_URL, OPERATOR_NAME } from "../lib/legalNotice";
  import { SHOP_DISCLOSURE } from "../lib/shopLinks";
  import { configuredShops } from "../lib/shopLinksStore";
  import { GITHUB_SPONSORS_URL, supportLinks } from "../lib/supportLinks";
  import { PRIVACY_HASH } from "../lib/privacy";
  import { TERMS_HASH } from "../lib/terms";

  const links = $derived(supportLinks($appConfig.support));
  const shopNames = $derived($configuredShops.map((s) => s.name).join(" and "));
</script>

<Page
  eyebrow="Funding"
  title={FUNDING_TITLE}
  lede="CMD & CTRL is free to play, and playing never requires paying. This page says what it costs to keep running and how you can help if you want to."
  width="prose"
>
  <div class="stack">
    <section aria-labelledby="free-title">
      <h2 id="free-title">Free to play</h2>
      <p>
        Playing CMD & CTRL costs nothing. You never need to pay or give anything to play, and nobody
        is required to.
      </p>
    </section>

    <section aria-labelledby="costs-title">
      <h2 id="costs-title">What it costs to run</h2>
      <p>
        CMD & CTRL runs on servers that have to stay online, and that is the main cost. The AI seats
        think on the server’s own model, which costs something every time one plays. There is also
        the time it takes to build and look after it.
      </p>
    </section>

    <section aria-labelledby="help-title">
      <h2 id="help-title">Ways to help</h2>
      <p>If you enjoy it and want to chip in, these are the ways. All of them are optional.</p>
      <ul class="ways">
        <li>
          <a href={GITHUB_SPONSORS_URL} target="_blank" rel="noopener noreferrer">GitHub Sponsors</a
          >
        </li>
        {#each links as l (l.id)}
          <li>
            <a href={l.href} target="_blank" rel="noopener noreferrer">{l.name}</a>
          </li>
        {/each}
      </ul>
      <p>
        Each of these opens a third party’s own site, under its own terms and privacy policy, and
        what you give there is between you and them.
      </p>
      {#if $configuredShops.length > 0}
        <p class="shops">
          You can also buy the real cards at {shopNames} from the <a href="#/decks">Decks</a> and
          <a href="#/catalog">Catalog</a>
          pages. {SHOP_DISCLOSURE}
        </p>
      {/if}
    </section>

    <section aria-labelledby="thanks-title">
      <h2 id="thanks-title">A thank-you, nothing more</h2>
      <p>
        Support is a thank-you. It unlocks nothing in the game: no cards, seats, features or
        advantages, and everyone gets the same CMD & CTRL whether they help or not.
      </p>
    </section>

    <section aria-labelledby="who-title">
      <h2 id="who-title">Who runs it</h2>
      <p>
        CMD & CTRL is run by {OPERATOR_NAME}. Questions about any of this, or about the
        <a href={TERMS_HASH}>terms</a> and the <a href={PRIVACY_HASH}>privacy notice</a>, go to
        <a href={LEGAL_CONTACT_URL}>{LEGAL_CONTACT_EMAIL}</a>.
      </p>
    </section>
  </div>
</Page>

<style>
  .stack {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 24px;
  }
  h2 {
    margin: 0 0 8px;
    font-family: var(--font-display);
    font-size: 19px;
    font-weight: 800;
    letter-spacing: -0.005em;
    text-transform: none;
    color: var(--fg);
  }
  p,
  li {
    max-width: 72ch;
    color: var(--fg-muted);
    font-size: 14px;
    line-height: 1.6;
    overflow-wrap: anywhere;
  }
  p {
    margin: 0 0 8px;
  }
  ul {
    margin: 0 0 8px;
    padding-left: 1.4em;
  }
  li + li {
    margin-top: 4px;
  }
  a {
    color: var(--accent);
  }
</style>
