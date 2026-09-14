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
  import ModalLayer from "../ModalLayer.svelte";

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
  <ModalLayer />
  <div class="prompt-backdrop" role="dialog" aria-modal="true" aria-labelledby="auto-tap-title">
    <div class="prompt-modal tap-modal">
      <h2 id="auto-tap-title">
        Auto-tap & cast
        {#if preview}
          <span class="prompt-src" aria-hidden="true">{preview.cost || "no cost"}</span>
        {/if}
      </h2>
      {#if loading && !preview}
        <p class="prompt-hint">planning…</p>
      {:else if fetchError}
        <p class="prompt-hint error">preview failed: {fetchError}</p>
      {:else if preview}
        {#if preview.ok}
          <p class="prompt-hint">
            These {preview.plan?.length ?? 0} permanent(s) tap to pay for it. Lock a source to keep it
            for a later cast.
          </p>
          <ul class="prompt-options plan">
            {#each planCards() as row (row.id)}
              <li class="prompt-opt src-row">
                <span class="card-name">{row.name}</span>
                <button
                  type="button"
                  class="ghost lock-btn"
                  onclick={() => toggleLock(row.id)}
                  title="reserve this source for a later cast"
                >
                  lock
                </button>
              </li>
            {/each}
          </ul>
        {:else}
          <p class="prompt-hint error">
            Insufficient mana
            {#if preview.missing && preview.missing.length > 0}
              — missing {preview.missing.join(" ")}
            {/if}
          </p>
        {/if}
        {#if lockedSources.length > 0}
          <p class="locked-label">locked sources</p>
          <ul class="prompt-options locked">
            {#each lockedCards() as row (row.id)}
              <li class="prompt-opt src-row on">
                <span class="card-name">{row.name}</span>
                <button
                  type="button"
                  class="ghost lock-btn"
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
      <div class="prompt-foot">
        <button type="button" class="ghost" onclick={onCancel}
          >Cancel <span class="kbd">Esc</span></button
        >
        <button type="button" class="primary" onclick={confirm} disabled={!preview?.ok || loading}>
          Cast <span class="kbd">↵</span>
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .tap-modal {
    width: min(460px, calc(100vw - 32px));
  }
  .src-row {
    justify-content: space-between;
    padding: 6px 6px 6px 12px;
    cursor: default;
  }
  .card-name {
    font-size: 13px;
    font-weight: 500;
  }
  .lock-btn {
    height: 26px;
    padding: 0 10px;
    font-size: 11px;
    font-family: var(--font-mono);
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }
  .locked-label {
    margin: 2px 0 -6px;
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--fg-dim);
    font-weight: 700;
  }
  .primary .kbd {
    color: var(--accent-fg);
    border-color: rgba(28, 21, 3, 0.35);
    opacity: 0.8;
  }
</style>
