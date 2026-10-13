<script lang="ts">
  // FrontPage — what a signed-out visitor sees at #/ and its alias
  // #/login (ADR 0156 §4, mockup A). A session never sees it: the router
  // sends every session here on to the Lobby (ADR 0112 §1).
  //
  //   - The hero: the pitch, Sign in with Discord and Join our Discord
  //     server side by side, Try a practice game under them, and beside
  //     them the steps to an account. With the gate on (GET /community's
  //     `gate`) those are join the C2 Discord, request access, sign in;
  //     with it off, only the last, and a line inviting people to the
  //     server.
  //   - Below: Have an invite? (the sign-in page's invite box, unchanged),
  //     Check a deck, Roadmap and About.
  //
  // Every Discord placement hides when the server names no invite. The
  // sign-in button shows only once /auth/discord/config says Discord
  // sign-in is configured, so a server without it never offers a 503.
  import { onMount } from "svelte";
  import SiteHeader from "../lib/components/SiteHeader.svelte";
  import InviteBox from "../lib/components/InviteBox.svelte";
  import Icon from "../lib/components/Icon.svelte";
  import { discordAuthEnabled, discordLoginHref } from "../lib/api";
  import { expiryNotice } from "../lib/session";
  import { community } from "../lib/community";
  import { PRIVACY_HASH } from "../lib/privacy";
  import { ABOUT_HASH } from "../lib/about";
  import { L } from "../lib/labels";

  let discordEnabled = $state(false);
  onMount(() => {
    void discordAuthEnabled().then((on) => {
      discordEnabled = on;
    });
  });

  const invite = $derived($community.invite_url);
  const gate = $derived($community.gate);
</script>

<SiteHeader />

