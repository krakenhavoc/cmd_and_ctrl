<script lang="ts">
  // LoopProposeSheet — ADR 0193 §9 (CR 732.2a): the recorder's sheet once
  // the loop closed. It lists the iteration and its net change and asks
  // how many more times to repeat it (1 … the cap, with Max). Propose
  // asks every other seat in turn order; "Not yet" folds the sheet away
  // with the recording kept (the dock's "propose…" brings it back), and
  // "Discard loop" cancels the recording.
  import type { LoopRecordingView } from "../../protocol";
  import { cancelAction, confirmAction, type DockAction } from "../../dock";
  import { clampLoopCount } from "../../loopShortcut";
  import { L } from "../../labels";
  import DockSheet from "./DockSheet.svelte";

  interface Props {
    // The viewer's ready recording; null closes the sheet.
    recording: LoopRecordingView | null;
    onPropose: (count: number) => void;
    onDiscard: () => void;
    onClose: () => void;
  }

  const { recording, onPropose, onDiscard, onClose }: Props = $props();

  const max = $derived(Math.max(1, recording?.max_iterations ?? 1000));
  let typed = $state(10);
  const count = $derived(clampLoopCount(typed, max));

  const discard: DockAction = $derived({
    id: "discard-loop",
    label: "Discard loop",
    ariaLabel: L.cancelLoop,
    onPress: onDiscard,
  });
</script>

{#if recording}
  <DockSheet
    label={L.proposeLoop}
    title={recording.name ?? "Your loop"}
    src="CR 732.2a"
    width={480}
    sheetKey="loop-propose"
    primary={confirmAction(`Propose ×${count}`, () => onPropose(count))}
    secondary={[cancelAction(onClose, "Not yet"), discard]}
  >
    <p class="loop-net">Each time: {recording.net ?? "no net change"}</p>
    <ol class="loop-steps">
      {#each recording.iteration ?? [] as step, i (i)}
        <li>{step.label}</li>
      {/each}
    </ol>
    <div class="loop-count">
      <label>
        Repeat
        <input
          type="number"
          min="1"
          {max}
          step="1"
          bind:value={typed}
          aria-label="iterations to propose"
        />
        more times
      </label>
      <button type="button" class="loop-max" onclick={() => (typed = max)}>Max ({max})</button>
    </div>
    <p class="loop-hint">
      Every other player is asked in turn order. Each may accept or shorten it; the replay then runs
      every iteration through the game and stops early if anything comes out differently.
    </p>
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
    max-height: 180px;
    overflow-y: auto;
    font-size: 13px;
  }

  .loop-count {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
  }

  .loop-count input {
    width: 80px;
    margin: 0 6px;
  }

  .loop-max {
    padding: 2px 8px;
    border-radius: 6px;
    border: 1px solid var(--border);
    background: var(--overlay-faint);
    color: var(--fg);
    cursor: pointer;
  }

  .loop-hint {
    margin: 10px 0 0;
    font-size: 12px;
    opacity: 0.8;
  }
</style>
