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
  import ModalLayer from "../ModalLayer.svelte";

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
  <ModalLayer />
  <div class="prompt-backdrop" role="dialog" aria-modal="true" aria-labelledby="discard-title">
    <div class="prompt-modal">
      <h2 id="discard-title">
        Discard {owedCount} card{owedCount === 1 ? "" : "s"}
        <span class="prompt-src" aria-hidden="true">{isCleanupContext ? "cleanup" : "effect"}</span>
      </h2>
      <p class="prompt-hint">
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
      <div class="prompt-foot">
        <span class="prompt-count">{selected.size} / {owedCount} selected</span>
        <button
          type="button"
          class="primary"
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
  .card-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(96px, 1fr));
    gap: 8px;
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
    border-color: var(--border-strong);
    background: transparent;
    transform: translateY(-2px);
  }
  .card-pick.selected {
    border-color: var(--gold);
    box-shadow: 0 0 16px rgba(217, 180, 92, 0.35);
  }
  .card-pick:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
</style>
