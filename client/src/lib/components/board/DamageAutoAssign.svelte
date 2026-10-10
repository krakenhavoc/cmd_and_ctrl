<script lang="ts">
  // DamageAutoAssign — "Auto-assign combat damage" (#2956, ADR 0147).
  //
  // When the viewer's next prompt is a damage assignment whose damage
  // covers lethal for every blocker (the server's `covers_lethal`) and
  // the setting is on, this sends the server's suggested split at once:
  // lethal to each blocker, the rest over with trample or onto the last
  // blocker without. ChoicePromptModal holds its sheet back meanwhile
  // (damageAssignment.ts autoAssignHolds). The dock then says what went
  // where for a few seconds, with Undo and Always ask.
  //
  // Each prompt is sent at most once. If it is still pending a few
  // seconds later (the server refused the split, or an undo brought it
  // back), its record expires and the sheet asks instead.
  import { onDestroy } from "svelte";

  import DockRequest from "./DockRequest.svelte";
  import type { ActionType, GameView } from "../../protocol";
  import { settings, updateSettings } from "../../settings";
  import {
    AUTO_ASSIGN_GRACE_MS,
    AUTO_ASSIGN_NOTICE_MS,
    autoAssignAnswer,
    autoAssignNoticeRequest,
    autoAssignRecords,
    autoAssignText,
  } from "../../damageAssignment";

  interface Props {
    view: GameView | null;
    viewerID: string | null;
    // False while a replay frame is on screen: a replay sends nothing.
    live: boolean;
    sendAction: (type: ActionType, params?: unknown, player?: string) => void;
    // Whether the viewer has an undo to spend.
    canUndo: boolean;
    onUndo: () => void;
  }
  const { view, viewerID, live, sendAction, canUndo, onUndo }: Props = $props();

  let notice = $state<{ choiceID: string; text: string } | null>(null);
  const timers: ReturnType<typeof setTimeout>[] = [];
  let noticeTimer: ReturnType<typeof setTimeout> | null = null;

  function nameOf(id: string): string {
    return view?.battlefield?.cards.find((c) => c.instance_id === id)?.name ?? "a creature";
  }

  function expire(choiceID: string): void {
    autoAssignRecords.update((r) =>
      r[choiceID] && !r[choiceID].expired
        ? { ...r, [choiceID]: { ...r[choiceID], expired: true } }
        : r,
    );
  }

  $effect(() => {
    if (!live || !view || !viewerID) return;
    const enabled = $settings.gameplay.autoAssignCombatDamage;
    // Only the viewer's next prompt: the queue is answered in order.
    const next = (view.pending_choices ?? []).find((c) => c.chooser === viewerID);
    const answer = autoAssignAnswer(next, viewerID, enabled);
    if (!next || !answer || $autoAssignRecords[next.id]) return;
    autoAssignRecords.update((r) => ({ ...r, [next.id]: { sentAt: Date.now(), expired: false } }));
    timers.push(setTimeout(() => expire(next.id), AUTO_ASSIGN_GRACE_MS));
    sendAction("resolve_choice", { choice_id: next.id, ...answer }, viewerID);

    const frame = next.damage_assignment!;
    const attacker = view.battlefield?.cards.find((c) => c.instance_id === frame.attacker_card_id);
    const defenderID = attacker?.attacking_target;
    const defender =
      view.seats.find((s) => s.id === defenderID)?.name ??
      (defenderID ? nameOf(defenderID) : "the defending player");
    notice = {
      choiceID: next.id,
      text: autoAssignText(attacker?.name ?? "Your attacker", answer, nameOf, defender),
    };
    if (noticeTimer) clearTimeout(noticeTimer);
    noticeTimer = setTimeout(() => {
      notice = null;
      noticeTimer = null;
    }, AUTO_ASSIGN_NOTICE_MS);
  });

  onDestroy(() => {
    for (const t of timers) clearTimeout(t);
    if (noticeTimer) clearTimeout(noticeTimer);
  });

  function dismiss(): void {
    notice = null;
    if (noticeTimer) clearTimeout(noticeTimer);
    noticeTimer = null;
  }

  const request = $derived.by(() => {
    if (!notice) return null;
    const choiceID = notice.choiceID;
    return autoAssignNoticeRequest(
      notice.text,
      canUndo,
      () => {
        // The undo brings the prompt back: ask it, don't assign it again.
        expire(choiceID);
        dismiss();
        onUndo();
      },
      () => {
        updateSettings("gameplay", "autoAssignCombatDamage", false);
        dismiss();
      },
    );
  });
</script>

{#if request}
  <DockRequest {request} />
{/if}
