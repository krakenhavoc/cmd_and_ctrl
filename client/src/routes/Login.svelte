<script lang="ts">
  import { onMount } from "svelte";
  import { adminLogin, discordAuthEnabled, discordLoginHref, joinByCode } from "../lib/api";
  import { navigate } from "../lib/router";
  import { expiryNotice, LobbyApiError, session } from "../lib/session";
  import Icon from "../lib/components/Icon.svelte";

  // Player-first landing: Discord sign-in and the invite box are
  // what a friend sees; the admin token form sits under them (the
  // #/admin route leads with it instead — `admin` prop).
  //
  // Two ways in, meeting at the same input. Signing in with Discord
  // mints an identity-only session and returns here to collect a
  // code, which the server resolves to a table. A pasted invite link
  // already names its table, so it goes straight to the Join page
  // and the flow that has always handled it.
  interface Props {
    admin?: boolean;
  }
  const { admin = false }: Props = $props();

  let token = $state("");
  let invite = $state("");
  let joinName = $state("");
  let error = $state("");
  let busy = $state(false);
  let joining = $state(false);

  // discordEnabled gates the sign-in button. Probed once from
  // /auth/discord/config: a deploy without the three CMDCTRL_DISCORD_*
  // vars hides the button rather than offering one that 503s. Same
  // posture as Join.svelte.
  let discordEnabled = $state(false);
  onMount(() => {
    void discordAuthEnabled().then((on) => {
      discordEnabled = on;
    });
  });

  // The signed-in-but-seatless principal, or null. Drives the copy on
  // the invite card — once we know who you are, the question stops
  // being "have an invite?" and becomes "which table?".
  const identity = $derived($session?.principal.role === "identified" ? $session.principal : null);

  // A bare code claims the seat from this page, so it needs a name to
  // put on it — unless Discord already supplied one. A pasted LINK
  // doesn't: it hands off to the Join page, which has its own name
  // field. Without this the manual path would post an empty name and
  // take a 400 the user could do nothing about.
  const needsName = $derived(!identity && invite.trim() !== "" && inviteHash(invite.trim()) === "");

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

  // inviteHash pulls the #/games/…/join?t=… fragment out of a pasted
  // invite URL. Returns "" for a bare code, which is the signal to
  // ask the server which table the code belongs to instead.
  function inviteHash(raw: string): string {
    if (raw.startsWith("#")) return raw;
    if (!/^https?:\/\//i.test(raw)) return "";
    try {
      return new URL(raw).hash;
    } catch {
      return "";
    }
  }

  async function submitInvite(e: SubmitEvent): Promise<void> {
    e.preventDefault();
    error = "";
    const raw = invite.trim();
    if (!raw) return;

    // A link carries its own game id, so the router can take it from
    // here — nothing for the server to resolve.
    const hash = inviteHash(raw);
    if (hash) {
      navigate(hash);
      return;
    }

    joining = true;
    try {
      await joinByCode(raw, joinName.trim());
      // Lobby first, exactly like the invite-link flow: that is where
      // a player imports a deck and sees the other seats before the
      // table itself (s085 / #43). Going straight to the game route
      // would skip the deck upload.
      navigate("#/lobby");
    } catch (err) {
      error = err instanceof LobbyApiError ? err.message : "could not join with that code";
    } finally {
      joining = false;
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
      {#if discordEnabled && !identity}
        <div class="card">
          <h2>Sign in</h2>
          <a class="primary lg discord-btn" href={discordLoginHref()}>
            Continue with Discord <Icon name="chevronRight" size={14} />
          </a>
          <p class="help">
            Your Discord name and avatar become your seat. You'll enter an invite code next.
          </p>
        </div>
      {/if}

      <div class="card">
        <h2>{identity ? "Join a table" : "Have an invite?"}</h2>
        {#if identity}
          <p class="signed-in" role="status">
            Signed in as {identity.name ?? "your Discord account"}.
          </p>
        {/if}
        <form class="fcol" onsubmit={submitInvite}>
          <div class="frow">
            <input
              class="mono"
              type="text"
              placeholder="invite code or link"
              aria-label="invite code or link"
              bind:value={invite}
            />
            <button
              type="submit"
              class="primary lg"
              disabled={joining || !invite.trim() || (needsName && !joinName.trim())}
            >
              {joining ? "…" : "join"}
              <Icon name="chevronRight" size={14} />
            </button>
          </div>
          {#if needsName}
            <input
              type="text"
              placeholder="your name"
              aria-label="your name"
              bind:value={joinName}
            />
          {/if}
        </form>
        <p class="help">
          Paste the code from your pod's invite, or the whole link — both work. A spectator link
          opens the table read-only.
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

    <!-- The one entry point to the public catalogue. Login is the
         page a signed-out visitor lands on, so a link that needs no
         session belongs here rather than in the lobby. -->
    <p class="foot">
      <a class="ghost-link" href="#/catalog">
        <Icon name="library" size={12} /> See which cards the engine plays
      </a>
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
  /* The invite card stacks: the code row, then the name field that
     appears only for a bare code. A pasted link hands off to the Join
     page, which asks for the name itself. */
  .fcol {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  /* Direct child only — the code input lives inside .frow and is
     already styled by the rule below it. */
  .fcol > input {
    margin: 0;
    height: 40px;
    padding: 0 12px;
    box-sizing: border-box;
    font-size: 13.5px;
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
  .discord-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    box-sizing: border-box;
    text-decoration: none;
    /* Discord blurple: the one hard-coded brand colour on the page.
       A "Continue with Discord" button in our own gold reads as a
       decoy rather than as the thing it is. */
    background: #5865f2;
    border-color: #4752c4;
    color: #fff;
  }
  .discord-btn:hover {
    background: #4752c4;
  }
  .signed-in {
    margin: 0;
    font-size: 12.5px;
    color: var(--fg-muted);
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
