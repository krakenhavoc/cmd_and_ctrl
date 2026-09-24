<script lang="ts">
  // Home — the site portal (#1386, ADR 0092). A sitemap-style grid of
  // link cards to everywhere else on the site, grouped by what you're
  // trying to do. Public: it links to session-gated pages but shows
  // none of their content itself, so a signed-out visitor can always
  // land here and see what exists before signing in.
  //
  // This is one half of the plan's dual-portal experiment — SiteHeader
  // is the other. The owner will pick between them later; until then,
  // both exist and both work on their own.
  import { onMount } from "svelte";
  import { fetchBugReportConfig, logout as apiLogout } from "../lib/api";
  import { navigate } from "../lib/router";
  import { session } from "../lib/session";
  import { signedInUserID } from "../lib/myGames";
  import Icon from "../lib/components/Icon.svelte";
  import SiteHeader from "../lib/components/SiteHeader.svelte";

  const isSignedIn = $derived($session !== null);
  const hasLinkedUser = $derived(signedInUserID($session) !== null);

  // The in-game "report a bug" button (BugReportModal.svelte) needs a
  // live GameClient — the protocol log, the current view, the seq —
  // none of which exists here. Rather than fake one, this links
  // straight to a new GitHub issue, gated on the same /bugreport/config
  // probe that hides the in-game button when the server has no GitHub
  // token configured.
  let bugReportsEnabled = $state(false);
  onMount(() => {
    void fetchBugReportConfig().then((cfg) => {
      bugReportsEnabled = cfg.enabled;
    });
  });

  async function signOut(): Promise<void> {
    await apiLogout();
    navigate("#/login");
  }

  interface Card {
    title: string;
    href: string;
    description: string;
    external?: boolean;
    // needsSession: what a signed-out visitor is missing, or null if
    // the card works either way.
    needsSession?: string;
    action?: () => void;
  }

  const playCards = $derived.by((): Card[] => [
    {
      title: "Lobby",
      href: "#/lobby",
      description: "Every table in play — create one, or jump back into yours.",
      needsSession: isSignedIn ? undefined : "Sign in first",
    },
    {
      title: "My games",
      href: hasLinkedUser ? "#/my-games" : "#/login",
      description: "Every table you've sat at with your Discord account, from any device.",
      // Same short label as every other gated card — the description
      // already says it's specifically a Discord-linked session.
      needsSession: hasLinkedUser ? undefined : "Sign in first",
    },
    {
      title: "Join with an invite",
      href: "#/login",
      description: "Have a code or a link from a friend? Enter it here.",
    },
  ]);

  const cardCards = $derived.by((): Card[] => [
    {
      title: "Catalog",
      href: "#/catalog",
      description: "Every card the rules engine automates, and how completely.",
      needsSession: isSignedIn ? undefined : "Sign in first",
    },
    {
      title: "Roadmap",
      href: "#/roadmap",
      description: "What's implemented, what's partial, what's missing, and what's next.",
    },
  ]);

  const helpCards = $derived.by((): Card[] => {
    const cards: Card[] = [
      {
        title: "Bot guide",
        href: "https://github.com/krakenhavoc/cmd_and_ctrl/blob/develop/docs/bot.md",
        description: "How the AI bot seats work, and how to play against or alongside one.",
        external: true,
      },
      {
        title: "Source",
        href: "https://github.com/krakenhavoc/cmd_and_ctrl",
        description: "The whole project — server, client, and every design decision.",
        external: true,
      },
    ];
    if (bugReportsEnabled) {
      cards.push({
        title: "Report a bug",
        href: "https://github.com/krakenhavoc/cmd_and_ctrl/issues/new/choose",
        description: "Found something broken? File it against the repo.",
        external: true,
      });
    }
    return cards;
  });
</script>

<SiteHeader />

