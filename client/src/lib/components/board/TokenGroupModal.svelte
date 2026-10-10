<script lang="ts">
  // TokenGroupModal — the list behind a token group (#1724).
  //
  // A group of identical tokens is drawn as at most two cards on the
  // board (lib/tokenGroups.ts). Clicking one opens this list of every
  // member, which is where the player picks one, N or all of them for
  // whatever the table is asking for right now:
  //
  //   - a targeting prompt: "Target" sends each selected member through
  //     the same handler a board click uses, so a multi-target prompt
  //     collects them and a single-target one takes the first;
  //   - declaring attackers: "Attack <seat> with N" is one
  //     declare_attackers action, the same one "attack with all" sends,
  //     restricted to the selection;
  //   - declaring blockers: "Block" declares each selected member as a
  //     blocker of the attacker chosen beside it;
  //   - otherwise, for the viewer's own group: "Tap N" / "Untap N".
  //
  // "Use this one" (with exactly one selected) is the board click on
  // that member — pick it as an attacker to aim at a planeswalker, tap
  // a Treasure for mana, block with the creature already selected. It
  // is what makes the list a complete stand-in for the cards it hides:
  // anything a click on one token did before, a click here still does.
  //
  // Each row shows the member's differences as badges — counters,
  // attachments, damage, summoning sickness — so the board can stay
  // compact while the detail is one click away. Rows the current action
  // cannot use stay visible and say why.

  import { onDestroy } from "svelte";
  import type { CardView, GameView, PlayerView } from "../../protocol";
  import { targeting, isLegalCardTarget } from "../../targeting";
  import { boardChoicePick, isBoardPickable, isBoardPicked } from "../../boardChoicePick";
  import { attackRefusal, BLOCKER_LABELS, planAttackAll, seatLabel } from "../../attackAll";
  import { NO_LEGAL_ACTIONS, type LegalActions } from "../../legalActions";
  import { attackersDefendedBy } from "../../attackTargets";
  import {
    bulkCandidates,
    liveSelection,
    rowBadges,
    selectFirstN,
    toggleSelection,
    type GroupListMode,
  } from "../../tokenGroups";
  import ModalLayer from "../ModalLayer.svelte";

  interface Props {
    // Which group is open; a change resets the selection.
    groupKey: string;
    // The group's live members; empty means the modal is closed.
    members: CardView[];
    attachmentsByHost?: Record<string, CardView[]>;
    view: GameView;
    viewerID: string | null;
    combatMode: "idle" | "attack" | "block";
    // The viewer may tap and untap these (their own panel, or an admin).
    canTap?: boolean;
    // The board click on one member (PlayerPanel.handleCardClick).
    onUse: (card: CardView, ev?: MouseEvent) => void;
    // The targeting handler; true when the prompt took the card.
    onTarget?: (card: CardView) => boolean;
    onTapToggle?: (card: CardView) => void;
    onAttack?: (attackerIDs: string[], defenderSeatID: string) => void;
    onBlock?: (blockerIDs: string[], attackerID: string) => void;
    onClose: () => void;
    // ADR 0105 sub-PR 5: the frame's FULL legal-action lookup, so the
    // members that can attack are the ones the server would declare.
    // Passed for the viewer's own group only; absent means no
    // information, and the row fields decide as they did before.
    legalGate?: LegalActions;
  }

  const {
    groupKey,
    members,
    attachmentsByHost = {},
    view,
    viewerID,
    combatMode,
    canTap = false,
    onUse,
    onTarget,
    onTapToggle,
    onAttack,
    onBlock,
    onClose,
    legalGate = NO_LEGAL_ACTIONS,
  }: Props = $props();

  const open = $derived(members.length > 0);
  const own = $derived(!!viewerID && members[0]?.controller === viewerID);

  // #2880: a pending choice's permanents are picked here too, through
  // the same handler a board click uses.
  const legal = (c: CardView) => {
    if (isBoardPickable($boardChoicePick, c.instance_id)) return true;
    const t = $targeting;
    return t !== null && isLegalCardTarget(t, c.instance_id);
  };
  const cantAttack = (c: CardView) => {
    const r = attackRefusal(c, legalGate);
    return r ? BLOCKER_LABELS[r] : null;
  };

  const incoming = $derived(viewerID ? attackersDefendedBy(view, viewerID) : []);
  const defenders = $derived<PlayerView[]>(planAttackAll(view, viewerID, legalGate).defenders);

  const mode = $derived.by((): GroupListMode => {
    if (($targeting !== null || $boardChoicePick !== null) && !!onTarget && members.some(legal)) {
      return "target";
    }
    if (own && combatMode === "attack" && !!onAttack) return "attack";
    if (own && combatMode === "block" && !!onBlock && incoming.length > 0) return "block";
    if (canTap && !!onTapToggle) return "tap";
    return "look";
  });

  const candidates = $derived(
    bulkCandidates(members, mode, { isLegalTarget: legal, attackBlocker: cantAttack }),
  );

  // Why a row can't take part in the current action, or null.
  function reasonFor(c: CardView): string | null {
    if (mode === "target") return legal(c) ? null : "not a legal target";
    if (mode === "attack") return c.attacking_target ? null : cantAttack(c);
    if (mode === "block") return c.tapped ? "tapped" : null;
    return null;
  }

  let selected = $state<string[]>([]);
  let count = $state(1);
  let blockAttackerID = $state("");

  // A different group resets the selection; a member leaving the
  // group while the list is open drops out of it.
  let lastGroupKey: string | null = null;
  $effect(() => {
    if (groupKey !== lastGroupKey) {
      lastGroupKey = groupKey;
      selected = [];
      count = 1;
    }
  });
  $effect(() => {
    const kept = liveSelection(selected, members);
    if (kept !== selected) selected = kept;
  });
  $effect(() => {
    if (!incoming.some((a) => a.instance_id === blockAttackerID)) {
      blockAttackerID = incoming[0]?.instance_id ?? "";
    }
  });

  const picked = $derived(members.filter((c) => selected.includes(c.instance_id)));
  const pickedForAction = $derived.by(() => {
    const ok = new Set(candidates.map((c) => c.instance_id));
    return picked.filter((c) => ok.has(c.instance_id));
  });
  const pickedUntapped = $derived(picked.filter((c) => !c.tapped));
  const pickedTapped = $derived(picked.filter((c) => !!c.tapped));

  function toggle(id: string): void {
    selected = toggleSelection(selected, id);
  }
  function selectN(): void {
    selected = selectFirstN(candidates, count);
  }
  function selectAll(): void {
    selected = candidates.map((c) => c.instance_id);
  }
  function selectNone(): void {
    selected = [];
  }

  function target(): void {
    if (!onTarget) return;
    for (const c of pickedForAction) {
      if ($boardChoicePick !== null) {
        // A choice's pick toggles, so one already picked stays picked.
        if (!isBoardPicked($boardChoicePick, c.instance_id)) onTarget(c);
        continue;
      }
      if ($targeting === null) break;
      onTarget(c);
    }
    onClose();
  }
  function attack(seatID: string): void {
    const ids = pickedForAction.map((c) => c.instance_id);
    if (!onAttack || ids.length === 0) return;
    onAttack(ids, seatID);
    onClose();
  }
  function block(): void {
    const ids = pickedForAction.map((c) => c.instance_id);
    if (!onBlock || ids.length === 0 || !blockAttackerID) return;
    onBlock(ids, blockAttackerID);
    onClose();
  }
  function tapAll(which: CardView[]): void {
    if (!onTapToggle) return;
    for (const c of which) onTapToggle(c);
    onClose();
  }
  function useOne(ev: MouseEvent): void {
    const c = picked[0];
    if (!c) return;
    onClose();
    onUse(c, ev);
  }

  function attackerLabel(a: CardView): string {
    const who = view.seats.find((s) => s.id === a.controller);
    const pt = a.power !== undefined ? ` ${a.power}/${a.toughness ?? 0}` : "";
    return `${a.name}${pt}${who ? ` (${seatLabel(who)})` : ""}`;
  }

  // The list is mounted by the panel that owns the group, and a panel
  // can sit under a transformed or filtered ancestor (the disabled
  // board's greyscale), which would make the backdrop's position:
  // fixed mean "fixed to that ancestor". So it moves to <body>, as the
  // hand's drag ghost does.
  function portal(node: HTMLElement): { destroy(): void } {
    document.body.appendChild(node);
    return {
      destroy() {
        node.remove();
      },
    };
  }

  function handleKey(e: KeyboardEvent): void {
    if (!open) return;
    if (e.key === "Escape") {
      e.preventDefault();
      onClose();
    }
  }
  $effect(() => {
    if (!open) return;
    document.addEventListener("keydown", handleKey);
    return () => document.removeEventListener("keydown", handleKey);
  });
  onDestroy(() => document.removeEventListener("keydown", handleKey));

  const title = $derived(members[0]?.name ?? "");
  const hint = $derived(
    mode === "target"
      ? $boardChoicePick !== null
        ? "Pick which of these to choose."
        : "Pick which of these to target."
      : mode === "attack"
        ? "Pick which of these attack. Select N takes untapped, non-summoning-sick tokens first."
        : mode === "block"
          ? "Pick which of these block."
          : "Pick one, a number, or all of them.",
  );
