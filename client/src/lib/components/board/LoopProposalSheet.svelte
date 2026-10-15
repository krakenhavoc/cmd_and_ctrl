<script lang="ts">
  // LoopProposalSheet — ADR 0193 §4 and §9 (CR 732.2b): another player
  // proposes repeating a loop they recorded, and this seat answers:
  // Accept, or Shorten to one of its own steps in some iteration — the
  // place where it will choose differently. The default is its first
  // step in iteration 1. The server asks each seat in turn order and
  // keeps the earliest ending point. A pending choice, so it has no
  // cancel: the table waits on the answer.
  import type { LoopPointView, PendingChoiceView, LoopProposalView } from "../../protocol";
  import { confirmAction, type DockAction } from "../../dock";
  import { defaultShorten, shortenAllowed } from "../../loopShortcut";
  import { L } from "../../labels";
  import DockSheet from "./DockSheet.svelte";

  interface Props {
    // The viewer's loop_proposal prompt; null closes the sheet.
    choice: (PendingChoiceView & { loop_proposal: LoopProposalView }) | null;
    proposerName: string;
    onAccept: (choiceID: string) => void;
    onShorten: (choiceID: string, at: LoopPointView) => void;
  }

  const { choice, proposerName, onAccept, onShorten }: Props = $props();

  const p = $derived(choice?.loop_proposal ?? null);
  let iteration = $state(1);
  let step = $state(-1);
  let seenID: string | null = null;
  $effect(() => {
    const id = choice?.id ?? null;
    if (id !== seenID) {
      seenID = id;
      const d = p ? defaultShorten(p) : null;
      iteration = d?.iteration ?? 1;
      step = d?.step ?? -1;
    }
  });
  const mine = $derived(new Set(p?.choice_points ?? []));
  const at = $derived({ iteration: Math.trunc(Number(iteration)), step });
  const canShorten = $derived(!!p && shortenAllowed(p, at));

  const shorten: DockAction = $derived({
    id: "shorten",
    label: "Shorten here",
    disabled: !canShorten,
    title: "End the loop before your marked step in that iteration; you get priority there",
    onPress: () => {
      if (choice && canShorten) onShorten(choice.id, at);
    },
  });
</script>

{#if choice && p}
  <DockSheet
    label={L.loopProposal}
    rank="choice"
    title={`${proposerName} proposes a loop`}
    src="CR 732.2b"
    width={480}
    sheetKey={`loop-proposal:${choice.id}`}
    primary={confirmAction("Accept", () => onAccept(choice.id))}
    secondary={[shorten]}
  >
    <p class="loop-net">
      Repeat {p.name ?? "the loop"}
      {p.count} more times. Each time: {p.net ?? "no net change"}.
    </p>
    {#if p.end.iteration < p.count || p.end.step < p.steps.length}
      <p class="loop-hint">
        Already shortened to iteration {p.end.iteration}, before step {p.end.step + 1}.
      </p>
    {/if}
    <ol class="loop-steps">
      {#each p.steps as s, i (i)}
        <li class:mine={mine.has(i)}>
          {#if mine.has(i)}
            <label>
              <input type="radio" name="loop-shorten-step" value={i} bind:group={step} />
              {s.label}
            </label>
          {:else}
            {s.label}
          {/if}
        </li>
      {/each}
    </ol>
    {#if mine.size > 0}
      <label class="loop-iteration">
        Shorten in iteration
        <input
          type="number"
          min="1"
          max={p.count}
          step="1"
          bind:value={iteration}
          aria-label="iteration to shorten in"
        />
        before the marked step
      </label>
    {:else}
      <p class="loop-hint">None of the steps is yours, so you can only accept.</p>
    {/if}
  </DockSheet>
{/if}

<style>
  .loop-net {
    margin: 0 0 8px;
    font-weight: 600;
  }

  .loop-steps {
    margin: 0 0 10px;
    padding-left: 22px;
    max-height: 200px;
    overflow-y: auto;
    font-size: 13px;
  }

  .loop-steps li.mine {
    color: var(--accent, #d9a441);
  }

  .loop-iteration input {
    width: 70px;
    margin: 0 6px;
  }

  .loop-hint {
    margin: 0 0 8px;
    font-size: 12px;
    opacity: 0.8;
  }
</style>
