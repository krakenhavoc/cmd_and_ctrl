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
  // §5): "Roll a d6", "Roll a d20" and "Flip a coin". Everyone has the
  // Help section after Table (ADR 0125 §6).
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
  import LegalNotice from "../LegalNotice.svelte";
  import {
    TABLE_ROLL_ITEMS,
    clampMulligan,
    concedeTitle,
    parseVote,
    tableRollTitle,
    type GameMenuOptions,
  } from "../../gameMenu";
  import { adminChipTitle, adminSwitchLabel } from "../../admin";
  import { HAND_SORTS } from "../../handOrder";
  import { L } from "../../labels";
  import {
    BLOCK_SPECTATOR_LABEL,
    REMOVE_SPECTATOR_LABEL,
    REPORT_INTRO,
    REPORT_MENU_LABEL,
    REPORT_REASON_MAX,
    REPORT_SEATS_GROUP,
    REPORT_SENT,
    REPORT_SPECTATORS_GROUP,
    REPORT_TITLE,
    SPECTATORS_EMPTY,
    SPECTATORS_HINT,
    SPECTATORS_MENU_LABEL,
    SPECTATORS_TITLE,
    SPECTATOR_BLOCKED_NOTE,
    SPECTATOR_NO_NAME_NOTE,
    SPECTATOR_REMOVED_NOTE,
    canSendReport,
    reportErrorMessage,
    spectatorErrorMessage,
    spectatorTargets,
  } from "../../moderation";
  import type { SpectatorLine } from "../../api";

  type Props = GameMenuOptions & {
    // "up": the dock's chip. "down": the command bar's icon.
    placement?: "up" | "down";
  };
  const props: Props = $props();
  const placement = $derived(props.placement ?? "up");

  type Mode = "closed" | "menu" | "vote" | "concede" | "legal" | "report" | "spectators";
  let mode = $state<Mode>("closed");
  let root: HTMLElement | undefined = $state();
  let trigger: HTMLButtonElement | undefined = $state();
  // The open popover, read from the root rather than bound: a swap from
  // the menu to the confirm unbinds the old one after the new one binds.
  function pop(): HTMLElement | null {
    return root?.querySelector<HTMLElement>(".pop") ?? null;
  }

  let mulliganTo = $state(7);
  // The report form (#2984). Left filled after a refusal so a retry costs
  // nothing; cleared once a report is sent.
  let reportTarget = $state("");
  let reportReason = $state("");
  let reportBusy = $state(false);
  let reportError = $state("");
  let reportSent = $state(false);
  // Who is watching (#3113): read when the report form or the Spectators
  // panel opens, so it is as fresh as the click, and not polled.
  let watchers = $state<SpectatorLine[]>([]);
  let watchError = $state("");
  let watchNote = $state("");
  let watchBusy = $state<string | null>(null);
  const reportTargetsAll = $derived([
    ...(props.reportTargets ?? []),
    ...spectatorTargets(watchers),
  ]);
  const spectatorReportable = $derived(watchers.length > 0);
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
    else if (next === "report") pop()?.querySelector<HTMLSelectElement>("select")?.focus();
    else if (next === "spectators") pop()?.querySelector<HTMLButtonElement>("button")?.focus();
    else if (next === "concede" || next === "legal")
      pop()?.querySelector<HTMLButtonElement>("button.keep")?.focus();
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
  async function loadWatchers(): Promise<void> {
    if (!props.loadSpectators) return;
    try {
      watchers = await props.loadSpectators();
      watchError = "";
    } catch (err) {
      // The list is a convenience for the report form: a failed read
      // leaves the seats reportable and says nothing there. The panel
      // shows it.
      watchers = [];
      watchError = spectatorErrorMessage(err);
    }
  }
  function openReport(): void {
    reportError = "";
    reportSent = false;
    const targets = props.reportTargets ?? [];
    if (!targets.some((t) => t.id === reportTarget)) reportTarget = targets[0]?.id ?? "";
    void show("report");
    void loadWatchers().then(() => {
      if (!reportTargetsAll.some((t) => t.id === reportTarget)) {
        reportTarget = reportTargetsAll[0]?.id ?? "";
      }
    });
  }
  async function openSpectators(): Promise<void> {
    watchNote = "";
    watchError = "";
    void show("spectators");
    await loadWatchers();
  }
  async function removeWatcher(w: SpectatorLine, block: boolean): Promise<void> {
    if (!props.onRemoveSpectator || watchBusy !== null) return;
    watchBusy = w.id;
    watchError = "";
    watchNote = "";
    try {
      const out = await props.onRemoveSpectator(w.id, block);
      watchers = out.spectators;
      watchNote = !block
        ? SPECTATOR_REMOVED_NOTE
        : out.blocked
          ? SPECTATOR_BLOCKED_NOTE
          : SPECTATOR_NO_NAME_NOTE;
    } catch (err) {
      // Read the list again (the spectator may simply be gone), then
      // put the refusal back: a successful read clears the error line.
      const refusal = spectatorErrorMessage(err);
      await loadWatchers();
      watchError = refusal;
    } finally {
      watchBusy = null;
    }
  }
  async function sendReport(e: SubmitEvent): Promise<void> {
    e.preventDefault();
    if (reportBusy || !props.onReport || !canSendReport(reportTarget, reportReason)) return;
    reportBusy = true;
    reportError = "";
    try {
      const picked = reportTargetsAll.find((t) => t.id === reportTarget);
      if (picked?.kind === "spectator") {
        await props.onReport(reportTarget, reportReason.trim(), "spectator");
      } else {
        await props.onReport(reportTarget, reportReason.trim());
      }
      reportSent = true;
      reportReason = "";
    } catch (err) {
      reportError = reportErrorMessage(err);
    } finally {
      reportBusy = false;
    }
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
        {#if props.sandbox !== false}
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
          {#if props.sandboxPassTurn && props.onSandboxPassTurn}
            <!-- ADR 0143 §4.2: the old Pass turn, a sandbox jump. The
               dock's End turn is the rules-faithful way to end a turn. -->
            <button
              class="mi"
              role="menuitem"
              onclick={via(props.onSandboxPassTurn)}
              title="jump straight to the next turn: no end step, no end-step triggers, no cleanup discard. The log shows the table you did it"
            >
              <Icon name="chevronRight" size={15} />
              {L.sandboxPassTurn}
            </button>
          {/if}
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
        {/if}
        <button class="mi" role="menuitem" onclick={via(props.onLifeHistory)}>
          <Icon name="drop" size={15} /> Life history
        </button>
        <div class="sep"></div>
      {/if}
      {#if props.seated && props.sortHand && props.onSortHand}
        <!-- #3028: the one-time hand sorts, moved here from a button on
             the hand. The result becomes the saved order. -->
        {@const onSort = props.onSortHand}
        <div class="menu-h">
          {L.sortHand}
          {#if props.sortKey}<span class="mi-r">{props.sortKey}</span>{/if}
        </div>
        {#each HAND_SORTS as s (s.key)}
          <button class="mi" role="menuitem" onclick={via(() => onSort(s.key))}>
            <Icon name="hand" size={15} />
            {s.label}
          </button>
        {/each}
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
          ? "the table's house rules — undos, life, commander damage, table pace, spawning"
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
      {#if props.seated && !props.voteOpen && props.sandbox !== false}
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
      {#if props.seated && props.onReport && ((props.reportTargets?.length ?? 0) > 0 || props.loadSpectators)}
        <!-- #2984: tell the people who run the server about a player. -->
        <button
          class="mi"
          role="menuitem"
          onclick={openReport}
          title="report a player's name, chat or behaviour to the server's admins"
        >
          <Icon name="flag" size={15} />
          {REPORT_MENU_LABEL}
        </button>
      {/if}
      {#if props.canRemoveSpectators && props.loadSpectators && props.onRemoveSpectator}
        <!-- #3113: see who is watching and remove (or block) one. The
             host's and the admin's; the server decides. Any game state. -->
        <button
          class="mi"
          role="menuitem"
          onclick={openSpectators}
          title="see who is watching, and remove or block a spectator"
        >
          <Icon name="scroll" size={15} />
          {SPECTATORS_MENU_LABEL}
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
      <!-- ADR 0125 §6: Help, after Table. The tutorial is replayed on
           the practice table only; a real table offers no practice
           game, which would take the player out of their seat. -->
      <div class="sep"></div>
      <div class="menu-h">Help</div>
      <button
        class="mi"
        role="menuitem"
        onclick={via(() => props.onTableTips?.())}
        title="show the table's tips again, at the next quiet moments"
      >
        <Icon name="help" size={15} />
        {L.tipsForTheTable}
      </button>
      <button class="mi" role="menuitem" onclick={via(() => props.onShortcuts?.())}>
        <Icon name="keys" size={15} />
        {L.keyboardShortcutsItem}
      </button>
      {#if props.practice}
        <button class="mi" role="menuitem" onclick={via(() => props.onReplayTutorial?.())}>
          <Icon name="robot" size={15} />
          {L.replayTutorial}
        </button>
      {/if}
      <!-- #2982: the site footer is not on the table, so the fan-content
           notice is reachable from here. -->
      <button class="mi" role="menuitem" onclick={() => void show("legal")}>
        <Icon name="help" size={15} />
        {L.legalNotice}
      </button>
      <!-- #2981: the About page, in a new tab so a game is never left. -->
      <a
        class="mi"
        role="menuitem"
        href="#/about"
        target="_blank"
        rel="noopener noreferrer"
        onclick={() => close(false)}
      >
        <Icon name="help" size={15} />
        About this project
      </a>
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
  {:else if mode === "report"}
    <form class="pop vote-form report-form" aria-label="report a player" onsubmit={sendReport}>
      <div class="menu-h">{REPORT_TITLE}</div>
      {#if reportSent}
        <p class="report-note" role="status">{REPORT_SENT}</p>
        <div class="pop-actions">
          <button type="button" class="keep" onclick={() => close(true)}>Close</button>
        </div>
      {:else}
        <p class="report-note">{REPORT_INTRO}</p>
        <select bind:value={reportTarget} aria-label="player to report" disabled={reportBusy}>
          {#if spectatorReportable}
            <optgroup label={REPORT_SEATS_GROUP}>
              {#each props.reportTargets ?? [] as t (t.id)}
                <option value={t.id}>{t.name}</option>
              {/each}
            </optgroup>
            <optgroup label={REPORT_SPECTATORS_GROUP}>
              {#each spectatorTargets(watchers) as t (t.id)}
                <option value={t.id}>{t.name}</option>
              {/each}
            </optgroup>
          {:else}
            {#each props.reportTargets ?? [] as t (t.id)}
              <option value={t.id}>{t.name}</option>
            {/each}
          {/if}
        </select>
        <textarea
          bind:value={reportReason}
          maxlength={REPORT_REASON_MAX}
          rows="3"
          placeholder="what's the problem?"
          aria-label="reason for the report"
          disabled={reportBusy}
        ></textarea>
        {#if reportError}
          <p class="report-note report-err" role="alert">{reportError}</p>
        {/if}
        <div class="pop-actions">
          <button type="button" onclick={() => close(true)}>cancel</button>
          <button
            type="submit"
            class="go"
            disabled={reportBusy || !canSendReport(reportTarget, reportReason)}
            >{reportBusy ? "sending…" : "send"}</button
          >
        </div>
      {/if}
    </form>
  {:else if mode === "spectators"}
    <div class="pop spectators" role="dialog" aria-label="spectators">
      <div class="menu-h">{SPECTATORS_TITLE}</div>
      {#if watchers.length === 0}
        <p class="report-note">{SPECTATORS_EMPTY}</p>
      {:else}
        <ul class="watchers" aria-label="spectators watching">
          {#each watchers as w (w.id)}
            <li>
              <span class="wname">{w.name}{w.signed_in ? " · signed in" : ""}</span>
              <span class="wbtns">
                <button
                  type="button"
                  disabled={watchBusy !== null}
                  aria-label={`${REMOVE_SPECTATOR_LABEL} ${w.name}`}
                  onclick={() => removeWatcher(w, false)}>{REMOVE_SPECTATOR_LABEL}</button
                >
                <button
                  type="button"
                  disabled={watchBusy !== null}
                  aria-label={`${BLOCK_SPECTATOR_LABEL} ${w.name}`}
                  onclick={() => removeWatcher(w, true)}>{BLOCK_SPECTATOR_LABEL}</button
                >
              </span>
            </li>
          {/each}
        </ul>
      {/if}
      {#if watchNote}
        <p class="report-note" role="status">{watchNote}</p>
      {/if}
      {#if watchError}
        <p class="report-note report-err" role="alert">{watchError}</p>
      {/if}
      <p class="report-note">{SPECTATORS_HINT}</p>
      <div class="pop-actions">
        <button type="button" class="keep" onclick={() => close(true)}>Close</button>
      </div>
    </div>
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
  {:else if mode === "legal"}
    <div class="pop confirm legal" role="dialog" aria-modal="true" aria-label="legal notice">
      <div class="confirm-title">Legal notice</div>
      <div class="confirm-body"><LegalNotice newTab /></div>
      <div class="pop-actions">
        <button type="button" class="keep" onclick={() => close(true)}>Close</button>
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
  .report-form select,
  .report-form textarea {
    background: var(--surface-sunken);
    border: 1px solid var(--border);
    color: var(--fg);
    border-radius: var(--radius-sm);
    padding: 6px 8px;
    font: inherit;
    font-size: 12.5px;
    resize: vertical;
  }
  .report-note {
    margin: 0;
    font-size: 12px;
    line-height: 1.45;
    color: var(--fg-muted);
  }
  .report-err {
    color: var(--danger);
  }
  .spectators {
    gap: 8px;
    padding: 12px 14px;
    width: 320px;
  }
  .watchers {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .watchers li {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 6px;
    font-size: 12.5px;
  }
  .wname {
    color: var(--fg);
    overflow-wrap: anywhere;
  }
  .wbtns {
    display: inline-flex;
    gap: 6px;
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
  .confirm.legal {
    width: 340px;
    max-height: min(70vh, 420px);
    overflow-y: auto;
  }
  .confirm-body :global(p) {
    margin: 0;
  }
  .confirm-body :global(a) {
    color: var(--fg);
    text-decoration: underline;
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
