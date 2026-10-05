<script lang="ts">
  // GameMenu — the ⋯ menu (ADR 0111 Delivery PR 7, owner decision 3).
  //
  // For a seated player it is the last chip on the action dock's toggles
  // row, and it opens UPWARD, right-aligned: the dock sits in the screen's
  // bottom-right corner, so down or right would leave the screen. A
  // viewer with no dock (a spectator, an admin with no seat here) keeps a
  // smaller one on the command bar, opening downward (lib/gameMenu.ts).
  //
  // A seat still in an active game also has the Dice section (ADR 0121
  // §5): "Roll a d6", "Roll a d20" and "Flip a coin".
  //
  // One trigger, one popover, which shows one of three things:
  //   - the menu itself (role="menu", "game actions");
  //   - the vote launcher's form (moved here from VotingPanel's corner
  //     button), reached from "Call a vote…";
  //   - Concede's confirm, reached from "Concede…". It opens where the
  //     menu was, so the answer is a short move away.
  //
  // It closes on Escape (focus goes back to the ⋯ button), on a pointer
  // press outside it, and when focus tabs out of it. While it is open a
  // ModalLayer stands the global shortcuts down, as the command-bar menu
  // did, so typing a vote topic or a hand size presses nothing, and the
  // dock's Enter / Escape handler stands down too (foreignModalOpen).
  // Arrow keys, Home and End move between the menu's items.

  import { tick } from "svelte";
  import Icon from "../Icon.svelte";
  import ModalLayer from "../ModalLayer.svelte";
  import {
    TABLE_ROLL_ITEMS,
    clampMulligan,
    concedeTitle,
    parseVote,
    tableRollTitle,
    type GameMenuOptions,
  } from "../../gameMenu";
  import { adminChipTitle, adminSwitchLabel } from "../../admin";
  import { L } from "../../labels";

  type Props = GameMenuOptions & {
    // "up": the dock's chip. "down": the command bar's icon.
    placement?: "up" | "down";
  };
  const props: Props = $props();
  const placement = $derived(props.placement ?? "up");

  type Mode = "closed" | "menu" | "vote" | "concede";
  let mode = $state<Mode>("closed");
  let root: HTMLElement | undefined = $state();
  let trigger: HTMLButtonElement | undefined = $state();
  // The open popover, read from the root rather than bound: a swap from
  // the menu to the confirm unbinds the old one after the new one binds.
  function pop(): HTMLElement | null {
    return root?.querySelector<HTMLElement>(".pop") ?? null;
  }

  let mulliganTo = $state(7);
  let topic = $state("");
  let optionsText = $state("yes, no");

  const cannotConcede = $derived(props.eliminated === true || props.gameEnded === true);

  function items(): HTMLElement[] {
    const el = pop();
    if (!el) return [];
    return [...el.querySelectorAll<HTMLElement>('[role="menuitem"]')].filter(
      (el) => !(el as HTMLButtonElement).disabled,
    );
  }

  async function show(next: Mode): Promise<void> {
    mode = next;
    await tick();
    if (next === "menu") items()[0]?.focus();
    else if (next === "vote") pop()?.querySelector<HTMLInputElement>("input")?.focus();
    else if (next === "concede") pop()?.querySelector<HTMLButtonElement>("button.keep")?.focus();
  }
  function close(refocus: boolean): void {
    if (mode === "closed") return;
    mode = "closed";
    if (refocus) trigger?.focus();
  }
  function toggle(): void {
    if (mode === "closed") void show("menu");
    else close(false);
  }
  // A menu entry: close, then act.
  function via(fn: () => void): () => void {
    return () => {
      close(false);
      fn();
    };
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
    if (e.key !== "Escape" || mode === "closed") return;
    e.preventDefault();
    e.stopPropagation();
    close(true);
  }
  function onWindowPointerDown(e: PointerEvent): void {
    if (mode !== "closed" && root && !root.contains(e.target as Node)) close(false);
  }
  function onFocusOut(e: FocusEvent): void {
    const to = e.relatedTarget as Node | null;
    if (mode !== "closed" && to && root && !root.contains(to)) close(false);
  }

  function requestConcede(): void {
    if (cannotConcede) return;
    void show("concede");
  }
  function confirmConcede(): void {
    close(false);
    if (cannotConcede) return;
    props.onConcede();
  }
  function startVote(e: SubmitEvent): void {
    e.preventDefault();
    const v = parseVote(topic, optionsText);
    if (!v) return;
    props.onStartVote(v.topic, v.options);
    topic = "";
    optionsText = "yes, no";
    close(true);
  }
