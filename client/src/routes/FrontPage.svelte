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
  //     on (GET /community's `gate`) those are join the CMD & CTRL Discord,
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

  // The table in the hero. The bot's attacker hits you for 3, then the
  // turn goes round to you, so the picture plays a moment of a real game
  // and lands on your turn. Under four seconds, played once, never on a
  // loop, and not at all under reduced motion (it starts on your turn).
  const seats = [
    { name: "You", life: 40, bot: false },
    { name: "Mara", life: 33, bot: false },
    { name: "Bot", life: 28, bot: true },
    { name: "Dev", life: 40, bot: false },
  ];
  let lives = $state(seats.map((s) => s.life));
  let turn = $state(0);
  let step = $state("Main phase");
  let hit = $state(-1);

  onMount(() => {
    void discordAuthEnabled().then((on) => {
      discordEnabled = on;
    });
    const still = window.matchMedia?.("(prefers-reduced-motion: reduce)").matches ?? true;
    if (still) return;
    turn = 2;
    step = "Combat";
    const beats: [number, () => void][] = [
      [
        1300,
        () => {
          hit = 0;
          lives[0] = 37;
        },
      ],
      [
        2300,
        () => {
          hit = -1;
          turn = 3;
          step = "Main phase";
        },
      ],
      [
        3500,
        () => {
          turn = 0;
          step = "Untap, upkeep, draw";
        },
      ],
      [4300, () => (step = "Main phase")],
    ];
    const timers = beats.map(([ms, fn]) => setTimeout(fn, ms));
    return () => timers.forEach(clearTimeout);
  });

  const invite = $derived($community.invite_url);
  const gate = $derived($community.gate);
</script>

<SiteHeader />

