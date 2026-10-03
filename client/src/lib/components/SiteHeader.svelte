<script lang="ts">
  // SiteHeader — the shared site nav (#1386, ADR 0092's dual-portal
  // experiment). One of two competing entries into the app; the other
  // is the #/home sitemap page. The owner will pick between them once
  // both exist, so this stays deliberately small: a wordmark, the
  // links, and an account control, reusing the session store and the
  // existing logout paths rather than inventing either.
  //
  // Session-aware: Lobby and Catalog need any session, "My games"
  // needs one tied to a Discord user — a guest or admin session would
  // otherwise land on a page that can only tell them to sign in.
  // Roadmap, Deck check and Home are public (ADR 0095 §5 adds the
  // second of those), so a signed-out visitor still sees all three
  // plus Sign in.
  //
  // Not shown on Game — the board has its own command bar, and a
  // second row of chrome above it would just eat table space.
  //
  // For a signed-in person the header is the portal (ADR 0112 §1 items
  // 4 and 5): the wordmark goes to the Lobby, the signed-in home, and
  // the account controls that used to live on the Lobby's own command
  // bar and the login page's signed-in card are one menu, opened from
  // the button that shows the person's name.
  import { onMount } from "svelte";
  import { route, navigate } from "../router";
  import { canSignOutEverywhere, LobbyApiError, session } from "../session";
  import { signedInUserID } from "../myGames";
  import {
    discordAuthEnabled,
    discordLoginHref,
    logout as apiLogout,
    logoutEverywhere as apiLogoutEverywhere,
  } from "../api";
  import { openSettings } from "../settings";
  import Icon from "./Icon.svelte";
  import AdminChip from "./AdminChip.svelte";

  let menuOpen = $state(false);
  let accountOpen = $state(false);
  let accountError = $state("");
  let accountEl = $state<HTMLElement | null>(null);

  const isSignedIn = $derived($session !== null);
  const hasLinkedUser = $derived(signedInUserID($session) !== null);

  // "Sign in with a different Discord account" needs Discord sign-in on
  // this server. Probed once, and only for a session that could use it.
  let discordEnabled = $state(false);
  onMount(() => {
    if (!hasLinkedUser) return;
    void discordAuthEnabled().then((on) => {
      discordEnabled = on;
    });
  });

  // Prefer the person's Discord display name; fall back to the
  // session's role (guest / admin / spectator) when there isn't one.
  const whoAmI = $derived.by(() => {
    const s = $session;
    if (!s) return "";
    const { name, role } = s.principal;
    return name && name !== role ? name : role;
  });

  // What kind of session this is, in a player's words: the line under
  // the name in the account menu. It replaces the Lobby's old role chip.
  const accountKind = $derived.by(() => {
    const s = $session;
    if (!s) return "";
    const { role } = s.principal;
    if (role === "admin") return "admin token";
    if (role === "identified" || hasLinkedUser) return "signed in with Discord";
    return role === "spectator" ? "guest spectator" : "guest seat";
  });

  // The wordmark is the way home: the Lobby for a session, the public
  // site map without one (ADR 0112 §1 item 5).
  const homeHref = $derived(isSignedIn ? "#/lobby" : "#/home");

  function closeMenu(): void {
    menuOpen = false;
  }

  function closeAccount(): void {
    accountOpen = false;
  }

  function toggleAccount(): void {
    accountError = "";
    accountOpen = !accountOpen;
    if (accountOpen) menuOpen = false;
  }

  // Escape closes the account menu and hands focus back to its button;
  // a click anywhere outside it closes it too.
  function onWindowKey(e: KeyboardEvent): void {
    if (e.key !== "Escape" || !accountOpen) return;
    closeAccount();
    accountEl?.querySelector<HTMLButtonElement>(".acct-btn")?.focus();
  }

  function onWindowClick(e: MouseEvent): void {
    if (!accountOpen || !accountEl) return;
    if (e.target instanceof Node && accountEl.contains(e.target)) return;
    closeAccount();
  }

  function current(name: string): boolean {
    return $route.name === name;
  }

  function settings(): void {
    closeAccount();
    openSettings();
  }

  async function signOut(): Promise<void> {
    closeMenu();
    closeAccount();
    await apiLogout();
    navigate("#/login");
  }

  // Sign out everywhere (ADR 0051 decision 6): every browser this
  // Discord account is signed in on, this one included. A failure stays
  // in the menu, so the person knows their other browsers are still
  // signed in.
  async function signOutEverywhere(): Promise<void> {
    accountError = "";
    try {
      await apiLogoutEverywhere();
      closeAccount();
      navigate("#/login");
    } catch (err) {
      accountError = err instanceof LobbyApiError ? err.message : "could not sign out everywhere";
    }
  }
