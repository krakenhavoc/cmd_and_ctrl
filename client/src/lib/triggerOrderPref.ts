import type { GameView } from "./protocol";
import type { TriggerOrderMode } from "./settings";

// #1530, #1968: when the game asks you to order your triggers. The
// setting lives in Settings (gameplay.triggerOrder: "when_it_matters",
// "always" or "never") but the skip decision it controls is the
// server's (seatNeedsTriggerOrder), so the server holds the seat's
// copy. The server persists it per seat (snapshot + undo), and the
// client reconciles instead of firing on connect: when the seat's own
// view disagrees with the local setting, send the action. That covers
// a change, a reconnect, a second device and a server that restarted
// from an older snapshot, with one rule.

export interface TriggerOrderPrefState {
  // The value this client last sent and has not yet seen reflected.
  // Stops a slow action from being re-sent on every frame. Cleared when
  // the server refuses the send (triggerOrderPrefRefused), so the next
  // frame tries again.
  lastSent: TriggerOrderMode | null;
  // The frame id of that send, to recognise its refusal.
  frameID: string | null;
}

export function newTriggerOrderPrefState(): TriggerOrderPrefState {
  return { lastSent: null, frameID: null };
}

// serverTriggerOrder reads a seat's mode off its view. A server from
// #1968 on sends trigger_order ("always" / "never", absent for the
// default); one from before it sends only the #1530 boolean.
export function serverTriggerOrder(seat: {
  trigger_order?: string;
  trigger_order_always_ask?: boolean;
}): TriggerOrderMode {
  if (seat.trigger_order === "always" || seat.trigger_order === "never") {
    return seat.trigger_order;
  }
  if (seat.trigger_order === undefined && seat.trigger_order_always_ask === true) {
    return "always";
  }
  return "when_it_matters";
}

// triggerOrderPrefToSend returns the mode to send, or null for nothing.
// Updates state. A viewer who is not seated (spectator, unseated admin)
// or an eliminated one never sends.
export function triggerOrderPrefToSend(
  state: TriggerOrderPrefState,
  view: GameView | null | undefined,
  viewerID: string | null,
  desired: TriggerOrderMode,
): TriggerOrderMode | null {
  if (!view || !viewerID) return null;
  const me = view.seats.find((s) => s.id === viewerID);
  if (!me || me.eliminated) return null;
  if (serverTriggerOrder(me) === desired) {
    state.lastSent = null;
    state.frameID = null;
    return null;
  }
  if (state.lastSent === desired) return null;
  state.lastSent = desired;
  state.frameID = null;
  return desired;
}

// triggerOrderPrefSent records the frame id sendAction returned for the
// send triggerOrderPrefToSend asked for. A null id (the socket was not
// open) is a send that never left, so the next frame retries it.
export function triggerOrderPrefSent(state: TriggerOrderPrefState, frameID: string | null): void {
  if (frameID === null) {
    state.lastSent = null;
    state.frameID = null;
    return;
  }
  state.frameID = frameID;
}

// triggerOrderPrefRefused clears the pending send when `err` is the
// server's refusal of it, so the reconcile sends again on the next
// frame instead of waiting for a reflection that will never come.
// Returns whether it was.
export function triggerOrderPrefRefused(
  state: TriggerOrderPrefState,
  err: { replyTo?: string } | null | undefined,
): boolean {
  if (!err?.replyTo || state.frameID === null || err.replyTo !== state.frameID) return false;
  state.lastSent = null;
  state.frameID = null;
  return true;
}
