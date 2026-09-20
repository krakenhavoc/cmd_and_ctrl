<script lang="ts">
  // The table settings panel (ADR 0075 §2.5).
  //
  // One component, two hosts: the lobby page, which PATCHes over
  // HTTP, and the in-game menu, which sends `set_table_settings` over
  // the socket. Neither transport is in here — the parent passes
  // `onpatch` and owns the round trip — because the two differ in
  // every respect except the patch body, and the panel's job is to
  // produce that body.
  //
  // Three things it is deliberately not:
  //
  // - It is NOT a form with a Save button. Every field is independent
  //   and the server takes a partial, so a change commits on the spot
  //   and is narrated in the game log. A Save button would collect
  //   five changes into one log line that names none of them.
  // - It is NOT hidden from the people who cannot change it. The
  //   settings are public on purpose (§2.2): how many take-backs this
  //   table allows, and whether the host can conjure a Treasure, are
  //   not the host's private business. A non-manager sees the same
  //   panel with the controls disabled.
  // - It does NOT enforce anything. The server refuses a non-manager
  //   whatever this renders. Disabling the control is so the click
  //   never happens, not so the rule holds.

  import { untrack } from "svelte";
  import {
    BOT_PACE_CHOICES,
    MAX_COMMANDER_DAMAGE,
    MAX_STARTING_LIFE,
    MAX_UNDO_LIMIT,
    MIN_COMMANDER_DAMAGE,
    MIN_STARTING_LIFE,
    UNDO_SCOPE_CHOICES,
    UNDO_UNLIMITED,
    describeUndoLimit,
    isUnlimitedUndo,
    settingsDiff,
    settingsPatchError,
    startingLifeLocked,
    type TableSettingsPatch,
  } from "../tableSettings";
  import type { TableSettingsView } from "../protocol";

  interface Props {
    settings: TableSettingsView;
    // Whether this viewer may change anything — canManageTable, which
    // is the host or the server admin.
    canManage: boolean;
    // The game's lifecycle state: "lobby" | "active" | "ended". Only
    // starting life reads it, and only to lock itself. Named
    // `gameState` rather than `state` because a local binding called
    // `state` turns every `$state(...)` rune in this file into a
    // store subscription.
    gameState: string;
    // Applies a partial. The parent decides how it travels and
    // reports the outcome back through `busy` / `error`.
    onpatch: (patch: TableSettingsPatch) => void;
    busy?: boolean;
    // The server's own message for the last failed patch, rendered
    // verbatim — "this table has spawning switched off" is advice,
    // and paraphrasing it loses the advice.
    error?: string | null;
  }

  const { settings, canManage, gameState, onpatch, busy = false, error = null }: Props = $props();

  // Numbers are edited, not chosen, so they need somewhere to live
  // between keystrokes. The draft follows the authoritative settings
  // whenever those change underneath us (another manager's patch
  // arriving on the socket), and is only read back on commit.
  // untrack at the initialiser: reading a prop there is a one-time
  // read by definition, and saying so is what keeps the compiler from
  // warning that the draft will not follow the prop. The $effect
  // below is what makes it follow.
  let lifeDraft = $state(untrack(() => settings.starting_life));
  let cmdDraft = $state(untrack(() => settings.commander_damage));
  let undoDraft = $state(untrack(() => settings.undo_limit));
  $effect(() => void (lifeDraft = settings.starting_life));
  $effect(() => void (cmdDraft = settings.commander_damage));
  $effect(() => void (undoDraft = settings.undo_limit));

  // A locally-caught range problem, shown under the field that
  // produced it instead of travelling to the server and coming back
  // as a 400.
  let localError = $state<string | null>(null);

  const lifeLocked = $derived(startingLifeLocked(gameState));
  const disabled = $derived(!canManage || busy);
  const shownError = $derived(localError ?? error);

  function apply(patch: TableSettingsPatch): void {
    // The disabled attribute is the affordance, not the guard. A
    // browser suppresses a click on a disabled control; a keyboard
    // handler, a test, or a future wrapper that forgets to pass
    // `disabled` down does not, and a read-only panel that emits a
    // patch is a control that lied.
    if (disabled) return;
    localError = null;
    const changed = settingsDiff(settings, patch);
    if (Object.keys(changed).length === 0) return;
    const problem = settingsPatchError(changed, gameState);
    if (problem) {
      localError = problem;
      return;
    }
    onpatch(changed);
  }

  // Number inputs commit on change (blur / Enter), never per
  // keystroke: "4" on the way to "40" is a legal starting life, and
  // sending it would deal every seat 4 life and write a log line
  // about it.
  function commitNumber(value: number, key: keyof TableSettingsView): void {
    if (!Number.isFinite(value)) return;
    apply({ [key]: Math.trunc(value) } as TableSettingsPatch);
  }

  // Lowering the commander-damage threshold is read at the next
  // state-based action check, so it can end somebody's game a beat
  // later (§2.3). Said out loud only when it can actually happen.
  const commanderWarning = $derived(
    gameState === "active" && cmdDraft < settings.commander_damage
      ? "Lowering this takes effect at the next check — a player already over the new number loses then."
      : null,
  );