<main id="main" class="site-page wide front">
  {#if $expiryNotice}
    <p class="notice" role="status"><Icon name="undo" size={14} /> {$expiryNotice}</p>
  {/if}

  <section class="hero" aria-labelledby="front-title">
    <div class="pitch">
      <h1 id="front-title" class="title rise" style="--d: 0ms">
        Your Commander pod’s table, <span class="hot">over the wire.</span>
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

    <div class="table rise" style="--d: 120ms" data-ui>
      <div class="table-bar">
        <span class="table-name">Table 1</span>
        <span class="table-meta">Commander · 4 seats · 40 life</span>
      </div>
      <div
        class="board"
        role="img"
        aria-label="A four-seat table in progress: you, two friends and a bot in the empty chair"
      >
        {#each seats as seat, i (seat.name)}
          <div
            class="seat"
            class:you={i === 0}
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
          <span class="step">{step}</span>
          <span class="whose"
            >{seats[turn].name === "You" ? "Your" : `${seats[turn].name}’s`} turn</span
          >
        </div>
      </div>
    </div>
  </section>

  <aside class="steps-panel" aria-label="how to get an account">
    {#if gate}
      <ol class="steps">
        <li class="step-item">
          <h2>Join the CMD & CTRL Discord</h2>
          <p>That’s where pods form, bugs get reported and new cards are announced.</p>
        </li>
        <li class="step-item">
          <h2>Request account access</h2>
          <p>
            Answer yes to “request account access” in the server’s onboarding, then say hi in
            <span class="nowrap">#account-requests</span>. A mod gives you the Elder Dragon
            Highlander role.
          </p>
        </li>
        <li class="step-item last">
          <h2>Sign in and sit down</h2>
          <p>Create a table, invite your pod, or start a practice game against a bot.</p>
        </li>
      </ol>
    {:else}
      <div class="solo">
        <div>
          <h2>Sign in and sit down</h2>
          <p>Create a table, invite your pod, or start a practice game against a bot.</p>
        </div>
        {#if invite}
          <p class="invite-line">
            <Icon name="discord" size={18} />
            <span
              >Looking for a pod? <a href={invite} target="_blank" rel="noopener noreferrer"
                >Come and say hello on the CMD & CTRL Discord.</a
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
    /* Its own stacking context, so the lamp behind the table stays above
       the page's background. */
    isolation: isolate;
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
    max-width: 17ch;
  }
  /* The brand's one accented phrase, in the brighter orange (10:1 on the
     page) so it carries at a glance. */
  .hot {
    color: var(--accent-strong);
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

  /* The table, laid out as the real board is: the three opponents across
     the top, the step strip, and your seat along the bottom. One raised
     surface with a visible edge holds all four, so it reads as one table
     rather than four tiles (Ian's review on #3839). */
  .table {
    position: relative;
    padding: 14px;
    border-radius: var(--radius-xl);
    background: var(--surface-raised);
    box-shadow:
      inset 0 0 0 2px color-mix(in srgb, var(--fg) 16%, transparent),
      0 30px 80px -30px color-mix(in srgb, var(--accent) 35%, transparent);
  }
  /* A lamp over the table: one warm pool of light behind it, so the
     picture sits in a room at night rather than on a page. */
  .table::before {
    content: "";
    position: absolute;
    /* Sideways no further than the page gutter, so the lamp never makes
       the page scroll sideways (#3877). */
    inset: -18% max(-14%, calc(var(--page-gutter) * -1)) -10%;
    z-index: -1;
    background: radial-gradient(
      closest-side,
      color-mix(in srgb, var(--accent) 16%, transparent),
      transparent
    );
    pointer-events: none;
  }
  .table-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 2px 4px 12px;
    font-size: 14px;
  }
  .table-name {
    font-family: var(--font-display);
    font-weight: 800;
    letter-spacing: -0.01em;
    color: var(--fg);
  }
  .table-meta {
    color: var(--fg-muted);
  }
  .board {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 10px;
  }
  .seat {
    order: 1;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    gap: 10px;
    min-width: 0;
    padding: 14px 16px;
    border-radius: var(--radius-lg);
    background: color-mix(in srgb, var(--seat) 18%, var(--bg-1));
    box-shadow: inset 0 4px 0 var(--seat);
    transition:
      background 300ms var(--ease),
      box-shadow 300ms var(--ease);
  }
  .seat .life {
    font-size: clamp(34px, 3vw, 44px);
  }
  .seat.you {
    order: 3;
    grid-column: 1 / -1;
    padding: 20px 24px;
    box-shadow: inset 0 -4px 0 var(--seat);
  }
  .seat.you .life {
    font-size: clamp(56px, 6vw, 88px);
  }
  .seat.turn {
    background: color-mix(in srgb, var(--seat) 28%, var(--bg-1));
    box-shadow:
      inset 0 4px 0 var(--seat),
      inset 0 0 0 2px var(--priority);
  }
  .seat.you.turn {
    box-shadow:
      inset 0 -4px 0 var(--seat),
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
  /* The step strip sits between the opponents and you, as the table's own
     step bar does. */
  .centre {
    order: 2;
    grid-column: 1 / -1;
    display: flex;
    align-items: baseline;
    justify-content: center;
    gap: 12px;
    padding: 10px 16px;
    border-radius: var(--radius);
    background: var(--bg);
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

  /* How to get an account: a band two tone steps up from the page
     (1.3:1), so it reads as its own section; one step (1.06:1) blended in. */
  .steps-panel {
    margin-top: 88px;
    padding: 36px 40px;
    border-radius: var(--radius-xl);
    background: var(--surface-raised);
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
  .nowrap {
    white-space: nowrap;
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
    background: var(--surface-raised);
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

  /* Below the tablet design target the hero and the bands stack. */
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
  }

  /* Phone width. The table itself stays tablet-minimum, but this page is
     what a link tapped in Discord opens, so it gets a layout of its own:
     opponents shrink to a name and a life total, the actions go full
     width, and the gaps tighten. */
  @media (max-width: 599px) {
    .front {
      padding-top: 40px;
      padding-bottom: 64px;
    }
    .title {
      font-size: clamp(36px, 10vw, 44px);
    }
    .lede {
      font-size: 17px;
    }
    .ctas .btn,
    .practice .btn {
      width: 100%;
    }
    .table {
      padding: 10px;
    }
    .board {
      gap: 8px;
    }
    .seat {
      padding: 10px 12px;
      gap: 6px;
    }
    .seat:not(.you) .row,
    .chip {
      display: none;
    }
    .seat .life {
      font-size: 30px;
    }
    .seat.you .life {
      font-size: 52px;
    }
    .who {
      font-size: 14px;
    }
    .steps-panel {
      margin-top: 56px;
    }
    .cards {
      margin-top: 72px;
    }
    .invite-band h2 {
      font-size: 26px;
    }
    .ways {
      gap: 28px;
      margin-top: 40px;
    }
  }
</style>
