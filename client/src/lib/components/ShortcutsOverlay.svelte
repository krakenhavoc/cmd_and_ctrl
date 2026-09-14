<script lang="ts">
  // The `?` cheat sheet.
  //
  // Every row is generated from SHORTCUTS, so the overlay physically
  // cannot drift from what the keys do: an action that is not in the
  // table is not dispatchable, and an action that is in the table
  // appears here. The chord shown is the EFFECTIVE one — defaults
  // with the user's overrides applied — not the shipped default.
  //
  // Rows grey out with a reason when the move is not legal right now.
  // That reason is the server's (`GameView.legal_moves`, ADR 0033 §1)
  // routed through shortcuts.shortcutAvailability; nothing here
  // re-derives a rule.
  import { settings, openSettings } from "../settings";
  import { shortcutContext, shortcutsHelpOpen, closeShortcutsHelp } from "../shortcutRuntime";
  import {
    SHORTCUTS,
    GROUP_ORDER,
    GROUP_LABELS,
    bindingConflicts,
    chordFrom,
    effectiveBindings,
    formatChord,
    isMacLike,
    shortcutAvailability,
    type ShortcutGroup,
  } from "../shortcuts";
  import ModalLayer from "./ModalLayer.svelte";
  import Icon from "./Icon.svelte";

  const mac = isMacLike();
  const bindings = $derived(effectiveBindings($settings.shortcuts.bindings));
  const conflicts = $derived(bindingConflicts(bindings));
  const ctx = $derived($shortcutContext);

  // One pre-rendered row per action, grouped for display.
  const groups = $derived(
    GROUP_ORDER.map((group: ShortcutGroup) => ({
      group,
      label: GROUP_LABELS[group],
      rows: SHORTCUTS.filter((s) => s.group === group).map((s) => {
        const chord = bindings[s.id];
        const availability = shortcutAvailability(s.id, ctx);
        return {
          id: s.id,
          label: s.label,
          hint: s.hint,
          chord,
          display: formatChord(chord, mac),
          availability,
          conflicted: chord !== "" && (conflicts.get(chord)?.length ?? 0) > 1,
        };
      }),
    })).filter((g) => g.rows.length > 0),
  );

  // The overlay owns the keyboard while it is up (it registers a
  // ModalLayer, so the global dispatcher has already stood down).
  // Escape closes; so does the key that opened it, which is why the
  // global layer only ever OPENS this thing.
  function onKey(ev: KeyboardEvent): void {
    if (!$shortcutsHelpOpen) return;
    // The press that OPENED this overlay is still propagating: the
    // global layer handled it and called preventDefault, and this
    // listener is next on window. Without this guard the help key
    // would open and close the panel on one keystroke.
    if (ev.defaultPrevented) return;
    if (ev.key === "Escape") {
      ev.preventDefault();
      closeShortcutsHelp();
      return;
    }
    const chord = chordFrom(ev);
    if (chord && chord === bindings.toggleHelp) {
      ev.preventDefault();
      closeShortcutsHelp();
    }
  }

  function onBackdropClick(e: MouseEvent): void {
    if (e.target === e.currentTarget) closeShortcutsHelp();
  }

  function customise(): void {
    closeShortcutsHelp();
    openSettings("shortcuts");
  }
</script>

<svelte:window onkeydown={onKey} />