</script>

{#if open}
  <ModalLayer />
  <div
    class="prompt-backdrop"
    role="dialog"
    aria-modal="true"
    aria-labelledby="tg-title"
    use:portal
  >
    <div class="prompt-modal tg-modal">
      <h2 id="tg-title">
        {title}
        <span class="prompt-src">{members.length} tokens</span>
      </h2>
      <p class="prompt-hint">{hint}</p>
      <div class="tg-bulk" role="group" aria-label="select several">
        <label class="tg-n">
          Select
          <input
            type="number"
            min="1"
            max={Math.max(1, candidates.length)}
            bind:value={count}
            aria-label="how many to select"
          />
        </label>
        <button
          type="button"
          class="ghost"
          disabled={candidates.length === 0}
          onclick={selectN}
          data-testid="tg-select-n">Select {count}</button
        >
        <button
          type="button"
          class="ghost"
          disabled={candidates.length === 0}
          onclick={selectAll}
          data-testid="tg-select-all">All {candidates.length}</button
        >
        <button type="button" class="ghost" onclick={selectNone}>None</button>
      </div>
      <ul class="prompt-options tg-list">
        {#each members as c (c.instance_id)}
          {@const on = selected.includes(c.instance_id)}
          {@const reason = reasonFor(c)}
          <li>
            <button
              type="button"
              class="prompt-opt"
              class:on
              class:unusable={!!reason}
              role="checkbox"
              aria-checked={on}
              data-instance={c.instance_id}
              onclick={() => toggle(c.instance_id)}
            >
              <span class="prompt-radio" aria-hidden="true"></span>
              <span class="tg-name">{c.name}</span>
              <span class="tg-badges">
                {#each rowBadges(c, attachmentsByHost[c.instance_id] ?? [], reason) as b, i (i)}
                  <span class="tg-badge {b.kind}">{b.text}</span>
                {/each}
              </span>
              {#if c.power !== undefined || c.toughness !== undefined}
                <span class="note pt">{c.power ?? 0}/{c.toughness ?? 0}</span>
              {/if}
            </button>
          </li>
        {/each}
      </ul>
      <div class="prompt-foot tg-actions">
        <span class="prompt-count" aria-live="polite" data-testid="tg-count"
          >{picked.length} selected</span
        >
        {#if mode === "target"}
          <button
            type="button"
            class="primary"
            disabled={pickedForAction.length === 0}
            onclick={target}
            >{$boardChoicePick !== null ? "Choose" : "Target"} {pickedForAction.length}</button
          >
        {:else if mode === "attack"}
          {#each defenders as seat (seat.id)}
            <button
              type="button"
              class="primary"
              disabled={pickedForAction.length === 0}
              onclick={() => attack(seat.id)}
              >Attack {seatLabel(seat)} with {pickedForAction.length}</button
            >
          {/each}
        {:else if mode === "block"}
          <select bind:value={blockAttackerID} aria-label="attacker to block">
            {#each incoming as a (a.instance_id)}
              <option value={a.instance_id}>{attackerLabel(a)}</option>
            {/each}
          </select>
          <button
            type="button"
            class="primary"
            disabled={pickedForAction.length === 0 || !blockAttackerID}
            onclick={block}>Block with {pickedForAction.length}</button
          >
        {:else if mode === "tap"}
          <button
            type="button"
            class="ghost"
            disabled={pickedUntapped.length === 0}
            onclick={() => tapAll(pickedUntapped)}>Tap {pickedUntapped.length}</button
          >
          <button
            type="button"
            class="ghost"
            disabled={pickedTapped.length === 0}
            onclick={() => tapAll(pickedTapped)}>Untap {pickedTapped.length}</button
          >
        {/if}
      </div>
      <div class="prompt-foot">
        {#if mode !== "target"}
          <button
            type="button"
            class="ghost"
            disabled={picked.length !== 1}
            title="Does what clicking this token on the board does"
            onclick={useOne}>Use this one</button
          >
        {/if}
        <button type="button" class="ghost" onclick={onClose}
          >Close <span class="kbd">Esc</span></button
        >
      </div>
    </div>
  </div>
{/if}

<style>
  .tg-modal {
    width: min(520px, calc(100vw - 32px));
  }
  .tg-bulk {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
  }
  .tg-n {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 12.5px;
    color: var(--fg-muted);
  }
  .tg-n input {
    width: 4.5em;
  }
  .tg-list {
    max-height: 48vh;
    overflow: auto;
  }
  .tg-list .prompt-opt {
    flex-wrap: wrap;
  }
  .tg-list .prompt-opt.unusable:not(.on) {
    opacity: 0.6;
  }
  .tg-name {
    font-weight: 600;
  }
  .tg-badges {
    display: inline-flex;
    flex-wrap: wrap;
    gap: 4px;
  }
  .tg-badge {
    font-family: var(--font-mono);
    font-size: 10px;
    padding: 1px 6px;
    border-radius: 8px;
    border: 1px solid var(--border);
    background: var(--surface-raised);
    color: var(--fg-muted);
    white-space: nowrap;
  }
  .tg-badge.counter,
  .tg-badge.attachment {
    border-color: color-mix(in srgb, var(--accent) 60%, transparent);
    color: var(--accent);
  }
  .tg-badge.damage,
  .tg-badge.reason {
    color: var(--danger);
  }
  .pt {
    font-family: var(--font-mono);
  }
  .tg-actions {
    flex-wrap: wrap;
  }
</style>
