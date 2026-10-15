<script lang="ts">
  // FrontPage — what a signed-out visitor sees at #/ and its alias
  // #/login (ADR 0156 §4, mockup A). A session never sees it: the router
  // sends every session here on to the Lobby (ADR 0112 §1).
  //
  // DESIGN TRIAL (design/plugin-trial, not for merge as is): the taste
  // skill's art direction and Vercel's interface guidelines applied to
  // this page. The idea: the four seats of a pod become the hero picture.
  // It is drawn in HTML from the table's own vocabulary (seat colours,
  // life totals, the turn ring) and shows no card art, which ADR 0092
  // keeps away from signed-out visitors.
  //
  //   - The hero: the pitch, Sign in with Discord and Join our Discord
  //     side by side, Try a practice game under them; beside them, the
  //     table.
  //   - The steps to an account, as a band under the hero. With the gate
  //     on (GET /community's `gate`) those are join the C2 Discord,
  //     request access, sign in; with it off, only the last, and a line
  //     inviting people to the server.
  //   - Below: Have an invite? as a band of its own, then Check a deck,
  //     Roadmap and New here? as plain links.
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
  import { ABOUT_HASH, GUIDE_HASH } from "../lib/about";
  import { L } from "../lib/labels";
  import { seatColor } from "../lib/colors";

  let discordEnabled = $state(false);

  // The table in the hero. One seat takes 3 from a bot's attacker and the
  // turn passes once, so the picture reads as a game in progress. Played
  // once, never on a loop, and not at all under reduced motion.
  const seats = [
    { name: "You", life: 40, bot: false },
    { name: "Mara", life: 33, bot: false },
    { name: "Bot", life: 28, bot: true },
    { name: "Dev", life: 40, bot: false },
  ];
  let lives = $state(seats.map((s) => s.life));
  let turn = $state(2);
  let hit = $state(-1);

  onMount(() => {
    void discordAuthEnabled().then((on) => {
      discordEnabled = on;
    });
    const still = window.matchMedia?.("(prefers-reduced-motion: reduce)").matches ?? true;
    if (still) return;
    const timers = [
      setTimeout(() => {
        hit = 0;
        lives[0] = 37;
      }, 1400),
      setTimeout(() => {
        hit = -1;
        turn = 3;
      }, 2600),
    ];
    return () => timers.forEach(clearTimeout);
  });

  const invite = $derived($community.invite_url);
  const gate = $derived($community.gate);
</script>

<SiteHeader />

