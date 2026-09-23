<script lang="ts">
  import { onMount } from "svelte";
  import { get } from "svelte/store";
  import {
    settings,
    settingsOpen,
    settingsTab,
    closeSettings,
    resetSettings,
    updateSettings,
    exportSettings,
    importSettings,
    fingerprintSettings,
  } from "../settings";
  import { STEP_IDS, STEP_LABELS, hasOwnStop, type StepID } from "../turn";
  import {
    SHORTCUTS,
    GROUP_ORDER,
    GROUP_LABELS,
    bindingConflicts,
    chordFrom,
    defaultBindings,
    effectiveBindings,
    evaluateBinding,
    formatChord,
    isMacLike,
    elementOf,
    isTypingTarget,
    shortcutDef,
    type ShortcutGroup,
    type ShortcutID,
  } from "../shortcuts";
  import { openShortcutsHelp } from "../shortcutRuntime";
  import ModalLayer from "./ModalLayer.svelte";

  // Steps that grant priority — the only ones the per-step stops UI
  // surfaces. Untap and Cleanup are filtered out since the server
  // sentinel (priority_holder = -1) makes them un-stoppable anyway.
  const STOPPABLE_STEPS: readonly StepID[] = STEP_IDS.filter((id) => hasOwnStop(id));

  // toggleStepStop flips one entry in the stepStops map and flashes
  // the saved indicator next to the row. Path uses the step ID as
  // the leaf so each row's flash is independent.
  function toggleStepStop(step: StepID, value: boolean): void {
    updateSettings("gameplay", "stepStops", {
      ...$settings.gameplay.stepStops,
      [step]: value,
    });
    flashSaved(`gameplay.stepStops.${step}`);
  }

  // Active sidebar tab. Reset to "audio" every time the modal
  // re-opens so the user doesn't land on a deep tab they forgot
  // about.
  type Tab =
    | "audio"
    | "animations"
    | "display"
    | "gameplay"
    | "shortcuts"
    | "accessibility"
    | "advanced";
  let activeTab = $state<Tab>("audio");

  // Transient "saved ✓" indicator keyed by field path. Surfaces
  // next to a changed control for 900ms, just long enough to
  // confirm the write landed. Keyed so rapid changes on different
  // controls don't stomp each other's fade-out.
  let savedFlash = $state<Record<string, number>>({});
  const SAVED_FLASH_MS = 900;
  function flashSaved(path: string): void {
    savedFlash = { ...savedFlash, [path]: Date.now() };
    setTimeout(() => {
      // Only clear if no newer write came in for this path.
      if (savedFlash[path] && Date.now() - savedFlash[path] >= SAVED_FLASH_MS) {
        const next = { ...savedFlash };
        delete next[path];
        savedFlash = next;
      }
    }, SAVED_FLASH_MS + 50);
  }
  function isFresh(path: string): boolean {
    const t = savedFlash[path];
    return t !== undefined && Date.now() - t < SAVED_FLASH_MS;
  }

  // Each control calls this with its path so the flash appears. The
  // second arg is the new value; we forward to updateSettings so the
  // tab handlers stay one-liners.
  function change<G extends keyof Omit<typeof $settings, "__version">>(
    group: G,
    key: keyof (typeof $settings)[G],
    value: (typeof $settings)[G][keyof (typeof $settings)[G]],
  ): void {
    updateSettings(group, key, value);
    flashSaved(`${String(group)}.${String(key)}`);
  }

  // Land on the tab the opener asked for — "audio" unless someone
  // deep-linked (the shortcuts overlay's "Customise…" does). Read
  // non-reactively so the pin-back-to-audio effect below can write
  // the same store without looping.
  $effect(() => {
    if ($settingsOpen) activeTab = get(settingsTab);
  });
  // …and forget the deep link on close, so the next plain open is
  // back on the first tab the way it always was.
  $effect(() => {
    if (!$settingsOpen) settingsTab.set("audio");
  });

  // ---- Shortcuts tab ------------------------------------------------
  //
  // The rebinding UI. All of the thinking is in lib/shortcuts.ts —
  // this is capture, show, write.
  const mac = isMacLike();
  const liveBindings = $derived(effectiveBindings($settings.shortcuts.bindings));
  const liveConflicts = $derived(bindingConflicts(liveBindings));
  const shortcutGroups = $derived(
    GROUP_ORDER.map((group: ShortcutGroup) => ({
      group,
      label: GROUP_LABELS[group],
      rows: SHORTCUTS.filter((s) => s.group === group),
    })).filter((g) => g.rows.length > 0),
  );

  // Which row is listening for a keypress, and the last refusal to
  // show under it.
  let capturing = $state<ShortcutID | null>(null);
  let captureNote = $state<string | null>(null);

  // writeBindings persists the override map. Only rows that differ
  // from today's default are stored — see the schema comment on
  // Settings["shortcuts"]["bindings"].
  function writeBindings(next: Record<string, string>): void {
    updateSettings("shortcuts", "bindings", next);
    flashSaved("shortcuts.bindings");
  }

  function assignBinding(id: ShortcutID, chord: string): void {
    const def = shortcutDef(id);
    const next = { ...$settings.shortcuts.bindings };
    if (def && chord === def.defaultBinding) delete next[id];
    else next[id] = chord;
    writeBindings(next);
  }

  // Unbind is an explicit act and has to survive a default change, so
  // it is stored as "" rather than by deleting the row.
  function unbindShortcut(id: ShortcutID): void {
    assignBinding(id, "");
    capturing = null;
    captureNote = null;
  }

  function resetBinding(id: ShortcutID): void {
    const next = { ...$settings.shortcuts.bindings };
    delete next[id];
    writeBindings(next);
    capturing = null;
    captureNote = null;
  }

  function resetAllBindings(): void {
    writeBindings({});
    capturing = null;
    captureNote = null;
  }

  function startCapture(id: ShortcutID): void {
    capturing = capturing === id ? null : id;
    captureNote = null;
  }

  // onCapture turns the next keypress into a binding. Tab is let
  // through so a keyboard user can always leave the control; Escape
  // cancels; a bare modifier keeps the row listening rather than
  // assigning a chord that can never be typed.
  function onCapture(id: ShortcutID, ev: KeyboardEvent): void {
    if (ev.key === "Tab") {
      capturing = null;
      return;
    }
    ev.preventDefault();
    ev.stopPropagation();
    if (ev.key === "Escape") {
      capturing = null;
      captureNote = null;
      return;
    }
    const candidate = evaluateBinding(liveBindings, id, ev);
    if (candidate.problem === "invalid") return; // still holding a modifier
    if (candidate.problem === "reserved") {
      captureNote = `${formatChord(candidate.chord, mac)} belongs to the dialog on screen — Esc closes, Enter confirms, Tab navigates. Pick another key.`;
      return;
    }
    assignBinding(id, candidate.chord);
    capturing = null;
    // Conflicts are allowed through on purpose: the user is usually
    // mid-swap and about to fix the other row. Both rows light up in
    // the list until they do.
    captureNote =
      candidate.conflicts.length > 0
        ? `${formatChord(candidate.chord, mac)} is now on two actions — the one higher in this list wins until you change the other.`
        : null;
  }

  function isConflicted(id: ShortcutID): boolean {
    const chord = liveBindings[id];
    if (!chord) return false;
    return (liveConflicts.get(chord)?.length ?? 0) > 1;
  }

  function conflictPartners(id: ShortcutID): string {
    const chord = liveBindings[id];
    const ids = liveConflicts.get(chord) ?? [];
    return ids
      .filter((other) => other !== id)
      .map((other) => shortcutDef(other)?.label ?? other)
      .join(", ");
  }

  function isCustomised(id: ShortcutID): boolean {
    return Object.prototype.hasOwnProperty.call($settings.shortcuts.bindings, id);
  }

  const anyCustomised = $derived(Object.keys($settings.shortcuts.bindings).length > 0);
  const DEFAULTS = defaultBindings();

  $effect(() => {
    if (!$settingsOpen) {
      capturing = null;
      captureNote = null;
    }
  });

  // This panel's own keys. Local on purpose: the global layer stands
  // down entirely while a modal is up (this one registers a
  // ModalLayer), so a modal that wants a key has to handle it itself.
  //
  // Escape closes — every modal in the game does, and the global layer
  // is forbidden from binding it (shortcuts.RESERVED_CHORDS).
  //
  // The settings key TOGGLES, which is what it has done since S11.5.
  // Opening is the global layer's job now (ADR 0047); closing has to
  // be handled here, because by the time the panel is up the global
  // layer no longer fires. Read from the effective binding map rather
  // than hard-coding "," so a rebound key still closes the panel.
  onMount(() => {
    const handler = (e: KeyboardEvent) => {
      if (!$settingsOpen) return;
      if (e.key === "Escape") {
        e.preventDefault();
        closeSettings();
        return;
      }
      if (e.defaultPrevented) return;
      if (!$settings.shortcuts.enabled) return;
      if (isTypingTarget(elementOf(e.target))) return;
      const chord = chordFrom(e);
      if (chord && chord === effectiveBindings($settings.shortcuts.bindings).openSettings) {
        e.preventDefault();
        closeSettings();
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  });

  function onBackdropClick(e: MouseEvent): void {
    if (e.target === e.currentTarget) closeSettings();
  }

  function onConfirmReset(): void {
    if (confirm("Reset every setting to its default?")) resetSettings();
  }

  // Advanced tab state — all scoped to the modal lifecycle so
  // stale text doesn't linger between open/close.
  let importText = $state("");
  let importStatus = $state<{ kind: "ok" | "err"; message: string } | null>(null);
  let copyStatus = $state<"idle" | "copied">("idle");
  // Re-derive the fingerprint whenever settings change so the
  // "my settings hash" line always reflects what's live.
  const fp = $derived($settings ? fingerprintSettings() : "");

  async function onCopyExport(): Promise<void> {
    try {
      await navigator.clipboard.writeText(exportSettings());
      copyStatus = "copied";
      setTimeout(() => (copyStatus = "idle"), 1500);
    } catch {
      // Some browsers (Firefox non-focused tab, HTTP contexts)
      // reject writeText. Fall back to no-op; the user can still
      // read/copy from a revealed <textarea> if we add one later.
      importStatus = { kind: "err", message: "clipboard write failed" };
    }
  }

  function onApplyImport(): void {
    const res = importSettings(importText);
    if (!res.ok) {
      importStatus = { kind: "err", message: res.error ?? "invalid JSON" };
      return;
    }
    importStatus = {
      kind: "ok",
      message: res.changed ? "settings imported" : "no changes — already matched",
    };
    importText = "";
  }

  $effect(() => {
    if (!$settingsOpen) {
      importText = "";
      importStatus = null;
      copyStatus = "idle";
    }
  });
</script>

{#if $settingsOpen}
  <!-- Registers a modal layer for as long as the panel is up, so the
       global keymap stands down and `d` does not draw a card while
       you are reading the shortcuts list. See lib/modalLayers.ts. -->
  <ModalLayer />
  <!-- Backdrop is role-less; click-outside delegates to
       onBackdropClick (which ignores clicks on the inner panel).
       role="dialog" lives on the panel itself so screen readers
       and focus rings target the right element. Keyboard users get
       Esc via the global handler in onMount; no role or onkeydown
       needed on the backdrop itself. -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="settings-backdrop" onclick={onBackdropClick}>
    <div
      class="settings-panel"
      role="dialog"
      aria-modal="true"
      aria-labelledby="settings-title"
      tabindex="-1"
    >
      <header>
        <h2 id="settings-title">settings</h2>
        <button class="close" aria-label="close settings" onclick={closeSettings}>×</button>
      </header>

      <div class="body">
        <nav aria-label="settings sections">
          <button class:active={activeTab === "audio"} onclick={() => (activeTab = "audio")}
            >Audio</button
          >
          <button
            class:active={activeTab === "animations"}
            onclick={() => (activeTab = "animations")}>Animations</button
          >
          <button class:active={activeTab === "display"} onclick={() => (activeTab = "display")}
            >Display</button
          >
          <button class:active={activeTab === "gameplay"} onclick={() => (activeTab = "gameplay")}
            >Gameplay</button
          >
          <button class:active={activeTab === "shortcuts"} onclick={() => (activeTab = "shortcuts")}
            >Shortcuts</button
          >
          <button
            class:active={activeTab === "accessibility"}
            onclick={() => (activeTab = "accessibility")}>Accessibility</button
          >
          <button class:active={activeTab === "advanced"} onclick={() => (activeTab = "advanced")}
            >Advanced</button
          >
        </nav>

        <section>
          {#if activeTab === "audio"}
            <h3>Audio</h3>
            <label>
              <input
                type="checkbox"
                checked={$settings.audio.muted}
                onchange={(e) => change("audio", "muted", e.currentTarget.checked)}
              />
              Mute all sounds
              {#if isFresh("audio.muted")}<span class="saved">✓ saved</span>{/if}
            </label>
            <p class="help">Toggle with <kbd>M</kbd> from anywhere in the game.</p>

            <label class="slider-row">
              <span>Master volume</span>
              <input
                type="range"
                min="0"
                max="100"
                value={$settings.audio.masterVolume}
                oninput={(e) => change("audio", "masterVolume", Number(e.currentTarget.value))}
              />
              <span class="value">{$settings.audio.masterVolume}</span>
              {#if isFresh("audio.masterVolume")}<span class="saved">✓</span>{/if}
            </label>

            <label class="slider-row">
              <span>Effects volume</span>
              <input
                type="range"
                min="0"
                max="100"
                value={$settings.audio.effectsVolume}
                oninput={(e) => change("audio", "effectsVolume", Number(e.currentTarget.value))}
              />
              <span class="value">{$settings.audio.effectsVolume}</span>
              {#if isFresh("audio.effectsVolume")}<span class="saved">✓</span>{/if}
            </label>

            <label class="slider-row disabled" title="No music track ships yet">
              <span>Music volume</span>
              <input
                type="range"
                min="0"
                max="100"
                value={$settings.audio.musicVolume}
                oninput={(e) => change("audio", "musicVolume", Number(e.currentTarget.value))}
              />
              <span class="value">{$settings.audio.musicVolume}</span>
            </label>
            <p class="help">Music slider is scaffolded — no music track ships yet.</p>
          {:else if activeTab === "animations"}
            <h3>Animations</h3>
            <label>
              <input
                type="checkbox"
                checked={$settings.animations.enabled}
                onchange={(e) => change("animations", "enabled", e.currentTarget.checked)}
              />
              Enable animations
              {#if isFresh("animations.enabled")}<span class="saved">✓ saved</span>{/if}
            </label>
            <p class="help">
              Defaults to off when your OS reports <em>prefers-reduced-motion: reduce</em>.
            </p>

            <label class="slider-row">
              <span>Animation speed</span>
              <select
                value={$settings.animations.speed}
                onchange={(e) =>
                  change("animations", "speed", Number(e.currentTarget.value) as 0.5 | 1 | 1.5 | 2)}
              >
                <option value={0.5}>0.5×</option>
                <option value={1}>1× (default)</option>
                <option value={1.5}>1.5×</option>
                <option value={2}>2×</option>
              </select>
              {#if isFresh("animations.speed")}<span class="saved">✓</span>{/if}
            </label>

            <fieldset disabled={!$settings.animations.enabled}>
              <legend>Per-effect toggles</legend>
              {#each ["cardDraw", "cardPlay", "cardTap", "cardUntap", "cardFlip", "particlesEtb", "damagePopups"] as k (k)}
                <label class="inline">
                  <input
                    type="checkbox"
                    checked={$settings.animations[
                      k as keyof typeof $settings.animations
                    ] as boolean}
                    onchange={(e) =>
                      change(
                        "animations",
                        k as keyof typeof $settings.animations,
                        e.currentTarget.checked as never,
                      )}
                  />
                  {k.replace(/([A-Z])/g, " $1").toLowerCase()}
                </label>
              {/each}
            </fieldset>
          {:else if activeTab === "display"}
            <h3>Display</h3>

            <!-- #956 / ADR 0077. First row in the tab on purpose: this
                 is the one a player goes looking for after an upgrade
                 changed what their opponents look like. -->
            <label class="slider-row">
              <span>Opponent boards</span>
              <select
                value={$settings.display.opponentDetail}
                onchange={(e) =>
                  change("display", "opponentDetail", e.currentTarget.value as "summary" | "full")}
              >
                <option value="summary">Summary (default)</option>
                <option value="full">Full boards</option>
              </select>
              {#if isFresh("display.opponentDetail")}<span class="saved">✓</span>{/if}
            </label>
            <p class="help">
              Summary draws each opponent as a read-out — life, the mana they could still tap by
              colour, one pip per creature with its power and toughness — and expands that player to
              a full board when you need to click a card there: targeting a spell, declaring
              blockers, or clicking their avatar to pin them open. Full boards draws every opponent
              as cards all the time, which is how the table used to work; at four players that means
              90px cards, which is where this started.
            </p>

            <label>
              <input
                type="checkbox"
                checked={$settings.display.expandActivePlayer}
                onchange={(e) => change("display", "expandActivePlayer", e.currentTarget.checked)}
              />
              Expand the active player's board on their turn
              {#if isFresh("display.expandActivePlayer")}<span class="saved">✓ saved</span>{/if}
            </label>
            <p class="help">
              Only does anything while opponent boards are summaries. Off means a board expands only
              from something you did — a targeting prompt, block or attack mode, or pinning a seat.
              It changes on a turn boundary either way, so it can't move the table under a click you
              have already started.
            </p>

            <label class="slider-row">
              <span>Expand style <span class="experimental">experimental</span></span>
              <select
                value={$settings.display.expandStyle}
                onchange={(e) =>
                  change("display", "expandStyle", e.currentTarget.value as "reflow" | "overlay")}
              >
                <option value="reflow">Reflow (default)</option>
                <option value="overlay">Overlay</option>
              </select>
              {#if isFresh("display.expandStyle")}<span class="saved">✓</span>{/if}
            </label>
            <p class="help">
              How an expanding board makes room. Reflow gives it a bigger share of the table and
              shrinks the others; Overlay floats it on top. Both ship so they can be compared in a
              real game — <strong>one of them will be removed</strong>, along with this setting,
              once it is clear which feels better. Worth flipping mid-game to see the difference.
            </p>

            <label class="slider-row disabled">
              <span>Theme</span>
              <select disabled value="dark">
                <option value="dark">Dark (default)</option>
              </select>
            </label>
            <p class="help">
              Light and high-contrast themes are scaffolded in CSS but most table panels still
              hardcode dark colours. Switching themes right now produces a broken-looking mix, so
              the toggle stays disabled until per-component <code>var()</code> migration ships.
            </p>

            <label class="slider-row">
              <span>Card size on battlefield</span>
              <select
                value={$settings.display.cardSize}
                onchange={(e) =>
                  change(
                    "display",
                    "cardSize",
                    e.currentTarget.value as "small" | "medium" | "large",
                  )}
              >
                <option value="small">Small</option>
                <option value="medium">Medium (default)</option>
                <option value="large">Large</option>
              </select>
              {#if isFresh("display.cardSize")}<span class="saved">✓</span>{/if}
            </label>

            <label class="slider-row">
              <span>Table layout</span>
              <select
                value={$settings.display.tableLayout}
                onchange={(e) =>
                  change("display", "tableLayout", e.currentTarget.value as "row" | "quadrant")}
              >
                <option value="quadrant">Quadrant (default)</option>
                <option value="row">Row</option>
              </select>
              {#if isFresh("display.tableLayout")}<span class="saved">✓</span>{/if}
            </label>
            <p class="help">
              Quadrant keeps the around-the-table seating. Row seats the opponents in turn order
              across the top and gives your board the full width. At three players the two are the
              same — you need the whole bottom row there either way.
            </p>

            <label class="slider-row">
              <span>Hand layout</span>
              <select
                value={$settings.display.handLayout}
                onchange={(e) =>
                  change("display", "handLayout", e.currentTarget.value as "fan" | "stacked")}
              >
                <option value="fan">Fan (default)</option>
                <option value="stacked">Stacked</option>
              </select>
              {#if isFresh("display.handLayout")}<span class="saved">✓</span>{/if}
            </label>

            <label class="slider-row">
              <span>Hover preview delay (ms)</span>
              <input
                type="range"
                min="0"
                max="1000"
                step="50"
                value={$settings.display.hoverDelayMs}
                oninput={(e) => change("display", "hoverDelayMs", Number(e.currentTarget.value))}
              />
              <span class="value">{$settings.display.hoverDelayMs}</span>
              {#if isFresh("display.hoverDelayMs")}<span class="saved">✓</span>{/if}
            </label>

            <label>
              <input
                type="checkbox"
                checked={$settings.display.showOpponentHandCount}
                onchange={(e) =>
                  change("display", "showOpponentHandCount", e.currentTarget.checked)}
              />
              Show opponent hand count
              {#if isFresh("display.showOpponentHandCount")}<span class="saved">✓ saved</span>{/if}
            </label>
          {:else if activeTab === "gameplay"}
            <h3>Gameplay</h3>
            <label>
              <input
                type="checkbox"
                checked={$settings.gameplay.confirmExit}
                onchange={(e) => change("gameplay", "confirmExit", e.currentTarget.checked)}
              />
              Confirm before leaving an active game
              {#if isFresh("gameplay.confirmExit")}<span class="saved">✓ saved</span>{/if}
            </label>

            <label>
              <input
                type="checkbox"
                checked={$settings.gameplay.autoPassPriority}
                onchange={(e) => change("gameplay", "autoPassPriority", e.currentTarget.checked)}
              />
              Auto-pass priority through unstopped steps
              {#if isFresh("gameplay.autoPassPriority")}<span class="saved">✓ saved</span>{/if}
            </label>
            <p class="help">
              Priority passes for you everywhere except the steps you tick below and, with smart
              auto-pass on, the moments you can actually respond: an opponent's spell or ability on
              the stack, combat once attackers are declared, and an opponent's end step. Click
              <strong>next</strong> in the phase widget to pass by hand, or click a step icon to stop
              there once.
            </p>

            <fieldset class="step-stops">
              <legend>Stop on these steps</legend>
              <p class="help">
                When auto-pass is on, priority stops here for your input. Untap and Cleanup are
                excluded — they don't grant priority (turn-based actions auto-fire).
              </p>
              <div class="step-stops-grid">
                {#each STOPPABLE_STEPS as step (step)}
                  <label class="step-stop-row">
                    <input
                      type="checkbox"
                      checked={$settings.gameplay.stepStops[step] === true}
                      onchange={(e) => toggleStepStop(step, e.currentTarget.checked)}
                    />
                    <span>{STEP_LABELS[step]}</span>
                    {#if isFresh(`gameplay.stepStops.${step}`)}<span class="saved">✓</span>{/if}
                  </label>
                {/each}
              </div>
            </fieldset>

            <label>
              <input
                type="checkbox"
                checked={$settings.gameplay.smartAutoPass}
                onchange={(e) => change("gameplay", "smartAutoPass", e.currentTarget.checked)}
              />
              Smart auto-pass (stop only when you can do something)
              {#if isFresh("gameplay.smartAutoPass")}<span class="saved">✓ saved</span>{/if}
            </label>
            <p class="help">
              Two things. A ticked step passes when you have nothing to play there, so &ldquo;stop
              on upkeep&rdquo; means &ldquo;stop if I have something to do,&rdquo; not &ldquo;stop
              every time.&rdquo; And outside the ticked steps, it stops you in the key windows
              &mdash; an opponent's spell or ability on the stack, declared attackers or blockers,
              an opponent's end step &mdash; only when you hold a response from the list below. Mana
              abilities and land drops never count as a response. Turn this off to stop on every
              ticked step and every opponent stack item.
            </p>

            <fieldset class="step-stops" disabled={!$settings.gameplay.smartAutoPass}>
              <legend>Stop for</legend>
              <label>
                <input
                  type="checkbox"
                  checked={$settings.gameplay.respondCounterspells}
                  onchange={(e) =>
                    change("gameplay", "respondCounterspells", e.currentTarget.checked)}
                />
                Counterspells (anything that targets a spell or ability on the stack)
                {#if isFresh("gameplay.respondCounterspells")}<span class="saved">✓</span>{/if}
              </label>
              <label>
                <input
                  type="checkbox"
                  checked={$settings.gameplay.respondInstants}
                  onchange={(e) => change("gameplay", "respondInstants", e.currentTarget.checked)}
                />
                Instants and flash spells
                {#if isFresh("gameplay.respondInstants")}<span class="saved">✓</span>{/if}
              </label>
              <label>
                <input
                  type="checkbox"
                  checked={$settings.gameplay.respondAbilities}
                  onchange={(e) => change("gameplay", "respondAbilities", e.currentTarget.checked)}
                />
                Activated abilities (not mana abilities)
                {#if isFresh("gameplay.respondAbilities")}<span class="saved">✓</span>{/if}
              </label>
              <label>
                <input
                  type="checkbox"
                  checked={$settings.gameplay.respondSpecialActions}
                  onchange={(e) =>
                    change("gameplay", "respondSpecialActions", e.currentTarget.checked)}
                />
                Special actions (foretell, suspend, turning a card face up)
                {#if isFresh("gameplay.respondSpecialActions")}<span class="saved">✓</span>{/if}
              </label>
              <label>
                <input
                  type="checkbox"
                  checked={$settings.gameplay.alwaysStopOpponentStack}
                  onchange={(e) =>
                    change("gameplay", "alwaysStopOpponentStack", e.currentTarget.checked)}
                />
                Always stop for opponents' spells and abilities
                {#if isFresh("gameplay.alwaysStopOpponentStack")}<span class="saved">✓</span>{/if}
              </label>
              <p class="help">
                The last one stops on every opponent item on the stack even when you can't answer
                it, which is how auto-pass worked before. Stopping every time also means a pause
                never tells the table you have an answer.
              </p>
            </fieldset>

            <label>
              <input
                type="checkbox"
                checked={$settings.gameplay.autoPassOwnStack}
                onchange={(e) => change("gameplay", "autoPassOwnStack", e.currentTarget.checked)}
              />
              Auto-pass your own spells and triggers on the stack
              {#if isFresh("gameplay.autoPassOwnStack")}<span class="saved">✓ saved</span>{/if}
            </label>
            <p class="help">
              Casting is already the decision, so the client doesn't ask &ldquo;Counter or
              Pass?&rdquo; about a stack holding only your own items — it passes and lets them
              resolve. A stack with <em>anything</em> an opponent put on it still stops for you.
              When you do want to respond to your own spell or trigger (stacking two effects,
              holding up a counter, responding to your own ETB), click <strong>hold</strong> in the
              phase widget or the stack card <em>before</em> you cast — the pass fires the instant the
              spell is announced. Turn this off to stop on every stack, always.
            </p>

            <label class="danger">
              <input
                type="checkbox"
                checked={$settings.gameplay.autopassPersistThroughTurns}
                onchange={(e) =>
                  change("gameplay", "autopassPersistThroughTurns", e.currentTarget.checked)}
              />
              Autopass persists through your own turns
              {#if isFresh("gameplay.autopassPersistThroughTurns")}
                <span class="saved">✓ saved</span>
              {/if}
            </label>
            <p class="help danger-help">
              <strong>WARNING: ENABLING THIS SETTING MAY CAUSE YOU TO SKIP YOUR OWN TURN.</strong>
              By default, the autopass toggle in the phase display auto-clears when the cursor reaches
              your own first main phase — a safety belt so a forgotten autopass doesn't cost you a turn.
              Flip this on to keep autopass engaged indefinitely (until you click it off).
            </p>

            <label>
              <input
                type="checkbox"
                checked={$settings.gameplay.strictMana}
                onchange={(e) => change("gameplay", "strictMana", e.currentTarget.checked)}
              />
              Strict mana enforcement
              {#if isFresh("gameplay.strictMana")}<span class="saved">✓ saved</span>{/if}
            </label>
            <p class="help">
              When on, the server checks your mana pool before letting a spell resolve and enforces
              commander tax (CR&nbsp;903.8). If you're short, a toast lets you cast anyway by
              overriding the gate for that one spell. Default is off — the sandbox treats mana as
              paper-tracked.
            </p>

            <label>
              <input
                type="checkbox"
                checked={$settings.gameplay.adminOverrides}
                onchange={(e) => change("gameplay", "adminOverrides", e.currentTarget.checked)}
              />
              Enable admin overrides (right-click menu)
              {#if isFresh("gameplay.adminOverrides")}<span class="saved">✓ saved</span>{/if}
            </label>
            <p class="help">
              Right-click any card for a menu of manual overrides: move it to another zone, add or
              remove counters, mark or clear damage, declare it as an attacker or blocker, sacrifice
              it. Every option maps to an action the server already supports, so a card the engine
              gets wrong can still be played by hand. You only get overrides on cards you control;
              admins get every card. Default is off, which leaves right-click showing a permanent's
              abilities — with it on, those abilities are the menu's first section.
            </p>

            <label>
              <input
                type="checkbox"
                checked={$settings.gameplay.showBotReasoning}
                onchange={(e) => change("gameplay", "showBotReasoning", e.currentTarget.checked)}
              />
              Show bot reasoning
              {#if isFresh("gameplay.showBotReasoning")}<span class="saved">✓ saved</span>{/if}
            </label>
            <p class="help">
              Bot seats explain each move they make in the table feed. Off by default — it's a lot
              of lines, and it's debug output rather than table talk. Note it can mention cards in
              the bot's own hand, which makes the game easier. This does <strong>not</strong> control
              improvisation announcements: when a bot plays a card the rules engine can't run, it says
              so every time, and no setting hides that.
            </p>
          {:else if activeTab === "shortcuts"}
            <h3>Keyboard shortcuts</h3>
            <label>
              <input
                type="checkbox"
                checked={$settings.shortcuts.enabled}
                onchange={(e) => change("shortcuts", "enabled", e.currentTarget.checked)}
              />
              Enable keyboard shortcuts
              {#if isFresh("shortcuts.enabled")}<span class="saved">✓ saved</span>{/if}
            </label>
            <p class="help">
              Shortcuts never fire while you're typing in a text field, and they stand down entirely
              while a dialog is open — the dialog owns the keyboard. <kbd>Esc</kbd>,
              <kbd>Enter</kbd> and <kbd>Tab</kbd> are reserved and can't be rebound: they close,
              confirm and navigate whatever is on screen. Press
              <kbd>{formatChord(liveBindings.toggleHelp, mac)}</kbd> anywhere for the cheat sheet,
              or
              <button class="linkish" onclick={openShortcutsHelp}>open it now</button>.
            </p>

            <div class="sc-actions">
              <button onclick={resetAllBindings} disabled={!anyCustomised}>
                Reset every binding
              </button>
              {#if isFresh("shortcuts.bindings")}<span class="saved">✓ saved</span>{/if}
            </div>

            {#if captureNote}
              <p class="sc-note" role="status">{captureNote}</p>
            {/if}

            {#each shortcutGroups as g (g.group)}
              <fieldset class="sc-group">
                <legend>{g.label}</legend>
                {#each g.rows as def (def.id)}
                  <div class="sc-row" class:conflict={isConflicted(def.id)}>
                    <span class="sc-label">
                      <span class="sc-name">{def.label}</span>
                      <span class="sc-desc">{def.hint}</span>
                      {#if isConflicted(def.id)}
                        <span class="sc-conflict"
                          >also bound to {conflictPartners(def.id)} — the one higher in this list wins</span
                        >
                      {/if}
                    </span>
                    <span class="sc-controls">
                      <button
                        class="sc-capture"
                        class:listening={capturing === def.id}
                        aria-label={capturing === def.id
                          ? `press a key for ${def.label}`
                          : `change the key for ${def.label}`}
                        onclick={() => startCapture(def.id)}
                        onkeydown={(e) => {
                          if (capturing === def.id) onCapture(def.id, e);
                        }}
                        onblur={() => {
                          if (capturing === def.id) capturing = null;
                        }}
                      >
                        {#if capturing === def.id}
                          press a key…
                        {:else if liveBindings[def.id] === ""}
                          unbound
                        {:else}
                          {formatChord(liveBindings[def.id], mac)}
                        {/if}
                      </button>
                      <button
                        class="sc-mini"
                        onclick={() => unbindShortcut(def.id)}
                        disabled={liveBindings[def.id] === ""}
                        title="turn this shortcut off">off</button
                      >
                      <button
                        class="sc-mini"
                        onclick={() => resetBinding(def.id)}
                        disabled={!isCustomised(def.id)}
                        title={`restore the default (${DEFAULTS[def.id] === "" ? "unbound" : formatChord(DEFAULTS[def.id], mac)})`}
                        >reset</button
                      >
                    </span>
                  </div>
                {/each}
              </fieldset>
            {/each}

            <p class="help">
              Bindings are matched on the character your layout produces, not on the physical key
              position, so <kbd>L</kbd> is the key printed L on your keyboard whatever layout you use.
              Only the rows you actually change are saved — everything else follows the shipped default,
              so if we retune one later you'll get the better key.
            </p>
          {:else if activeTab === "accessibility"}
            <h3>Accessibility</h3>
            <label>
              <input
                type="checkbox"
                checked={$settings.accessibility.reduceMotion}
                onchange={(e) => change("accessibility", "reduceMotion", e.currentTarget.checked)}
              />
              Reduce motion
              {#if isFresh("accessibility.reduceMotion")}<span class="saved">✓ saved</span>{/if}
            </label>
            <p class="help">
              Follows your OS <em>prefers-reduced-motion</em> setting by default. Flipping this here overrides
              that for this site only.
            </p>

            <label class="slider-row">
              <span>Text size</span>
              <select
                value={$settings.accessibility.textScale}
                onchange={(e) =>
                  change(
                    "accessibility",
                    "textScale",
                    Number(e.currentTarget.value) as 0.9 | 1.0 | 1.2 | 1.5,
                  )}
              >
                <option value={0.9}>0.9× (compact)</option>
                <option value={1.0}>1.0× (default)</option>
                <option value={1.2}>1.2×</option>
                <option value={1.5}>1.5× (large)</option>
              </select>
              {#if isFresh("accessibility.textScale")}<span class="saved">✓</span>{/if}
            </label>

            <label>
              <input
                type="checkbox"
                checked={$settings.accessibility.colorblindPalette}
                onchange={(e) =>
                  change("accessibility", "colorblindPalette", e.currentTarget.checked)}
              />
              Use colour-blind-friendly seat palette
              {#if isFresh("accessibility.colorblindPalette")}<span class="saved">✓ saved</span
                >{/if}
            </label>

            <label>
              <input
                type="checkbox"
                checked={$settings.accessibility.alwaysShowFocus}
                onchange={(e) =>
                  change("accessibility", "alwaysShowFocus", e.currentTarget.checked)}
              />
              Always show focus outlines
              {#if isFresh("accessibility.alwaysShowFocus")}<span class="saved">✓ saved</span>{/if}
            </label>
          {:else if activeTab === "advanced"}
            <h3>Advanced</h3>

            <div class="adv-section">
              <h4>Export</h4>
              <p class="help">
                Copies your current settings as JSON. Paste into the Import box below on another
                device to carry your prefs across without a server sync.
              </p>
              <button onclick={onCopyExport}>
                {copyStatus === "copied" ? "✓ copied to clipboard" : "Copy settings to clipboard"}
              </button>
            </div>

            <div class="adv-section">
              <h4>Import</h4>
              <p class="help">
                Paste a settings JSON blob. Unknown fields are dropped, missing fields fall back to
                defaults — safe to paste an older export.
              </p>
              <textarea
                rows="6"
                placeholder="paste a settings JSON blob here"
                bind:value={importText}
                spellcheck="false"
              ></textarea>
              <div class="adv-row">
                <button onclick={onApplyImport} disabled={!importText.trim()}>Apply import</button>
                {#if importStatus}
                  <span class={`import-status ${importStatus.kind}`}>{importStatus.message}</span>
                {/if}
              </div>
            </div>

            <div class="adv-section">
              <h4>Settings fingerprint</h4>
              <p class="help">
                Short hash (FNV-1a, non-cryptographic) of your current settings. Useful to quote in
                a bug report so another user can tell whether they're running the same config.
              </p>
              <code class="fingerprint">{fp}</code>
            </div>

            <div class="adv-section">
              <h4>Reset</h4>
              <p class="help">
                Storage key: <code>localStorage["cmdctrl.settings.v1"]</code>. Reset scraps every
                value and restores defaults (including OS-pref sensing for reduced motion).
              </p>
              <button class="danger" onclick={onConfirmReset}>Reset all settings</button>
            </div>
          {/if}
        </section>
      </div>
    </div>
  </div>
{/if}

<style>
  /* Sept 2026 redesign: the settings shell shares the prompt-modal
     vocabulary (dimmed blur backdrop, flat raised panel), a left
     nav of quiet rows, and one grid row per setting — label and help
     on the left, the control on the right. Checkboxes render as
     switches, ranges get the gold accent. Markup is unchanged. */
  .settings-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(11, 10, 9, 0.72);
    backdrop-filter: blur(6px);
    -webkit-backdrop-filter: blur(6px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 200;
  }
  .settings-panel {
    background: var(--surface);
    color: var(--fg);
    border: 1px solid var(--border-strong);
    border-radius: 16px;
    width: min(780px, 92vw);
    max-height: min(88vh, 640px);
    display: flex;
    flex-direction: column;
    box-shadow: var(--shadow-lg);
    overflow: hidden;
  }
  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 14px 18px 14px 22px;
    border-bottom: 1px solid var(--border);
  }
  header h2 {
    margin: 0;
    font-family: var(--font-display);
    font-size: 17px;
    font-weight: 700;
    letter-spacing: -0.01em;
    text-transform: none;
    color: var(--fg);
  }
  .close {
    width: 30px;
    height: 30px;
    padding: 0;
    border-radius: 8px;
    background: transparent;
    border: 1px solid transparent;
    color: var(--fg-muted);
    font-size: 20px;
    line-height: 1;
  }
  .close:hover {
    color: var(--fg);
    background: rgba(255, 255, 255, 0.06);
    border-color: var(--border);
  }
  .body {
    display: flex;
    flex: 1;
    min-height: 0;
  }
  nav {
    display: flex;
    flex-direction: column;
    gap: 2px;
    width: 180px;
    border-right: 1px solid var(--border);
    padding: 10px;
    flex-shrink: 0;
  }
  nav button {
    display: flex;
    align-items: center;
    height: 34px;
    padding: 0 10px;
    border-radius: 8px;
    border: 1px solid transparent;
    background: transparent;
    color: var(--fg-muted);
    text-align: left;
    font-size: 13px;
    font-weight: 500;
    justify-content: flex-start;
  }
  nav button:hover {
    background: rgba(255, 255, 255, 0.04);
    color: var(--fg);
  }
  nav button.active {
    background: var(--surface-raised);
    color: var(--fg);
    font-weight: 600;
    border-color: var(--border);
  }
  section {
    flex: 1;
    padding: 18px 26px 22px;
    overflow-y: auto;
    min-width: 0;
  }
  section h3 {
    margin: 0 0 6px;
    font-family: var(--font-mono);
    font-size: 10.5px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--fg-dim);
    font-weight: 600;
  }
  section h3:not(:first-child) {
    margin-top: 18px;
  }
  /* One row per setting. Switches (checkbox inputs) go to the right
     via flex order; the text nodes stay on the left. */
  label {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 9px 0;
    cursor: pointer;
    font-size: 13px;
    font-weight: 600;
    color: var(--fg);
    border-bottom: 1px solid var(--border);
  }
  label > input[type="checkbox"] {
    order: 2;
    margin-left: auto;
  }
  label .saved {
    order: 1;
    margin-left: auto;
  }
  label > input[type="checkbox"] ~ .saved,
  label.slider-row .saved {
    margin-left: 0;
  }
  label.inline {
    display: inline-flex;
    border-bottom: none;
    padding: 4px 0;
    margin-right: 14px;
    font-weight: 500;
  }
  label.inline > input[type="checkbox"] {
    order: 0;
    margin-left: 0;
  }
  label.slider-row > span:first-child {
    min-width: 11rem;
    flex: 1;
  }
  label.slider-row .value {
    min-width: 2.5rem;
    text-align: right;
    font-family: var(--font-mono);
    font-size: 12px;
    color: var(--fg-muted);
  }
  label.disabled {
    opacity: 0.5;
  }
  /* Switch: the native checkbox with appearance:none. */
  input[type="checkbox"] {
    appearance: none;
    -webkit-appearance: none;
    width: 38px;
    height: 22px;
    border-radius: 999px;
    background: var(--surface-hover);
    border: 1px solid var(--border-strong);
    position: relative;
    cursor: pointer;
    flex: 0 0 auto;
    margin: 0;
    transition:
      background 140ms var(--ease),
      border-color 140ms var(--ease);
  }
  input[type="checkbox"]::after {
    content: "";
    position: absolute;
    top: 2px;
    left: 2px;
    width: 16px;
    height: 16px;
    border-radius: 50%;
    background: var(--fg-muted);
    transition:
      left 140ms var(--ease),
      background 140ms var(--ease);
  }
  input[type="checkbox"]:checked {
    background: var(--gold);
    border-color: var(--gold-strong);
  }
  input[type="checkbox"]:checked::after {
    left: 18px;
    background: var(--accent-fg);
  }
  input[type="checkbox"]:focus-visible {
    outline: 2px solid var(--accent-strong);
    outline-offset: 2px;
  }
  input[type="range"] {
    flex: 0 0 180px;
    accent-color: var(--gold);
    margin: 0;
  }
  select {
    flex: 0 0 auto;
    min-width: 160px;
    padding: 6px 10px;
    font-size: 12.5px;
    margin: 0;
  }
  fieldset {
    border: 1px solid var(--border);
    border-radius: 10px;
    padding: 8px 14px 12px;
    margin: 12px 0;
    background: var(--surface-sunken);
  }
  fieldset[disabled] {
    opacity: 0.45;
  }
  legend {
    color: var(--fg-muted);
    font-size: 12px;
    font-weight: 600;
    padding: 0 6px;
  }
  .help {
    color: var(--fg-dim);
    font-size: 11.5px;
    line-height: 1.45;
    margin: 4px 0 8px;
  }
  /* Danger-flagged settings get gold framing so an opt-in that might
     cost the player a turn can't be mistaken for a routine
     preference. */
  label.danger {
    color: var(--gold-strong);
  }
  .danger-help {
    color: var(--fg-muted);
    border-left: 2px solid var(--gold);
    padding: 6px 10px;
    background: var(--gold-soft);
    border-radius: 0 8px 8px 0;
  }
  .danger-help strong {
    color: var(--gold-strong);
    letter-spacing: 0.03em;
  }
  /* An option that exists only to be compared against another one.
     Flagged in the UI as well as in the code, because a setting that
     is going to disappear should not look permanent. */
  .experimental {
    margin-left: 6px;
    padding: 1px 5px;
    border-radius: 4px;
    border: 1px solid var(--gold);
    background: var(--gold-soft);
    color: var(--gold-strong);
    font-family: var(--font-mono);
    font-size: 9px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    font-weight: 600;
    vertical-align: middle;
  }
  .saved {
    color: var(--mint);
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }
  kbd {
    background: var(--surface-raised);
    border: 1px solid var(--border);
    padding: 0 5px;
    border-radius: 4px;
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--fg-muted);
  }
  .danger {
    background: rgba(255, 107, 107, 0.1);
    border: 1px solid rgba(255, 107, 107, 0.4);
    color: var(--danger);
    padding: 0.5rem 0.9rem;
    border-radius: var(--radius);
    cursor: pointer;
  }
  .danger:hover {
    background: rgba(255, 107, 107, 0.18);
  }
  .adv-section {
    margin: 0 0 18px;
  }
  .adv-section h4 {
    margin: 0 0 4px;
    font-size: 13px;
    font-weight: 600;
    color: var(--fg);
  }
  .adv-section button {
    padding: 0.4rem 0.8rem;
    font-size: 12.5px;
  }
  .adv-section button.danger {
    background: rgba(255, 107, 107, 0.1);
    border-color: rgba(255, 107, 107, 0.4);
    color: var(--danger);
  }
  .adv-section button.danger:hover:not(:disabled) {
    background: rgba(255, 107, 107, 0.18);
  }
  .adv-section textarea {
    width: 100%;
    box-sizing: border-box;
    padding: 8px 10px;
    font-family: var(--font-mono);
    font-size: 11.5px;
    resize: vertical;
    margin: 0;
  }
  .adv-row {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-top: 0.4rem;
  }
  .import-status.ok {
    color: var(--mint);
    font-size: 12px;
  }
  .import-status.err {
    color: var(--danger);
    font-size: 12px;
  }
  .fingerprint {
    background: var(--surface-sunken);
    border: 1px solid var(--border);
    padding: 0.25rem 0.5rem;
    border-radius: 6px;
    font-family: var(--font-mono);
    font-size: 12px;
    user-select: all;
  }
  .step-stops {
    margin-top: 12px;
  }
  .step-stops-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
    gap: 2px 14px;
    margin-top: 4px;
  }
  .step-stop-row {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 12.5px;
    font-weight: 500;
    padding: 5px 0;
    border-bottom: none;
  }
  .step-stop-row > input[type="checkbox"] {
    order: 0;
    margin-left: 0;
    width: 30px;
    height: 18px;
  }
  .step-stop-row > input[type="checkbox"]::after {
    width: 12px;
    height: 12px;
  }
  .step-stop-row > input[type="checkbox"]:checked::after {
    left: 14px;
  }
  .step-stop-row .saved {
    margin-left: auto;
  }

  /* ---- Shortcuts tab ------------------------------------------- */
  /* Every selector below is new and scoped to this component; nothing
     borrows a class from another component, which is how a previous
     PR shipped styles that silently did not apply. */
  .linkish {
    background: none;
    border: none;
    padding: 0;
    color: var(--accent);
    font: inherit;
    cursor: pointer;
    text-decoration: underline;
  }
  .sc-actions {
    display: flex;
    align-items: center;
    gap: 10px;
    margin: 6px 0 10px;
  }
  .sc-note {
    margin: 0 0 10px;
    padding: 7px 10px;
    border-left: 2px solid var(--gold);
    background: var(--gold-soft);
    border-radius: 0 8px 8px 0;
    color: var(--fg-muted);
    font-size: 11.5px;
    line-height: 1.45;
  }
  fieldset.sc-group {
    border: 1px solid var(--border);
    border-radius: 10px;
    padding: 4px 12px 8px;
    margin: 0 0 12px;
  }
  fieldset.sc-group legend {
    padding: 0 6px;
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--fg-dim);
  }
  .sc-row {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 7px 0;
    border-bottom: 1px solid var(--border);
  }
  .sc-row:last-child {
    border-bottom: none;
  }
  .sc-label {
    display: flex;
    flex-direction: column;
    gap: 1px;
    min-width: 0;
    flex: 1 1 auto;
  }
  .sc-name {
    font-size: 12.5px;
    font-weight: 600;
    color: var(--fg);
  }
  .sc-desc {
    font-size: 11px;
    color: var(--fg-dim);
    line-height: 1.35;
  }
  .sc-conflict {
    font-size: 11px;
    color: var(--danger);
  }
  .sc-controls {
    display: flex;
    align-items: center;
    gap: 6px;
    flex: 0 0 auto;
  }
  .sc-capture {
    min-width: 92px;
    padding: 5px 10px;
    border-radius: 6px;
    border: 1px solid var(--border-strong);
    background: var(--surface-raised);
    color: var(--fg);
    font-family: var(--font-mono);
    font-size: 11px;
    cursor: pointer;
  }
  .sc-capture:hover {
    background: var(--surface-hover);
  }
  .sc-capture.listening {
    border-color: var(--gold);
    background: var(--gold-soft);
    color: var(--gold-strong);
  }
  .sc-row.conflict .sc-capture {
    border-color: var(--danger);
  }
  .sc-mini {
    padding: 4px 7px;
    border-radius: 6px;
    border: 1px solid var(--border);
    background: transparent;
    color: var(--fg-muted);
    font-size: 10px;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    cursor: pointer;
  }
  .sc-mini:hover:not(:disabled) {
    color: var(--fg);
    border-color: var(--border-strong);
  }
  .sc-mini:disabled {
    opacity: 0.35;
    cursor: default;
  }
</style>
