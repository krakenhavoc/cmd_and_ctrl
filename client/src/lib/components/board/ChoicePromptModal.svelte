<script lang="ts">
  // ChoicePromptModal opens when the wire's pending_choices queue
  // has an entry addressed to the viewer. Generic counterpart to
  // DiscardPromptModal: that one handles the S13.4 cleanup-specific
  // map where chooser == owner; this one handles the S14+ queue
  // where chooser can differ from the pool owner (Thoughtseize:
  // caster picks from target's revealed hand).
  //
  // The options[] slice on each entry is pre-filtered by the
  // server's per-viewer KnownBy projection — cards the viewer
  // is legally allowed to see arrive face-up; the rest arrive
  // redacted (backs). So this modal just renders options[] as-is.

  import type { CardView, GameView, PendingChoiceView } from "../../protocol";
  import Card from "./Card.svelte";

  interface Props {
    snap: GameView;
    viewerID: string | null;
    sendAction: (type: string, params?: unknown, player?: string) => void;
  }

  const { snap, viewerID, sendAction }: Props = $props();

  // First choice addressed to the viewer. Queue ordering: front of
  // list is "what the chooser sees next." One modal at a time; when
  // they resolve this one, the next pops automatically on the
  // snapshot after the server drains the entry.
  const active = $derived.by((): PendingChoiceView | null => {
    if (!viewerID || !snap.pending_choices) return null;
    for (const c of snap.pending_choices) {
      if (c.chooser === viewerID) return c;
    }
    return null;
  });

  const open = $derived(active !== null);

  // Source player (whose hand the picks come from). Used for the
  // modal header copy.
  const fromName = $derived.by(() => {
    if (!active) return "";
    const seat = snap.seats.find((s) => s.id === active.from_player);
    return seat?.name ?? "opponent";
  });

  const isSelfSource = $derived(active && viewerID && active.from_player === viewerID);

  let selected = $state<Set<string>>(new Set());

  // Reset selection whenever the modal opens fresh (active changes
  // from null → non-null, or the choice ID changes).
  let lastChoiceID: string | null = null;
  $effect(() => {
    const nextID = active?.id ?? null;
    if (nextID !== lastChoiceID) {
      selected = new Set();
      lastChoiceID = nextID;
    }
  });

  function toggle(id: string): void {
    if (!active) return;
    const next = new Set(selected);
    if (next.has(id)) next.delete(id);
    else if (next.size < active.count) next.add(id);
    selected = next;
  }

  function submit(): void {
    if (!active || !viewerID) return;
    if (selected.size !== active.count) return;
    sendAction(
      "resolve_choice",
      { choice_id: active.id, card_ids: Array.from(selected) },
      viewerID,
    );
  }

  // Options come redacted for non-knowers; filter to cards the
  // viewer can identify so we don't render a grid of anonymous
  // backs the viewer can't meaningfully pick from. Unknown
  // options here would mean the server misrouted — the chooser
  // should always be a knower of the revealed cards.
  const optionCards = $derived<CardView[]>(active?.options ?? []);
</script>

{#if open && active}
  <div class="backdrop" role="dialog" aria-modal="true" aria-labelledby="choice-title">
    <div class="modal">
      <h2 id="choice-title">
        {active.reason || "Choose"} — pick {active.count} card{active.count === 1 ? "" : "s"}
      </h2>
      <p class="hint">
        {#if isSelfSource}
          Pick {active.count} card{active.count === 1 ? "" : "s"} from your hand to discard.
        {:else}
          Pick {active.count} card{active.count === 1 ? "" : "s"} from
          <strong>{fromName}</strong>'s revealed hand.
          <strong>{fromName}</strong> will discard your pick{active.count === 1 ? "" : "s"}.
        {/if}
      </p>
      <div class="card-grid">
        {#each optionCards as c (c.instance_id)}
          <button
            type="button"
            class="card-pick"
            class:selected={selected.has(c.instance_id)}
            disabled={!selected.has(c.instance_id) && selected.size >= active.count}
            onclick={() => toggle(c.instance_id)}
            aria-pressed={selected.has(c.instance_id)}
            aria-label={`select ${c.name || "card"}`}
          >
            <Card card={c} />
          </button>
        {/each}
      </div>
      <div class="footer">
        <span class="counter">{selected.size} / {active.count} selected</span>
        <button
          type="button"
          class="submit"
          disabled={selected.size !== active.count}
          onclick={submit}
        >
          Confirm
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
  .hint strong {
    color: #cfd6ee;
    font-weight: 600;
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
