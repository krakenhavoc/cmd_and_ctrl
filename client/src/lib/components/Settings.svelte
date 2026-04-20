<script lang="ts">
  import { onMount } from "svelte";
  import {
    settings,
    settingsOpen,
    closeSettings,
    resetSettings,
    updateSettings,
  } from "../settings";

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
            <label class="slider-row">
              <span>Theme</span>
              <select
                value={$settings.display.theme}
                onchange={(e) =>
                  change(
                    "display",
                    "theme",
                    e.currentTarget.value as "dark" | "light" | "high-contrast",
                  )}
              >
                <option value="dark">Dark (default)</option>
                <option value="light">Light</option>
                <option value="high-contrast">High contrast</option>
              </select>
              {#if isFresh("display.theme")}<span class="saved">✓</span>{/if}
            </label>
            <p class="help">Theme scaffolding is in place; full palettes ship in a follow-up.</p>

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
              Auto-pass priority when I have nothing playable
              {#if isFresh("gameplay.autoPassPriority")}<span class="saved">✓ saved</span>{/if}
            </label>
            <p class="help">
              Hold <kbd>Shift</kbd> while clicking the pass button to override for one step.
            </p>

            <p class="help">
              Per-step stops grid will land with the client-side timing affordance in S13.3.
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
            <p class="help">
              Settings live under <code>localStorage["cmdctrl.settings.v1"]</code>. Export / import
              + per-setting reset land in the next S11.5 PR.
            </p>
            <button class="danger" onclick={onConfirmReset}>Reset all settings</button>
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
</style>
