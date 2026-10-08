<script lang="ts">
  // AutoAnswerNotice — the dock's short notice after the server answered
  // a prompt for the viewer with their standing answer (ADR 0127 §6):
  // "Rhystic Study: paid {1} for you", with Undo and Ask me next time.
  // A status, not a question: a `step` request, so it takes no keys and
  // blocks nothing, and any other request outranks it. It stays about
  // six seconds.
  //
  // It is driven by the viewer's own `auto_answer` log lines, which
  // carry the rule's key on the chooser's view only. Undo works exactly
  // while the room stamps that line's seq as the viewer's top undo entry
  // (`undo_auto_answer`, the owner's 2026-10-07 amendment).
  import { onDestroy } from "svelte";

  import DockRequest from "./DockRequest.svelte";
  import { findCardView } from "../../commanderReturn";
  import type { GameView } from "../../protocol";
  import { settings, updateSettings } from "../../settings";
  import {
    AUTO_ANSWER_NOTICE_MS,
    autoAnswerNotice,
    autoAnswerNoticeRequest,
    newAutoAnswerEntries,
    withoutRule,
  } from "../../autoAnswerPref";

  interface Props {
    view: GameView | null;
    viewerID: string | null;
    // False while a replay frame is on screen: a replay never notifies.
    live: boolean;
    onUndo: () => void;
  }
  let { view, viewerID, live, onUndo }: Props = $props();

  // The highest auto_answer seq already accounted for. Null until the
  // first live frame, which sets it without a notice: lines already in
  // the log when the table opens are history, not news.
  let since: number | null = null;
  let current = $state<{ seq: number; key: string } | null>(null);
  let timer: ReturnType<typeof setTimeout> | null = null;

  function clearTimer(): void {
    if (timer) clearTimeout(timer);
    timer = null;
  }

  $effect(() => {
    if (!live || !view || !viewerID) return;
    const lastSeq = Math.max(0, ...(view.log ?? []).map((e) => e.seq));
    if (since === null) {
      since = lastSeq;
      return;
    }
    // An undo rewinds the event sequence: later lines may reuse seqs.
    if (since > lastSeq) since = lastSeq;
    const fresh = newAutoAnswerEntries(view, viewerID, since);
    const top = fresh.length > 0 ? fresh[fresh.length - 1] : null;
    if (!top) return;
    since = top.seq;
    current = { seq: top.seq, key: top.auto_answer_key ?? "" };
    clearTimer();
    timer = setTimeout(() => {
      current = null;
      timer = null;
    }, AUTO_ANSWER_NOTICE_MS);
  });
  onDestroy(clearTimer);

  const request = $derived.by(() => {
    if (!current || !view) return null;
    const entry = (view.log ?? []).find((e) => e.seq === current?.seq && e.kind === "auto_answer");
    if (!entry) return null;
    const card = findCardView(view, entry.card_id)?.name ?? "";
    const notice = autoAnswerNotice(entry, card, view, viewerID);
    const key = current.key;
    return autoAnswerNoticeRequest(
      notice,
      () => {
        current = null;
        clearTimer();
        onUndo();
      },
      () => {
        updateSettings("gameplay", "autoAnswers", withoutRule($settings.gameplay.autoAnswers, key));
        current = null;
        clearTimer();
      },
    );
  });
</script>

{#if request}
  <DockRequest {request} />
{/if}