{#if $shortcutsHelpOpen}
  <ModalLayer />
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="sc-backdrop" onclick={onBackdropClick}>
    <div class="sc-panel" role="dialog" aria-modal="true" aria-labelledby="sc-title" tabindex="-1">
      <header>
        <h2 id="sc-title">keyboard shortcuts</h2>
        <button class="sc-close" aria-label="close keyboard shortcuts" onclick={closeShortcutsHelp}
          >×</button
        >
      </header>

      {#if !$settings.shortcuts.enabled}
        <p class="sc-off" role="status">
          Shortcuts are switched off. Nothing below will fire until you turn them back on in
          <button class="sc-link" onclick={customise}>Settings → Shortcuts</button>.
        </p>
      {/if}

      <div class="sc-body">
        {#each groups as g (g.group)}
          <section class="sc-group">
            <h3>{g.label}</h3>
            <ul>
              {#each g.rows as row (row.id)}
                <li class:off={!row.availability.enabled} class:unbound={row.chord === ""}>
                  <span class="sc-keys">
                    {#if row.chord === ""}
                      <span class="sc-none">unbound</span>
                    {:else}
                      <kbd>{row.display}</kbd>
                    {/if}
                  </span>
                  <span class="sc-text">
                    <b>{row.label}</b>
                    <span class="sc-hint">{row.hint}</span>
                    {#if row.conflicted}
                      <span class="sc-warn"
                        ><Icon name="bug" size={11} /> shared with another action</span
                      >
                    {:else if !row.availability.enabled && row.availability.reason}
                      <span class="sc-why">{row.availability.reason}</span>
                    {/if}
                  </span>
                </li>
              {/each}
            </ul>
          </section>
        {/each}
      </div>

      <footer>
        <p class="sc-note">
          <kbd>Esc</kbd> closes whatever is open and <kbd>Enter</kbd> confirms the current prompt. Both
          belong to the dialog on screen and are never taken over.
        </p>
        <button class="sc-customise" onclick={customise}>Customise…</button>
      </footer>
    </div>
  </div>
{/if}

<style>
  /* Styles are scoped to this component — no class is borrowed from
     anywhere else, which is the failure a previous PR shipped when a
     component reused another's selectors and Svelte scoped them
     away. Transitions here are plain CSS, so the global
     :root[data-reduce-motion="1"] rule in app.css already flattens
     them for a user who asked for reduced motion. */
  .sc-backdrop {
    position: fixed;
    inset: 0;
    z-index: 200;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 2rem 1rem;
    background: rgba(0, 0, 0, 0.62);
    backdrop-filter: blur(2px);
  }
  .sc-panel {
    width: min(760px, 100%);
    max-height: 100%;
    display: flex;
    flex-direction: column;
    background: var(--bg-1);
    border: 1px solid var(--border-strong);
    border-radius: 10px;
    box-shadow: 0 18px 48px rgba(0, 0, 0, 0.55);
    overflow: hidden;
  }
  .sc-panel > header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.7rem 1rem;
    border-bottom: 1px solid var(--border);
  }
  .sc-panel > header h2 {
    margin: 0;
    font-family: var(--font-display);
    font-size: 14px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--fg);
  }
  .sc-close {
    background: none;
    border: none;
    color: var(--fg-muted);
    font-size: 20px;
    line-height: 1;
    cursor: pointer;
    padding: 0 0.25rem;
  }
  .sc-close:hover {
    color: var(--fg);
  }
  .sc-off {
    margin: 0;
    padding: 0.55rem 1rem;
    background: var(--accent-soft);
    border-bottom: 1px solid var(--border);
    font-size: 12px;
    color: var(--fg);
  }
  .sc-link {
    background: none;
    border: none;
    padding: 0;
    color: var(--accent);
    font: inherit;
    cursor: pointer;
    text-decoration: underline;
  }
  .sc-body {
    overflow-y: auto;
    padding: 0.5rem 1rem 0.9rem;
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 0 1.4rem;
  }
  .sc-group h3 {
    margin: 0.9rem 0 0.35rem;
    font-size: 11px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: var(--fg-dim);
  }
  .sc-group ul {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .sc-group li {
    display: flex;
    gap: 0.7rem;
    align-items: baseline;
    padding: 0.3rem 0;
    border-bottom: 1px solid var(--border);
  }
  .sc-group li:last-child {
    border-bottom: none;
  }
  .sc-group li.off {
    opacity: 0.45;
  }
  .sc-keys {
    flex: 0 0 5.5rem;
    text-align: right;
  }
  .sc-text {
    display: flex;
    flex-direction: column;
    gap: 1px;
    min-width: 0;
  }
  .sc-text b {
    font-size: 13px;
    color: var(--fg);
    font-weight: 600;
  }
  .sc-hint {
    font-size: 11px;
    color: var(--fg-muted);
  }
  .sc-why {
    font-size: 11px;
    color: var(--fg-dim);
    font-style: italic;
  }
  .sc-warn {
    font-size: 11px;
    color: var(--danger);
    display: inline-flex;
    align-items: center;
    gap: 4px;
  }
  .sc-none {
    font-size: 11px;
    color: var(--fg-dim);
    font-style: italic;
  }
  kbd {
    display: inline-block;
    font-family: var(--font-mono);
    font-size: 11px;
    line-height: 1;
    padding: 4px 6px;
    color: var(--fg);
    background: var(--bg-2);
    border: 1px solid var(--border-strong);
    border-bottom-width: 2px;
    border-radius: 4px;
    white-space: nowrap;
  }
  .sc-panel > footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    padding: 0.6rem 1rem;
    border-top: 1px solid var(--border);
  }
  .sc-note {
    margin: 0;
    font-size: 11px;
    color: var(--fg-muted);
  }
  .sc-customise {
    flex: 0 0 auto;
    cursor: pointer;
  }
</style>
