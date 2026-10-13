<script lang="ts">
  // DeleteAccountDialog — the account menu's "Delete my account" (#2983,
  // ADR 0149): DeleteAccount in a modal frame. The form is the same one
  // the privacy page carries; this only supplies the backdrop, the
  // keyboard (Escape cancels) and the modal-layer registration that stops
  // the global shortcuts firing behind it.
  import { onMount } from "svelte";
  import DeleteAccount from "./DeleteAccount.svelte";
  import ModalLayer from "./ModalLayer.svelte";
  import { PRIVACY_HASH } from "../privacy";

  interface Props {
    onclose: () => void;
  }
  const { onclose }: Props = $props();

  let box: HTMLElement | undefined = $state();

  onMount(() => {
    box?.querySelector<HTMLInputElement>("input")?.focus();
  });

  function onKey(e: KeyboardEvent): void {
    if (e.key === "Escape") {
      e.preventDefault();
      onclose();
    }
  }
</script>

<svelte:window onkeydown={onKey} />

<ModalLayer />

<div class="backdrop" role="dialog" aria-modal="true" aria-label="Delete my account">
  <div class="modal" bind:this={box}>
    <DeleteAccount oncancel={onclose} ondeleted={onclose} />
    <p class="more">
      <a href={PRIVACY_HASH} onclick={onclose}>What else we store, and for how long</a>
    </p>
  </div>
</div>

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
  }
  .modal {
    background: linear-gradient(180deg, var(--surface) 0%, var(--bg-2) 100%);
    border: 1px solid var(--danger-line);
    border-radius: var(--radius-xl);
    padding: 22px 26px;
    width: min(560px, 92vw);
    max-height: 86vh;
    overflow: auto;
    box-shadow: 0 30px 80px rgba(0, 0, 0, 0.7);
  }
  .more {
    margin: 12px 0 0;
    font-size: 12px;
  }
  .more a {
    color: var(--fg-muted);
    text-decoration: underline;
  }
</style>
