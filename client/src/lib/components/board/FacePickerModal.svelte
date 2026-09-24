<script lang="ts">
  // FacePickerModal — ADR 0034: choose which printed face of a modal
  // double-faced card you are playing.
  //
  // This is the one announce-time prompt that is not a cost, a mode
  // or a target: it decides WHAT THE CARD IS. Sea Gate Restoration
  // and Sea Gate, Reborn have different names, different types,
  // different costs and — through the composite catalog key — a
  // different rules implementation. Everything else the cast chain
  // asks about is a property of a card whose identity is already
  // settled, which is why this prompt opens FIRST, ahead of even the
  // alternative-cost one.
  //
  // Unlike every other prompt in the chain it shows ART. A modal DFC
  // is picked from hand under time pressure, both halves share one
  // physical card, and "the land one" is how players actually think
  // about it — a text list of two type lines is the wrong affordance
  // for a decision the printed card makes visually.
  import { onDestroy } from "svelte";
  import type { CardView } from "../../protocol";
  import { cardImageURL } from "../../cardImage";
  import { cardArt } from "../../cardArt";
  import { LAYOUT_ADVENTURE, castableFaceIndex } from "../../faces";
  import { hasSatisfiableTargets } from "../../timing";
  import ModalLayer from "../ModalLayer.svelte";

  interface Props {
    // The card being played; null closes the modal. Only ever
    // opened for a card whose `faces` has more than one entry and
    // whose layout offers a choice — see needsFacePicker.
    card: CardView | null;
    // Fires with the chosen face index.
    onConfirm: (face: number) => void;
    onCancel: () => void;
  }

  const { card, onConfirm, onCancel }: Props = $props();

  const faces = $derived(card?.faces ?? []);

  // #1173 / #1168: default to the face this cast can actually make,
  // not always the front. A printed graveyard permission on one half
  // only (`castable_here`) is the stronger signal — it names a zone
  // where casting the OTHER half is a rejected click, not a live
  // choice — so it wins when present; a hand or command-zone cast
  // never carries it (it is never stamped there at all), so this
  // falls back to a face whose own target clause isn't stuck on
  // nothing to point at. Falls back to the front face when nothing
  // distinguishes them, which is every ordinary case: it is the half
  // the card is named after, it is what a mis-click should land on,
  // and it matches the server's default for a cast that arrives with
  // no face at all.
  function preferredFace(c: CardView): number {
    return castableFaceIndex(
      c,
      (f) => f.castable_here === true || (!f.cant_cast && hasSatisfiableTargets(f.legal_targets)),
    );
  }
  let chosen = $state(0);
  let lastCardID: string | null = null;
  $effect(() => {
    const id = card?.instance_id ?? null;
    if (id !== lastCardID) {
      lastCardID = id;
      chosen = card ? preferredFace(card) : 0;
    }
  });

  function confirm(): void {
    onConfirm(chosen);
  }

  function handleKey(e: KeyboardEvent): void {
    if (!card) return;
    if (e.key === "Enter") {
      e.preventDefault();
      confirm();
    } else if (e.key === "Escape") {
      e.preventDefault();
      onCancel();
    } else if (e.key === "ArrowLeft" || e.key === "ArrowUp") {
      e.preventDefault();
      chosen = Math.max(0, chosen - 1);
    } else if (e.key === "ArrowRight" || e.key === "ArrowDown") {
      e.preventDefault();
      chosen = Math.min(faces.length - 1, chosen + 1);
    } else if (e.key >= "1" && e.key <= "9") {
      const n = Number(e.key) - 1;
      if (n < faces.length) {
        e.preventDefault();
        chosen = n;
      }
    }
  }
  $effect(() => {
    if (!card) return;
    document.addEventListener("keydown", handleKey);
    return () => document.removeEventListener("keydown", handleKey);
  });
  onDestroy(() => document.removeEventListener("keydown", handleKey));

  // "Play" for a land face, "Cast" for a spell — the CR 305.1 /
  // 601.2 distinction, and the difference a player is actually
  // choosing between on a land-backed MDFC.
  const verb = $derived(
    (faces[chosen]?.type_line ?? "").toLowerCase().includes("land") ? "Play" : "Cast",
  );

  // The rule the choice comes from, named for the player. A modal DFC
  // and an adventure card ask the same question — "which half?" — out
  // of two different rules, and the caption is the only place the
  // modal says which one it is looking at.
  const provenance = $derived(
    card?.layout === LAYOUT_ADVENTURE ? "adventure · CR 715.3" : "modal double-faced · CR 712.12",
  );
