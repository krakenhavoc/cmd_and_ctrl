<script lang="ts">
  import Wordmark from "../lib/components/Wordmark.svelte";
  import { adminLogin } from "../lib/api";
  import { navigate } from "../lib/router";
  import { expiryNotice, LobbyApiError, session } from "../lib/session";
  import { signedInUserID } from "../lib/myGames";
  import { isAdminViewReturn, takeAfterSignIn } from "../lib/decksPage";
  import Icon from "../lib/components/Icon.svelte";

  // AdminLogin is #/admin, the shared token's page and the only place
  // its form appears (ADR 0112 §2 item 8). It was the sign-in page's
  // admin mode; the sign-in page itself became the front page (ADR 0156
  // §4, routes/FrontPage.svelte) and this form is unchanged.
  //
  // The token's own session never sees it (the router sends it on); any
  // other session does, and is told the token replaces it.

  let token = $state("");
  let error = $state("");
  let busy = $state(false);

  // A signed-in person's session is set aside, not lost (ADR 0110 §1
  // item 6), and comes back when the token's session ends.
  const replaces = $derived($session !== null);
  const setsAside = $derived(signedInUserID($session) !== null);

  async function submit(e: SubmitEvent): Promise<void> {
    e.preventDefault();
    error = "";
    busy = true;
    try {
      await adminLogin(token);
      // An admin view opened signed out (a Grafana link, ADR 0124 §7)
      // comes back after the token's sign-in too; otherwise the Lobby.
      const back = takeAfterSignIn();
      navigate(isAdminViewReturn(back) ? back : "#/lobby");
    } catch (err) {
      error = err instanceof LobbyApiError ? err.message : "login failed";
    } finally {
      busy = false;
    }
  }
</script>

<section class="entry">
  <div class="fan" aria-hidden="true">
    <i></i><i></i><i></i><i></i><i></i>
  </div>
  <div class="stack">
    <div class="brand">
      <h1 class="wm"><Wordmark size="hero" /></h1>
      <p class="tagline">The Commander table, over the wire.</p>
    </div>

    {#if $expiryNotice}
      <p class="notice" role="status"><Icon name="undo" size={14} /> {$expiryNotice}</p>
    {/if}

    <div class="card">
      <h2>Admin log in</h2>
      <form class="frow" onsubmit={submit}>
        <input
          class="mono"
          type="password"
          placeholder="admin token"
          bind:value={token}
          autocomplete="current-password"
          required
        />
        <button type="submit" class="primary lg" disabled={busy || !token}>
          {busy ? "…" : "log in"}
        </button>
      </form>
      <p class="help">
        The shared admin token from <span class="mono">CMDCTRL_ADMIN_TOKEN</span>. Players never
        need this — they arrive through an invite link.
      </p>
      {#if replaces}
        <p class="help replaces" role="note">
          Logging in with the token replaces this browser's session.
          {#if setsAside}
            Your Discord sign-in is set aside, and comes back when the token's session ends.
          {/if}
        </p>
      {/if}
      {#if error}
        <p class="error" role="alert">{error}</p>
      {/if}
    </div>

    <p class="foot">
      <a class="ghost-link" href={$session ? "#/lobby" : "#/"}
        ><Icon name="chevronLeft" size={12} /> Back</a
      >
    </p>
  </div>
</section>

<style>
  .entry {
    position: relative;
    min-height: calc(100vh - 3rem);
    display: flex;
    justify-content: center;
    overflow: hidden;
    margin: -1.5rem;
    padding: 1.5rem;
  }
  /* Five card backs fanned behind the stack — the only decoration. */
  .fan {
    position: absolute;
    left: 50%;
    bottom: -150px;
    transform: translateX(-50%);
    width: 900px;
    height: 420px;
    pointer-events: none;
    opacity: 0.35;
  }
  .fan i {
    position: absolute;
    left: 50%;
    bottom: 0;
    width: 210px;
    height: 294px;
    border-radius: 12px;
    border: 1px solid color-mix(in srgb, var(--accent) 25%, transparent);
    background:
      radial-gradient(60% 50% at 50% 40%, var(--accent-soft), transparent 70%), var(--bg-2);
    transform-origin: 50% 120%;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.6);
  }
  .fan i:nth-child(1) {
    transform: translateX(-50%) rotate(-16deg);
  }
  .fan i:nth-child(2) {
    transform: translateX(-50%) rotate(-8deg);
  }
  .fan i:nth-child(3) {
    transform: translateX(-50%);
  }
  .fan i:nth-child(4) {
    transform: translateX(-50%) rotate(8deg);
  }
  .fan i:nth-child(5) {
    transform: translateX(-50%) rotate(16deg);
  }
  .stack {
    position: relative;
    width: min(460px, 100%);
    display: flex;
    flex-direction: column;
    gap: 22px;
    margin-top: clamp(40px, 14vh, 150px);
  }
  .brand {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    text-align: center;
  }
  h1.wm {
    margin: 0;
  }
  .tagline {
    margin: 0;
    font-size: 14px;
    color: var(--fg-muted);
  }
  .card {
    background: var(--surface);
    border: 1px solid var(--border-strong);
    border-radius: 16px;
    padding: 22px 24px 24px;
    display: flex;
    flex-direction: column;
    gap: 14px;
    box-shadow: var(--shadow-lg);
  }
  .card h2 {
    margin: 0;
    font-family: var(--font-display);
    font-size: 17px;
    font-weight: 700;
    color: var(--fg);
    text-transform: none;
    letter-spacing: -0.01em;
  }
  .frow {
    display: flex;
    gap: 8px;
  }
  .frow input {
    flex: 1;
    min-width: 0;
    margin: 0;
    height: 40px;
    padding: 0 12px;
    box-sizing: border-box;
    font-size: 13.5px;
  }
  .frow input.mono {
    font-family: var(--font-mono);
    font-size: 12.5px;
  }
  .lg {
    height: 40px;
    padding: 0 16px;
    font-size: 13.5px;
    flex: 0 0 auto;
  }
  .help {
    margin: 0;
    font-size: 12px;
    color: var(--fg-muted);
    line-height: 1.5;
  }
  .help .mono {
    font-family: var(--font-mono);
    font-size: 11px;
  }
  .notice {
    margin: 0;
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
  .replaces {
    padding-top: 10px;
    border-top: 1px solid var(--border);
  }
  .error {
    margin: 0;
    color: var(--danger);
    font-size: 12.5px;
  }
  .foot {
    margin: 0;
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 6px;
    font-size: 12.5px;
    color: var(--fg-dim);
  }
  .ghost-link {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 4px 6px;
    border-radius: 6px;
    color: var(--fg-muted);
    text-decoration: none;
    font-weight: 600;
  }
  .ghost-link:hover {
    color: var(--fg);
    background: var(--overlay);
  }
</style>
