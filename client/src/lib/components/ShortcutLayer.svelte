<script lang="ts">
  // The one and only global keydown listener for shortcuts.
  //
  // Mounted at the app shell so there is exactly one of it. Every
  // decision it makes is `dispatchShortcut`'s — this file reads the
  // event, hands the facts over, and calls whatever handler comes
  // back. Nothing about typing detection, modal precedence or move
  // legality is decided here, because none of that is testable here.
  //
  // What it deliberately does NOT touch: Escape, Enter and Tab. Every
  // modal in the game closes on Escape and confirms on Enter, and
  // Game.svelte cancels targeting on Escape and confirms a
  // multi-target pick on Enter. Those local handlers keep working
  // exactly as they did; shortcuts.RESERVED_CHORDS makes it
  // impossible to bind over them by accident.
  import { settings, settingsOpen } from "../settings";
  import { modalOpen } from "../modalLayers";
  import { currentHandlers, shortcutContext, openShortcutsHelp } from "../shortcutRuntime";
  import { toggleMuted } from "../sounds";
  import {
    chordFrom,
    dispatchShortcut,
    effectiveBindings,
    elementOf,
    isInteractiveFocus,
    isTypingTarget,
    type ShortcutID,
  } from "../shortcuts";
  import ShortcutsOverlay from "./ShortcutsOverlay.svelte";

  const bindings = $derived(effectiveBindings($settings.shortcuts.bindings));

  // Actions that work with no game mounted. A route that can do
  // better registers its own handler for the same id and wins — the
  // game route does this for `toggleMute` so the command bar's
  // speaker icon re-renders with the new state.
  const globalActions: Partial<Record<ShortcutID, () => void>> = {
    openSettings: () => settingsOpen.update((v) => !v),
    // Only ever opens. The overlay owns its own close, because by the
    // time it is up it has registered a modal layer and this
    // dispatcher has stood down — which is exactly the behaviour we
    // want from every modal.
    toggleHelp: () => openShortcutsHelp(),
    toggleMute: () => {
      toggleMuted();
    },
  };

  // Transient "that key did nothing, and here's why" line. Only for a
  // key that IS bound and IS unavailable — an unbound key stays
  // silent, and so does a game action pressed outside a game, because
  // "only at a game table" is noise on the lobby screen.
  let hint = $state<string | null>(null);
  let hintTimer: ReturnType<typeof setTimeout> | null = null;
  function flashHint(message: string): void {
    hint = message;
    if (hintTimer) clearTimeout(hintTimer);
    hintTimer = setTimeout(() => {
      hint = null;
      hintTimer = null;
    }, 1600);
  }

  function onKeydown(ev: KeyboardEvent): void {
    const result = dispatchShortcut({
      chord: chordFrom(ev),
      repeat: ev.repeat,
      typing: isTypingTarget(elementOf(ev.target)),
      modalOpen: $modalOpen,
      interactiveFocus: isInteractiveFocus(elementOf(document.activeElement)),
      enabled: $settings.shortcuts.enabled,
      bindings,
      ctx: $shortcutContext,
    });

    if (result.skipped === "unavailable") {
      const reason = result.availability?.reason;
      if (reason && reason !== "only at a game table") {
        // The key is real and the player pressed it on purpose;
        // swallowing the press AND the explanation is how a keymap
        // earns a reputation for being broken.
        ev.preventDefault();
        flashHint(reason);
      }
      return;
    }

    if (!result.action) return;
    const handler = currentHandlers()[result.action] ?? globalActions[result.action];
    if (!handler) return;
    ev.preventDefault();
    handler();
  }
</script>

<svelte:window onkeydown={onKeydown} />

<ShortcutsOverlay />

{#if hint}
  <div class="sc-toast" role="status" aria-live="polite">{hint}</div>
{/if}

<style>
  /* Bottom-centre, pointer-transparent, gone in under two seconds.
     Deliberately not in the board's attention strip: it is about the
     key you just pressed, not about the game state, and it must not
     push the table around. No transition, so there is nothing for the
     reduced-motion rule to flatten. */
  .sc-toast {
    position: fixed;
    left: 50%;
    bottom: 12px;
    transform: translateX(-50%);
    z-index: 300;
    pointer-events: none;
    padding: 5px 11px;
    border-radius: 999px;
    background: rgba(0, 0, 0, 0.82);
    border: 1px solid var(--border-strong);
    color: var(--fg-muted);
    font-size: 11px;
    font-family: var(--font-ui);
    white-space: nowrap;
  }
</style>