<main class="site-page wide front">
  {#if $expiryNotice}
    <p class="notice" role="status"><Icon name="undo" size={14} /> {$expiryNotice}</p>
  {/if}

  <section class="hero" aria-labelledby="front-title">
    <div class="pitch">
      <p class="eyebrow">Free, unofficial, four seats</p>
      <h1 id="front-title" class="title">
        Your Commander pod's table, <span class="hot">over the wire.</span>
      </h1>
      <p class="lede">
        Four seats, a rules engine that resolves thousands of cards for you, and bots to fill an
        empty chair. Free and unofficial.
      </p>

      {#if discordEnabled || invite}
        <div class="ctas">
          {#if discordEnabled}
            <a class="btn discord lg" href={discordLoginHref()}>
              <Icon name="discord" size={18} />
              {L.signInWithDiscord}
            </a>
          {/if}
          {#if invite}
            <a class="btn ghost lg" href={invite} target="_blank" rel="noopener noreferrer"
              >{L.joinDiscordServer}</a
            >
          {/if}
        </div>
      {/if}

      <p class="practice">
        <a class="btn ghost" href="#/practice"
          ><Icon name="robot" size={15} /> {L.tryPracticeGame}</a
        >
        <span
          >No account needed: five minutes against a bot that shows you where everything is.</span
        >
      </p>

      <p class="fine">
        {#if discordEnabled}
          Your Discord name and avatar become your seat, and your tables, games and decks follow you
          to any device.
        {/if}
        <a href={PRIVACY_HASH}>What we store, and how to delete it.</a>
      </p>
    </div>

    <aside class="steps-panel" aria-label="how to get an account">
      {#if gate}
        <ol class="steps">
          <li class="step">
            <span class="n" aria-hidden="true">1</span>
            <div>
              <h2>Join the C2 Discord</h2>
              <p>That's where pods form, bugs get reported and new cards are announced.</p>
            </div>
          </li>
          <li class="step">
            <span class="n" aria-hidden="true">2</span>
            <div>
              <h2>Request account access</h2>
              <p>
                Answer yes to “request account access” in the server's onboarding, and a mod gives
                you the Elder Dragon Highlander role.
              </p>
            </div>
          </li>
          <li class="step">
            <span class="n hot" aria-hidden="true">3</span>
            <div>
              <h2>Sign in and sit down</h2>
              <p>Create a table, invite your pod, or start a practice game against bots.</p>
            </div>
          </li>
        </ol>
      {:else}
        <div class="step solo">
          <span class="n hot" aria-hidden="true"><Icon name="check" size={16} /></span>
          <div>
            <h2>Sign in and sit down</h2>
            <p>Create a table, invite your pod, or start a practice game against bots.</p>
          </div>
        </div>
        {#if invite}
          <p class="invite-line">
            <Icon name="discord" size={16} />
            <span
              >Looking for a pod? <a href={invite} target="_blank" rel="noopener noreferrer"
                >Come and say hello on the C2 Discord.</a
              ></span
            >
          </p>
        {/if}
      {/if}
    </aside>
  </section>

  <section class="cards" aria-label="more ways in">
    <div class="card invite-card">
      <h2>Have an invite?</h2>
      <p>A host's link or code seats you as a guest. No account needed.</p>
      <InviteBox />
    </div>
    <div class="card">
      <h2>Check a deck</h2>
      <p>Paste a list and see how much of it plays as printed.</p>
      <a class="link-arrow" href="#/decks">Decks</a>
    </div>
    <div class="card">
      <h2>Roadmap</h2>
      <p>What the engine does today, what's partial, and what's next.</p>
      <a class="link-arrow" href="#/roadmap">Roadmap</a>
    </div>
    <div class="card">
      <h2>About</h2>
      <p>What this project is, its limits, and how to send feedback.</p>
      <a class="link-arrow" href={ABOUT_HASH}>About</a>
    </div>
  </section>
</main>

<style>
  /* Page.svelte's wide frame (its rules are scoped to it): the hero is
     not a page head, so this page draws the frame itself. */
  .front {
    flex: 1 0 auto;
    width: 100%;
    max-width: calc(1200px + 2 * var(--page-gutter));
    margin: 0 auto;
    padding: 64px var(--page-gutter) 64px;
    box-sizing: border-box;
    min-width: 0;
  }
  .notice {
    margin: 0 0 24px;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 14px;
    border-radius: 10px;
    background: var(--accent-soft);
    border: 1px solid color-mix(in srgb, var(--accent) 40%, transparent);
    color: var(--accent-strong);
    font-size: 12.5px;
  }

  .hero {
    display: grid;
    grid-template-columns: minmax(0, 1.25fr) minmax(0, 1fr);
    gap: 40px 64px;
    align-items: start;
  }
  .eyebrow {
    margin: 0 0 16px;
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--accent);
  }
  .title {
    margin: 0;
    font-family: var(--font-display);
    font-size: clamp(38px, 3.9vw, 54px);
    font-weight: 800;
    line-height: 1.04;
    letter-spacing: -0.035em;
    text-transform: none;
    color: var(--fg);
    max-width: 19ch;
  }
  .hot {
    color: var(--accent);
  }
  .lede {
    margin: 22px 0 0;
    max-width: 52ch;
    font-size: 17px;
    line-height: 1.55;
    color: var(--fg-muted);
  }
  .ctas {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    margin-top: 32px;
  }
  .btn.lg {
    height: 50px;
    padding: 0 24px;
    font-size: 15px;
  }
  .practice {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 10px 14px;
    margin: 18px 0 0;
    font-size: 13px;
    color: var(--fg-dim);
  }
  .practice span {
    flex: 1 1 22ch;
    max-width: 40ch;
  }
  .fine {
    margin: 18px 0 0;
    max-width: 60ch;
    font-size: 12px;
    line-height: 1.5;
    color: var(--fg-dim);
  }
  .fine a {
    color: var(--fg-muted);
  }

  .steps {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  .step {
    display: flex;
    gap: 16px;
    padding: 20px 22px;
    border-radius: var(--radius-lg);
    border: 1px solid var(--border-strong);
    background: var(--surface);
  }
  .n {
    flex: none;
    display: grid;
    place-items: center;
    width: 36px;
    height: 36px;
    border-radius: 10px;
    background: var(--surface-raised);
    color: var(--fg);
    font-family: var(--font-display);
    font-weight: 800;
    font-size: 15px;
  }
  .n.hot {
    color: var(--accent);
    background: var(--accent-soft);
  }
  .step h2,
  .card h2 {
    margin: 0;
    font-family: var(--font-display);
    font-size: 17px;
    font-weight: 700;
    letter-spacing: -0.01em;
    text-transform: none;
    color: var(--fg);
  }
  .step p,
  .card p {
    margin: 6px 0 0;
    font-size: 13.5px;
    line-height: 1.5;
    color: var(--fg-muted);
  }
  .invite-line {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    margin: 14px 4px 0;
    font-size: 13.5px;
    line-height: 1.5;
    color: var(--fg-muted);
  }
  .invite-line :global(svg) {
    flex: none;
    margin-top: 2px;
    color: #aab3ff;
  }
  .invite-line a {
    color: #c9cfff;
    font-weight: 600;
  }

  .cards {
    display: grid;
    grid-template-columns: minmax(320px, 1.4fr) repeat(3, minmax(0, 1fr));
    gap: 16px;
    margin-top: 56px;
  }
  .card {
    display: flex;
    flex-direction: column;
    gap: 0;
  }
  .card .link-arrow {
    margin-top: 14px;
    font-size: 13.5px;
  }
  .invite-card p {
    margin-bottom: 16px;
  }

  /* Below the tablet design target the hero and the cards stack; there is
     no layout of their own (the tablet-minimum rule). */
  @media (max-width: 819px) {
    .hero,
    .cards {
      grid-template-columns: minmax(0, 1fr);
    }
  }
</style>
