<script lang="ts">
  // BlockListSheet — the block declaration as a list (#3898).
  //
  // Blocks are made on the board: click your creature, then the attacker
  // (lib/blockPick.ts). This is the same declaration one row per
  // creature, one button per attacker, opened from the block request's
  // "Show as list". It is the keyboard path and the fallback for a board
  // too crowded to click: every pile is spelled out member by member, and
  // every button sends exactly what the board click would, through the
  // same state machine. An attacker a creature can't block is a disabled
  // button that says why.

  import { L } from "../../labels";
  import { cancelAction, confirmAction } from "../../dock";
  import { attackersDefendedBy, blockedAttackersOf } from "../../attackTargets";
  import {
    blockTargets,
    canSelectBlocker,
    type BlockContext,
    type BlockPickState,
    type BlockTarget,
  } from "../../blockPick";
  import type { CardView } from "../../protocol";
  import DockSheet from "./DockSheet.svelte";

  interface Props {
    ctx: BlockContext;
    pick: BlockPickState;
    // How many of the viewer's creatures block something now.
    staged: number;
    // The button for creature `blocker` on attacker `attacker`.
    onPair: (blocker: string, attacker: string) => void;
    // "No block" on a creature that blocks something.
    onClear: (blocker: string) => void;
    onFinish: () => void;
    onClose: () => void;
  }

  const { ctx, pick, staged, onPair, onClear, onFinish, onClose }: Props = $props();

  const attackers = $derived(attackersDefendedBy(ctx.view, ctx.viewerID));
  const blockers = $derived(
    (ctx.view.battlefield?.cards ?? []).filter((c) => canSelectBlocker(ctx, c.instance_id)),
  );

  // Copies of one attacker (a pile of tokens) are told apart by a
  // number, in board order.
  function labelOf(c: CardView, list: readonly CardView[]): string {
    const same = list.filter((x) => x.name === c.name);
    const pt = c.power !== undefined ? ` ${c.power}/${c.toughness ?? 0}` : "";
    const n = same.length > 1 ? ` (${same.indexOf(c) + 1})` : "";
    return `${c.name || "unknown creature"}${pt}${n}`;
  }

  function targetsOf(id: string): Map<string, BlockTarget> {
    return blockTargets(ctx, id);
  }

  const group = $derived(pick.group);
  const groupName = $derived(
    group ? (attackers.find((a) => a.instance_id === group.attacker)?.name ?? "it") : "",
  );
</script>

<DockSheet
  label={L.chooseBlockers}
  title="Choose blockers"
  width={620}
  sheetKey="blocks-list"
  count={`${staged} ${staged === 1 ? "blocker" : "blockers"} declared`}
  primary={confirmAction(staged > 0 ? "Done blocking" : "No blocks", onFinish, {
    id: "finish-blocks",
    enter: staged > 0,
    title:
      staged > 0
        ? "finish declaring blockers"
        : "declare no blockers — and pass priority, if you hold it",
  })}
  secondary={[cancelAction(onClose, "Back to board")]}
>
  <p class="prompt-hint">
    For each creature, pick the attacker it blocks. The same blocks show on the board.
  </p>
  {#if group}
    <p class="group-note" role="status">
      {groupName} needs {group.need} blockers: pick it for {group.need - group.blockers.length} more.
    </p>
  {/if}
  {#if blockers.length === 0}
    <p class="prompt-hint">None of your creatures can block right now.</p>
  {/if}
  <ul class="rows" role="list">
    {#each blockers as b (b.instance_id)}
      {@const t = targetsOf(b.instance_id)}
      {@const blocking = blockedAttackersOf(b).length > 0}
      {@const name = labelOf(b, blockers)}
      <li class="row">
        <span class="blocker">{name}</span>
        <span class="choices" role="group" aria-label={`${name} blocks`}>
          {#each attackers as a (a.instance_id)}
            {@const tt = t.get(a.instance_id)}
            {@const pending =
              !!group && group.attacker === a.instance_id && group.blockers.includes(b.instance_id)}
            {@const on = tt?.kind === "remove" || pending}
            {@const refused = tt?.kind === "refused" ? tt.reason : ""}
            <button
              type="button"
              class="prompt-opt choice"
              class:on
              class:pending
              aria-pressed={on}
              disabled={!!refused || !tt}
              title={refused
                ? `can't block: ${refused}`
                : tt?.kind === "group"
                  ? `takes ${tt.need} blockers`
                  : on
                    ? "blocking it — press to take the block back"
                    : "block it"}
              onclick={() => onPair(b.instance_id, a.instance_id)}
            >
              {labelOf(a, attackers)}{#if tt?.kind === "group"}<span class="note">
                  needs {tt.need}</span
                >{/if}
            </button>
          {/each}
          {#if blocking}
            <button
              type="button"
              class="ghost clear"
              onclick={() => onClear(b.instance_id)}
              title="take this creature's block back">No block</button
            >
          {/if}
        </span>
      </li>
    {/each}
  </ul>
</DockSheet>

<style>
  .rows {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .row {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 6px 10px;
    padding-bottom: 8px;
    border-bottom: 1px solid var(--border);
  }
  .blocker {
    font-size: 13px;
    font-weight: 600;
    min-width: 9em;
  }
  .choices {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    flex: 1;
    min-width: 0;
  }
  .choice {
    width: auto;
    padding: 4px 10px;
    font-size: 12px;
  }
  .choice.on {
    border-color: var(--block);
    background: color-mix(in srgb, var(--block) 18%, transparent);
  }
  .choice.pending {
    border-style: dashed;
  }
  .choice:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
  .note {
    margin-left: 6px;
    font-family: var(--font-mono);
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--fg-muted);
  }
  .clear {
    height: 26px;
    padding: 0 10px;
    font-size: 11px;
  }
  .group-note {
    margin: 0;
    font-size: 13px;
    color: var(--fg);
  }
</style>
