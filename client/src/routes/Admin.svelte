<script lang="ts">
  // Admin: the admin views (ADR 0124 §7). One page with a tab strip
  // (`navigation "admin views"`: Live now, Games, Accounts) and one child
  // per view under lib/components/admin/: Live now, Games, one table,
  // Accounts and one account. The Grafana Overview's tiles link here.
  //
  // Who sees what:
  //   - signed out: the router sends the visitor to #/login, storing the
  //     view so the sign-in comes back to it (App.svelte);
  //   - an admin (the token, or an allowlisted person in admin mode):
  //     the view;
  //   - an allowlisted person in player mode: "Admin mode is off", with
  //     the same switch the account menu has, and no request at all;
  //   - anyone else: "This page is for admins.", and no request.
  //
  // The server is the gate. A view that hears 403 (a stale admin: true)
  // shows the same message and asks /me again (refreshAdminStatus).
  import SiteHeader from "../lib/components/SiteHeader.svelte";
  import AdminChip from "../lib/components/AdminChip.svelte";
  import LiveNow from "../lib/components/admin/LiveNow.svelte";
  import GamesList from "../lib/components/admin/GamesList.svelte";
  import GameDetail from "../lib/components/admin/GameDetail.svelte";
  import AccountsList from "../lib/components/admin/AccountsList.svelte";
  import AccountDetail from "../lib/components/admin/AccountDetail.svelte";
  import { isAdmin, isAdminAllowed, refreshAdminStatus } from "../lib/admin";
  import {
    accountsHash,
    adminViewHash,
    gamesHash,
    LIVE_HASH,
    tabOf,
    type AdminView,
  } from "../lib/adminViews";
  import { session } from "../lib/session";

  interface Props {
    view: AdminView;
  }
  const { view }: Props = $props();

  const admin = $derived(isAdmin($session));
  const allowed = $derived(isAdminAllowed($session));

  // A 403 is remembered for the session as it was when refused: a switch
  // of admin mode (a new end time) or another session clears it.
  const sessionKey = $derived(`${$session?.token ?? ""}|${$session?.admin_mode_ends_at ?? ""}`);
  let refusedFor = $state<string | null>(null);
  const refused = $derived(refusedFor !== null && refusedFor === sessionKey);
  const showData = $derived(admin && !refused);

  function onforbidden(): void {
    refusedFor = sessionKey;
    void refreshAdminStatus();
  }

  const tab = $derived(tabOf(view));
  const title = $derived(
    view.view === "live"
      ? "live now"
      : view.view === "games"
        ? "games"
        : view.view === "game"
          ? "one table"
          : view.view === "accounts"
            ? "accounts"
            : "one account",
  );

  const TABS = [
    { id: "live", label: "Live now", href: LIVE_HASH },
    { id: "games", label: "Games", href: gamesHash() },
    { id: "accounts", label: "Accounts", href: accountsHash() },
  ] as const;
</script>

<SiteHeader />

