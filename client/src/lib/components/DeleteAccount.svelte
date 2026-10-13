<script lang="ts">
  // DeleteAccount — "Delete my account" (#2983, ADR 0149): what a
  // deletion does, a typed confirmation, and the one call. It renders
  // inside the account menu's dialog (DeleteAccountDialog) and on the
  // privacy page, so the words are one copy.
  //
  // Only a signed-in Discord account has anything to delete: a guest
  // seat or spectator link has no stored profile, and the shared admin
  // token is not a person. For those the component says so rather than
  // showing a button the server would refuse.
  //
  // The server decides everything that matters (the typed phrase, the
  // live-table rule); this checks the phrase first only so the button is
  // not live until the person has typed it.
  import { deleteMyAccount } from "../api";
  import { LobbyApiError, session } from "../session";
  import { signedInUserID } from "../myGames";
  import { DELETE_PHRASE } from "../privacy";
  import { navigate } from "../router";

  interface Props {
    /** Called after the account is gone, before the move to the sign-in page. */
    ondeleted?: () => void;
    /** Shown as a Cancel button when set (the dialog); the page has none. */
    oncancel?: () => void;
  }
  const { ondeleted, oncancel }: Props = $props();

  const canDelete = $derived(signedInUserID($session) !== null);

  let typed = $state("");
  let busy = $state(false);
  let error = $state("");

  const matches = $derived(typed.trim().toLowerCase() === DELETE_PHRASE);

  async function submit(e: Event): Promise<void> {
    e.preventDefault();
    if (!matches || busy) return;
    busy = true;
    error = "";
    try {
      await deleteMyAccount(typed.trim());
      ondeleted?.();
      navigate("#/login");
    } catch (err) {
      error =
        err instanceof LobbyApiError ? err.message : "could not delete the account; try again";
      busy = false;
    }
  }
</script>

<section class="delete" aria-labelledby="delete-title">
  <h2 id="delete-title">Delete my account</h2>
  {#if !canDelete}
    <p class="note">
      Only a signed-in Discord account can be deleted. A guest seat or a spectator link keeps no
      profile, and signing out ends it. To have a guest seat's name removed from a past game, use
      the contact on this page.
    </p>
  {:else}
    <p class="note">This is permanent and there is no undo. It signs you out everywhere at once.</p>
    <ul class="what">
      <li>
        <strong>Deleted:</strong> your saved decks, deck-request history, playmats and their images, your
        saved settings and last table setup, and your Discord sign-in (id, name, avatar and token).
      </li>
      <li>
        <strong>Anonymised:</strong> your seat in every table that has ended becomes “Deleted player”,
        in the game list, the replay and the saved game. The other players keep the game and its result.
      </li>
      <li>
        <strong>Not deletable by us:</strong> issues you already filed on GitHub, which are public. Screenshots
        from a bug report are not tied to your account and expire on their own. Server backups roll off
        over several months.
      </li>
      <li>
        <strong>Refused while you hold a seat at a table that has not ended.</strong> Leave it or wait
        for it to finish first.
      </li>
    </ul>
    <form onsubmit={submit}>
      <label for="delete-confirm">Type <code>{DELETE_PHRASE}</code> to confirm</label>
      <input
        id="delete-confirm"
        type="text"
        autocomplete="off"
        autocapitalize="off"
        spellcheck="false"
        bind:value={typed}
        disabled={busy}
      />
      {#if error}
        <p class="error" role="alert">{error}</p>
      {/if}
      <div class="actions">
        {#if oncancel}
          <button type="button" class="ghost" onclick={oncancel} disabled={busy}>Cancel</button>
        {/if}
        <button type="submit" class="danger" disabled={!matches || busy}>
          {busy ? "Deleting…" : "Delete my account"}
        </button>
      </div>
    </form>
  {/if}
</section>

<style>
  .delete h2 {
    margin: 0 0 8px;
    font-size: 18px;
    letter-spacing: -0.01em;
    text-transform: none;
    color: var(--danger);
  }
  .note {
    margin: 0 0 10px;
    color: var(--fg-muted);
    font-size: 13px;
    line-height: 1.5;
  }
  .what {
    margin: 0 0 14px;
    padding-left: 18px;
    color: var(--fg-muted);
    font-size: 13px;
    line-height: 1.5;
  }
  .what li {
    margin-bottom: 6px;
  }
  .what strong {
    color: var(--fg);
  }
  form {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  label {
    font-size: 13px;
    color: var(--fg);
  }
  code {
    font-family: var(--font-mono);
    font-size: 12.5px;
  }
  input {
    width: 100%;
  }
  .error {
    margin: 0;
    color: var(--danger);
    font-size: 13px;
  }
  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    margin-top: 4px;
  }
</style>
