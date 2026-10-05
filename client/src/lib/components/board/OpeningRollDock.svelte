<script lang="ts">
  // OpeningRollDock — what the action dock asks a seat during the
  // opening roll (ADR 0121 §6). It renders nothing where it is mounted:
  // it opens dock requests (lib/dock.ts), as every prompt does.
  //
  //   - A seat that owes a die: the non-modal dialog `roll for the
  //     first turn`, primary `Roll` (Enter). The question reads "Roll a
  //     d20. The highest roll chooses who goes first." or, in a reroll,
  //     "You tied with 17. Roll again." It is the viewer's own: it never
  //     blocks another seat's dock.
  //   - The host also gets the secondary `Roll for everyone left` while
  //     some other seat has not rolled. Never the primary, never Enter.
  //     With nothing of their own to roll, the host's request reads
  //     "Waiting for Bob and Dave to roll" and offers only that button.
  //   - The chooser: the sheet `choose who takes the first turn`, one
  //     button per seat in turn order ("<name> goes first", "I go first"
  //     for their own). No primary, so Enter chooses nothing. "I go
  //     first" sends at once. Any other seat opens a confirm in its
  //     place (owner decision 6): the dialog `Let <name> take the first
  //     turn?`, primary `Confirm` (no Enter: the choice has no undo) and
  //     secondary `Cancel` (Escape), which goes back to the sheet.
  //
  // Everyone else reads who the table is waiting for in the dock's
  // status line (dockHint.ts).
  import type { GameView } from "../../protocol";
  import { cancelAction, type DockAction, type DockRequest as Req } from "../../dock";
  import {
    chooserLead,
    giveAwayQuestion,
    openingRollAsk,
    openingRollModel,
    startingChoices,
    type StartingChoice,
  } from "../../openingRoll";
  import { seatColor } from "../../colors";
  import { L } from "../../labels";
  import DockRequest from "./DockRequest.svelte";
  import DockSheet from "./DockSheet.svelte";

  interface Props {
    view: GameView;
    // The viewer's seat index; null for a viewer with no seat.
    viewerSeat: number | null;
    // The table's host or the admin (lib/tableSettings canManageTable):
    // who may press "Roll for everyone left". The server gates it too.
    isHost: boolean;
    onRoll: () => void;
    onRollForEveryone: () => void;
    onChoose: (seat: number) => void;
  }

  const { view, viewerSeat, isHost, onRoll, onRollForEveryone, onChoose }: Props = $props();

  const model = $derived(openingRollModel(view));
  const ask = $derived(openingRollAsk(model, view.seats, viewerSeat, isHost));

  // A button pressed waits for the next frame: a second press before
  // the server answers would only be refused. Any new view re-enables
  // it (the request is gone by then if the press landed).
  let pressedOn: GameView | null = $state.raw(null);
  const waiting = $derived(pressedOn !== null && pressedOn === view);
  function press(fn: () => void): () => void {
    return () => {
      if (waiting) return;
      pressedOn = view;
      fn();
    };
  }

  const rollRequest = $derived.by((): Req | null => {
    if (ask.kind !== "roll" && ask.kind !== "wait") return null;
    const secondary: DockAction[] =
      ask.others.length > 0
        ? [
            {
              id: "roll-everyone",
              label: L.rollForEveryone,
              title: "roll a d20 for every seat that has not rolled yet",
              disabled: waiting,
              onPress: press(onRollForEveryone),
            },
          ]
        : [];
    return {
      rank: "choice",
      label: L.rollForFirstTurn,
      tag: "d20",
      tone: "gold",
      question: ask.question,
      primary:
        ask.kind === "roll"
          ? {
              id: "roll",
              label: L.roll,
              title: "roll your d20",
              disabled: waiting,
              keyShortcuts: "Enter",
              cap: "⏎",
              onPress: press(onRoll),
            }
          : null,
      secondary,
    };
  });

  // The chooser's confirm (owner decision 6), by seat.
  let confirming: StartingChoice | null = $state.raw(null);
  // A new chooser (or none) drops a confirm left open.
  $effect(() => {
    if (ask.kind !== "choose") confirming = null;
  });
  const choices = $derived(
    model && model.chooser !== null ? startingChoices(view.seats, model.chooser) : [],
  );

  function choose(c: StartingChoice): void {
    if (c.self) {
      press(() => onChoose(c.seat))();
      return;
    }
    confirming = c;
  }

  const confirmRequest = $derived.by((): Req | null => {
    const c = confirming;
    if (!c) return null;
    const q = giveAwayQuestion(c.name);
    return {
      rank: "choice",
      label: q,
      question: q,
      hint: "There is no undo: the first turn is theirs.",
      // The dialog takes focus, not Confirm: a stray Enter must not
      // give the first turn away.
      focus: "dialog",
      primary: {
        id: "confirm",
        label: L.confirm,
        disabled: waiting,
        onPress: press(() => {
          confirming = null;
          onChoose(c.seat);
        }),
      },
      secondary: [cancelAction(() => (confirming = null))],
    };
  });
</script>

{#if rollRequest}
  <DockRequest request={rollRequest} />
{:else if ask.kind === "choose"}
  {#if confirmRequest}
    <DockRequest request={confirmRequest} />
  {:else}
    <DockSheet
      rank="choice"
      label={L.chooseFirstTurn}
      title="Choose who goes first"
      width={440}
      sheetKey="opening-roll-choice"
      question={chooserLead(ask.result)}
    >
      <div class="choices">
        {#each choices as c (c.seat)}
          <button
            type="button"
            class="choice"
            class:self={c.self}
            style="--seat-color: {seatColor(c.seat)}"
            disabled={waiting}
            onclick={() => choose(c)}
          >
            <span class="seat-dot" style="background:{seatColor(c.seat)}"></span>
            {c.label}
          </button>
        {/each}
      </div>
    </DockSheet>
  {/if}
{/if}

<style>
  .choices {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .choice {
    display: flex;
    align-items: center;
    gap: 8px;
    min-height: 36px;
    padding: 6px 12px;
    border-radius: 8px;
    border: 1px solid color-mix(in srgb, var(--seat-color) 45%, var(--border));
    background: color-mix(in srgb, var(--seat-color) 8%, transparent);
    color: var(--fg);
    font-size: 13px;
    font-weight: 600;
    text-align: left;
    cursor: pointer;
  }
  .choice:hover:not(:disabled) {
    background: color-mix(in srgb, var(--seat-color) 18%, transparent);
  }
  .choice.self {
    border-color: var(--accent);
    background: var(--accent-soft);
    color: var(--accent-strong);
  }
  .choice:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
  .choice:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .seat-dot {
    flex: none;
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }
</style>
