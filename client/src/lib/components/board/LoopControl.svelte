<script lang="ts">
  // LoopControl — ADR 0193 §9 (CR 732.2): the action dock's loop control.
  // Two parts, so the dock keeps its height:
  //   toggle  "record loop" on the toggles row, while the viewer holds
  //           priority with no recording standing;
  //   status  the status row's line once a recording stands: the
  //           recorder's step count with Stop and Cancel, "propose…"
  //           once the loop closed, and for every other seat what is
  //           being built (CR 732.1a).
  // The state is lib/loopShortcut.ts's loopControl; this only draws it.
  import type { LoopControlState } from "../../loopShortcut";
  import { L } from "../../labels";

  interface Props {
    part: "toggle" | "status";
    state: LoopControlState;
    onRecord: () => void;
    onStop: () => void;
    onCancel: () => void;
    onPropose: () => void;
  }

  const { part, state, onRecord, onStop, onCancel, onPropose }: Props = $props();
</script>

{#if part === "toggle"}
  {#if state.mode === "record"}
    <button
      type="button"
      class="loop-btn"
      aria-label={L.recordLoop}
      title="Record a loop (CR 732.2a): play one iteration by hand, then propose repeating it"
      onclick={onRecord}>loop</button
    >
  {/if}
{:else if state.mode === "recording"}
  <span class="loop-line" role="status">
    <span class="loop-text"
      >Recording a loop · {state.steps} {state.steps === 1 ? "step" : "steps"}</span
    >
    <button
      type="button"
      class="loop-btn"
      aria-label={L.stopLoop}
      title="Stop: the server checks the board came back to where it started"
      disabled={state.steps === 0}
      onclick={onStop}>stop</button
    >
    <button
      type="button"
      class="loop-btn"
      aria-label={L.cancelLoop}
      title="Discard this recording"
      onclick={onCancel}>✕</button
    >
  </span>
{:else if state.mode === "ready"}
  <span class="loop-line" role="status">
    <span class="loop-text">Loop ready · {state.steps} {state.steps === 1 ? "step" : "steps"}</span>
    <button
      type="button"
      class="loop-btn on"
      title="Choose how many more times to repeat it"
      onclick={onPropose}>propose…</button
    >
    <button
      type="button"
      class="loop-btn"
      aria-label={L.cancelLoop}
      title="Discard this recording"
      onclick={onCancel}>✕</button
    >
  </span>
{:else if state.mode === "proposed"}
  <span class="loop-line" role="status">
    <span class="loop-text">Loop proposed · waiting for answers</span>
    <button
      type="button"
      class="loop-btn"
      aria-label={L.cancelLoop}
      title="Withdraw the proposal"
      onclick={onCancel}>✕</button
    >
  </span>
{:else if state.mode === "watching"}
  <span class="loop-line" role="status"><span class="loop-text">{state.text}</span></span>
{/if}

<style>
  .loop-line {
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
    width: 100%;
  }

  .loop-text {
    flex: 1 1 auto;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--accent, #d9a441);
    font-size: 0.85em;
  }

  .loop-btn {
    flex: 0 0 auto;
    padding: 2px 8px;
    border-radius: 6px;
    border: 1px solid var(--border);
    background: var(--overlay-faint);
    color: var(--fg);
    font-size: 0.85em;
    font-weight: 600;
    cursor: pointer;
    white-space: nowrap;
  }

  .loop-btn.on {
    border-color: var(--accent, #d9a441);
    color: var(--accent, #d9a441);
  }

  .loop-btn:hover:not(:disabled) {
    background: color-mix(in srgb, var(--overlay-ink) 8%, transparent);
  }

  .loop-btn:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
</style>
