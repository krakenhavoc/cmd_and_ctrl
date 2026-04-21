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

  import type { GameView } from "../../protocol";
  import Card from "./Card.svelte";

  interface Props {
    snap: GameView;
    viewerID: string | null;
    sendAction: (type: string, params?: unknown, player?: string) => void;
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
        Your hand size exceeds your maximum ({viewerSeat?.max_hand_size ?? 7}). Pick exactly
        {owedCount} card{owedCount === 1 ? "" : "s"} to send to the graveyard. Cleanup resumes once you
        submit.
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
    background: rgba(0, 0, 0, 0.65);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 200;
  }
  .modal {
    background: #131a2c;
    border: 1px solid #4a5270;
    border-radius: 12px;
    padding: 16px 20px;
    max-width: 720px;
    max-height: 86vh;
    overflow: auto;
    box-shadow: 0 12px 32px rgba(0, 0, 0, 0.7);
  }
  h2 {
    margin: 0 0 4px;
    font-size: 16px;
    color: #ffd07a;
  }
  .hint {
    color: #9aa5cd;
    font-size: 12px;
    margin: 0 0 12px;
  }
  .card-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(96px, 1fr));
    gap: 8px;
  }
  .card-pick {
    background: transparent;
    border: 2px solid transparent;
    border-radius: 6px;
    padding: 2px;
    cursor: pointer;
    transition: border-color 80ms ease;
  }
  .card-pick:hover:not(:disabled) {
    border-color: #6c7794;
  }
  .card-pick.selected {
    border-color: #ffd07a;
    box-shadow: 0 0 8px rgba(255, 208, 122, 0.4);
  }
  .card-pick:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
  .footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 12px;
    padding-top: 10px;
    border-top: 1px solid #2a3148;
  }
  .counter {
    font-size: 12px;
    color: #cfd6ee;
  }
  .submit {
    padding: 6px 18px;
    border-radius: 4px;
    background: #ffd07a;
    color: #0c1426;
    border: none;
    font-weight: 700;
    cursor: pointer;
  }
  .submit:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
</style>
