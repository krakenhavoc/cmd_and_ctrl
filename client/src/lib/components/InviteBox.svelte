<script lang="ts">
  // InviteBox — "Have an invite?" for a signed-out visitor (ADR 0112 §1,
  // ADR 0156 §4): the front page's first card. It was the sign-in page's
  // invite card, moved here unchanged, labels and behaviour included.
  //
  // A pasted invite LINK already names its table, so it goes straight to
  // the Join page, which asks for the name itself. A bare CODE joins here
  // as a guest (POST /join), so it needs a name: the field appears only
  // then, pre-filled with the name this browser last joined with.
  import { joinByCode } from "../api";
  import { navigate } from "../router";
  import { LobbyApiError } from "../session";
  import { inviteHash } from "../signedInHome";
  import { loadGuestName, rememberGuestName } from "../guestName";
  import Icon from "./Icon.svelte";

  let invite = $state("");
  let joinName = $state(loadGuestName());
  let error = $state("");
  let joining = $state(false);

  const needsName = $derived(invite.trim() !== "" && inviteHash(invite.trim()) === "");

  async function submitInvite(e: SubmitEvent): Promise<void> {
    e.preventDefault();
    error = "";
    const raw = invite.trim();
    if (!raw) return;

    const hash = inviteHash(raw);
    if (hash) {
      navigate(hash);
      return;
    }

    joining = true;
    try {
      await joinByCode(raw, joinName.trim());
      rememberGuestName(joinName);
      // Lobby first, exactly like the invite-link flow: that is where a
      // player imports a deck and sees the other seats (s085 / #43).
      navigate("#/lobby");
    } catch (err) {
      error = err instanceof LobbyApiError ? err.message : "could not join with that code";
    } finally {
      joining = false;
    }
  }
</script>

<form class="invite" onsubmit={submitInvite}>
  <div class="row">
    <input
      class="mono"
      type="text"
      placeholder="invite code or link"
      aria-label="invite code or link"
      bind:value={invite}
    />
    <button
      type="submit"
      class="btn primary"
      disabled={joining || !invite.trim() || (needsName && !joinName.trim())}
    >
      {joining ? "…" : "join"}
      <Icon name="chevronRight" size={14} />
    </button>
  </div>
  {#if needsName}
    <input type="text" placeholder="your name" aria-label="your name" bind:value={joinName} />
  {/if}
  {#if error}
    <p class="error" role="alert">{error}</p>
  {/if}
</form>

<style>
  .invite {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .row {
    display: flex;
    gap: 8px;
  }
  input {
    flex: 1;
    min-width: 0;
    margin: 0;
    height: 40px;
    padding: 0 12px;
    box-sizing: border-box;
    font-size: 13.5px;
  }
  input.mono {
    font-family: var(--font-mono);
    font-size: 12.5px;
  }
  .btn {
    flex: 0 0 auto;
    gap: 4px;
    padding: 0 12px 0 14px;
    text-transform: capitalize;
  }
  .btn:disabled {
    opacity: 0.55;
    cursor: default;
  }
  .error {
    margin: 0;
    color: var(--danger);
    font-size: 12.5px;
  }
</style>
