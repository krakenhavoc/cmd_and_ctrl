<script lang="ts">
  import { adminLogin } from "../lib/api";
  import { navigate } from "../lib/router";
  import { expiryNotice, LobbyApiError } from "../lib/session";
  import Icon from "../lib/components/Icon.svelte";

  // Player-first landing: the invite paste is the only field a
  // friend sees; the admin token form sits under it (the #/admin
  // route leads with it instead — `admin` prop). Regular players
  // arrive via an invite link and hit /routes/Join; the paste box
  // is for a share link opened without its hash fragment.
  interface Props {
    admin?: boolean;
  }
  const { admin = false }: Props = $props();

  let token = $state("");
  let inviteURL = $state("");
  let error = $state("");
  let busy = $state(false);

  async function submit(e: SubmitEvent): Promise<void> {
    e.preventDefault();
    error = "";
    busy = true;
    try {
      await adminLogin(token);
      navigate("#/lobby");
    } catch (err) {
      error = err instanceof LobbyApiError ? err.message : "login failed";
    } finally {
      busy = false;
    }
  }

  function goToInvite(e: SubmitEvent): void {
    e.preventDefault();
    // Accept either a full URL or just the path+hash fragment.
    try {
      const url = new URL(inviteURL, location.origin);
      if (url.hash) {
        navigate(url.hash);
      } else {
        error = "invite URL does not contain a hash fragment";
      }
    } catch {
      // Treat as a raw hash.
      navigate(inviteURL.startsWith("#") ? inviteURL : "#" + inviteURL);
    }
  }
</script>

<section class="entry" class:admin>
  <div class="fan" aria-hidden="true">
    <i></i><i></i><i></i><i></i><i></i>
  </div>
  <div class="stack">
    <div class="brand">
      <span class="mark" aria-hidden="true"><i></i></span>
      <h1 class="wm">CMD &amp; CTRL</h1>
      <p class="tagline">The Commander table, over the wire.</p>
    </div>

    {#if $expiryNotice}
      <p class="notice" role="status"><Icon name="undo" size={14} /> {$expiryNotice}</p>
    {/if}

    {#if !admin}
      <div class="card">
        <h2>Have an invite?</h2>
        <form class="frow" onsubmit={goToInvite}>
          <input
            class="mono"
            type="text"
            placeholder="https://.../#/games/.../join?t=..."
            aria-label="invite link"
            bind:value={inviteURL}
          />
          <button type="submit" class="primary lg" disabled={!inviteURL}>
            open <Icon name="chevronRight" size={14} />
          </button>
        </form>
        <p class="help">
          Invite links look like <span class="mono">…/#/games/…/join?t=…</span> and take you straight
          to the table's lobby. A spectator link opens the table read-only.
        </p>
      </div>
    {/if}

    <div class="card" class:secondary={!admin}>
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
      {#if admin}
        <p class="help">
          The shared admin token from <span class="mono">CMDCTRL_ADMIN_TOKEN</span>. Players never
          need this — they arrive through an invite link.
        </p>
      {/if}
      {#if error}
        <p class="error" role="alert">{error}</p>
      {/if}
    </div>

    <p class="foot">
      {#if admin}
        <a class="ghost-link" href="#/login"><Icon name="chevronLeft" size={12} /> Back</a>
      {:else}
        Players never need a token — they arrive through an invite link.
      {/if}
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
    border: 1px solid rgba(217, 180, 92, 0.25);
    background:
      radial-gradient(60% 50% at 50% 40%, rgba(217, 180, 92, 0.16), transparent 70%), var(--bg-2);
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
  .mark {
    width: 56px;
    height: 56px;
    border: 3px solid var(--gold);
    transform: rotate(45deg);
    border-radius: 8px;
    box-sizing: border-box;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 0 40px rgba(217, 180, 92, 0.25);
  }
  .mark i {
    width: 14px;
    height: 14px;
    background: var(--gold);
    border-radius: 2px;
  }
  h1.wm {
    margin: 8px 0 0;
    font-family: var(--font-display);
    font-weight: 800;
    font-size: 34px;
    letter-spacing: 0.22em;
    color: var(--fg);
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
  .card.secondary {
    background: transparent;
    border-color: var(--border);
    box-shadow: none;
    padding: 16px 20px;
    gap: 10px;
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
  .card.secondary h2 {
    font-size: 14px;
    color: var(--fg-muted);
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
    background: var(--gold-soft);
    border: 1px solid rgba(217, 180, 92, 0.4);
    color: var(--gold-strong);
    font-size: 12.5px;
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
    background: rgba(255, 255, 255, 0.06);
  }
</style>