<main class="site-page wide front" id="main">
  {#if $expiryNotice}
    <p class="notice" role="status"><Icon name="undo" size={14} /> {$expiryNotice}</p>
  {/if}

  <section class="hero" aria-labelledby="front-title">
    <div class="pitch">
      <h1 id="front-title" class="title rise" style="--d: 0ms">
        Your Commander pod’s table, online
      </h1>
      <p class="lede rise" style="--d: 80ms">
        Four seats, a rules engine that resolves thousands of cards for you, and bots to fill an
        empty chair. Free and unofficial.
      </p>

      {#if discordEnabled || invite}
        <div class="ctas rise" style="--d: 160ms">
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

      <p class="practice rise" style="--d: 220ms">
        <a class="btn ghost" href="#/practice"
          ><Icon name="robot" size={15} /> {L.tryPracticeGame}</a
        >
        <span
          >No account needed: 5&nbsp;minutes against a bot that shows you where everything is.</span
        >
      </p>

      <p class="fine rise" style="--d: 280ms">
        {#if discordEnabled}
          Your Discord name and avatar become your seat, and your tables, games and decks follow you
          to any device.
        {/if}
        <a href={PRIVACY_HASH}>What we store, and how to delete it.</a>
      </p>
    </div>

    <div
      class="table rise"
      style="--d: 120ms"
      data-ui
      role="img"
      aria-label="A four-seat table in progress: you, two friends and a bot in the empty chair"
    >
      {#each seats as seat, i (seat.name)}
        <div
          class="seat"
          class:turn={turn === i}
          class:hit={hit === i}
          style="--seat: {seatColor(i)}"
        >
          <span class="who">
            {seat.name}{#if seat.bot}<span class="chip">bot</span>{/if}
          </span>
          <span class="life">{lives[i]}</span>
          <span class="row">
            <span class="zone">Hand {7 - i}</span>
            <span class="zone">Library {88 - 3 * i}</span>
          </span>
        </div>
      {/each}
      <div class="centre">
        <span class="step">Combat</span>
        <span class="whose"
          >{seats[turn].name === "You" ? "Your" : `${seats[turn].name}’s`} turn</span
        >
      </div>
    </div>
  </section>

  <aside class="steps-panel" aria-label="how to get an account">
    {#if gate}
      <ol class="steps">
        <li class="step-item">
          <h2>Join the C2 Discord</h2>
          <p>That’s where pods form, bugs get reported and new cards are announced.</p>
        </li>
        <li class="step-item">
          <h2>Request account access</h2>
          <p>
            Answer yes to “request account access” in the server’s onboarding, and a mod gives you
            the Elder Dragon Highlander role.
          </p>
        </li>
        <li class="step-item last">
          <h2>Sign in and sit down</h2>
          <p>Create a table, invite your pod, or start a practice game against bots.</p>
        </li>
      </ol>
    {:else}
      <div class="solo">
        <div>
          <h2>Sign in and sit down</h2>
          <p>Create a table, invite your pod, or start a practice game against bots.</p>
        </div>
        {#if invite}
          <p class="invite-line">
            <Icon name="discord" size={18} />
            <span
              >Looking for a pod? <a href={invite} target="_blank" rel="noopener noreferrer"
                >Come and say hello on the C2 Discord.</a
              ></span
            >
          </p>
        {/if}
      </div>
    {/if}
  </aside>

  <section class="cards" aria-label="more ways in">
    <div class="invite-band">
      <div>
        <h2>Have an invite?</h2>
        <p>A host’s link or code seats you as a guest. No account needed.</p>
      </div>
      <InviteBox />
    </div>
    <div class="ways">
      <div class="way">
        <h2>Check a deck</h2>
        <p>Paste a list and see how much of it plays as printed.</p>
        <a class="link-arrow" href="#/decks">Decks</a>
      </div>
      <div class="way">
        <h2>Roadmap</h2>
        <p>What the engine does today, what’s partial, and what’s next.</p>
        <a class="link-arrow" href="#/roadmap">Roadmap</a>
      </div>
      <div class="way">
        <h2>New here?</h2>
        <p>How to get in, find a pod and play your first game; and what this project is.</p>
        <p class="links">
          <a class="link-arrow" href={GUIDE_HASH}>Getting started</a>
          <a class="link-arrow" href={ABOUT_HASH}>About</a>
        </p>
      </div>
    </div>
  </section>
</main>

<style>
  .front {
    flex: 1 0 auto;
    width: 100%;
    max-width: calc(1240px + 2 * var(--page-gutter));
    margin: 0 auto;
    padding: 72px var(--page-gutter) 96px;
    box-sizing: border-box;
    min-width: 0;
  }
  .notice {
    margin: 0 0 24px;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 16px;
    border-radius: var(--radius);
    background: var(--accent-soft);
    color: var(--accent-strong);
    font-size: 15px;
  }

  /* The hero: the pitch on the left, the table as its picture. */
  .hero {
    display: grid;
    grid-template-columns: minmax(0, 1.1fr) minmax(0, 1fr);
    gap: 48px 72px;
    align-items: center;
  }
  .title {
    margin: 0;
    font-family: var(--font-display);
    font-size: clamp(44px, 5.6vw, 84px);
    font-weight: 800;
    line-height: 0.98;
    letter-spacing: -0.04em;
    text-transform: none;
    text-wrap: balance;
    color: var(--fg);
    max-width: 15ch;
  }
  .lede {
    margin: 28px 0 0;
    max-width: 46ch;
    font-size: 18px;
    line-height: 1.55;
    color: var(--fg);
  }
  .ctas {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    margin-top: 36px;
  }
  .btn.lg {
    height: 52px;
    padding: 0 26px;
    font-size: 16px;
  }
  .practice {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 10px 16px;
    margin: 20px 0 0;
    font-size: 16px;
    line-height: 1.5;
    color: var(--fg-muted);
  }
  .practice span {
    flex: 1 1 24ch;
    max-width: 40ch;
  }
  .fine {
    margin: 20px 0 0;
    max-width: 56ch;
    font-size: 16px;
    line-height: 1.5;
    color: var(--fg-muted);
  }
  .fine a {
    color: var(--fg);
  }

  /* The table: four seats round a centre, drawn with fills, not lines. */
  .table {
    position: relative;
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
    padding: 10px;
    border-radius: var(--radius-xl);
    background: var(--bg-1);
    grid-template-rows: 1fr auto 1fr;
    aspect-ratio: 1 / 0.86;
  }
  .seat {
    position: relative;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    padding: 18px 20px;
    border-radius: var(--radius-lg);
    background: color-mix(in srgb, var(--seat) 12%, var(--bg-2));
    box-shadow: inset 4px 0 0 var(--seat);
    transition:
      background 300ms var(--ease),
      box-shadow 300ms var(--ease);
  }
  .seat.turn {
    background: color-mix(in srgb, var(--seat) 22%, var(--bg-2));
    box-shadow:
      inset 4px 0 0 var(--seat),
      inset 0 0 0 2px var(--priority);
  }
  .who {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 16px;
    font-weight: 700;
    color: var(--fg);
  }
  .chip {
    padding: 2px 8px;
    border-radius: 999px;
    background: color-mix(in srgb, var(--seat) 30%, transparent);
    font-size: 13px;
    font-weight: 600;
  }
  .life {
    font-family: var(--font-display);
    font-size: clamp(48px, 5vw, 76px);
    font-weight: 800;
    line-height: 1;
    letter-spacing: -0.04em;
    font-variant-numeric: tabular-nums;
    color: var(--fg);
  }
  .seat.hit .life {
    color: var(--attack);
    animation: knock 420ms var(--ease);
  }
  .row {
    display: flex;
    flex-wrap: wrap;
    gap: 4px 14px;
    font-size: 14px;
    color: color-mix(in srgb, var(--fg) 78%, transparent);
    font-variant-numeric: tabular-nums;
  }
  /* The step strip sits between the two rows, as the table's own step
     bar does, so it never covers a seat. */
  .centre {
    order: 1;
    grid-column: 1 / -1;
    display: flex;
    align-items: baseline;
    justify-content: center;
    gap: 12px;
    padding: 10px 16px;
    border-radius: var(--radius);
    background: var(--bg);
  }
  .seat:nth-child(n + 3) {
    order: 2;
  }
  .step {
    font-family: var(--font-display);
    font-size: 18px;
    font-weight: 800;
    color: var(--priority-strong);
  }
  .whose {
    font-size: 14px;
    color: var(--fg-muted);
  }

  /* How to get an account: a band one tone up from the page. */
  .steps-panel {
    margin-top: 88px;
    padding: 36px 40px;
    border-radius: var(--radius-xl);
    background: var(--bg-1);
  }
  .steps {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 32px;
  }
  .steps-panel h2,
  .cards h2 {
    margin: 0;
    font-family: var(--font-display);
    font-size: 22px;
    font-weight: 800;
    letter-spacing: -0.02em;
    text-transform: none;
    text-wrap: balance;
    color: var(--fg);
  }
  .step-item.last h2 {
    color: var(--accent-strong);
  }
  .steps-panel p,
  .cards p {
    margin: 8px 0 0;
    font-size: 16px;
    line-height: 1.55;
    color: var(--fg-muted);
  }
  .solo {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 16px 48px;
  }
  .invite-line {
    display: flex;
    align-items: center;
    gap: 10px;
    margin: 0;
  }
  .invite-line :global(svg) {
    flex: none;
    color: #aab3ff;
  }
  .invite-line a {
    color: #c9cfff;
    font-weight: 600;
  }

  /* More ways in: the invite box gets a band, the rest are plain links. */
  .cards {
    margin-top: 120px;
  }
  .invite-band {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(0, 1.1fr);
    gap: 24px 64px;
    align-items: center;
    padding: 40px;
    border-radius: var(--radius-xl);
    background: var(--surface);
  }
  .invite-band h2 {
    font-size: 32px;
  }
  .ways {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 40px;
    margin-top: 64px;
  }
  .way .link-arrow {
    margin-top: 14px;
    font-size: 16px;
  }
  .way .links {
    margin: 14px 0 0;
    display: flex;
    flex-wrap: wrap;
    gap: 6px 20px;
  }
  .way .links .link-arrow {
    margin-top: 0;
  }

  /* Entrances: each part rises once, staggered. Nothing stays hidden
     without motion: the animation only runs where motion is allowed. */
  @media (prefers-reduced-motion: no-preference) {
    :global(:root:not([data-reduce-motion="1"])) .rise {
      animation: rise 560ms var(--ease) both;
      animation-delay: var(--d, 0ms);
    }
  }
  @keyframes rise {
    from {
      opacity: 0;
      transform: translateY(20px);
    }
    to {
      opacity: 1;
      transform: none;
    }
  }
  @keyframes knock {
    30% {
      transform: translateX(-6px);
    }
    60% {
      transform: translateX(4px);
    }
  }

  /* Below the tablet design target everything stacks; there is no layout
     of its own (the tablet-minimum rule). */
  @media (max-width: 819px) {
    .hero,
    .steps,
    .invite-band,
    .ways {
      grid-template-columns: minmax(0, 1fr);
    }
    .steps-panel,
    .invite-band {
      padding: 28px 24px;
    }
    .table {
      aspect-ratio: auto;
    }
    .seat {
      min-height: 150px;
      gap: 12px;
    }
    .life {
      font-size: 52px;
    }
  }
</style>
