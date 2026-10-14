<script lang="ts">
  // AutoPassNotice — the dock's short chip after the client passed
  // priority for the viewer (#3343): "Passed for you: nothing to respond
  // with", with a one-tap Undo while the existing undo would take the
  // pass back. A status, not a question: a `step` request, so it takes
  // no keys and blocks nothing, and any other request outranks it. It
  // stays about six seconds.
  //
  // Game.svelte hands it each automatic pass as a new `notice` object;
  // the rules for the wording and for when Undo is offered live in
  // lib/passReason.ts. Undo is the ordinary `undo` action: same budget,
  // same stack gates, nothing new on the server.
  import { onDestroy } from "svelte";

  import DockRequest from "./DockRequest.svelte";
  import {
    PASS_NOTICE_MS,
    passNoticeRequest,
    passNoticeUndoable,
    type PassNotice,
    type PassNoticeFacts,
  } from "../../passReason";

  interface Props {
    // The newest automatic pass, or null. A new object replaces the chip.
    notice: PassNotice | null;
    // What passNoticeUndoable reads, as the frame on screen shows it.
    facts: PassNoticeFacts;
    // The frame id of an action the server refused (lastError.replyTo):
    // a refused pass was not made, so its chip goes away.
    refusedFrame: string | null;
    // False while a replay frame is on screen: a replay never notifies.
    live: boolean;
    onUndo: () => void;
  }
  let { notice, facts, refusedFrame, live, onUndo }: Props = $props();

  let current = $state<PassNotice | null>(null);
  let seen: PassNotice | null = null;
  let timer: ReturnType<typeof setTimeout> | null = null;

  function clearTimer(): void {
    if (timer) clearTimeout(timer);
    timer = null;
  }
  function dismiss(): void {
    current = null;
    clearTimer();
  }

  $effect(() => {
    const n = notice;
    if (!n || n === seen) return;
    seen = n;
    current = n;
    clearTimer();
    timer = setTimeout(dismiss, PASS_NOTICE_MS);
  });
  onDestroy(clearTimer);

  const request = $derived.by(() => {
    if (!live || !current || refusedFrame === current.frame) return null;
    return passNoticeRequest(current, passNoticeUndoable(current, facts), () => {
      dismiss();
      onUndo();
    });
  });
</script>

{#if request}
  <DockRequest {request} />
{/if}