<section class="admin-page">
  <div class="head">
    <p class="eyebrow">Admin</p>
    <h1>{title}</h1>
  </div>

  {#if $session}
    {#if showData}
      <nav class="tabs" aria-label="admin views">
        {#each TABS as t (t.id)}
          <a
            href={t.href}
            class:current={tab === t.id}
            aria-current={tab === t.id ? "page" : undefined}>{t.label}</a
          >
        {/each}
      </nav>

      <!-- Keyed on the view's hash: a new filter, page or id is a new
           load, never a stale list under a new heading. -->
      {#key adminViewHash(view)}
        {#if view.view === "live"}
          <LiveNow {onforbidden} />
        {:else if view.view === "games"}
          <GamesList filter={view.filter} {onforbidden} />
        {:else if view.view === "game"}
          <GameDetail id={view.id} {onforbidden} />
        {:else if view.view === "accounts"}
          <AccountsList filter={view.filter} {onforbidden} />
        {:else}
          <AccountDetail id={view.id} {onforbidden} />
        {/if}
      {/key}
    {:else}
      <div class="card gate" role="status">
        {#if allowed && !admin}
          <p>Admin mode is off. Switch it on to see this page.</p>
          <div class="switch"><AdminChip /></div>
        {:else}
          <p>This page is for admins.</p>
        {/if}
        <p class="back"><a href="#/lobby">Back to the Lobby</a></p>
      </div>
    {/if}
  {/if}
</section>

<style>
  .admin-page {
    width: min(1080px, 100%);
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: 16px;
    box-sizing: border-box;
    min-width: 0;
  }
  .head {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .eyebrow {
    margin: 0;
    font-family: var(--font-mono);
    font-size: 10.5px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--fg-dim);
  }
  h1 {
    margin: 0;
    font-family: var(--font-display);
    font-size: 26px;
    font-weight: 800;
    letter-spacing: -0.02em;
    color: var(--fg);
  }
  .tabs {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    padding-bottom: 8px;
    border-bottom: 1px solid var(--border);
  }
  .tabs a {
    display: inline-flex;
    align-items: center;
    height: 32px;
    padding: 0 12px;
    border-radius: var(--radius);
    color: var(--fg-muted);
    font-size: 13px;
    font-weight: 600;
    text-decoration: none;
  }
  .tabs a:hover {
    background: var(--surface-hover);
    color: var(--fg);
  }
  .tabs a.current {
    color: var(--accent-strong);
    background: var(--accent-soft);
  }
  .gate p {
    margin: 0;
    font-size: 14px;
    color: var(--fg);
  }
  .gate .switch {
    margin-top: 10px;
    max-width: 260px;
  }
  .gate .back {
    margin-top: 12px;
    font-size: 12.5px;
  }
  .gate .back a {
    color: var(--fg-muted);
  }

  /* Shared by the views under lib/components/admin/: one look for every
     card, filter row, table, chip and notice on these pages. */
  .admin-page :global(.view) {
    display: flex;
    flex-direction: column;
    gap: 12px;
    min-width: 0;
  }
  .admin-page :global(.card) {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 14px;
    padding: 14px 16px;
    min-width: 0;
    box-sizing: border-box;
  }
  .admin-page :global(.thead) {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
  }
  .admin-page :global(.filters) {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-end;
    gap: 10px;
  }
  .admin-page :global(.filters label) {
    display: flex;
    flex-direction: column;
    gap: 3px;
    min-width: 0;
  }
  .admin-page :global(.filters label > span) {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--fg-dim);
  }
  .admin-page :global(.filters select),
  .admin-page :global(.filters input) {
    box-sizing: border-box;
    max-width: 100%;
    margin: 0;
    padding: 0.4rem 0.6rem;
    font-size: 12.5px;
    line-height: 1.3;
    font-family: inherit;
    background: var(--surface-sunken);
    color: var(--fg);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius);
  }
  .admin-page :global(.filters input:focus) {
    outline: none;
    border-color: var(--accent);
    box-shadow: 0 0 0 3px var(--accent-soft);
  }
  .admin-page :global(button.ghost.sm) {
    height: 30px;
    padding: 0 12px;
    font-size: 12px;
  }
  /* A table scrolls inside its own box, never the page (390px wide). */
  .admin-page :global(.scroll) {
    overflow-x: auto;
    border: 1px solid var(--border);
    border-radius: 12px;
    background: var(--surface);
  }
  .admin-page :global(table) {
    width: 100%;
    border-collapse: collapse;
    font-size: 13px;
  }
  .admin-page :global(th) {
    text-align: left;
    padding: 9px 12px;
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--fg-dim);
    border-bottom: 1px solid var(--border);
    white-space: nowrap;
  }
  .admin-page :global(td) {
    padding: 9px 12px;
    border-top: 1px solid var(--border);
    vertical-align: top;
    color: var(--fg);
  }
  .admin-page :global(tbody tr:first-child td) {
    border-top: none;
  }
  .admin-page :global(td.num),
  .admin-page :global(th.num) {
    text-align: right;
  }
  .admin-page :global(.tcell > div) {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px;
    min-width: 9rem;
  }
  .admin-page :global(.tname) {
    font-weight: 700;
    color: var(--fg);
    text-decoration: none;
    overflow-wrap: anywhere;
  }
  .admin-page :global(.tname:hover) {
    text-decoration: underline;
  }
  .admin-page :global(.chip) {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    padding: 2px 7px;
    border-radius: 999px;
    border: 1px solid var(--border-strong);
    color: var(--fg-muted);
    white-space: nowrap;
  }
  .admin-page :global(.chip.live) {
    color: var(--accent);
    border-color: var(--accent);
  }
  .admin-page :global(.chip.dim) {
    opacity: 0.7;
  }
  .admin-page :global(.chip.warn) {
    color: var(--priority-strong);
    border-color: var(--priority);
  }
  .admin-page :global(.dim) {
    color: var(--fg-muted);
  }
  .admin-page :global(.facts) {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
    gap: 10px 16px;
    margin: 12px 0 0;
  }
  .admin-page :global(.facts dt) {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--fg-dim);
  }
  .admin-page :global(.facts dd) {
    margin: 3px 0 0;
    font-size: 13px;
    color: var(--fg);
  }
  .admin-page :global(.empty) {
    margin: 0;
    padding: 18px;
    text-align: center;
    font-size: 13px;
    color: var(--fg-muted);
    border: 1px dashed var(--border-strong);
    border-radius: 12px;
  }
  .admin-page :global(.notice) {
    margin: 0;
    font-size: 13px;
  }
  .admin-page :global(.notice.err) {
    color: var(--danger);
  }
  .admin-page :global(.notice.ok) {
    margin-top: 10px;
    color: var(--mint);
  }
  .admin-page :global(.back) {
    margin: 0;
    font-size: 12.5px;
  }
  .admin-page :global(.back a) {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    color: var(--fg-muted);
    text-decoration: none;
  }
  .admin-page :global(.back a:hover) {
    color: var(--fg);
  }
  .admin-page :global(.pages) {
    display: flex;
    gap: 14px;
    font-size: 12.5px;
  }
  .admin-page :global(.pages a) {
    color: var(--accent-strong);
  }
  .admin-page :global(.btn-link) {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: 34px;
    padding: 0 14px;
    border-radius: var(--radius);
    border: 1px solid var(--border-strong);
    background: var(--surface-raised);
    color: var(--fg);
    font-size: 12.5px;
    font-weight: 600;
    text-decoration: none;
    box-sizing: border-box;
  }
  .admin-page :global(.btn-link:hover) {
    background: var(--surface-hover);
  }
  .admin-page :global(.confirm) {
    margin-top: 12px;
    padding: 12px 14px;
    border: 1px solid var(--border-strong);
    border-radius: 12px;
    background: var(--surface-raised);
  }
  .admin-page :global(.confirm p) {
    margin: 0;
    font-size: 13px;
    line-height: 1.5;
  }
  .admin-page :global(.confirm-actions) {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 10px;
  }

  /* Phone width: the less needed columns go, the rest stay readable. */
  @media (max-width: 600px) {
    .admin-page :global(th.wide),
    .admin-page :global(td.wide) {
      display: none;
    }
    .admin-page :global(.filters label) {
      flex: 1 1 40%;
    }
    .admin-page :global(.filters select),
    .admin-page :global(.filters input) {
      width: 100%;
    }
  }
</style>
