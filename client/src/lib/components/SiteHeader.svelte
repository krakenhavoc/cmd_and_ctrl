<script lang="ts">
  // SiteHeader — the site nav on every page outside the table (ADR 0156
  // §3 item 3). Page.svelte draws it at the top of the page frame.
  //
  // Grouped, not flat:
  //   - Signed in: the wordmark, then Tables | Decks, Catalog and
  //     Roadmap. (My games is Tables' History since ADR 0158 §1.) On the right the CMD & CTRL Discord, Help, and the
  //     account menu. In admin mode an accent-tinted Admin link follows
  //     as its own group, and the account menu keeps its Admin item.
  //   - Signed out: the wordmark, then Decks, Roadmap and About. On the
  //     right the CMD & CTRL Discord, Help, and Sign in.
  //
  // "Tables" is the Lobby (the route stays #/lobby); its name is a
  // contract label (L.navTables). The Discord link shows only when the
  // server names an invite (GET /community, lib/community.ts).
  //
  // Not shown on Game: the board has its own command bar.
  //
  // For a signed-in person the header is the portal (ADR 0112 §1 items
  // 4 and 5): the wordmark goes to the Lobby, the signed-in home, and
  // the account controls are one menu, opened from the button that
  // shows the person's name.
  //
  // Below 600px (a public page's phone layout, #3877) the header is one
  // row: the wordmark, Help, Sign in or the account button, and a menu
  // button. The nav links, the CMD & CTRL Discord and the bug report move into
  // a panel the menu button opens (a disclosure: aria-expanded, focus
  // goes into it, Escape closes it and hands focus back). From 600px up
  // the links sit in the row as before and the button is hidden. The
  // table and the app screens stay tablet-minimum; only the site frame
  // has a phone layout.
  import { onMount, tick } from "svelte";
  import Wordmark from "./Wordmark.svelte";
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
  import ServerPing from "./ServerPing.svelte";
  import HelpMenu from "./HelpMenu.svelte";
  import BugReportModal from "./BugReportModal.svelte";
  import BugReportDraftPill from "./BugReportDraftPill.svelte";
  import { currentBugDraft } from "../bugReportDraft.svelte";
  import { bugReportConfig } from "../bugReportConfig";
  import DeleteAccountDialog from "./DeleteAccountDialog.svelte";
  import { PRIVACY_HASH } from "../privacy";
  import { ABOUT_HASH, GUIDE_HASH } from "../about";
  import { isAdmin } from "../admin";
  import { community } from "../community";
  import { L } from "../labels";

  let accountOpen = $state(false);
  let accountError = $state("");
  // "Delete my account" (#2983, ADR 0149) opens a confirmation dialog.
  let deleteOpen = $state(false);
  let accountEl = $state<HTMLElement | null>(null);
  // The phone layout's nav panel and the button that opens it.
  let menuOpen = $state(false);
  let navEl = $state<HTMLElement | null>(null);
  let menuBtnEl = $state<HTMLButtonElement | null>(null);

  const isSignedIn = $derived($session !== null);
  const hasLinkedUser = $derived(signedInUserID($session) !== null);
  const invite = $derived($community.invite_url);

  // Whether this server has Discord sign-in. Probed once, for a session
  // that could use "Sign in with a different Discord account", and for a
  // signed-out visitor, whose Sign in starts Discord sign-in directly
  // (the hero's OAuth start) and otherwise goes to the front page.
  let discordEnabled = $state(false);
  onMount(() => {
    if (isSignedIn && !hasLinkedUser) return;
    void discordAuthEnabled().then((on) => {
      discordEnabled = on;
    });
  });

  // Report a bug on every page outside the table (#3492), for someone
  // signed in with Discord: the report names them, and a guest's
  // reports have nobody to follow up with. The table keeps its own
  // button, which attaches the game. Shown only when this server files
  // reports at all.
  let bugReportOn = $state(false);
  let bugAttachments = $state(false);
  let bugOpen = $state(false);
  // Minimize (#4029): the draft outlives the form, and this pill, beside
  // the bug button, opens it again. Focus goes to the pill that replaces
  // the form.
  let bugDraftPill = $state<BugReportDraftPill | null>(null);
  function minimizeBugReport(): void {
    bugOpen = false;
    void tick().then(() => bugDraftPill?.focus());
  }
  $effect(() => {
    if (!hasLinkedUser) return;
    void bugReportConfig().then((cfg) => {
      bugReportOn = cfg.enabled;
      bugAttachments = cfg.attachments;
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
  // the name in the account menu.
  const accountKind = $derived.by(() => {
    const s = $session;
    if (!s) return "";
    const { role } = s.principal;
    if (role === "admin") return "admin token";
    if (role === "identified" || hasLinkedUser) return "signed in with Discord";
    return role === "spectator" ? "guest spectator" : "guest seat";
  });

  // ADR 0112 follow-up (#1992): admin mode is visible on the button
  // itself, so it can be seen without opening the menu. isAdmin is true
  // for the token and for an allowlisted person whose mode is on; the
  // lapse timer in lib/admin.ts clears it on the same store.
  const adminOn = $derived(isAdmin($session));
  const acctLabel = $derived(`account menu: ${whoAmI}${adminOn ? " (admin mode on)" : ""}`);

  // The wordmark is the way home: the Lobby for a session, the
  // front page without one (ADR 0112 §1 item 5, ADR 0156 §4).
  const homeHref = $derived(isSignedIn ? "#/lobby" : "#/");

  function closeAccount(): void {
    accountOpen = false;
  }

  // Opening the menu moves focus into it, so a keyboard user is not left
  // on the button with the items a Tab away behind other controls.
  async function toggleAccount(): Promise<void> {
    accountError = "";
    menuOpen = false;
    accountOpen = !accountOpen;
    if (!accountOpen) return;
    await tick();
    accountEl?.querySelector<HTMLElement>(".acct-menu .acct-item")?.focus();
  }

  // The phone layout's nav panel: opening it moves focus to its first
  // link, for the same reason as the account menu.
  async function toggleMenu(): Promise<void> {
    menuOpen = !menuOpen;
    if (!menuOpen) return;
    accountOpen = false;
    await tick();
    navEl?.querySelector<HTMLElement>("a, button")?.focus();
  }

  // A page change closes the panel, so it does not hang over the next page.
  $effect(() => {
    void $route;
    menuOpen = false;
  });

  // Focus leaving the panel for anywhere but its own button closes it.
  function onNavFocusOut(e: FocusEvent): void {
    const to = e.relatedTarget as Node | null;
    if (!menuOpen || !to || !navEl || navEl.contains(to) || to === menuBtnEl) return;
    menuOpen = false;
  }

  function openBugReport(): void {
    menuOpen = false;
    bugOpen = true;
  }

  // The skip link (first focusable thing in the shell). A plain
  // href="#main" would rewrite the hash and re-route, so it focuses the
  // page's <main> itself.
  function skipToMain(e: MouseEvent): void {
    const main = document.getElementById("main");
    if (!main) return;
    e.preventDefault();
    main.setAttribute("tabindex", "-1");
    main.focus();
    main.scrollIntoView({ block: "start" });
  }

  // Escape closes the account menu and hands focus back to its button;
  // a click anywhere outside it closes it too.
  function onWindowKey(e: KeyboardEvent): void {
    if (e.key !== "Escape") return;
    if (menuOpen) {
      menuOpen = false;
      menuBtnEl?.focus();
      return;
    }
    if (!accountOpen) return;
    closeAccount();
    accountEl?.querySelector<HTMLButtonElement>(".acct-btn")?.focus();
  }

  function onWindowClick(e: MouseEvent): void {
    const target = e.target instanceof Node ? e.target : null;
    // The panel closes on a click outside it and its button, and on a
    // click on one of its links (a link to the page already open changes
    // no route, so the route effect would not close it).
    if (menuOpen && target && !menuBtnEl?.contains(target)) {
      const inLink = target instanceof Element && target.closest("a") !== null;
      if (!navEl?.contains(target) || inLink) menuOpen = false;
    }
    if (!accountOpen || !accountEl) return;
    if (target && accountEl.contains(target)) return;
    closeAccount();
  }

  const onAdminRoute = $derived($route.name === "adminViews" || $route.name === "adminLogin");

  function current(name: string): boolean {
    return $route.name === name;
  }

  function settings(): void {
    closeAccount();
    openSettings();
  }

  function openDelete(): void {
    closeAccount();
    deleteOpen = true;
  }

  async function signOut(): Promise<void> {
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
      accountError =
        err instanceof LobbyApiError
          ? err.message
          : "Couldn’t sign out everywhere. Check your connection and try again.";
    }
  }
</script>

<svelte:window onkeydown={onWindowKey} onclick={onWindowClick} />

{#snippet navLink(href: string, name: string, text: string)}
  <a {href} class:current={current(name)} aria-current={current(name) ? "page" : undefined}
    >{text}</a
  >
{/snippet}

<a class="skip-link" href="#main" onclick={skipToMain}>Skip to main content</a>

<header class="site-header">
  <div class="row">
    <a class="wordmark" href={homeHref} aria-label="cmd_and_ctrl home">
      <Wordmark size="header" />
    </a>

    <nav
      class="site-nav"
      class:open={menuOpen}
      id="site-nav"
      aria-label="Site"
      bind:this={navEl}
      onfocusout={onNavFocusOut}
    >
      {#if isSignedIn}
        <!-- Play: your tables now, and the games you have played (ADR
             0158 §1: My games is Tables' History). -->
        <div class="grp">
          {@render navLink("#/lobby", "lobby", L.navTables)}
          {#if hasLinkedUser}
            {@render navLink("#/stats", "stats", "Stats")}
          {/if}
        </div>
        <!-- Look things up: decks, the cards the engine plays, what's next. -->
        <div class="grp">
          {@render navLink("#/decks", "decks", "Decks")}
          {@render navLink("#/catalog", "catalog", "Catalog")}
          {@render navLink("#/roadmap", "roadmap", "Roadmap")}
        </div>
        {#if adminOn}
          <!-- Admin mode is loud: its own group, in the accent. The same
               condition as the account menu's Admin item (#3221). -->
          <div class="grp">
            <a
              class="admin-link"
              href="#/admin/live"
              class:current={onAdminRoute}
              aria-current={onAdminRoute ? "page" : undefined}>Admin</a
            >
          </div>
        {/if}
      {:else}
        <div class="grp">
          {@render navLink("#/decks", "decks", "Decks")}
          {@render navLink("#/catalog", "catalog", "Catalog")}
          {@render navLink("#/roadmap", "roadmap", "Roadmap")}
          {@render navLink(GUIDE_HASH, "guide", "Guide")}
          {@render navLink(ABOUT_HASH, "about", "About")}
        </div>
      {/if}
      <!-- Phone layout only (#3877): the CMD & CTRL Discord and the bug report as
           labelled rows, since their icon buttons leave the row below
           600px. Hidden from 600px up, where the icons show. -->
      {#if invite || (hasLinkedUser && bugReportOn)}
        <div class="panel-util">
          {#if invite}
            <a href={invite} target="_blank" rel="noopener noreferrer">
              <Icon name="discord" size={18} /> CMD & CTRL Discord
            </a>
          {/if}
          {#if hasLinkedUser && bugReportOn}
            <button type="button" onclick={openBugReport}>
              <Icon name="bug" size={17} />
              {L.reportABug}
            </button>
          {/if}
        </div>
      {/if}
    </nav>

    <div class="account">
      <span class="ping-slot"><ServerPing /></span>
      {#if invite}
        <!-- ADR 0156 §6: the CMD & CTRL Discord on every page outside the table.
             Hidden when the server names no invite. -->
        <a
          class="discord-link"
          href={invite}
          target="_blank"
          rel="noopener noreferrer"
          aria-label="the CMD & CTRL Discord server"
          title="Join the CMD & CTRL Discord server"
        >
          <Icon name="discord" size={18} />
        </a>
      {/if}
      {#if hasLinkedUser && bugReportOn}
        {#if currentBugDraft() && !bugOpen}
          <BugReportDraftPill bind:this={bugDraftPill} onrestore={() => (bugOpen = true)} />
        {/if}
        <button
          type="button"
          class="bug-btn"
          aria-label={L.reportABug}
          title="Report a bug or suggest an improvement"
          onclick={() => (bugOpen = true)}
        >
          <Icon name="bug" size={17} />
        </button>
      {/if}
      <!-- Help (ADR 0125 §6), for everyone: tips, a practice game, the keymap. -->
      <HelpMenu />
      <div class="acct" bind:this={accountEl}>
        {#if isSignedIn}
          <button
            type="button"
            class="acct-btn"
            aria-expanded={accountOpen}
            aria-controls="account-menu"
            class:admin-on={adminOn}
            aria-label={acctLabel}
            title={adminOn ? "admin mode on" : undefined}
            onclick={toggleAccount}
          >
            {#if adminOn}<span class="admin-dot" aria-hidden="true"></span>{/if}
            <span class="who">{whoAmI}</span>
            <span class="caret" class:open={accountOpen}
              ><Icon name="chevronRight" size={12} /></span
            >
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
              {#if adminOn}
                <!-- The admin views (ADR 0124 §7), for an admin only (the
                     token, or admin mode on). In the account menu since
                     ADR 0156 §3; it goes when admin mode lapses. -->
                <a class="acct-item" href="#/admin/live" onclick={closeAccount}>
                  <Icon name="flag" size={14} /> Admin
                </a>
              {/if}
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
              <div class="acct-sep" aria-hidden="true"></div>
              <a class="acct-item" href={PRIVACY_HASH} onclick={closeAccount}>
                Privacy and your data
              </a>
              {#if hasLinkedUser}
                <button type="button" class="acct-item acct-danger" onclick={openDelete}>
                  Delete my account
                </button>
              {/if}
              {#if accountError}
                <p class="acct-error" role="alert">{accountError}</p>
              {/if}
            </div>
          {/if}
        {:else}
          <a class="btn sm ghost sign-in" href={discordEnabled ? discordLoginHref() : "#/"}
            >Sign in</a
          >
        {/if}
      </div>
    </div>

    <button
      type="button"
      class="menu-btn"
      class:on={menuOpen}
      bind:this={menuBtnEl}
      aria-label={L.siteMenu}
      aria-expanded={menuOpen}
      aria-controls="site-nav"
      onclick={() => void toggleMenu()}
    >
      <Icon name={menuOpen ? "x" : "menu"} size={20} />
    </button>
  </div>
</header>

{#if deleteOpen}
  <DeleteAccountDialog onclose={() => (deleteOpen = false)} />
{/if}

{#if bugOpen}
  <BugReportModal
    gameID=""
    view={null}
    seq={0}
    connection=""
    wsLog={[]}
    attachments={bugAttachments}
    page={$route.name}
    onclose={() => (bugOpen = false)}
    onminimize={minimizeBugReport}
  />
{/if}

<style>
  /* Edge to edge, its row centred at the page frame's wide width. Sticks
     to the top so the nav is a glance away on a long page. */
  .site-header {
    position: sticky;
    top: 0;
    z-index: 30;
    background: color-mix(in srgb, var(--bg) 86%, transparent);
    backdrop-filter: blur(10px);
    border-bottom: 1px solid var(--border);
  }
  /* Magic's five colours, W U B R G, along the top edge (the colour-
     identity tokens). Hard stops, so it is five flat bands, not a
     gradient. */
  .site-header::before {
    content: "";
    display: block;
    height: 3px;
    background: linear-gradient(
      90deg,
      var(--identity-w) 0 20%,
      var(--identity-u) 20% 40%,
      var(--identity-b) 40% 60%,
      var(--identity-r) 60% 80%,
      var(--identity-g) 80% 100%
    );
  }
  .ping-slot {
    display: inline-flex;
    margin-right: 4px;
  }
  /* At phone width the header needs the room more than the ping. */
  @media (max-width: 599px) {
    .ping-slot {
      display: none;
    }
  }
  .row {
    max-width: 1200px;
    min-height: 60px;
    margin: 0 auto;
    padding: 8px var(--page-gutter);
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px 24px;
    box-sizing: content-box;
  }
  .wordmark {
    display: inline-flex;
    align-items: center;
    text-decoration: none;
    white-space: nowrap;
    border-radius: var(--radius);
  }

  .site-nav {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 4px;
    margin-right: auto;
  }
  /* The groups: a hairline between them, none before the first. */
  .grp {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 2px;
  }
  .grp + .grp {
    margin-left: 10px;
    padding-left: 14px;
    border-left: 1px solid var(--border);
  }
  .site-nav a {
    display: inline-flex;
    align-items: center;
    height: 34px;
    padding: 0 11px;
    border-radius: var(--radius);
    color: var(--fg-muted);
    font-size: 14px;
    font-weight: 600;
    text-decoration: none;
    white-space: nowrap;
  }
  .site-nav a:hover {
    background: var(--surface-hover);
    color: var(--fg);
  }
  .site-nav a.current {
    color: var(--fg);
    background: var(--surface-raised);
  }

  /* Admin mode, visible from the nav: an accent-tinted pill. */
  .site-nav a.admin-link {
    color: var(--accent-strong);
    background: var(--accent-soft);
    border: 1px solid var(--accent-line);
  }
  .site-nav a.admin-link:hover,
  .site-nav a.admin-link.current {
    color: var(--accent-strong);
    background: color-mix(in srgb, var(--accent) 28%, transparent);
    border-color: var(--accent);
  }

  .account {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-left: auto;
  }
  .discord-link {
    display: inline-grid;
    place-items: center;
    width: 34px;
    height: 34px;
    border-radius: var(--radius);
    color: #aab3ff;
    text-decoration: none;
  }
  .discord-link:hover {
    color: #c9cfff;
    background: color-mix(in srgb, #5865f2 18%, transparent);
  }
  .discord-link:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
  .bug-btn {
    display: inline-grid;
    place-items: center;
    width: 34px;
    height: 34px;
    padding: 0;
    border: 0;
    border-radius: var(--radius);
    background: transparent;
    color: var(--fg-muted);
    cursor: pointer;
  }
  .bug-btn:hover {
    color: var(--fg);
    background: var(--overlay);
  }
  .bug-btn:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
  .acct {
    position: relative;
    display: flex;
    align-items: center;
  }
  /* The name, in a pill. */
  .acct-btn {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: 34px;
    padding: 0 10px 0 12px;
    border-radius: 999px;
    border: 1px solid var(--border-strong);
    background: transparent;
    color: var(--fg);
    font-size: 13px;
    font-weight: 600;
  }
  .acct-btn:hover {
    background: var(--surface-hover);
  }
  .acct-btn.admin-on {
    border-color: var(--accent);
  }
  .who {
    white-space: nowrap;
    max-width: 22ch;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .admin-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--accent);
    box-shadow: 0 0 0 2px var(--accent-soft);
    flex: none;
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
  .acct-danger {
    color: var(--danger);
  }
  .acct-error {
    margin: 4px 10px 6px;
    font-size: 12px;
    color: var(--danger);
  }

  /* The phone layout's controls, hidden from 600px up. */
  .menu-btn,
  .panel-util {
    display: none;
  }

  /* Public pages at phone width (#3877): one row, the wordmark and the
     controls, about 60px with the colour strip. The nav links, the C2
     Discord and the bug report move into a panel under the row, opened
     by the menu button. Tap targets are 44px. The one phone breakpoint
     the table's own rules also use (599px); DESIGN.md §5. */
  @media (max-width: 599px) {
    .row {
      flex-wrap: nowrap;
      min-height: 44px;
      padding-top: 6px;
      padding-bottom: 6px;
      gap: 4px;
    }
    .wordmark {
      flex: none;
      min-height: 44px;
      margin-right: auto;
    }
    .account {
      flex: 0 1 auto;
      min-width: 0;
      margin-left: 0;
      gap: 4px;
    }
    /* Their rows in the panel stand in for these icons. */
    .account .discord-link,
    .account .bug-btn {
      display: none;
    }
    .account .sign-in {
      height: 44px;
      padding: 0 14px;
      font-size: 14px;
    }
    /* The account menu and Help's menu hang from the header's right edge,
       not from their buttons, which sit too far left to hold them. */
    .acct {
      position: static;
      min-width: 0;
    }
    .acct-btn {
      height: 44px;
      min-width: 0;
      max-width: 100%;
    }
    .who {
      min-width: 0;
      max-width: 16ch;
    }
    .acct-menu {
      top: calc(100% + 4px);
      right: var(--page-gutter);
      max-width: calc(100vw - 2 * var(--page-gutter));
    }
    .acct-item {
      min-height: 44px;
    }
    .menu-btn {
      display: inline-grid;
      place-items: center;
      flex: none;
      width: 44px;
      height: 44px;
      padding: 0;
      border: 1px solid transparent;
      border-radius: var(--radius);
      background: transparent;
      color: var(--fg);
      cursor: pointer;
    }
    .menu-btn:hover,
    .menu-btn.on {
      background: var(--surface-hover);
      border-color: var(--border);
    }
    .menu-btn:focus-visible {
      outline: 2px solid var(--accent);
      outline-offset: 2px;
    }
    /* Under the sticky header, edge to edge. */
    .site-nav {
      display: none;
      position: absolute;
      top: 100%;
      left: 0;
      right: 0;
      flex-direction: column;
      align-items: stretch;
      gap: 0;
      margin: 0;
      padding: 6px var(--page-gutter) 10px;
      max-height: calc(100dvh - 64px);
      overflow-y: auto;
      overscroll-behavior: contain;
      background: var(--surface-raised);
      border-bottom: 1px solid var(--border-strong);
      box-shadow: var(--shadow);
    }
    .site-nav.open {
      display: flex;
    }
    .grp {
      flex-direction: column;
      align-items: stretch;
      gap: 0;
    }
    .grp + .grp,
    .panel-util {
      margin: 6px 0 0;
      padding: 6px 0 0;
      border-left: 0;
      border-top: 1px solid var(--border);
    }
    .panel-util {
      display: flex;
      flex-direction: column;
    }
    .site-nav a,
    .panel-util button {
      height: 48px;
      padding: 0 12px;
      font-size: 16px;
    }
    .panel-util button {
      display: flex;
      align-items: center;
      gap: 8px;
      border: 0;
      border-radius: var(--radius);
      background: transparent;
      color: var(--fg-muted);
      font-weight: 600;
      text-align: left;
      cursor: pointer;
    }
    .panel-util a {
      gap: 8px;
    }
    .panel-util button:hover {
      background: var(--surface-hover);
      color: var(--fg);
    }
  }
  /* At 359px and under the wordmark drops its keys for the mark alone
     (its name stays for a screen reader), to keep the row to one line. */
  @media (max-width: 359px) {
    .wordmark {
      min-width: 44px;
    }
    .wordmark :global(.keys) {
      display: none;
    }
  }
</style>
