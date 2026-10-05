import type { GameView } from "./protocol";

// #1530: "Always ask me to order my triggers". The setting lives in
// Settings (gameplay.alwaysAskTriggerOrder) but the skip decision it
// controls is the server's (seatNeedsTriggerOrder), so the server holds
// the seat's copy. The server persists it per seat (snapshot + undo),
// and the client reconciles instead of firing on connect: when the
// seat's own view disagrees with the local setting, send the action.
// That covers a toggle, a reconnect, a second device and a server that
// restarted from an older snapshot, with one rule.

export interface TriggerOrderPrefState {
  // The value this client last sent and has not yet seen reflected.
  // Stops a refused or slow action from being re-sent on every frame.
  lastSent: boolean | null;
}

export function newTriggerOrderPrefState(): TriggerOrderPrefState {
  return { lastSent: null };
}

// triggerOrderPrefToSend returns the value to send, or null for
// nothing. Updates state. A viewer who is not seated (spectator,
// unseated admin) or an eliminated one never sends.
export function triggerOrderPrefToSend(
  state: TriggerOrderPrefState,
  view: GameView | null | undefined,
  viewerID: string | null,
  desired: boolean,
): boolean | null {
  if (!view || !viewerID) return null;
  const me = view.seats.find((s) => s.id === viewerID);
  if (!me || me.eliminated) return null;
  const server = me.trigger_order_always_ask === true;
  if (server === desired) {
    state.lastSent = null;
    return null;
  }
  if (state.lastSent === desired) return null;
  state.lastSent = desired;
  return desired;
}
