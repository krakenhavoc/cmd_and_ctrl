<script lang="ts">
  // DivideDamageModal — #1563, CR 601.2d: "N damage divided as you
  // choose among …". Opens when a divided target step completes with
  // two or more picks: one stepper per target, seeded with the even
  // split (the remainder to the earliest picks, which is also what the
  // bot announces), and Confirm enabled only when every target has at
  // least 1 and the shares add up to the amount — the same two rules
  // the engine's announce gate enforces, via divisionProblem.
  //
  // One pick needs no question (it takes the whole amount), so the
  // walk never opens this for one. The answer rides the SAME
  // cast_spell / activate_ability / resolve_choice as the targets, as
  // `distribution`. Cancel returns to the target picker with the picks
  // intact, so a player can change who is hit before dividing again.

  import { onDestroy } from "svelte";
  import { divisionProblem, evenSplit } from "../../targeting";
  import ModalLayer from "../ModalLayer.svelte";

  interface Props {
    // The spell or ability source, for the title. Null closes the modal.
    sourceName: string | null;
    // The picks, in the order they were picked, with a display name.
    targets: { id: string; name: string }[];
    // The amount being divided, already resolved against X.
    // #1657: "up to that many" — the shares may add up to less.
    upTo?: boolean;
    total: number;
    onConfirm: (dist: Record<string, number>) => void;
    onCancel: () => void;
  }

  const { sourceName, targets, total, upTo = false, onConfirm, onCancel }: Props = $props();

  let shares = $state<Record<string, number>>({});

  // Reseed with the even split whenever a different set of picks opens
  // the prompt.
  let lastKey: string | null = null;
  $effect(() => {
    const key = sourceName === null ? null : `${total}:${targets.map((t) => t.id).join(",")}`;
    if (key !== lastKey) {
      lastKey = key;
      const picked = targets.map((t) => t.id);
      shares = key === null ? {} : evenSplit(picked, total);
    }
  });

  const ids = $derived(targets.map((t) => t.id));
  const assigned = $derived(ids.reduce((n, id) => n + (shares[id] ?? 0), 0));
  const problem = $derived(divisionProblem(ids, total, shares, upTo));

  function step(id: string, delta: number): void {
    const next = (shares[id] ?? 0) + delta;
    if (next < 1) return;
    shares = { ...shares, [id]: next };
  }

  function confirm(): void {
    if (sourceName === null || problem !== null) return;
    const out: Record<string, number> = {};
    for (const id of ids) out[id] = shares[id];
    onConfirm(out);
  }

  // Enter and Escape stop here: Game.svelte's window handler would
  // otherwise also confirm the (still open) target walk on Enter, and
  // cancel the whole cast on Escape rather than going Back to the picks.
  function handleKey(e: KeyboardEvent): void {
    if (sourceName === null) return;
    if (e.key === "Enter") {
      e.preventDefault();
      e.stopPropagation();
      confirm();
    } else if (e.key === "Escape") {
      e.preventDefault();
      e.stopPropagation();
      onCancel();
    }
  }
  $effect(() => {
    if (sourceName === null) return;
    document.addEventListener("keydown", handleKey);
    return () => document.removeEventListener("keydown", handleKey);
  });
  onDestroy(() => document.removeEventListener("keydown", handleKey));
</script>

{#if sourceName !== null}
  <ModalLayer />
  <div class="prompt-backdrop" role="dialog" aria-modal="true" aria-labelledby="divide-title">
    <div class="prompt-modal divide-modal">
      <h2 id="divide-title">Divide {upTo ? "up to " : ""}{total} — {sourceName}</h2>
      <p class="prompt-hint">
        Assign each target at least 1. The shares must add up to {upTo
          ? `at most ${total}`
          : total}.
      </p>
      <ul class="rows">
        {#each targets as t (t.id)}
          <li class="row">
            <span class="name">{t.name}</span>
            <button
              type="button"
              class="ghost step"
              disabled={(shares[t.id] ?? 0) <= 1}
              onclick={() => step(t.id, -1)}
              aria-label={`one less to ${t.name}`}>−</button
            >
            <output class="count" aria-label={`share for ${t.name}`}>{shares[t.id] ?? 0}</output>
            <button
              type="button"
              class="ghost step"
              disabled={assigned >= total}
              onclick={() => step(t.id, 1)}
              aria-label={`one more to ${t.name}`}>+</button
            >
          </li>
        {/each}
      </ul>
      <p class="status" class:bad={problem !== null} aria-live="polite">
        {problem ?? `${assigned} of ${total} assigned`}
      </p>
      <div class="prompt-foot">
        <button type="button" class="ghost" onclick={onCancel}
          >Back <span class="kbd">Esc</span></button
        >
        <button type="button" class="primary" disabled={problem !== null} onclick={confirm}>
          Confirm <span class="kbd">↵</span>
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .divide-modal {
    width: min(440px, calc(100vw - 32px));
  }
  .rows {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .row {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 14px;
  }
  .name {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--fg);
  }
  .step {
    min-width: 2rem;
    font-family: var(--font-mono);
    font-size: 16px;
    line-height: 1;
    padding: 4px 8px;
  }
  .count {
    font-family: var(--font-mono);
    font-size: 15px;
    min-width: 2.5rem;
    text-align: center;
  }
  .status {
    font-size: 12px;
    color: var(--mint);
    margin: 8px 0 0;
  }
  .status.bad {
    color: var(--danger);
  }
  .primary .kbd {
    color: var(--accent-fg);
    border-color: rgba(28, 21, 3, 0.35);
    opacity: 0.8;
  }
</style>
