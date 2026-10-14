<script lang="ts">
  // About — what this project is, what the engine does for you and what
  // it leaves to you, where to send feedback, and the credits (#2981).
  //
  // Public on purpose (lib/signedInHome.ts): it is what a stranger from a
  // post reads before deciding to sign in. It holds no account data.
  //
  // EVERY STATEMENT HERE IS TRUE OF THE CODE, and the card count is not a
  // number written down here: it is read from the public GET /roadmap
  // (fetchRoadmap), the same figure the roadmap page shows. If that fetch
  // fails the page says nothing about a count rather than a stale one.
  // Legal text is LegalNotice.svelte, not a second copy.
  import { onMount } from "svelte";
  import Page from "../lib/components/Page.svelte";
  import LegalNotice from "../lib/components/LegalNotice.svelte";
  import { fetchRoadmap, type RoadmapCardCounts } from "../lib/roadmap";
  import { ISSUES_URL, NEVER_PAYWALLED_URL, REPO_URL } from "../lib/about";
  import { PRIVACY_HASH } from "../lib/privacy";
  import { L } from "../lib/labels";
  import { community } from "../lib/community";

  let cards = $state<RoadmapCardCounts | null>(null);

  onMount(() => {
    const ac = new AbortController();
    fetchRoadmap(ac.signal)
      .then((r) => (cards = r.counts.cards))
      .catch(() => {
        // The count is a nicety; the page reads fine without it.
      });
    return () => ac.abort();
  });
</script>

<Page
  eyebrow="About"
  title="What this is"
  lede="A private-table Magic: The Gathering Commander sandbox for four players, built as a hobby. It is in beta: expect rough edges, and expect things to change."
  width="prose"
>
  <div class="stack">
    <section aria-labelledby="free-title">
      <h2 id="free-title">Free and unofficial</h2>
      <ul class="points">
        <li>It is a free fan project. Nothing is sold, and there are no payments.</li>
        <li>
          That includes the AI bot seats: the ones that call a paid model must never be paywalled (<a
            href={NEVER_PAYWALLED_URL}
            target="_blank"
            rel="noopener noreferrer">design decision 0033</a
          >).
        </li>
        <li>It is not made, approved or endorsed by Wizards of the Coast.</li>
      </ul>
    </section>

    <section aria-labelledby="engine-title">
      <h2 id="engine-title">What the engine does, and what you do</h2>
      <ul class="points">
        <li>
          Cards the engine implements resolve on their own, and show an
          <span class="chip">AUTO</span> badge on the table.
          {#if cards}
            <strong data-testid="about-card-count">{cards.total}</strong> cards are in the catalogue right
            now.
          {/if}
        </li>
        <li>
          A card that is not implemented yet still plays: it goes on the stack, resolves with no
          effect, and you resolve it by hand, together, at the table. Some implemented cards leave
          one part to you, and the card's entry in the catalogue says which.
        </li>
        <li>
          See <a href="#/catalog">the catalogue</a> and <a href="#/roadmap">the roadmap</a> for what is
          built, partial and missing.
        </li>
        <li>
          <a href="#/decks">Check a deck</a> to see how much of it plays as printed before you sit down.
        </li>
      </ul>
    </section>

    <section aria-labelledby="limits-title">
      <h2 id="limits-title">Known limits</h2>
      <ul class="points">
        <li>Beta: games can hit bugs, and a table may need to be restarted.</li>
        <li>
          Not every card or rule is automated, so a table works best with players who are happy to
          sort out an odd interaction together.
        </li>
        <li>What is and is not built is on the roadmap, which is kept honest by tests.</li>
      </ul>
    </section>

    <section aria-labelledby="feedback-title">
      <h2 id="feedback-title">Feedback</h2>
      <ul class="points">
        <li>
          The bug button (“{L.reportABug}”), in a game's toolbar or, when you are signed in, at the
          top of every other page, lets you report a bug, suggest an improvement, or ask a question.
          It files an issue in the project's GitHub repository, so what you write is
          <strong>public</strong>
          (see <a href={PRIVACY_HASH}>Privacy and your data</a>).
        </li>
        <li>
          You can also open an issue yourself on
          <a href={ISSUES_URL} target="_blank" rel="noopener noreferrer">GitHub</a>.
        </li>
      </ul>
      {#if $community.invite_url}
        <p class="discord" data-testid="about-discord">
          <a
            class="btn discord"
            href={$community.invite_url}
            target="_blank"
            rel="noopener noreferrer">Join the C2 Discord</a
          >
          <span>Chat with the people who play here.</span>
        </p>
      {/if}
    </section>

    <section aria-labelledby="credits-title">
      <h2 id="credits-title">Credits</h2>
      <ul class="points">
        <li>
          Card data and card images come from
          <a href="https://scryfall.com" target="_blank" rel="noopener noreferrer">Scryfall</a>,
          which is not affiliated with this site.
        </li>
        <li>
          Card art belongs to its artists: where a card is shown as an art crop, the artist is
          credited on it.
        </li>
        <li>
          The source is open at
          <a href={REPO_URL} target="_blank" rel="noopener noreferrer">GitHub</a>.
        </li>
      </ul>
      <div class="legal"><LegalNotice /></div>
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
  /* The shared section head: the display face, as the roadmap's. app.css
     styles a bare h2 as a small mono label, so say what this one is. */
  h2 {
    margin: 0 0 10px;
    font-family: var(--font-display);
    font-size: 19px;
    font-weight: 800;
    letter-spacing: -0.005em;
    text-transform: none;
    color: var(--fg);
  }
  .discord {
    margin: 14px 0 0;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px 14px;
    color: var(--fg-muted);
    font-size: 13.5px;
  }
  .points {
    margin: 0;
    padding-left: 1.2em;
    display: flex;
    flex-direction: column;
    gap: 8px;
    max-width: 72ch;
    color: var(--fg-muted);
    font-size: 13.5px;
    line-height: 1.6;
  }
  .points a,
  .legal :global(a) {
    color: var(--accent);
  }
  .chip {
    font-family: var(--font-mono);
    font-size: 10.5px;
    padding: 1px 5px;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    color: var(--fg);
  }
  .legal {
    margin-top: 14px;
    max-width: 72ch;
    color: var(--fg-dim);
    font-size: 12px;
    line-height: 1.55;
  }
</style>