</script>

<svelte:window onkeydown={onWindowKey} onpointerdown={onWindowPointerDown} />

<div
  class="game-menu"
  class:up={placement === "up"}
  class:down={placement === "down"}
  bind:this={root}
  role="presentation"
  onfocusout={onFocusOut}
>
  <button
    type="button"
    class="more-btn"
    class:on={mode !== "closed"}
    bind:this={trigger}
    aria-haspopup="menu"
    aria-expanded={mode !== "closed"}
    aria-label={L.moreActions}
    title={props.seated ? "sandbox actions and more" : "life history, the table and more"}
    onclick={toggle}><Icon name="more" size={placement === "up" ? 15 : 17} /></button
  >

  {#if mode !== "closed"}
    <ModalLayer />
  {/if}

  {#if mode === "menu"}
    <div
      class="pop menu"
      role="menu"
      aria-label={L.gameActions}
      tabindex="-1"
      onkeydown={onMenuKey}
    >
      {#if props.seated}
        <div class="menu-h">Sandbox</div>
        <button class="mi" role="menuitem" onclick={via(props.onDraw)} title={props.drawTitle}>
          <Icon name="draw" size={15} /> Draw a card
          {#if props.drawKey}<span class="mi-r">{props.drawKey}</span>{/if}
        </button>
        <button class="mi" role="menuitem" onclick={via(props.onUntapAll)}>
          <Icon name="untap" size={15} /> Untap all
        </button>
        <button class="mi" role="menuitem" onclick={via(props.onShuffle)}>
          <Icon name="shuffle" size={15} /> Shuffle library
        </button>
        <div class="mi mi-row">
          <Icon name="hand" size={15} />
          <span>Mulligan to</span>
          <input
            type="number"
            min="0"
            max="20"
            bind:value={mulliganTo}
            aria-label="mulligan hand size"
          />
          <button class="sm" onclick={via(() => props.onMulligan(clampMulligan(mulliganTo)))}
            >Go</button
          >
        </div>
        <button class="mi" role="menuitem" onclick={via(props.onLifeHistory)}>
          <Icon name="drop" size={15} /> Life history
        </button>
        <div class="sep"></div>
      {/if}
      {#if props.tableRoll}
        <!-- ADR 0121 §5: a die or a coin at the table, for fun. Never a
             game roll and never undone. Disabled for 2 s after the
             viewer's own roll, matching the server's rate. -->
        {@const ready = props.tableRoll.ready}
        <div class="menu-h">{L.dice}</div>
        {#each TABLE_ROLL_ITEMS as item (item.die)}
          <button
            class="mi"
            role="menuitem"
            disabled={!ready}
            title={tableRollTitle(ready)}
            onclick={via(() => props.onTableRoll?.(item.die))}
          >
            <Icon name={item.die === "coin" ? "coin" : "die"} size={15} />
            {item.label}
          </button>
        {/each}
        <div class="sep"></div>
      {/if}
      <div class="menu-h">Table</div>
      {#if !props.seated}
        <button class="mi" role="menuitem" onclick={via(props.onLifeHistory)}>
          <Icon name="drop" size={15} /> Life history
        </button>
      {/if}
      <!-- ADR 0075 §2.5. Open to everyone, because the settings are
           public on purpose; the panel disables its own controls for
           anyone who is not the host or the admin. -->
      <button
        class="mi"
        role="menuitem"
        onclick={via(props.onTableSettings)}
        title={props.canManage
          ? "the table's house rules — undos, life, commander damage, bot speed, spawning"
          : "the table's house rules (only the host can change them)"}
      >
        <Icon name="gear" size={15} /> Table settings…
        {#if !props.canManage}<span class="mi-r">view</span>{/if}
      </button>
      {#if props.spawnAvailable}
        <!-- Both of the server's gates, checked together: the host or
             admin, AND the table's spawn switch. Absent rather than
             disabled when the switch is off. -->
        <button
          class="mi"
          role="menuitem"
          onclick={via(props.onSpawn)}
          title="put a card or a token on the table — announced in the game log, and undoable"
        >
          <Icon name="spark" size={15} /> Spawn a card or token…
        </button>
      {/if}
      {#if props.seated && !props.voteOpen}
        <!-- The vote launcher, out of the board's top-left corner (ADR
             0111 inventory, "Vote launcher"). An open vote is the dock's
             own request, so this stands down while one is open. -->
        <button
          class="mi"
          role="menuitem"
          onclick={() => void show("vote")}
          title="call a vote — a topic and the options, for every seat to answer"
        >
          <Icon name="check" size={15} /> Call a vote…
        </button>
      {/if}
      {#if props.discordLink}
        <!-- A navigation, not a fetch: the server answers with a 302 to
             Discord's consent screen and comes back to this table with
             the seat linked. -->
        <a
          class="mi"
          role="menuitem"
          href={props.discordLink.href}
          title="sign in with Discord and put your Discord name and avatar on this seat"
        >
          <Icon name="link" size={15} />
          {props.discordLink.label}
        </a>
      {/if}
      {#if props.myGames}
        <button class="mi" role="menuitem" onclick={via(props.onMyGames)}>
          <Icon name="library" size={15} /> My games
        </button>
      {/if}
      {#if props.adminMode && props.onAdminMode}
        <!-- ADR 0112 §2 item 9: the header's Admin chip, here because the
             table has no header. The connection reconnects with the
             new mode's binding (a 4001 from the server). -->
        <button
          class="mi"
          role="menuitem"
          onclick={via(props.onAdminMode)}
          title={adminChipTitle(props.adminMode.on)}
        >
          <Icon name="bolt" size={15} />
          {adminSwitchLabel(props.adminMode.on)}
          <span class="mi-r">{props.adminMode.on ? "admin" : "player"}</span>
        </button>
      {/if}
      <button class="mi" role="menuitem" onclick={via(props.onBack)}>
        <Icon name="chevronLeft" size={15} /> Back to lobby
      </button>
      {#if props.seated}
        <div class="sep"></div>
        <!-- Last, and confirmed: Concede is irreversible. Disabled, not
             hidden, once it means nothing. -->
        <button
          class="mi danger"
          role="menuitem"
          onclick={requestConcede}
          disabled={cannotConcede}
          title={concedeTitle(props)}
        >
          <Icon name="flag" size={15} /> Concede…
        </button>
      {/if}
    </div>
  {:else if mode === "vote"}
    <form class="pop vote-form" aria-label="call a vote" onsubmit={startVote}>
      <div class="menu-h">Call a vote</div>
      <input
        type="text"
        bind:value={topic}
        placeholder="topic (e.g. 'monarchy?')"
        aria-label="vote topic"
      />
      <input
        type="text"
        bind:value={optionsText}
        placeholder="options, comma-separated"
        aria-label="vote options"
      />
      <div class="pop-actions">
        <button type="button" onclick={() => close(true)}>cancel</button>
        <button type="submit" class="go" disabled={parseVote(topic, optionsText) === null}
          >start</button
        >
      </div>
    </form>
  {:else if mode === "concede"}
    <div class="pop confirm" role="dialog" aria-modal="true" aria-label="concede the game?">
      <div class="confirm-title">Concede the game?</div>
      <p class="confirm-body">
        You'll be eliminated and keep watching as a spectator. This can't be undone.
      </p>
      <div class="pop-actions">
        <button type="button" class="keep" onclick={() => close(true)}>Keep playing</button>
        <button type="button" class="danger" onclick={confirmConcede}
          ><Icon name="flag" size={14} /> Concede</button
        >
      </div>
    </div>
  {/if}
</div>

<style>
  .game-menu {
    position: relative;
    flex: 0 0 auto;
    display: flex;
  }

  /* The dock's chip: the toggles row's quiet button look (ActionDock's
     .action), icon only; its name is its aria-label. */
  .up .more-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 30px;
    padding: 4px 7px;
    border-radius: 6px;
    border: 1px solid var(--border);
    background: var(--overlay-faint);
    color: var(--fg);
    cursor: pointer;
    box-shadow: none;
  }
  .up .more-btn:hover,
  .up .more-btn.on {
    background: color-mix(in srgb, var(--overlay-ink) 8%, transparent);
    border-color: color-mix(in srgb, var(--border) 60%, var(--overlay-ink) 40%);
  }
  /* The command bar's icon button (Game.svelte's .ibtn). */
  .down .more-btn {
    width: 32px;
    height: 32px;
    padding: 0;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 8px;
    border: 1px solid transparent;
    background: transparent;
    color: var(--fg-muted);
    box-shadow: none;
  }
  .down .more-btn:hover,
  .down .more-btn.on {
    color: var(--fg);
    background: var(--overlay);
    border-color: var(--border);
  }
  .more-btn:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }

  /* Right-aligned with the trigger. Upward from the dock, downward from
     the command bar. Never taller than the room it has: it scrolls. */
  .pop {
    position: absolute;
    right: 0;
    z-index: 60;
    width: 284px;
    max-width: calc(100vw - 24px);
    box-sizing: border-box;
    background: var(--surface);
    border: 1px solid var(--border-strong);
    border-radius: 12px;
    box-shadow: var(--shadow-lg);
    padding: 6px;
    display: flex;
    flex-direction: column;
    gap: 2px;
    color: var(--fg);
    overflow-y: auto;
  }
  .up .pop {
    bottom: calc(100% + 8px);
    max-height: min(560px, calc(100vh - 120px));
    transform-origin: bottom right;
    animation: pop-up 140ms var(--ease);
  }
  .down .pop {
    top: calc(100% + 6px);
    max-height: calc(100vh - 80px);
  }
  @keyframes pop-up {
    from {
      opacity: 0;
      transform: translateY(6px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
  .pop:focus {
    outline: none;
  }
  .menu-h {
    font-family: var(--font-mono);
    font-size: 9.5px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--fg-dim);
    padding: 8px 10px 4px;
  }
  .mi {
    display: flex;
    align-items: center;
    gap: 10px;
    min-height: 32px;
    padding: 0 10px;
    border-radius: 8px;
    border: 1px solid transparent;
    background: transparent;
    color: var(--fg);
    font-size: 12.5px;
    font-weight: 500;
    width: 100%;
    justify-content: flex-start;
    text-align: left;
    box-shadow: none;
    box-sizing: border-box;
    flex: none;
  }
  a.mi {
    text-decoration: none;
  }
  .mi:hover:not(:disabled),
  .mi:focus-visible {
    background: var(--overlay);
    border-color: transparent;
    outline: none;
  }
  .mi:focus-visible {
    box-shadow: inset 0 0 0 1px var(--accent);
  }
  .mi:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
  .mi.danger {
    color: var(--danger);
  }
  .mi-r {
    margin-left: auto;
    font-family: var(--font-mono);
    font-size: 10.5px;
    color: var(--fg-muted);
    background: var(--surface-raised);
    border: 1px solid var(--border);
    border-radius: 6px;
    padding: 2px 6px;
  }
  .mi-row {
    cursor: default;
  }
  .mi-row input {
    width: 44px;
    padding: 3px 6px;
    margin: 0 0 0 auto;
    font-size: 12px;
    font-family: var(--font-mono);
    border-radius: 6px;
  }
  .mi-row .sm {
    height: 24px;
    padding: 0 8px;
    font-size: 11.5px;
    border-radius: 6px;
  }
  .sep {
    flex: none;
    height: 1px;
    background: var(--border);
    margin: 4px 6px;
  }

  /* The vote launcher's form and Concede's confirm open where the menu
     was. */
  .vote-form {
    gap: 6px;
    padding: 6px 10px 10px;
  }
  .vote-form .menu-h {
    padding: 4px 0 0;
  }
  .vote-form input {
    background: var(--surface-sunken);
    border: 1px solid var(--border);
    color: var(--fg);
    border-radius: var(--radius-sm);
    padding: 6px 8px;
    font: inherit;
    font-size: 12.5px;
  }
  .vote-form input:focus {
    outline: none;
    border-color: var(--magenta);
    box-shadow: 0 0 0 2px color-mix(in srgb, var(--magenta) 25%, transparent);
  }
  .confirm {
    gap: 8px;
    padding: 14px 14px 12px;
    width: 300px;
  }
  .confirm-title {
    font-family: var(--font-display);
    font-size: 15px;
    font-weight: 700;
    color: var(--fg);
  }
  .confirm-body {
    margin: 0;
    font-size: 12.5px;
    color: var(--fg-muted);
    line-height: 1.45;
  }
  .pop-actions {
    display: flex;
    justify-content: flex-end;
    gap: 6px;
    margin-top: 4px;
  }
  .vote-form .go {
    color: var(--magenta);
    border-color: color-mix(in srgb, var(--magenta) 60%, transparent);
  }

  @media (max-width: 599px) {
    /* A phone's toggles row has five chips in 358px. */
    .up .more-btn {
      min-width: 36px;
      padding: 4px 6px;
    }
    .mi {
      min-height: 40px;
    }
  }
</style>