<section class="home">
  <div class="head">
    <p class="eyebrow">Everywhere else on the site</p>
    <h1>home</h1>
    <p class="lede">A map of every page. Sign in for the ones that need a seat.</p>
  </div>

  <div class="groups">
    <section class="group">
      <h2>Play</h2>
      <ul class="cards">
        {#each playCards as c (c.title)}
          <li>
            <a class="tile" href={c.href}>
              <div class="tile-head">
                <span class="tile-title">{c.title}</span>
                {#if c.needsSession}
                  <span class="pill">{c.needsSession}</span>
                {/if}
              </div>
              <p class="tile-desc">{c.description}</p>
            </a>
          </li>
        {/each}
      </ul>
    </section>

    <section class="group">
      <h2>Cards</h2>
      <ul class="cards">
        {#each cardCards as c (c.title)}
          <li>
            <a class="tile" href={c.href}>
              <div class="tile-head">
                <span class="tile-title">{c.title}</span>
                {#if c.needsSession}
                  <span class="pill">{c.needsSession}</span>
                {/if}
              </div>
              <p class="tile-desc">{c.description}</p>
            </a>
          </li>
        {/each}
      </ul>
    </section>

    <section class="group">
      <h2>Help</h2>
      <ul class="cards">
        {#each helpCards as c (c.title)}
          <li>
            <a
              class="tile"
              href={c.href}
              target={c.external ? "_blank" : undefined}
              rel={c.external ? "noreferrer noopener" : undefined}
            >
              <div class="tile-head">
                <span class="tile-title">{c.title}</span>
                {#if c.external}<Icon name="link" size={12} />{/if}
              </div>
              <p class="tile-desc">{c.description}</p>
            </a>
          </li>
        {/each}
      </ul>
    </section>

    <section class="group">
      <h2>Account</h2>
      <ul class="cards">
        {#if isSignedIn}
          <li>
            <button type="button" class="tile as-button" onclick={signOut}>
              <div class="tile-head">
                <span class="tile-title">Sign out</span>
              </div>
              <p class="tile-desc">End this session on this browser.</p>
            </button>
          </li>
        {:else}
          <li>
            <a class="tile" href="#/login">
              <div class="tile-head">
                <span class="tile-title">Sign in</span>
              </div>
              <p class="tile-desc">With Discord, or a shared invite code.</p>
            </a>
          </li>
        {/if}
      </ul>
    </section>
  </div>
</section>

<style>
  .home {
    width: min(1080px, 100%);
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: 28px;
  }
  .head {
    padding-top: 6px;
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
    max-width: 60ch;
    color: var(--fg-muted);
    font-size: 14px;
    line-height: 1.55;
  }

  .groups {
    display: flex;
    flex-direction: column;
    gap: 26px;
  }
  .group h2 {
    margin: 0 0 12px;
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: var(--fg-dim);
  }
  .cards {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
    gap: 14px;
  }
  .tile {
    display: flex;
    flex-direction: column;
    gap: 6px;
    height: 100%;
    box-sizing: border-box;
    padding: 16px 16px 18px;
    border-radius: var(--radius-lg);
    background: var(--surface);
    border: 1px solid var(--border);
    color: var(--fg);
    text-decoration: none;
    transition:
      border-color 120ms var(--ease),
      background 120ms var(--ease),
      transform 120ms var(--ease);
  }
  .tile:hover {
    border-color: var(--gold);
    background: var(--surface-hover);
    transform: translateY(-1px);
  }
  .tile.as-button {
    width: 100%;
    text-align: left;
    /* app.css gives every button align-items/justify-content: center,
       which in this column flexbox centres the title and description. */
    align-items: stretch;
    justify-content: flex-start;
    font: inherit;
    letter-spacing: normal;
    cursor: pointer;
  }
  .tile-head {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    justify-content: space-between;
    gap: 6px;
  }
  .tile-title {
    font-family: var(--font-display);
    font-weight: 700;
    font-size: 15px;
  }
  .tile-desc {
    margin: 0;
    color: var(--fg-muted);
    font-size: 12.5px;
    line-height: 1.5;
  }
  .pill {
    /* Wraps rather than overflows if a future label runs long — the
       grid can go as narrow as 160px per tile (#1386 phone check). */
    max-width: 100%;
    display: inline-flex;
    align-items: center;
    height: 20px;
    padding: 0 8px;
    border-radius: 999px;
    background: var(--accent-soft);
    color: var(--gold-strong);
    font-family: var(--font-mono);
    font-size: 9.5px;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    font-weight: 700;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  @media (max-width: 700px) {
    .cards {
      grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    }
  }
</style>
