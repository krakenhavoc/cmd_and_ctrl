<script lang="ts">
  // DiscardPromptModal opens when the wire's discard_pending map
  // names the viewer (S13.4 — interactive cleanup discard). The
  // modal is non-dismissible until the right number of hand cards
  // is selected; submitting dispatches discard_selection and the
  // server resumes the cleanup auto-advance.
  //
  // Per CR 402.2 the discard happens at cleanup; the engine pauses
  // the cursor at Cleanup with PriorityHolder=NoPriority and waits.
  // Multi-player simultaneous discard (Mindslicer) drains in seat
  // order — only one viewer's modal is open at a time on each
  // client, since viewer != active-seat is read-only.

  import type { ActionType, GameView } from "../../protocol";
  import Card from "./Card.svelte";

  interface Props {
    snap: GameView;
    viewerID: string | null;
    sendAction: (type: ActionType, params?: unknown, player?: string) => void;
  }

  const { snap, viewerID, sendAction }: Props = $props();

  // owedCount is the number of cards the viewer must discard. Zero
  // / undefined → modal hidden. Reactive so the modal auto-closes
  // when the server drains the entry.
  const owedCount = $derived.by(() => {
    if (!viewerID || !snap.discard_pending) return 0;
    return snap.discard_pending[viewerID] ?? 0;
  });
  const open = $derived(owedCount > 0);

  // S14: the same DiscardPending map carries cleanup-step max-
  // hand overflow AND mid-game effect-driven discards (Mind Rot,
  // future effect cards). Tell them apart by looking at the turn
  // step. Cleanup → cleanup copy; any other step → "effect
  // resolving" copy. The behaviour is identical (pick N, submit)
  // either way.
  const isCleanupContext = $derived(snap.turn.step === "cleanup");

  // viewerSeat is the seated PlayerView for the viewer. Used to
  // read the hand contents — opponents' DiscardPromptModal
  // wouldn't have visibility anyway since the wire filters
  // opponent hands out of the snapshot.
  const viewerSeat = $derived(snap.seats.find((s) => s.id === viewerID) ?? null);
  const handCards = $derived(viewerSeat?.hand?.cards ?? []);

  let selected = $state<Set<string>>(new Set());

  // Reset selection whenever the modal opens fresh (owedCount goes
  // from 0 → positive). Without the reset, a stale selection from
  // a previous turn's discard would survive.
  let lastOpen = false;
  $effect(() => {
    const wasOpen = lastOpen;
    lastOpen = open;
    if (open && !wasOpen) {
      selected = new Set();
    }
  });

  function toggle(id: string): void {
    const next = new Set(selected);
    if (next.has(id)) next.delete(id);
    else if (next.size < owedCount) next.add(id);
    selected = next;
  }

  function submit(): void {
    if (selected.size !== owedCount) return;
    sendAction("discard_selection", { card_ids: Array.from(selected) }, viewerID ?? undefined);
  }
</script>

{#if open}
  <div class="backdrop" role="dialog" aria-modal="true" aria-labelledby="discard-title">
    <div class="modal">
      <h2 id="discard-title">
        Discard {owedCount} card{owedCount === 1 ? "" : "s"}
      </h2>
      <p class="hint">
        {#if isCleanupContext}
          Your hand size exceeds your maximum ({viewerSeat?.max_hand_size ?? 7}). Pick exactly
          {owedCount} card{owedCount === 1 ? "" : "s"} to send to the graveyard. Cleanup resumes once
          you submit.
        {:else}
          An effect is asking you to discard. Pick {owedCount} card{owedCount === 1 ? "" : "s"} to send
          to the graveyard.
        {/if}
      </p>
      <div class="card-grid">
        {#each handCards as c (c.instance_id)}
          <button
            type="button"
            class="card-pick"
            class:selected={selected.has(c.instance_id)}
            disabled={!selected.has(c.instance_id) && selected.size >= owedCount}
            onclick={() => toggle(c.instance_id)}
            aria-pressed={selected.has(c.instance_id)}
            aria-label={`select ${c.name}`}
          >
            <Card card={c} />
          </button>
        {/each}
      </div>
      <div class="footer">
        <span class="counter">{selected.size} / {owedCount} selected</span>
        <button
          type="button"
          class="submit"
          disabled={selected.size !== owedCount}
          onclick={submit}
        >
          Discard
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .backdrop {
    position: fixed;
    inset: 0;
    background: rgba(4, 8, 16, 0.7);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 200;
    animation: fade-in 160ms var(--ease);
  }
  @keyframes fade-in {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }
  .modal {
    background: linear-gradient(180deg, var(--surface) 0%, var(--bg-2) 100%);
    border: 1px solid rgba(122, 167, 255, 0.22);
    border-radius: var(--radius-xl);
    padding: 22px 26px;
    max-width: 760px;
    max-height: 86vh;
    overflow: auto;
    box-shadow:
      0 30px 80px rgba(0, 0, 0, 0.7),
      0 0 0 1px rgba(0, 0, 0, 0.4),
      inset 0 1px 0 rgba(255, 255, 255, 0.05);
    animation: modal-in 220ms var(--ease);
  }
  @keyframes modal-in {
    from {
      opacity: 0;
      transform: translateY(12px) scale(0.98);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }
  h2 {
    margin: 0 0 6px;
    font-size: 18px;
    letter-spacing: -0.01em;
    color: var(--gold);
    text-transform: none;
    font-weight: 700;
  }
  .hint {
    color: var(--fg-muted);
    font-size: 13px;
    line-height: 1.4;
    margin: 0 0 14px;
  }
  .card-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(96px, 1fr));
    gap: 10px;
  }
  .card-pick {
    background: transparent;
    border: 2px solid transparent;
    border-radius: var(--radius);
    padding: 3px;
    cursor: pointer;
    box-shadow: none;
    transition:
      border-color 120ms var(--ease),
      transform 120ms var(--ease),
      box-shadow 120ms var(--ease);
  }
  .card-pick:hover:not(:disabled) {
    border-color: var(--accent);
    box-shadow: 0 6px 18px rgba(0, 0, 0, 0.4);
    transform: translateY(-2px);
  }
  .card-pick.selected {
    border-color: var(--gold);
    box-shadow:
      0 0 18px rgba(255, 208, 122, 0.55),
      0 6px 18px rgba(0, 0, 0, 0.4);
  }
  .card-pick:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
  .footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 16px;
    padding-top: 14px;
    border-top: 1px solid rgba(255, 255, 255, 0.06);
    gap: 12px;
  }
  .counter {
    font-size: 12px;
    color: var(--fg-muted);
    font-variant-numeric: tabular-nums;
    letter-spacing: 0.02em;
  }
  .submit {
    padding: 8px 22px;
    border-radius: 999px;
    background: linear-gradient(180deg, #ffe59a 0%, #e6b85f 100%);
    color: #231806;
    border: 1px solid rgba(255, 230, 160, 0.6);
    font-weight: 800;
    letter-spacing: 0.02em;
    cursor: pointer;
    box-shadow:
      0 6px 18px rgba(255, 208, 122, 0.25),
      inset 0 1px 0 rgba(255, 255, 255, 0.4);
  }
  .submit:hover:not(:disabled) {
    filter: brightness(1.04);
  }
  .submit:disabled {
    opacity: 0.4;
    cursor: not-allowed;
    box-shadow: none;
  }
</style>