</script>

<div class="tsp" class:readonly={!canManage}>
  <p class="tsp-lede">
    {#if canManage}
      House rules for this table. Everyone at the table can see them, and every change you make is
      written into the game log.
    {:else}
      House rules for this table. Only the host and the server admin can change them — they are
      shown here so everybody knows what this table is playing.
    {/if}
  </p>

  <section class="tsp-field" data-field="undo_limit">
    <div class="tsp-row">
      <label class="tsp-label" for="tsp-undo">Take-backs</label>
      <span class="tsp-control">
        <input
          id="tsp-undo"
          type="number"
          min={UNDO_UNLIMITED}
          max={MAX_UNDO_LIMIT}
          {disabled}
          bind:value={undoDraft}
          onchange={() => commitNumber(undoDraft, "undo_limit")}
        />
        <button
          type="button"
          class="tsp-chip"
          class:on={isUnlimitedUndo(settings.undo_limit)}
          {disabled}
          aria-label="unlimited take-backs"
          aria-pressed={isUnlimitedUndo(settings.undo_limit)}
          title="no undo budget at all — nothing is ever refused"
          onclick={() => apply({ undo_limit: UNDO_UNLIMITED })}>∞</button
        >
      </span>
    </div>
    <p class="tsp-hint">{describeUndoLimit(settings.undo_limit)}</p>
  </section>

  <section class="tsp-field" data-field="undo_scope">
    <div class="tsp-row">
      <span class="tsp-label" id="tsp-scope-label">Who can take a move back</span>
      <span class="tsp-control" role="group" aria-labelledby="tsp-scope-label">
        {#each UNDO_SCOPE_CHOICES as choice (choice.value)}
          <button
            type="button"
            class="tsp-seg"
            class:on={settings.undo_scope === choice.value}
            aria-pressed={settings.undo_scope === choice.value}
            {disabled}
            onclick={() => apply({ undo_scope: choice.value })}>{choice.label}</button
          >
        {/each}
      </span>
    </div>
    <p class="tsp-hint">
      {UNDO_SCOPE_CHOICES.find((c) => c.value === settings.undo_scope)?.hint}
    </p>
  </section>

  <section class="tsp-field" data-field="starting_life" class:locked={lifeLocked}>
    <div class="tsp-row">
      <label class="tsp-label" for="tsp-life">Starting life</label>
      <span class="tsp-control">
        <input
          id="tsp-life"
          type="number"
          min={MIN_STARTING_LIFE}
          max={MAX_STARTING_LIFE}
          disabled={disabled || lifeLocked}
          bind:value={lifeDraft}
          onchange={() => commitNumber(lifeDraft, "starting_life")}
        />
      </span>
    </div>
    <p class="tsp-hint">
      {#if lifeLocked}
        Fixed — the life totals have already been dealt. Use the ± controls on a seat to change
        somebody's life now.
      {:else}
        What each seat starts on. Commander is 40.
      {/if}
    </p>
  </section>

  <section class="tsp-field" data-field="commander_damage">
    <div class="tsp-row">
      <label class="tsp-label" for="tsp-cmd">Commander damage</label>
      <span class="tsp-control">
        <input
          id="tsp-cmd"
          type="number"
          min={MIN_COMMANDER_DAMAGE}
          max={MAX_COMMANDER_DAMAGE}
          {disabled}
          bind:value={cmdDraft}
          onchange={() => commitNumber(cmdDraft, "commander_damage")}
        />
      </span>
    </div>
    <p class="tsp-hint">Damage from one commander that knocks a player out. Commander is 21.</p>
    {#if commanderWarning}
      <p class="tsp-warn" role="status">{commanderWarning}</p>
    {/if}
  </section>

  <section class="tsp-field" data-field="bot_pace">
    <div class="tsp-row">
      <span class="tsp-label" id="tsp-pace-label">Bot speed</span>
      <span class="tsp-control" role="group" aria-labelledby="tsp-pace-label">
        {#each BOT_PACE_CHOICES as choice (choice.value)}
          <button
            type="button"
            class="tsp-seg"
            class:on={settings.bot_pace === choice.value}
            aria-pressed={settings.bot_pace === choice.value}
            {disabled}
            onclick={() => apply({ bot_pace: choice.value })}>{choice.label}</button
          >
        {/each}
      </span>
    </div>
    <p class="tsp-hint">{BOT_PACE_CHOICES.find((c) => c.value === settings.bot_pace)?.hint}</p>
  </section>

  <section class="tsp-field" data-field="allow_spawn">
    <div class="tsp-row">
      <label class="tsp-label" for="tsp-spawn">Spawning</label>
      <span class="tsp-control">
        <input
          id="tsp-spawn"
          type="checkbox"
          {disabled}
          checked={settings.allow_spawn}
          onchange={(e) => apply({ allow_spawn: e.currentTarget.checked })}
        />
      </span>
    </div>
    <p class="tsp-hint">
      Lets the host put cards and tokens straight onto the table — for the cards the engine can't
      make yet, and for fixing a misplay. While it is on, everyone at the table sees a badge saying
      so, and every spawn is named in the game log.
    </p>
  </section>

  {#if shownError}
    <p class="tsp-error" role="alert">{shownError}</p>
  {/if}
</div>

<style>
  .tsp {
    display: flex;
    flex-direction: column;
    gap: 0.05rem;
    color: var(--fg, #e6ecff);
  }
  .tsp-lede {
    margin: 0 0 0.5rem;
    color: var(--fg-muted, #9aa5cd);
    font-size: 0.82em;
    line-height: 1.45;
  }
  .tsp-field {
    padding: 0.45rem 0;
    border-top: 1px solid var(--border, #273049);
  }
  .tsp-field:first-of-type {
    border-top: none;
  }
  .tsp-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
  }
  .tsp-label {
    font-weight: 600;
  }
  .tsp-control {
    display: inline-flex;
    align-items: center;
    gap: 0.3rem;
  }
  .tsp-control input[type="number"] {
    width: 4.5rem;
    padding: 0.2rem 0.35rem;
    border: 1px solid var(--border, #273049);
    border-radius: var(--radius-sm, 4px);
    background: var(--surface-sunken, #0a1122);
    color: inherit;
    font: inherit;
    text-align: right;
  }
  .tsp-control input[type="checkbox"] {
    width: 1.05rem;
    height: 1.05rem;
    accent-color: var(--gold, #ffd07a);
  }
  .tsp-seg,
  .tsp-chip {
    padding: 0.2rem 0.6rem;
    border: 1px solid var(--border, #273049);
    border-radius: var(--radius-sm, 4px);
    background: var(--surface, #111a2e);
    color: var(--fg-muted, #9aa5cd);
    font: inherit;
    font-size: 0.85em;
    cursor: pointer;
  }
  .tsp-seg.on,
  .tsp-chip.on {
    border-color: var(--gold, #ffd07a);
    background: var(--gold-soft, rgba(255, 208, 122, 0.18));
    color: var(--gold, #ffd07a);
  }
  .tsp-seg:disabled,
  .tsp-chip:disabled,
  .tsp-control input:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .tsp-hint {
    margin: 0.2rem 0 0;
    color: var(--fg-dim, #6c7a99);
    font-size: 0.78em;
    line-height: 1.45;
    max-width: 44ch;
  }
  .tsp-warn {
    margin: 0.25rem 0 0;
    color: var(--gold, #ffd07a);
    font-size: 0.78em;
  }
  .tsp-error {
    margin: 0.5rem 0 0;
    color: var(--danger, #ff7a7a);
    font-size: 0.82em;
  }
</style>
