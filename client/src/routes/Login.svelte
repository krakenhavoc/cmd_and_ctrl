<script lang="ts">
  import { adminLogin } from "../lib/api";
  import { navigate } from "../lib/router";
  import { expiryNotice, LobbyApiError } from "../lib/session";

  // Admin login is the only direct-auth path at S04. Regular players
  // arrive via an invite link and hit /routes/Join instead. The
  // invite-URL bar at the bottom of this page is a convenience for
  // pasting a full share link (e.g. when opening the app with no
  // hash fragment).
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

<section>
  <div class="hero">
    <span class="brandmark" aria-hidden="true">◆</span>
    <h1>cmd_and_ctrl</h1>
    <p class="tagline">the commander table, over the wire.</p>
  </div>

  {#if $expiryNotice}
    <p class="notice" role="status">{$expiryNotice}</p>
  {/if}

  <div class="card">
    <h2>admin login</h2>
    <form onsubmit={submit}>
      <input
        type="password"
        placeholder="admin token"
        bind:value={token}
        autocomplete="current-password"
        required
      />
      <button type="submit" disabled={busy || !token}>
        {busy ? "…" : "log in"}
      </button>
    </form>

    {#if error}
      <p class="error" role="alert">{error}</p>
    {/if}
  </div>

  <div class="card">
    <h2>have an invite link?</h2>
    <p class="muted">Paste the full invite URL to join a game.</p>
    <form onsubmit={goToInvite}>
      <input type="text" placeholder="https://.../#/games/.../join?t=..." bind:value={inviteURL} />
      <button type="submit" disabled={!inviteURL}>open</button>
    </form>
  </div>
</section>

<style>
  section {
    max-width: 480px;
    margin: 3rem auto;
    padding: 1.5rem;
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }
  .hero {
    text-align: center;
    margin-bottom: 0.5rem;
  }
  .brandmark {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 52px;
    height: 52px;
    border-radius: 14px;
    background:
      radial-gradient(120% 120% at 20% 0%, rgba(167, 196, 255, 0.4), transparent 55%),
      linear-gradient(135deg, #2f4fb8 0%, #6a3fb0 100%);
    color: #fff;
    font-size: 22px;
    margin-bottom: 14px;
    box-shadow:
      0 10px 30px rgba(47, 79, 184, 0.35),
      inset 0 1px 0 rgba(255, 255, 255, 0.2);
  }
  .hero h1 {
    font-size: 2rem;
    letter-spacing: -0.03em;
    background: linear-gradient(180deg, #ffffff 0%, #b9c6ea 100%);
    -webkit-background-clip: text;
    background-clip: text;
    color: transparent;
  }
  .tagline {
    margin: 0.25rem 0 0;
    color: var(--fg-muted);
    font-size: 0.95rem;
  }
  .card {
    background: linear-gradient(180deg, var(--surface) 0%, var(--bg-2) 100%);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: 1.25rem 1.25rem 1.1rem;
    box-shadow: var(--shadow);
  }
  form {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 0.25rem;
  }
  input {
    flex: 1;
  }
  .error {
    margin-top: 0.75rem;
    color: var(--danger);
    font-size: 0.9em;
  }
  .notice {
    padding: 0.65rem 0.9rem;
    margin: 0;
    background: var(--gold-soft);
    border: 1px solid rgba(255, 208, 122, 0.35);
    border-radius: var(--radius);
    color: var(--gold);
    font-size: 0.9em;
  }
  .muted {
    color: var(--fg-muted);
    font-size: 0.9em;
    margin: 0 0 0.75rem;
  }
</style>