</script>

<svelte:window onkeydown={onWindowKey} onclick={onWindowClick} />

<header class="site-header">
  <div class="row">
    <a class="wordmark" href={homeHref} aria-label="cmd_and_ctrl home" onclick={closeMenu}>
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
      {#if hasLinkedUser}
        <a
          href="#/decks"
          class:current={current("myDecks")}
          aria-current={current("myDecks") ? "page" : undefined}
          onclick={closeMenu}>My decks</a
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

    <div class="account" bind:this={accountEl}>
      {#if isSignedIn}
        <button
          type="button"
          class="ghost sm acct-btn"
          aria-haspopup="true"
          aria-expanded={accountOpen}
          aria-controls="account-menu"
          aria-label={`account menu: ${whoAmI}`}
          onclick={toggleAccount}
        >
          <span class="who">{whoAmI}</span>
          <span class="caret" class:open={accountOpen}><Icon name="chevronRight" size={12} /></span>
        </button>
        {#if accountOpen}
          <div class="acct-menu" id="account-menu" role="group" aria-label="account">
            <div class="acct-id">
              <span class="acct-name">{whoAmI}</span>
              <span class="acct-kind">{accountKind}</span>
            </div>
            <!-- ADR 0112 §2 item 9: the Admin chip for an allowlisted
                 person, or the token's static "Admin token" badge, first
                 in the menu. Renders nothing for anyone else. -->
            <AdminChip />
            <button type="button" class="acct-item" onclick={settings}>
              <Icon name="gear" size={14} /> Settings
            </button>
            {#if discordEnabled && hasLinkedUser}
              <!-- ADR 0110 §2 item 3: a repeat sign-in skips Discord's
                   screen and uses whichever account the browser is
                   signed in to. This asks for the screen, which has
                   Discord's own account switcher. -->
              <a class="acct-item" href={discordLoginHref({ consent: true })}>
                Sign in with a different Discord account
              </a>
            {/if}
            <div class="acct-sep" aria-hidden="true"></div>
            <button type="button" class="acct-item" onclick={signOut}>Sign out</button>
            {#if canSignOutEverywhere($session)}
              <button
                type="button"
                class="acct-item"
                title="sign out of every browser signed in with this Discord account"
                onclick={signOutEverywhere}>Sign out everywhere</button
              >
            {/if}
            {#if accountError}
              <p class="acct-error" role="alert">{accountError}</p>
            {/if}
          </div>
        {/if}
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
    position: relative;
    display: flex;
    align-items: center;
    gap: 10px;
    margin-left: auto;
  }
  .who {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.04em;
    color: var(--fg-muted);
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
  .acct-btn {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
  .caret {
    display: inline-flex;
    color: var(--fg-dim);
    transform: rotate(90deg);
    transition: transform 120ms ease;
  }
  .caret.open {
    transform: rotate(-90deg);
  }
  /* The account menu drops from the name button, right-aligned with
     it, over whatever page is underneath. */
  .acct-menu {
    position: absolute;
    top: calc(100% + 6px);
    right: 0;
    z-index: 40;
    min-width: 240px;
    max-width: calc(100vw - 28px);
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 6px;
    border-radius: 12px;
    border: 1px solid var(--border-strong);
    background: var(--surface-raised);
    box-shadow: var(--shadow-lg);
  }
  .acct-id {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 8px 10px 10px;
    margin-bottom: 4px;
    border-bottom: 1px solid var(--border);
  }
  .acct-name {
    font-size: 13px;
    font-weight: 700;
    color: var(--fg);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .acct-kind {
    font-family: var(--font-mono);
    font-size: 10.5px;
    letter-spacing: 0.06em;
    color: var(--fg-dim);
  }
  .acct-item {
    display: flex;
    align-items: center;
    justify-content: flex-start;
    gap: 8px;
    width: 100%;
    min-height: 34px;
    padding: 0 10px;
    border: 0;
    border-radius: var(--radius);
    background: transparent;
    color: var(--fg);
    font-size: 12.5px;
    font-weight: 600;
    text-align: left;
    text-decoration: none;
    box-sizing: border-box;
    cursor: pointer;
  }
  .acct-item:hover,
  .acct-item:focus-visible {
    background: var(--surface-hover);
  }
  .acct-sep {
    height: 1px;
    margin: 4px 6px;
    background: var(--border);
  }
  .acct-error {
    margin: 4px 10px 6px;
    font-size: 12px;
    color: var(--danger);
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
      /* After the account control, at the far right: the two used to
         split the free space and leave the toggle mid-row. */
      order: 3;
      margin-left: 0;
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
