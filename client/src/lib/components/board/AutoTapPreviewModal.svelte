<script lang="ts">
  // S15 sub-PR 5 auto-tap-and-cast preview modal.
  //
  // Opens when the viewer chooses "Auto-tap & cast" from the
  // insufficient-mana toast (or any future right-click → cast
  // affordance). On mount it fetches `/games/:id/auto-tap-preview`
  // for the chosen card and renders the proposed plan: the ordered
  // list of permanents the server would tap. Enter confirms
  // (re-fires cast_spell with auto_tap: true); ESC cancels.
  //
  // Lock-tap UI: each row in the proposed plan has a "lock" toggle.
  // Clicking it adds the source to the excluded set and re-fetches
  // — letting the player reserve a land for a later cast and watch
  // the planner re-route around it. When the planner can no longer
  // satisfy the cost, the modal shows the missing-symbols list and
  // disables the confirm button.

  import { onDestroy } from "svelte";
  import { fetchAutoTapPreview, type AutoTapPreview } from "../../api";
  import type { CardView, GameView } from "../../protocol";

  interface Props {
    gameID: string;
    snap: GameView;
    cardID: string | null;
    onConfirm: (lockedSources: string[]) => void;
    onCancel: () => void;
  }

  const { gameID, snap, cardID, onConfirm, onCancel }: Props = $props();

  let preview = $state<AutoTapPreview | null>(null);
  let loading = $state(false);
  let fetchError = $state<string | null>(null);
  let lockedSources = $state<string[]>([]);

  // Reset locked set whenever a new card opens the modal. Without
  // this, an excluded source from a prior cast would carry over
  // and silently sabotage the next plan.
  let lastCardID: string | null = null;
  $effect(() => {
    if (cardID !== lastCardID) {
      lastCardID = cardID;
      lockedSources = [];
      preview = null;
      fetchError = null;
    }
  });

  // Refetch whenever the card or the locked set changes. The server
  // is read-only on this endpoint so the round-trip is cheap; a
  // small debounce isn't necessary because lock toggles are user-
  // driven (one click → one fetch).
  //
  // fetchSeq is a monotonically increasing request id (deliberately a
  // plain let, not $state — it must not re-trigger this effect). Each
  // run claims the next id; a response only lands if it's still the
  // latest. Guarding by cardID alone isn't enough: two quick lock
  // toggles for the same card leave two fetches in flight and the
  // last to RESOLVE would win, possibly showing a plan computed for
  // an outdated lockedSources set. Bumping before the early return
  // also invalidates in-flight responses when the modal closes.
  let fetchSeq = 0;
  $effect(() => {
    const reqID = ++fetchSeq;
    if (!cardID) return;
    const card = cardID;
    const locked = lockedSources.slice();
    loading = true;
    fetchError = null;
    fetchAutoTapPreview(gameID, card, { excluded: locked })
      .then((p) => {
        if (reqID === fetchSeq) {
          preview = p;
        }
      })
      .catch((err: unknown) => {
        if (reqID === fetchSeq) {
          fetchError = err instanceof Error ? err.message : "preview failed";
        }
      })
      .finally(() => {
        if (reqID === fetchSeq) loading = false;
      });
  });

  // Build a lookup from instance_id → CardView from the snapshot's
  // battlefield zone. The plan only references cards on the
  // battlefield, so this covers every entry.
  const battlefieldByID = $derived.by(() => {
    const m = new Map<string, CardView>();
    for (const c of snap.battlefield?.cards ?? []) {
      m.set(c.instance_id, c);
    }
    return m;
  });

  function planCards(): { id: string; name: string; locked: boolean }[] {
    if (!preview?.plan) return [];
    return preview.plan.map((id) => {
      const card = battlefieldByID.get(id);
      return {
        id,
        name: card?.name ?? id.slice(0, 8),
        locked: lockedSources.includes(id),
      };
    });
  }

  function lockedCards(): { id: string; name: string }[] {
    return lockedSources.map((id) => {
      const card = battlefieldByID.get(id);
      return { id, name: card?.name ?? id.slice(0, 8) };
    });
  }

  function toggleLock(id: string): void {
    if (lockedSources.includes(id)) {
      lockedSources = lockedSources.filter((s) => s !== id);
    } else {
      lockedSources = [...lockedSources, id];
    }
  }

  function confirm(): void {
    if (!preview?.ok || loading) return;
    onConfirm(lockedSources.slice());
  }

  // Keyboard-only confirm/cancel. Captured at the document level so
  // the modal works even when focus has drifted off the buttons —
  // a reflex-Enter-then-cast flow shouldn't require a tab dance.
  function handleKey(e: KeyboardEvent): void {
    if (!cardID) return;
    if (e.key === "Enter") {
      e.preventDefault();
      confirm();
    } else if (e.key === "Escape") {
      e.preventDefault();
      onCancel();
    }
  }

  $effect(() => {
    if (!cardID) return;
    document.addEventListener("keydown", handleKey);
    return () => document.removeEventListener("keydown", handleKey);
  });

  onDestroy(() => {
    document.removeEventListener("keydown", handleKey);
  });
