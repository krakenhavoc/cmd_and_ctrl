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
  <h1>cmd_and_ctrl</h1>

  {#if $expiryNotice}
    <p class="notice">{$expiryNotice}</p>
  {/if}

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
    <p class="error">{error}</p>
  {/if}

  <hr />

  <h2>have an invite link?</h2>
  <p class="muted">Paste the full invite URL to join a game.</p>
  <form onsubmit={goToInvite}>
    <input type="text" placeholder="https://.../#/games/.../join?t=..." bind:value={inviteURL} />
    <button type="submit" disabled={!inviteURL}>open</button>
  </form>
</section>

<style>
  section {
    max-width: 480px;
    margin: 2rem auto;
    padding: 1.5rem;
  }
  form {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 0.75rem;
  }
  input {
    flex: 1;
    padding: 0.5rem;
  }
  .error {
    color: #c00;
  }
  .notice {
    padding: 0.5rem 0.75rem;
    margin-bottom: 1rem;
    background: #fff6d6;
    border: 1px solid #e0c75a;
    border-radius: 4px;
  }
  .muted {
    color: #666;
    font-size: 0.9em;
  }
  hr {
    margin: 1.5rem 0;
    border: none;
    border-top: 1px solid #ddd;
  }
</style>
