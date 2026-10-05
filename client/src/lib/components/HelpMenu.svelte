<script lang="ts">
  // HelpMenu — the site header's Help (ADR 0125 §6): a `button "help"`
  // that opens `menu "help"`, for everyone, signed in or not.
  //
  //   - Tips for this page: forgets this page's tips (and the site's,
  //     after them) and shows them one after another, even with tips off
  //     and past the one-per-visit rule, because the person asked.
  //   - Show all tips again: forgets every dismissed tip and switches
  //     tips back on. They then appear as pages are visited.
  //   - Practice game: #/practice. A practice table needs a session, so
  //     signed out it reads "Practice game (sign in first)" and goes to
  //     #/login.
  //   - Keyboard shortcuts: the `?` overlay.
  //
  // The menu is named by its button (aria-labelledby), so the one
  // contract name "help" covers both. It closes on Escape (focus goes
  // back to the button), on a click outside it, and when focus tabs out
  // of it. Arrow keys, Home and End move between its items, as in the ⋯
  // menu at the table.

  import { tick } from "svelte";
  import { route, navigate } from "../router";
  import { session } from "../session";
  import { openShortcutsHelp } from "../shortcutRuntime";
  import { HINTS } from "../hints";
  import { placeOfRoute, type Hint } from "../hints/hint";
  import { replayableTips, replayTips, showAllTipsAgain } from "../hints/runtime";
  import { L } from "../labels";
  import Icon from "./Icon.svelte";

  interface Props {
    /** Every hint; the collected ones unless a test hands it others. */
    hints?: readonly Hint[];
  }
  const { hints = HINTS }: Props = $props();

  let open = $state(false);
  let root = $state<HTMLElement | null>(null);
  let button = $state<HTMLButtonElement | null>(null);

  const place = $derived(placeOfRoute($route.name));
  // Login, the invite and reclaim doors and the practice door have no
  // tips (ADR 0125 §3.8), and neither has a place with none to show.
  const pageTips = $derived(place === null ? [] : replayableTips(place, hints));
  const hasSession = $derived($session !== null);

  function items(): HTMLElement[] {
    return [...(root?.querySelectorAll<HTMLElement>('[role="menuitem"]') ?? [])].filter(
      (el) => !(el as HTMLButtonElement).disabled,
    );
  }

  async function toggle(): Promise<void> {
    open = !open;
    if (!open) return;
    await tick();
    items()[0]?.focus();
  }

  function close(refocus: boolean): void {
    if (!open) return;
    open = false;
    if (refocus) button?.focus();
  }

  function tipsForThisPage(): void {
    close(false);
    if (place !== null) replayTips(place, hints);
  }

  function allTipsAgain(): void {
    close(false);
    showAllTipsAgain();
  }

  function practice(e: MouseEvent): void {
    e.preventDefault();
    close(false);
    navigate(hasSession ? "#/practice" : "#/login");
  }

  function shortcuts(): void {
    close(false);
    openShortcutsHelp();
  }

  function onMenuKey(e: KeyboardEvent): void {
    const list = items();
    if (list.length === 0) return;
    const at = list.indexOf(document.activeElement as HTMLElement);
    let next = -1;
    if (e.key === "ArrowDown") next = at < 0 ? 0 : (at + 1) % list.length;
    else if (e.key === "ArrowUp")
      next = at < 0 ? list.length - 1 : (at - 1 + list.length) % list.length;
    else if (e.key === "Home") next = 0;
    else if (e.key === "End") next = list.length - 1;
    else return;
    e.preventDefault();
    list[next]?.focus();
  }

  function onWindowKey(e: KeyboardEvent): void {
    if (e.key !== "Escape" || !open) return;
    e.preventDefault();
    close(true);
  }

  function onWindowClick(e: MouseEvent): void {
    if (!open || !root) return;
    if (e.target instanceof Node && root.contains(e.target)) return;
    close(false);
  }

  function onFocusOut(e: FocusEvent): void {
    const to = e.relatedTarget as Node | null;
    if (open && to && root && !root.contains(to)) close(false);
  }
</script>

<svelte:window onkeydown={onWindowKey} onclick={onWindowClick} />

<div class="help" bind:this={root} role="presentation" onfocusout={onFocusOut}>
  <button
    type="button"
    class="help-btn"
    class:on={open}
    id="help-button"
    bind:this={button}
    aria-label={L.help}
    aria-haspopup="menu"
    aria-expanded={open}
    aria-controls="help-menu"
    title="Help: tips, a practice game and the keyboard shortcuts"
    onclick={() => void toggle()}
  >
    <Icon name="help" size={18} />
  </button>

  {#if open}
    <div
      class="help-menu"
      id="help-menu"
      role="menu"
      aria-labelledby="help-button"
      tabindex="-1"
      onkeydown={onMenuKey}
    >
      <button
        type="button"
        class="mi"
        role="menuitem"
        disabled={pageTips.length === 0}
        title={pageTips.length === 0 ? "this page has no tips" : undefined}
        onclick={tipsForThisPage}
      >
        {L.tipsForThisPage}
      </button>
      <button type="button" class="mi" role="menuitem" onclick={allTipsAgain}>
        {L.showAllTipsAgain}
      </button>
      <a class="mi" role="menuitem" href={hasSession ? "#/practice" : "#/login"} onclick={practice}>
        {hasSession ? L.practiceGame : `${L.practiceGame} (sign in first)`}
      </a>
      <button type="button" class="mi" role="menuitem" onclick={shortcuts}>
        {L.keyboardShortcutsItem}
      </button>
    </div>
  {/if}
</div>

<style>
  .help {
    position: relative;
    display: flex;
    align-items: center;
  }
  /* The account button's ghost look, icon only; its name is its
     aria-label. */
  .help-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 30px;
    height: 30px;
    padding: 0;
    border-radius: var(--radius);
    border: 1px solid transparent;
    background: transparent;
    color: var(--fg-muted);
    box-shadow: none;
    cursor: pointer;
  }
  .help-btn:hover,
  .help-btn.on {
    color: var(--fg);
    background: var(--surface-hover);
    border-color: var(--border);
  }
  .help-btn:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
  /* Drops from the button, right-aligned with it, like the account
     menu beside it. */
  .help-menu {
    position: absolute;
    top: calc(100% + 6px);
    right: 0;
    z-index: 40;
    min-width: 220px;
    max-width: calc(100vw - 28px);
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 6px;
    border-radius: 12px;
    border: 1px solid var(--border-strong);
    background: var(--surface-raised);
    box-shadow: var(--shadow-lg);
  }
  .mi {
    display: flex;
    align-items: center;
    justify-content: flex-start;
    width: 100%;
    min-height: 34px;
    padding: 0 10px;
    border: 0;
    border-radius: var(--radius);
    background: transparent;
    color: var(--fg);
    font-size: 12.5px;
    font-weight: 600;
    text-align: left;
    text-decoration: none;
    box-sizing: border-box;
    cursor: pointer;
  }
  .mi:hover:not(:disabled),
  .mi:focus-visible {
    background: var(--surface-hover);
  }
  .mi:disabled {
    color: var(--fg-dim);
    cursor: default;
  }
</style>
