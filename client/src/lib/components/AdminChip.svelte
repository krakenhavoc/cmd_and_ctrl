<script lang="ts">
  // AdminChip — the Admin chip in the header's account menu (ADR 0112
  // §2 item 9, Delivery PR 4).
  //
  // For an allowlisted person (GET /me's `admin_allowed`) it is a toggle
  // button: "Admin · until 23:40" on the accent colour while admin mode
  // is on, "Player" in a muted outline while it is off. A click sends
  // PUT /me/admin-mode and the session takes the answer; a refusal
  // leaves the chip as it was and shows the server's message. Admin
  // mode lasts 12 hours (owner answer 1), and lib/admin.ts's lapse timer
  // turns the chip back to "Player" when it ends.
  //
  // The shared admin token gets a static "Admin token" badge instead:
  // it is an admin on every request and has no mode to switch. To play
  // as a person, sign out of the token, which puts back the saved
  // Discord session if there is one (ADR 0110 §1 item 6).
  //
  // Everyone else gets nothing. The server is the gate; this mirrors it.
  import { session } from "../session";
  import { adminChipFor, adminChipLabel, adminChipTitle, switchAdminMode } from "../admin";

  let busy = $state(false);
  let error = $state("");

  const chip = $derived(adminChipFor($session));

  async function toggle(): Promise<void> {
    if (chip?.kind !== "switch" || busy) return;
    busy = true;
    error = "";
    try {
      await switchAdminMode(!chip.on);
    } catch (err) {
      error = err instanceof Error && err.message ? err.message : "could not switch admin mode";
    } finally {
      busy = false;
    }
  }
</script>

{#if chip}
  <div class="acct-mode">
    <span class="mode-label">Mode</span>
    {#if chip.kind === "token"}
      <span
        class="mode-chip token"
        title="The shared admin token is an admin on every request. Sign out of it to play as yourself."
        >Admin token</span
      >
    {:else}
      {@const label = adminChipLabel(chip.on, chip.endsAt)}
      <button
        type="button"
        class="mode-chip"
        class:on={chip.on}
        aria-pressed={chip.on}
        aria-label={`admin mode: ${label}`}
        title={adminChipTitle(chip.on)}
        disabled={busy}
        onclick={toggle}>{label}</button
      >
    {/if}
  </div>
  {#if error}
    <p class="mode-error" role="alert">{error}</p>
  {/if}
{/if}

<style>
  .acct-mode {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    min-height: 34px;
    padding: 0 6px 0 10px;
  }
  .mode-label {
    font-family: var(--font-mono);
    font-size: 10.5px;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: var(--fg-dim);
  }
  .mode-chip {
    display: inline-flex;
    align-items: center;
    height: 26px;
    padding: 0 11px;
    border-radius: 999px;
    border: 1px solid var(--border-strong);
    background: transparent;
    color: var(--fg-muted);
    font-size: 11.5px;
    font-weight: 700;
    letter-spacing: 0.02em;
    white-space: nowrap;
    cursor: pointer;
  }
  button.mode-chip:hover:not(:disabled) {
    border-color: var(--accent);
    color: var(--fg);
  }
  .mode-chip.on {
    background: var(--accent);
    border-color: var(--accent);
    color: var(--accent-fg);
  }
  button.mode-chip.on:hover:not(:disabled) {
    background: var(--accent-strong);
    border-color: var(--accent-strong);
    color: var(--accent-fg);
  }
  .mode-chip:disabled {
    opacity: 0.6;
    cursor: progress;
  }
  .mode-chip.token {
    cursor: default;
    border-color: var(--accent);
    color: var(--accent-strong);
    background: var(--accent-soft);
  }
  .mode-error {
    margin: 2px 10px 6px;
    font-size: 12px;
    color: var(--danger);
  }
</style>
