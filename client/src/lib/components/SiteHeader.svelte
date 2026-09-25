<script lang="ts">
  // SiteHeader — the shared site nav (#1386, ADR 0092's dual-portal
  // experiment). One of two competing entries into the app; the other
  // is the #/home sitemap page. The owner will pick between them once
  // both exist, so this stays deliberately small: a wordmark, six
  // links, and a sign in/out control, reusing the session store and
  // the existing logout path rather than inventing either.
  //
  // Session-aware: Lobby and Catalog need any session, "My games"
  // needs one tied to a Discord user (the same gate Lobby's own "my
  // games" button already uses) — a guest or admin session would
  // otherwise land on a page that can only tell them to sign in.
  // Roadmap, Deck check and Home are public (ADR 0095 §5 adds the
  // second of those), so a signed-out visitor still sees all three
  // plus Sign in.
  //
  // Not shown on Game — the board has its own command bar, and a
  // second row of chrome above it would just eat table space.
  import { route, navigate } from "../router";
  import { session } from "../session";
  import { signedInUserID } from "../myGames";
  import { logout as apiLogout } from "../api";
  import Icon from "./Icon.svelte";

  let menuOpen = $state(false);

  const isSignedIn = $derived($session !== null);
  const hasLinkedUser = $derived(signedInUserID($session) !== null);

  // Prefer the person's Discord display name; fall back to the
  // session's role (guest / admin / spectator) when there isn't one —
  // the same rule Lobby's own header chip uses.
  const whoAmI = $derived.by(() => {
    const s = $session;
    if (!s) return "";
    const { name, role } = s.principal;
    return name && name !== role ? name : role;
  });

  function closeMenu(): void {
    menuOpen = false;
  }

  function current(name: string): boolean {
    return $route.name === name;
  }

  async function signOut(): Promise<void> {
    closeMenu();
    await apiLogout();
    navigate("#/login");
  }
</script>

<header class="site-header">
  <div class="row">
    <a class="wordmark" href="#/home" aria-label="cmd_and_ctrl home" onclick={closeMenu}>
      <i></i>CMD &amp; CTRL
    </a>

    <button
      type="button"
      class="menu-toggle"
      aria-expanded={menuOpen}
      aria-controls="site-nav"
      onclick={() => (menuOpen = !menuOpen)}
    >
      <Icon
        name={menuOpen ? "x" : "menu"}
        size={18}
        label={menuOpen ? "close menu" : "open menu"}
      />
    </button>

    <nav class="site-nav" id="site-nav" aria-label="Site" class:open={menuOpen}>
      {#if isSignedIn}
        <a
          href="#/lobby"
          class:current={current("lobby")}
          aria-current={current("lobby") ? "page" : undefined}
          onclick={closeMenu}>Lobby</a
        >
      {/if}
      {#if hasLinkedUser}
        <a
          href="#/my-games"
          class:current={current("myGames")}
          aria-current={current("myGames") ? "page" : undefined}
          onclick={closeMenu}>My games</a
        >
      {/if}
      {#if isSignedIn}
        <a
          href="#/catalog"
          class:current={current("catalog")}
          aria-current={current("catalog") ? "page" : undefined}
          onclick={closeMenu}>Catalog</a
        >
      {/if}
      <a
        href="#/roadmap"
        class:current={current("roadmap")}
        aria-current={current("roadmap") ? "page" : undefined}
        onclick={closeMenu}>Roadmap</a
      >
      <a
        href="#/deck-check"
        class:current={current("deckCheck")}
        aria-current={current("deckCheck") ? "page" : undefined}
        onclick={closeMenu}>Deck check</a
      >
      <a
        href="#/home"
        class:current={current("home")}
        aria-current={current("home") ? "page" : undefined}
        onclick={closeMenu}>Home</a
      >
    </nav>

    <div class="account">
      {#if isSignedIn}
        <span class="who">{whoAmI}</span>
        <button type="button" class="ghost sm" onclick={signOut}>Sign out</button>
      {:else}
        <a class="ghost-link" href="#/login">Sign in</a>
      {/if}
    </div>
  </div>
</header>

<style>
  /* Bleeds to the edge of #app's 1080px column, the same negative-
     margin convention Lobby's own `.bar` uses — this replaces that
     bar's job on every page except Lobby, which keeps its own for
     now (the dual-header overlap is deliberate; see ADR 0092). */
  .site-header {
    margin: -1.5rem -1.5rem 20px;
    padding: 0 14px;
    background: var(--bg-1);
    border-bottom: 1px solid var(--border);
  }
  .row {
    min-height: 48px;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 14px;
    padding: 6px 0;
  }
  .wordmark {
    font-family: var(--font-display);
    font-weight: 800;
    font-size: 13px;
    letter-spacing: 0.16em;
    color: var(--fg);
    display: inline-flex;
    align-items: center;
    gap: 8px;
    text-decoration: none;
    white-space: nowrap;
  }
  .wordmark i {
    display: inline-block;
    width: 12px;
    height: 12px;
    border: 2px solid var(--gold);
    transform: rotate(45deg);
    border-radius: 3px;
    box-sizing: border-box;
  }

  .menu-toggle {
    display: none;
    margin-left: auto;
    padding: 0.4rem 0.55rem;
  }

  .site-nav {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 4px;
    margin-right: auto;
  }
  .site-nav a {
    display: inline-flex;
    align-items: center;
    height: 30px;
    padding: 0 10px;
    border-radius: var(--radius);
    color: var(--fg-muted);
    font-size: 12.5px;
    font-weight: 600;
    letter-spacing: 0.01em;
    text-decoration: none;
    white-space: nowrap;
  }
  .site-nav a:hover {
    background: var(--surface-hover);
    color: var(--fg);
  }
  .site-nav a.current {
    color: var(--gold-strong);
    background: var(--accent-soft);
  }

  .account {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-left: auto;
  }
  .who {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.04em;
    color: var(--fg-dim);
    white-space: nowrap;
    max-width: 22ch;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  button.ghost.sm {
    height: 28px;
    padding: 0 10px;
    font-size: 11.5px;
  }
  .ghost-link {
    display: inline-flex;
    align-items: center;
    height: 28px;
    padding: 0 10px;
    border-radius: var(--radius);
    color: var(--gold-strong);
    font-size: 12px;
    font-weight: 600;
    text-decoration: none;
    white-space: nowrap;
  }
  .ghost-link:hover {
    background: var(--accent-soft);
  }

  /* Phone width and below: the wordmark and the account control stay
     put (sign in/out is one tap, never buried), the links collapse
     behind the toggle. */
  @media (max-width: 700px) {
    .menu-toggle {
      display: inline-flex;
    }
    .site-nav {
      display: none;
      order: 10;
      width: 100%;
      margin: 0;
      flex-direction: column;
      align-items: stretch;
      gap: 2px;
      padding: 6px 0 10px;
      border-top: 1px solid var(--border);
    }
    .site-nav.open {
      display: flex;
    }
    .site-nav a {
      height: 36px;
      padding: 0 12px;
    }
  }
</style>
