<script lang="ts">
  // Guide — getting started (#/guide): how to get in, find a pod, set up a
  // table, what to do at it, decks, and where things are on the C2
  // Discord. The same guide is posted in the Discord's #welcome channel;
  // keep the two saying the same things.
  //
  // Public on purpose (lib/signedInHome.ts), like About: it is written for
  // someone who has not signed in yet. Every button it names is a real
  // label in the app; when one is renamed, rename it here too.
  //
  // The Discord section, and every Discord button, shows only when the
  // server names an invite (GET /community), as on About and the footer.
  import Page from "../lib/components/Page.svelte";
  import Icon from "../lib/components/Icon.svelte";
  import { ABOUT_HASH } from "../lib/about";
  import { L } from "../lib/labels";
  import { community } from "../lib/community";

  const invite = $derived($community.invite_url);
</script>

<Page
  eyebrow="Guide"
  title="Getting started"
  lede="CMD & CTRL is a four-player Magic: The Gathering Commander table in your browser. Here's how to get in, find a pod and play your first game."
  width="prose"
>
  <div class="stack">
    <section aria-labelledby="what-title">
      <h2 id="what-title">What it is</h2>
      <ul class="points">
        <li>
          A free hobby project, built for a small group of friends. Nothing to install, nothing to
          buy.
        </li>
        <li>
          The game runs the rules for you: priority, the stack, combat, triggers, commander damage.
          Cards the engine knows resolve on their own and show an <span class="chip">AUTO</span> badge.
        </li>
        <li>
          A card the engine doesn't know yet still plays: you resolve it by hand, together, as you
          would at a real table.
        </li>
        <li>
          It's in beta, so expect rough edges and frequent changes. <a href={ABOUT_HASH}>About</a>
          has the details.
        </li>
      </ul>
    </section>

    <section aria-labelledby="in-title">
      <h2 id="in-title">Getting in</h2>
      <p class="para">You can play without an account:</p>
      <ul class="points">
        <li>
          A <a href="#/practice">practice game</a> seats you against a bot, no sign-in needed.
        </li>
        <li>A host's invite link seats you at their table as a guest.</li>
      </ul>
      <p class="para">
        To <strong>host your own tables and save your decks</strong>, you need an account. Accounts
        are for members of the C2 Discord with the <strong>Elder Dragon Highlander</strong> role:
      </p>
      <ol class="points">
        <li>
          Join the C2 Discord. Everyone who joins is a <strong>Planeswalker</strong>.
        </li>
        <li>
          Answer <strong>yes</strong> to “request account access” in the server's onboarding.
          Already answered no? Change it under <strong>Channels &amp; Roles</strong>.
        </li>
        <li>Say hi in <strong>#account-requests</strong>, and a mod gives you the role.</li>
        <li>Sign in here with Discord.</li>
      </ol>
      {#if invite}
        <p class="actions" data-testid="guide-discord">
          <a class="btn discord" href={invite} target="_blank" rel="noopener noreferrer"
            ><Icon name="discord" size={15} /> Join the C2 Discord</a
          >
          <a class="btn ghost" href="#/practice"
            ><Icon name="robot" size={15} /> Try a practice game</a
          >
        </p>
      {/if}
    </section>

    <section aria-labelledby="play-title">
      <h2 id="play-title">Playing a game</h2>
      <dl class="terms">
        <dt>Find a pod</dt>
        <dd>
          Ask in <strong>#looking-for-game</strong> on the Discord. Commander is best at four, but any
          empty seat can be filled with a bot.
        </dd>
        <dt>Host a table</dt>
        <dd>
          <strong>Tables → + New table</strong>, then <strong>Copy invite</strong> and send the link
          to your pod. Tick <strong>Create a voice room</strong> to give the table its own private voice
          channel on the Discord, open only to its players. It goes away 10 minutes after the game ends.
        </dd>
        <dt>Join a table</dt>
        <dd>Open the invite link. You don't need an account to sit as a guest.</dd>
        <dt>Fill a seat with a bot</dt>
        <dd>
          <strong>Add bot</strong> on an open seat, then pick how strong it plays and which deck it brings.
        </dd>
        <dt>Choose your deck</dt>
        <dd>
          Pick a pre-built deck, one of your saved decks, or paste an Archidekt link or a plain
          decklist. On Moxfield, open the deck, then Export and Copy plain text, and paste it here.
          Every deck is checked against the Commander rules: 100 cards, singleton, colour identity.
        </dd>
        <dt>Start</dt>
        <dd>Once every seat has a deck, the host presses <strong>Start</strong>.</dd>
      </dl>
    </section>

    <section aria-labelledby="table-title">
      <h2 id="table-title">At the table</h2>
      <ul class="points">
        <li>
          The game only asks what a player at a real table would have to answer. Most of the time
          you press <strong>Pass</strong> in the action dock to let things move on.
        </li>
        <li>
          <strong>Skip to my turn</strong> passes and resolves everything until your next main phase.
          It stops only for a choice you have to make.
        </li>
        <li>Misclicked? There's <strong>undo</strong>.</li>
        <li>
          A card without the <span class="chip">AUTO</span> badge is yours to resolve: move the pieces
          by hand, and agree on it with the table.
        </li>
        <li>
          The <strong>Help</strong> button (the <strong>?</strong> in the header) has tips for each page,
          a practice game and the keyboard shortcuts.
        </li>
      </ul>
    </section>

    <section aria-labelledby="decks-title">
      <h2 id="decks-title">Decks</h2>
      <ul class="points">
        <li>
          <a href="#/decks">Check a deck</a> before you sit down: it shows which cards play automatically,
          which are simplified, and which you'd resolve by hand.
        </li>
        <li>
          Want a deck's missing cards automated? Press <strong>Request missing cards</strong> on the decks
          page.
        </li>
        <li>
          The <a href="#/roadmap">roadmap</a> shows what the engine does today, what's partial and what's
          next.
        </li>
      </ul>
    </section>

    {#if invite}
      <section aria-labelledby="discord-title" data-testid="guide-discord-section">
        <h2 id="discord-title">On the Discord</h2>
        <dl class="terms">
          <dt>#welcome, #announcements</dt>
          <dd>This guide, and news: new features and downtime.</dd>
          <dt>#looking-for-game</dt>
          <dd>Find a pod.</dd>
          <dt>#decks</dt>
          <dd>Brewing, swapping lists, asking what works.</dd>
          <dt>#account-requests</dt>
          <dd>Ask for the Elder Dragon Highlander role.</dd>
          <dt>#bugs-and-ideas</dt>
          <dd>
            Report a bug or suggest something. The bug button (<strong>{L.reportABug}</strong>), in
            a game's toolbar or, signed in, at the top of every other page, files it on GitHub,
            where what you write is public.
          </dd>
        </dl>
        <p class="para">The C2 bot's slash commands:</p>
        <dl class="terms cmds">
          <dt><code>/c2-invite</code></dt>
          <dd>Start a new table (you host it) and post its invite link.</dd>
          <dt><code>/c2-invite-dm</code></dt>
          <dd>DM someone an invite, to a new table or to one you host.</dd>
          <dt><code>/c2-games</code></dt>
          <dd>List the tables in the lobby and in play.</dd>
          <dt><code>/c2-deck-check</code></dt>
          <dd>Check how much of a decklist the engine automates.</dd>
          <dt><code>/c2-deck-req</code></dt>
          <dd>Ask for a deck's missing cards to be added.</dd>
        </dl>
      </section>
    {/if}
  </div>
</Page>

<style>
  .stack {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 24px;
  }
  /* The same section head as About's: app.css styles a bare h2 as a
     small mono label, so say what this one is. */
  h2 {
    margin: 0 0 10px;
    font-family: var(--font-display);
    font-size: 19px;
    font-weight: 800;
    letter-spacing: -0.005em;
    text-transform: none;
    color: var(--fg);
  }
  .para {
    margin: 0 0 8px;
    max-width: 72ch;
    color: var(--fg-muted);
    font-size: 13.5px;
    line-height: 1.6;
  }
  .points {
    margin: 0 0 10px;
    padding-left: 1.2em;
    display: flex;
    flex-direction: column;
    gap: 8px;
    max-width: 72ch;
    color: var(--fg-muted);
    font-size: 13.5px;
    line-height: 1.6;
  }
  .points:last-child {
    margin-bottom: 0;
  }
  .terms {
    margin: 0 0 10px;
    display: grid;
    grid-template-columns: max-content minmax(0, 1fr);
    gap: 8px 16px;
    max-width: 80ch;
    font-size: 13.5px;
    line-height: 1.6;
  }
  .terms:last-child {
    margin-bottom: 0;
  }
  dt {
    font-weight: 700;
    color: var(--fg);
  }
  dd {
    margin: 0;
    color: var(--fg-muted);
  }
  /* A phone has no room for two columns: each term sits over its text. */
  @media (max-width: 560px) {
    .terms {
      grid-template-columns: minmax(0, 1fr);
      gap: 2px;
    }
    dd + dt {
      margin-top: 8px;
    }
  }
  code {
    font-family: var(--font-mono);
    font-size: 12.5px;
    color: var(--fg);
  }
  strong {
    color: var(--fg);
    font-weight: 600;
  }
  a:not(.btn) {
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
  .actions {
    margin: 12px 0 0;
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
</style>