</script>

{#if cardID}
  <div class="backdrop" role="dialog" aria-modal="true" aria-labelledby="auto-tap-title">
    <div class="modal">
      <h2 id="auto-tap-title">Auto-tap & cast</h2>
      {#if loading && !preview}
        <p class="muted">planning…</p>
      {:else if fetchError}
        <p class="error">preview failed: {fetchError}</p>
      {:else if preview}
        <p class="cost">cost: <span class="mono">{preview.cost || "(none)"}</span></p>
        {#if preview.ok}
          <p class="hint">tap these {preview.plan?.length ?? 0} permanent(s):</p>
          <ul class="plan">
            {#each planCards() as row (row.id)}
              <li>
                <span class="card-name">{row.name}</span>
                <button
                  type="button"
                  class="lock-btn"
                  onclick={() => toggleLock(row.id)}
                  title="reserve this source for a later cast"
                >
                  lock
                </button>
              </li>
            {/each}
          </ul>
        {:else}
          <p class="error">
            insufficient mana
            {#if preview.missing && preview.missing.length > 0}
              — missing {preview.missing.join(" ")}
            {/if}
          </p>
        {/if}
        {#if lockedSources.length > 0}
          <p class="hint locked-label">locked sources:</p>
          <ul class="locked">
            {#each lockedCards() as row (row.id)}
              <li>
                <span class="card-name">{row.name}</span>
                <button
                  type="button"
                  class="lock-btn unlock"
                  onclick={() => toggleLock(row.id)}
                  title="release this source back to the auto-tapper"
                >
                  unlock
                </button>
              </li>
            {/each}
          </ul>
        {/if}
      {/if}
      <div class="actions">
        <button type="button" class="cancel" onclick={onCancel}>cancel (esc)</button>
        <button type="button" class="confirm" onclick={confirm} disabled={!preview?.ok || loading}>
          cast (enter)
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
    z-index: 1100;
  }
  .modal {
    background: #1c1f2a;
    color: #e6e8ee;
    border: 1px solid #3a4055;
    border-radius: 6px;
    padding: 1.25rem 1.5rem;
    min-width: 340px;
    max-width: 480px;
    box-shadow: 0 12px 28px rgba(0, 0, 0, 0.4);
  }
  h2 {
    margin: 0 0 0.5rem 0;
    font-size: 1.05rem;
    color: #f4ead5;
  }
  .cost {
    margin: 0 0 0.75rem 0;
    color: #aab2c8;
    font-size: 0.9rem;
  }
  .mono {
    font-family: "Menlo", "Monaco", monospace;
    color: #e6e8ee;
  }
  .hint {
    margin: 0.5rem 0 0.25rem 0;
    color: #aab2c8;
    font-size: 0.85rem;
  }
  .locked-label {
    margin-top: 1rem;
    color: #ff9a85;
  }
  .error {
    margin: 0.5rem 0;
    color: #ff9a85;
    font-size: 0.9rem;
  }
  .muted {
    margin: 0.5rem 0;
    color: #6b7280;
    font-size: 0.85rem;
  }
  ul.plan,
  ul.locked {
    list-style: none;
    margin: 0.25rem 0 0 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }
  ul.plan li,
  ul.locked li {
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: #252a37;
    padding: 0.4rem 0.6rem;
    border-radius: 4px;
    font-size: 0.9rem;
  }
  ul.locked li {
    background: #3a2a2a;
  }
  .card-name {
    color: #e6e8ee;
  }
  .lock-btn {
    background: transparent;
    color: #aab2c8;
    border: 1px solid #3a4055;
    border-radius: 3px;
    padding: 0.15rem 0.5rem;
    font-size: 0.75rem;
    cursor: pointer;
  }
  .lock-btn:hover {
    color: #f4ead5;
    border-color: #6b7280;
  }
  .lock-btn.unlock {
    color: #ff9a85;
    border-color: #6b3030;
  }
  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
    margin-top: 1rem;
  }
  .actions button {
    padding: 0.4rem 0.9rem;
    border-radius: 4px;
    border: 1px solid #3a4055;
    background: #252a37;
    color: #e6e8ee;
    cursor: pointer;
    font-size: 0.9rem;
  }
  .actions .confirm {
    background: #3b6f3b;
    border-color: #4f9a4f;
  }
  .actions .confirm:disabled {
    background: #2a3a2a;
    border-color: #3a4055;
    color: #6b7280;
    cursor: not-allowed;
  }
  .actions .cancel:hover {
    border-color: #6b7280;
  }
</style>
