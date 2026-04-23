<script lang="ts">
  import { onMount } from "svelte";
  import {
    settings,
    settingsOpen,
    closeSettings,
    resetSettings,
    updateSettings,
    exportSettings,
    importSettings,
    fingerprintSettings,
  } from "../settings";
  import { STEP_IDS, STEP_LABELS, NO_PRIORITY_STEPS, type StepID } from "../turn";

  // Steps that grant priority — the only ones the per-step stops UI
  // surfaces. Untap and Cleanup are filtered out since the server
  // sentinel (priority_holder = -1) makes them un-stoppable anyway.
  const STOPPABLE_STEPS: readonly StepID[] = STEP_IDS.filter((id) => !NO_PRIORITY_STEPS.has(id));

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
  type Tab = "audio" | "animations" | "display" | "gameplay" | "accessibility" | "advanced";
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

  // Reset tab state when the modal reopens.
  $effect(() => {
    if ($settingsOpen) activeTab = "audio";
  });

  // Global keyboard shortcuts: Esc to close (when open), `,` to
  // toggle (when idle — not typing in an input). Installed once.
  onMount(() => {
    const handler = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement | null;
      const isEditable =
        target?.tagName === "INPUT" || target?.tagName === "TEXTAREA" || target?.isContentEditable;
      if (e.key === "Escape" && $settingsOpen) {
        e.preventDefault();
        closeSettings();
        return;
      }
      if (e.key === "," && !isEditable && !e.metaKey && !e.ctrlKey && !e.altKey) {
        e.preventDefault();
        settingsOpen.update((v) => !v);
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
              Pairs with the stops grid below: priority auto-passes through any step you haven't
              pinned. Use the &ldquo;&rarr; next stop&rdquo; toolbar button to fast-forward one stop
              at a time without enabling auto-pass globally.
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
              Smart auto-pass (skip stops with no legal response)
              {#if isFresh("gameplay.smartAutoPass")}<span class="saved">✓ saved</span>{/if}
            </label>
            <p class="help">
              Pairs with the stops grid. Auto-passes even at a pinned stop when the legality engine
              can't find anything you could cast or activate — so &ldquo;stop on upkeep&rdquo; means
              &ldquo;stop here if I have something to consider,&rdquo; not &ldquo;stop every time
              regardless.&rdquo; Turn off to demand a click at every stop.
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
  .settings-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.55);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 200;
  }
  .settings-panel {
    background: #1f2024;
    color: #eee;
    border-radius: 6px;
    width: min(720px, 92vw);
    max-height: min(88vh, 720px);
    display: flex;
    flex-direction: column;
    box-shadow: 0 12px 32px rgba(0, 0, 0, 0.4);
  }
  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.75rem 1rem;
    border-bottom: 1px solid #333;
  }
  header h2 {
    margin: 0;
    font-size: 1.1rem;
    letter-spacing: 0.02em;
  }
  .close {
    background: none;
    border: none;
    color: #aaa;
    font-size: 1.4rem;
    cursor: pointer;
    line-height: 1;
    padding: 0 0.25rem;
  }
  .close:hover {
    color: #fff;
  }
  .body {
    display: flex;
    flex: 1;
    min-height: 0;
  }
  nav {
    display: flex;
    flex-direction: column;
    width: 160px;
    border-right: 1px solid #333;
    padding: 0.5rem 0;
    flex-shrink: 0;
  }
  nav button {
    background: none;
    border: none;
    color: #bbb;
    text-align: left;
    padding: 0.5rem 1rem;
    cursor: pointer;
    font-size: 0.9rem;
  }
  nav button:hover {
    background: #2a2b30;
    color: #fff;
  }
  nav button.active {
    background: #2f6fb8;
    color: #fff;
  }
  section {
    flex: 1;
    padding: 1rem 1.25rem;
    overflow-y: auto;
  }
  section h3 {
    margin: 0 0 0.75rem 0;
    font-size: 1rem;
  }
  label {
    display: block;
    margin: 0.4rem 0;
    cursor: pointer;
  }
  label.inline {
    display: inline-block;
    margin-right: 1rem;
  }
  label.slider-row {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }
  label.slider-row > span:first-child {
    min-width: 11rem;
  }
  label.slider-row .value {
    min-width: 2.5rem;
    text-align: right;
    color: #aaa;
  }
  label.disabled {
    opacity: 0.5;
  }
  input[type="range"] {
    flex: 1;
  }
  select {
    background: #2a2b30;
    color: #eee;
    border: 1px solid #444;
    padding: 0.25rem 0.5rem;
  }
  fieldset {
    border: 1px solid #333;
    border-radius: 4px;
    padding: 0.5rem 0.75rem;
    margin: 0.75rem 0;
  }
  fieldset[disabled] {
    opacity: 0.45;
  }
  legend {
    color: #888;
    font-size: 0.85rem;
    padding: 0 0.25rem;
  }
  .help {
    color: #888;
    font-size: 0.85rem;
    margin: 0.25rem 0 0.75rem 0;
  }
  /* Danger-flagged settings get amber/red framing so an opt-in
     that might cost the player a turn can't be mistaken for a
     routine preference. */
  label.danger {
    color: #f0a868;
    font-weight: 600;
  }
  .danger-help {
    color: #e8a95a;
    border-left: 3px solid #d99a2e;
    padding-left: 0.55rem;
    background: rgba(217, 154, 46, 0.06);
    border-radius: 0 3px 3px 0;
  }
  .danger-help strong {
    color: #ffd07a;
    letter-spacing: 0.03em;
  }
  .saved {
    color: #6cc07a;
    font-size: 0.8rem;
    margin-left: 0.5rem;
  }
  kbd {
    background: #2a2b30;
    border: 1px solid #444;
    padding: 0 0.3rem;
    border-radius: 3px;
    font-family: ui-monospace, monospace;
    font-size: 0.85em;
  }
  .danger {
    background: #6b2020;
    color: #fff;
    border: 1px solid #8b2828;
    padding: 0.5rem 0.9rem;
    border-radius: 3px;
    cursor: pointer;
  }
  .danger:hover {
    background: #8b2828;
  }
  .adv-section {
    margin: 0.25rem 0 1.25rem 0;
  }
  .adv-section h4 {
    margin: 0 0 0.25rem 0;
    font-size: 0.95rem;
    color: #ddd;
  }
  .adv-section button {
    background: #2a2b30;
    color: #eee;
    border: 1px solid #444;
    padding: 0.4rem 0.8rem;
    border-radius: 3px;
    cursor: pointer;
  }
  .adv-section button:hover:not(:disabled) {
    background: #34363c;
  }
  .adv-section button:disabled {
    opacity: 0.5;
    cursor: default;
  }
  .adv-section button.danger {
    background: #6b2020;
    border-color: #8b2828;
  }
  .adv-section button.danger:hover {
    background: #8b2828;
  }
  .adv-section textarea {
    width: 100%;
    box-sizing: border-box;
    background: #15161a;
    color: #eee;
    border: 1px solid #333;
    border-radius: 3px;
    padding: 0.5rem;
    font-family: ui-monospace, monospace;
    font-size: 0.85em;
    resize: vertical;
  }
  .adv-row {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-top: 0.4rem;
  }
  .import-status.ok {
    color: #6cc07a;
    font-size: 0.85rem;
  }
  .import-status.err {
    color: #d77;
    font-size: 0.85rem;
  }
  .fingerprint {
    background: #15161a;
    border: 1px solid #333;
    padding: 0.25rem 0.5rem;
    border-radius: 3px;
    font-family: ui-monospace, monospace;
    font-size: 0.95em;
    user-select: all;
  }
  .step-stops {
    margin-top: 1rem;
  }
  .step-stops-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
    gap: 0.35rem 0.75rem;
    margin-top: 0.4rem;
  }
  .step-stop-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.95em;
  }
  .step-stop-row .saved {
    color: #6c9;
    font-size: 0.85em;
  }
</style>
