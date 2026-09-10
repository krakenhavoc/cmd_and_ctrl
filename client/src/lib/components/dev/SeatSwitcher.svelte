<script lang="ts">
  // Dev-only seat swap (ADR 0023): view and act as any seat from one
  // browser, so a four-player game is testable by one person.
  //
  // Entirely client-side. An admin session can already bind to any
  // seat — WSAuthorizer accepts ?player= for RoleAdmin and the hub
  // stamps that seat onto every action it sends — so this control
  // does nothing the URL bar could not already do. What it adds is
  // one click instead of hand-editing a WebSocket URL, which is the
  // difference between the capability existing and it being used.
  //
  // Switching seats reconnects: ws.ts treats a URL change as a
  // retarget and resets the seq watermark and rendered snapshot,
  // which is exactly right here — the new seat sees a different
  // filtered view and starts from its own first snapshot.
  //
  // Lives in the dock strip rather than behind a tab because it is a
  // mode you sit in while testing, not a panel you open.
  import type { PlayerView } from "../../protocol";

  interface Props {
    seats: PlayerView[];
    // Currently-viewed seat id, or null for the admin spectator view.
    current: string | null;
    onchange: (seatID: string | null) => void;
  }
  const { seats, current, onchange }: Props = $props();

  const SPECTATE = "__spectate__";

  // Bound to the <select>; SPECTATE stands in for null because an
  // option value cannot be null.
  const value = $derived(current ?? SPECTATE);

  function pick(e: Event) {
    const v = (e.currentTarget as HTMLSelectElement).value;
    onchange(v === SPECTATE ? null : v);
  }
</script>

<label class="seat-switcher">
  <span class="label">SEAT</span>
  <select {value} onchange={pick} aria-label="View and act as seat">
    <option value={SPECTATE}>— spectate —</option>
    {#each seats as p (p.id)}
      <option value={p.id}>{p.name}</option>
    {/each}
  </select>
</label>

<style>
  .seat-switcher {
    display: flex;
    align-items: center;
    gap: 0.5em;
    padding: 0.3rem 0.7rem;
    border: 1px solid var(--border-strong, #3a4570);
    border-bottom: none;
    background: var(--bg-1, #0b1220);
    color: var(--fg-muted, #9aa5cd);
  }
  .label {
    letter-spacing: 0.1em;
  }
  select {
    max-width: 9rem;
    padding: 0.1rem 0.3rem;
    border: 1px solid var(--border, #273049);
    border-radius: var(--radius-sm, 4px);
    background: var(--surface, #111a2e);
    color: var(--gold, #ffd07a);
    font: inherit;
    cursor: pointer;
  }
</style>