</script>

{#if card && faces.length > 1}
  <ModalLayer />
  <div class="prompt-backdrop" role="dialog" aria-modal="true" aria-labelledby="face-title">
    <div class="prompt-modal face-modal">
      <h2 id="face-title">
        {card.name}
        <span class="prompt-src" aria-hidden="true">{provenance}</span>
      </h2>
      <p class="prompt-hint">Which half are you playing?</p>
      <ul class="face-options">
        {#each faces as face, i (face.name)}
          {@const art = cardImageURL(card, "normal", i)}
          <li>
            <button
              type="button"
              class="face-opt"
              class:on={chosen === i}
              aria-pressed={chosen === i}
              onclick={() => (chosen = i)}
              ondblclick={() => {
                chosen = i;
                confirm();
              }}
            >
              {#if art}
                <img class="face-art" src={art} alt="" use:cardArt={art} />
              {:else}
                <div class="face-art face-art-blank" aria-hidden="true"></div>
              {/if}
              <span class="face-name">{face.name}</span>
              <span class="face-type">{face.type_line ?? ""}</span>
              <span class="face-cost">{face.mana_cost || "—"}</span>
            </button>
          </li>
        {/each}
      </ul>
      {#if faces[chosen]?.oracle_text}
        <p class="face-text">{faces[chosen].oracle_text}</p>
      {/if}
      <div class="prompt-foot">
        <button type="button" class="ghost" onclick={onCancel}
          >Cancel <span class="kbd">Esc</span></button
        >
        <button type="button" class="primary" onclick={confirm}>
          {verb}
          {faces[chosen]?.name ?? ""} <span class="kbd">↵</span>
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .face-modal {
    width: min(560px, calc(100vw - 32px));
  }

  .face-options {
    display: flex;
    gap: 12px;
    list-style: none;
    margin: 0;
    padding: 0;
  }

  .face-options li {
    flex: 1 1 0;
    min-width: 0;
  }

  .face-opt {
    /* positioned for the failed-art pip (#33); the offsets are the
       button's padding plus the pip's usual inset, so it lands on the
       art's corner rather than the button's */
    position: relative;
    --art-error-top: 11px;
    --art-error-right: 11px;
    display: grid;
    grid-template-rows: auto auto auto auto;
    gap: 2px;
    width: 100%;
    padding: 8px;
    border-radius: 10px;
    border: 2px solid transparent;
    background: var(--surface-sunken);
    color: inherit;
    cursor: pointer;
    text-align: left;
  }

  .face-opt:hover {
    border-color: var(--border-strong, rgba(255, 255, 255, 0.25));
  }

  .face-opt.on {
    border-color: var(--accent, #d9a441);
  }

  .face-art {
    width: 100%;
    aspect-ratio: 63 / 88;
    object-fit: cover;
    border-radius: 6px;
    display: block;
  }

  .face-art-blank {
    background: var(--surface-raised, rgba(255, 255, 255, 0.06));
  }

  .face-name {
    font-weight: 600;
    margin-top: 4px;
    overflow-wrap: anywhere;
  }

  .face-type {
    font-size: 12px;
    opacity: 0.75;
    overflow-wrap: anywhere;
  }

  .face-cost {
    font-family: var(--font-mono);
    font-size: 12px;
    opacity: 0.85;
  }

  .face-text {
    margin: 10px 0 0;
    font-size: 12px;
    line-height: 1.4;
    opacity: 0.85;
    white-space: pre-wrap;
    max-height: 8em;
    overflow-y: auto;
  }

  .primary .kbd {
    color: var(--accent-fg);
    border-color: rgba(28, 21, 3, 0.35);
    opacity: 0.8;
  }
</style>
