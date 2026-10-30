<script lang="ts">
  // Guide — getting started (#/guide): three ways in, getting an
  // account, a first game step by step, what to do at the table, decks,
  // and where things are on the CMD & CTRL Discord. The same guide is
  // posted in the Discord's #welcome channel; keep the two saying the
  // same things.
  //
  // Public on purpose (lib/signedInHome.ts), like About: it is written for
  // someone who has not signed in yet. Every button it names is a real
  // label in the app; when one is renamed, rename it here too.
  //
  // What it says follows GET /community, as the front page does: the
  // account steps only when the sign-up gate is on, the voice room and
  // the #looking-for-game post only when the server offers them, and the
  // Discord section and every Discord button only when it names an
  // invite.
  import Page from "../lib/components/Page.svelte";
  import Icon from "../lib/components/Icon.svelte";
  import { ABOUT_HASH, GUIDE_HASH } from "../lib/about";
  import { L } from "../lib/labels";
  import { community } from "../lib/community";

  const invite = $derived($community.invite_url);
  const gate = $derived($community.gate);
  const voice = $derived($community.voice === true);
  const lfg = $derived($community.lfg === true);

  // The router owns the hash, so an in-page #anchor would be read as a
  // route: scroll to the section instead.
  function toAccount(e: MouseEvent): void {
    e.preventDefault();
    document.getElementById("in-title")?.scrollIntoView({ behavior: "smooth", block: "start" });
  }
</script>

<Page
  eyebrow="Guide"
  title="Getting started"
  lede="CMD & CTRL is a free Commander table in your browser: four seats, a rules engine that plays thousands of cards for you, and bots for the empty chairs. Here's how to sit down and play."
  width="prose"
>
  <div class="stack">
    <section aria-labelledby="ways-title">
      <h2 id="ways-title">Three ways to start</h2>
      <ul class="ways">
        <li class="way">
          <Icon name="robot" size={20} />
          <h3>Try a practice game</h3>
          <p>No account needed. A short game against a bot that shows you where everything is.</p>
          <a class="link-arrow" href="#/practice">{L.tryPracticeGame}</a>
        </li>
        <li class="way">
          <Icon name="link" size={20} />
          <h3>Join a friend's table</h3>
          <p>
            Open the invite link your host sent you. You sit down as a guest, no account needed.
          </p>
        </li>
        <li class="way">
          <Icon name="crown" size={20} />
          <h3>Host your own</h3>
          <p>Sign in with Discord to create tables and keep your decks on every device.</p>
          <a class="link-arrow" href={GUIDE_HASH} onclick={toAccount}>Get an account</a>
        </li>
      </ul>
    </section>

    <section aria-labelledby="in-title">
      <h2 id="in-title">Getting an account</h2>
      {#if gate}
        <p class="para">
          Accounts are for members of the CMD & CTRL Discord with the
          <strong>Elder Dragon Highlander</strong> role. Getting it takes a minute:
        </p>
        <ol class="steps">
          <li class="step">
            <span class="n" aria-hidden="true">1</span>
            <div class="body">
              <h3>Join the CMD & CTRL Discord</h3>
              <p>Everyone who joins is a <strong>Planeswalker</strong>.</p>
            </div>
            {#if invite}
              <a class="btn discord sm" href={invite} target="_blank" rel="noopener noreferrer"
                ><Icon name="discord" size={15} /> Join</a
              >
            {/if}
          </li>
          <li class="step">
            <span class="n" aria-hidden="true">2</span>
            <div class="body">
              <h3>Answer yes to “request account access”</h3>
              <p>
                It's a question in the server's onboarding. Already answered no? Change it under
                <strong>Channels &amp; Roles</strong>.
              </p>
            </div>
          </li>
          <li class="step">
            <span class="n" aria-hidden="true">3</span>
            <div class="body">
              <h3>Say hi in #account-requests</h3>
              <p>A mod approves your request and gives you the role.</p>
            </div>
          </li>
          <li class="step">
            <span class="n hot" aria-hidden="true">4</span>
            <div class="body">
              <h3>Sign in here with Discord</h3>
              <p>Your Discord name and avatar become your seat at the table.</p>
            </div>
          </li>
        </ol>
      {:else}
        <p class="para">
          Press <span class="ui">{L.signInWithDiscord}</span>. Your Discord name and avatar become
          your seat, and your tables and decks follow you to any device.
        </p>
      {/if}
    </section>

    <section aria-labelledby="play-title">
      <h2 id="play-title">Your first game</h2>
      <ol class="steps">
        <li class="step">
          <span class="n" aria-hidden="true">1</span>
          <div class="body">
            <h3>Create a table</h3>
            <p>
              On <strong>Tables</strong>, press <span class="ui">+ New table</span>.
              {#if voice}
                Tick <span class="ui">Create a voice room</span> for a private voice channel on the Discord,
                open only to the table's players. It closes 10 minutes after the game ends.
              {/if}
            </p>
          </div>
        </li>
        <li class="step">
          <span class="n" aria-hidden="true">2</span>
          <div class="body">
            <h3>Invite your pod</h3>
            <p>
              Press <span class="ui">{L.inviteButton}</span>, then
              <span class="ui">{L.copyInvite}</span>, and send the link to your friends.
              {#if lfg}
                Short of players? When you create the table, tick
                <strong>Post to the CMD & CTRL Discord (#looking-for-game)</strong> and anyone on the
                Discord can join.
              {:else if invite}
                Short of players? Ask in <strong>#looking-for-game</strong> on the Discord.
              {/if}
            </p>
          </div>
        </li>
        <li class="step">
          <span class="n" aria-hidden="true">3</span>
          <div class="body">
            <h3>Fill empty seats with bots</h3>
            <p>
              Commander is best at four. Press <span class="ui">Add bot</span> on an open seat, then pick
              how strong it plays and which deck it brings.
            </p>
          </div>
        </li>
        <li class="step">
          <span class="n" aria-hidden="true">4</span>
          <div class="body">
            <h3>Choose your deck</h3>
            <p>
              Your seat opens the deck picker. Pick a pre-built deck or one of your saved decks, or
              paste an Archidekt link or a plain-text decklist. From Moxfield, open the deck, choose <strong
                >Export</strong
              >
              then
              <strong>Copy plain text</strong>, and paste that. Every deck is checked against the
              Commander rules: 100 cards, singleton, colour identity.
            </p>
          </div>
        </li>
        <li class="step">
          <span class="n hot" aria-hidden="true">5</span>
          <div class="body">
            <h3>Start</h3>
            <p>
              Once every seated player has a deck, the host presses <span class="ui"
                >Start game</span
              >.
            </p>
          </div>
        </li>
      </ol>
    </section>

    <section aria-labelledby="table-title">
      <h2 id="table-title">At the table</h2>
      <ul class="points">
        <li>
          <strong>The game runs the rules for you</strong>: priority, the stack, combat, triggers,
          commander damage. Cards it knows resolve on their own and carry an
          <span class="chip">AUTO</span> badge.
        </li>
        <li>
          <strong>It passes for you when there's nothing to do</strong>, and stops when you can
          respond. When it's waiting on you, press <span class="ui">{L.next}</span> to move on, or
          <span class="ui">{L.pass}</span> to let the top of the stack resolve.
        </li>
        <li>
          <span class="ui">{L.skipToMyTurn}</span> passes everything until your next main phase. It stops
          only for a choice you have to make.
        </li>
        <li>
          Misclicked? <span class="ui">Undo</span> takes it back. You get one each turn, unless the host
          allows more.
        </li>
        <li>
          <strong>No <span class="chip">AUTO</span> badge?</strong> That card is yours to resolve: move
          the pieces by hand and agree on it with the table, as you would in person.
        </li>
        <li>
          Stuck? The <strong>?</strong> in the site header has tips for each page, the practice game
          and the keyboard shortcuts. At the table, they're under <strong>Help</strong> in the
          <strong>⋯</strong> menu.
        </li>
      </ul>
    </section>

    <section aria-labelledby="decks-title">
      <h2 id="decks-title">Decks</h2>
      <ul class="points">
        <li>
          <a href="#/decks">Check a deck</a> before you sit down: it shows which cards are fully automated,
          which are automated with caveats, and which you'd resolve by hand.
        </li>
        <li>
          Missing something? Sign in with Discord and press
          <span class="ui">Request missing cards</span> on the decks page.
        </li>
        <li>
          The <a href="#/roadmap">roadmap</a> shows what the engine does today, what's partial and what's
          next.
        </li>
      </ul>
    </section>

    {#if invite}
      <section aria-labelledby="discord-title" data-testid="guide-discord-section">
        <h2 id="discord-title">On the CMD & CTRL Discord</h2>
        <dl class="terms">
          <dt>#welcome</dt>
          <dd>House rules, and this guide in short.</dd>
          <dt>#announcements</dt>
          <dd>Releases and downtime.</dd>
          <dt>#looking-for-game</dt>
          <dd>Find a pod, or ping <strong>@LFG</strong> when you need a fourth.</dd>
          <dt>#decks</dt>
          <dd>Lists, brews and asking what works.</dd>
          <dt>#account-requests</dt>
          <dd>Ask for the Elder Dragon Highlander role.</dd>
          <dt>#bug-reports</dt>
          <dd>
            A forum for anything broken: open a post. Or use the bug button (<strong
              >{L.reportABug}</strong
            >): it's in a game's toolbar and, once you're signed in, at the top of every other page.
            It files the report on GitHub, where what you write is public.
          </dd>
          <dt>#requests-and-ideas</dt>
          <dd>Anything you wish it did, and cards you'd like automated.</dd>
          <dt>#general</dt>
          <dd>Everything else.</dd>
        </dl>
        <h3 class="sub">The bot's slash commands</h3>
        <dl class="terms">
          <dt><code>/c2-invite</code></dt>
          <dd>Start a new table (you host it) and post its invite link.</dd>
          <dt><code>/c2-invite-dm</code></dt>
          <dd>DM someone an invite, to a new table or to one you host.</dd>
          <dt><code>/c2-games</code></dt>
          <dd>List the tables in the lobby and in play.</dd>
          <dt><code>/c2-card</code></dt>
          <dd>Show a card and whether the engine plays it.</dd>
          <dt><code>/c2-deck-check</code></dt>
          <dd>Check how much of a decklist the engine automates.</dd>
          <dt><code>/c2-deck-req</code></dt>
          <dd>Ask for a deck's missing cards to be added.</dd>
        </dl>
        <p class="actions" data-testid="guide-discord">
          <a class="btn discord" href={invite} target="_blank" rel="noopener noreferrer"
            ><Icon name="discord" size={15} /> Join the CMD & CTRL Discord</a
          >
          <a class="btn ghost" href="#/practice"
            ><Icon name="robot" size={15} /> {L.tryPracticeGame}</a
          >
        </p>
      </section>
    {/if}

    <p class="note">
      CMD & CTRL is a free, unofficial hobby project, and it's in beta: expect rough edges and
      frequent changes. <a href={ABOUT_HASH}>About</a> has the details.
    </p>
  </div>
</Page>

<style>
  .stack {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 40px;
  }
  /* The same section head as About's: app.css styles a bare h2 as a
     small mono label, so say what this one is. */
  h2 {
    margin: 0 0 14px;
    font-family: var(--font-display);
    font-size: 22px;
    font-weight: 800;
    letter-spacing: -0.01em;
    text-transform: none;
    color: var(--fg);
  }
  h3 {
    margin: 0;
    font-family: var(--font-display);
    font-size: 16px;
    font-weight: 700;
    letter-spacing: -0.01em;
    text-transform: none;
    color: var(--fg);
  }
  .para,
  .points,
  .terms,
  .note {
    font-size: 15px;
    line-height: 1.6;
    color: var(--fg-muted);
  }
  .para {
    margin: 0 0 14px;
  }

  /* Three ways to start: a row of three, stacked on a narrow screen. */
  .ways {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 12px;
  }
  .way {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 18px;
    border-radius: var(--radius-lg);
    border: 1px solid var(--border-strong);
    background: var(--surface);
  }
  .way :global(svg) {
    color: var(--accent);
    margin-bottom: 4px;
  }
  .way p {
    margin: 0;
    flex: 1;
    font-size: 14px;
    line-height: 1.5;
    color: var(--fg-muted);
  }
  .way .link-arrow {
    margin-top: 8px;
    font-size: 14px;
  }

  /* Numbered steps, as on the gate page (RequestAccess). */
  .steps {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }
  .step {
    display: flex;
    align-items: flex-start;
    gap: 16px;
    padding: 16px 18px;
    border-radius: var(--radius-lg);
    border: 1px solid var(--border-strong);
    background: var(--surface);
  }
  .n {
    flex: none;
    display: grid;
    place-items: center;
    width: 32px;
    height: 32px;
    border-radius: 9px;
    background: var(--surface-raised);
    color: var(--fg);
    font-family: var(--font-display);
    font-weight: 800;
    font-size: 14px;
  }
  .n.hot {
    color: var(--accent);
    background: var(--accent-soft);
  }
  .body {
    flex: 1;
    min-width: 0;
    padding-top: 5px;
  }
  .step p {
    margin: 4px 0 0;
    font-size: 14.5px;
    line-height: 1.55;
    color: var(--fg-muted);
  }
  .step .btn {
    flex: none;
    align-self: center;
  }

  .points {
    margin: 0;
    padding-left: 1.2em;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }
  .terms {
    margin: 0;
    display: grid;
    grid-template-columns: max-content minmax(0, 1fr);
    gap: 8px 20px;
  }
  .sub {
    margin: 24px 0 10px;
  }
  /* Narrower than three cards fit: the ways stack. */
  @media (max-width: 680px) {
    .ways {
      grid-template-columns: minmax(0, 1fr);
    }
  }
  /* A phone has no room for two columns: each term sits over its text. */
  @media (max-width: 560px) {
    .terms {
      grid-template-columns: minmax(0, 1fr);
      gap: 2px;
    }
    dd + dt {
      margin-top: 10px;
    }
  }
  dt {
    font-weight: 700;
    color: var(--fg);
  }
  dd {
    margin: 0;
  }
  code {
    font-family: var(--font-mono);
    font-size: 13.5px;
    color: var(--fg);
  }
  strong {
    color: var(--fg);
    font-weight: 600;
  }
  a:not(.btn) {
    color: var(--accent);
  }
  /* A button's name, drawn as a small button so it is easy to spot. */
  .ui {
    display: inline-block;
    padding: 0 7px;
    border: 1px solid var(--border-strong);
    border-radius: var(--radius);
    background: var(--surface-raised);
    color: var(--fg);
    font-size: 0.92em;
    font-weight: 600;
    line-height: 1.55;
    white-space: nowrap;
  }
  .chip {
    font-family: var(--font-mono);
    font-size: 11px;
    padding: 1px 5px;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    color: var(--fg);
  }
  .actions {
    margin: 20px 0 0;
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .note {
    margin: 0;
    padding-top: 22px;
    border-top: 1px solid var(--border);
    font-size: 14px;
  }
</style>
