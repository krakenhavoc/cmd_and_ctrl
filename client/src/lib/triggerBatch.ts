// triggerBatch — how many of one trigger are waiting, and which one a
// prompt is answering (#3817).
//
// A big board turns one event into dozens of triggers: twenty Blood
// Artists and two deaths are forty "target player loses 1 life"
// abilities, each with its own target prompt (CR 603.3d). The server
// holds the whole batch off the stack until every one of them has a
// target and an order (#1529, CR 603.3b), so while the player answers,
// the stack does not change and no life total moves. Each prompt looked
// exactly like the last one (the same pile, the same arrow, the same
// sentence), and a player answering the third of forty read it as the
// click not working.
//
// So the prompt says which one it is ("trigger 3 of 40"), and the stack
// panel says what is waiting ("Blood Artist ×40 waiting to go on the
// stack · 37 need your target").
//
// Everything is read from the snapshot, so it needs no memory of its
// own: the triggers still asking for a target are the pick_target
// prompts, and the ones already answered sit in `pending_triggers`
// (held there until the batch is complete). A batch is the triggers of
// one player from sources with the same name: twenty copies of Blood
// Artist are one batch, which is what the player sees as one pile.

import type { CardView, GameView, PendingChoiceView, StackItemView } from "./protocol";

const EM_DASH = "—";

/** One card name's triggers waiting to go on the stack for one player. */
export interface WaitingTriggers {
  /** The source's name ("Blood Artist"). */
  name: string;
  /** The player who puts them on the stack (who answers the prompts). */
  player: string;
  /** How many there are in all: still asking for a target, plus answered. */
  count: number;
  /** How many still need a target. */
  needTarget: number;
}

type Names = ReadonlyMap<string, string>;

/** cardNames maps every card the snapshot shows to its name. */
export function cardNames(
  view: Pick<GameView, "battlefield" | "stack" | "exile" | "seats">,
): Map<string, string> {
  const out = new Map<string, string>();
  const add = (cards: readonly CardView[] | undefined) => {
    for (const c of cards ?? []) if (c.name) out.set(c.instance_id, c.name);
  };
  add(view.battlefield?.cards);
  add(view.stack?.cards);
  add(view.exile?.cards);
  for (const s of view.seats ?? []) {
    add(s.graveyard?.cards);
    add(s.hand?.cards);
  }
  return out;
}

function labelSource(label: string | undefined): string | null {
  if (!label) return null;
  const at = label.indexOf(EM_DASH);
  if (at <= 0) return null;
  return label.slice(0, at).trim() || null;
}

/** The player a prompt is for: the controlled player, when another answers it (ADR 0166). */
function playerOf(c: PendingChoiceView): string {
  return c.on_behalf_of ?? c.chooser;
}

function isTargetPrompt(c: PendingChoiceView | null | undefined): c is PendingChoiceView {
  return !!c && c.kind === "pick_target";
}

function promptName(c: PendingChoiceView, names: Names): string | null {
  return (c.source && names.get(c.source)) || null;
}

function triggerName(it: StackItemView, names: Names): string | null {
  return names.get(it.source_card_id) ?? labelSource(it.label);
}

/**
 * waitingTriggers groups the triggers waiting to go on the stack by
 * player and source name, in the order they first appear: the target
 * prompts first (they are what the table waits on), then the triggers
 * already answered. A trigger whose source has no name the viewer can
 * read is left out.
 */
export function waitingTriggers(
  choices: readonly PendingChoiceView[] | undefined,
  pending: readonly StackItemView[] | undefined,
  names: Names,
): WaitingTriggers[] {
  const out: WaitingTriggers[] = [];
  const byKey = new Map<string, WaitingTriggers>();
  const group = (player: string, name: string): WaitingTriggers => {
    const key = `${player}\u0000${name}`;
    let g = byKey.get(key);
    if (!g) {
      g = { name, player, count: 0, needTarget: 0 };
      byKey.set(key, g);
      out.push(g);
    }
    return g;
  };
  for (const c of choices ?? []) {
    if (!isTargetPrompt(c)) continue;
    const name = promptName(c, names);
    if (!name) continue;
    const g = group(playerOf(c), name);
    g.count++;
    g.needTarget++;
  }
  for (const it of pending ?? []) {
    if (it.kind !== "triggered") continue;
    const name = triggerName(it, names);
    if (!name) continue;
    group(it.controller, name).count++;
  }
  return out;
}

/** Where one target prompt sits in its batch: 1-based, and the batch's size. */
export interface BatchPosition {
  index: number;
  total: number;
}

/**
 * batchPosition is "trigger `index` of `total`" for the target prompt
 * `choiceID`, or null when it is not a target prompt or its batch is
 * just itself. The answered siblings come first (they are in
 * `pending_triggers`), then the prompts in the server's queue order.
 */
export function batchPosition(
  view: Pick<
    GameView,
    "battlefield" | "stack" | "exile" | "seats" | "pending_choices" | "pending_triggers"
  >,
  choiceID: string | undefined,
): BatchPosition | null {
  if (!choiceID) return null;
  const choices = view.pending_choices ?? [];
  const mine = choices.find((c) => c.id === choiceID);
  if (!isTargetPrompt(mine)) return null;
  const names = cardNames(view);
  const name = promptName(mine, names);
  if (!name) return null;
  const player = playerOf(mine);
  const siblings = choices.filter(
    (c) => isTargetPrompt(c) && playerOf(c) === player && promptName(c, names) === name,
  );
  const answered = (view.pending_triggers ?? []).filter(
    (it) => it.kind === "triggered" && it.controller === player && triggerName(it, names) === name,
  ).length;
  const total = answered + siblings.length;
  if (total < 2) return null;
  return { index: answered + siblings.indexOf(mine) + 1, total };
}

/**
 * waitingLine is the stack panel's sentence for what is waiting, one
 * clause per batch, or "" with nothing waiting. `you` is the viewer's
 * seat id (null for a spectator); `playerName` names anyone else.
 */
export function waitingLine(
  groups: readonly WaitingTriggers[],
  you: string | null,
  playerName: (id: string) => string | undefined,
): string {
  return groups
    .map((g) => {
      const head =
        g.count === 1
          ? `${g.name}'s trigger is waiting to go on the stack`
          : `${g.name} ×${g.count} waiting to go on the stack`;
      if (g.needTarget === 0) return head;
      const whose = g.player === you ? "your" : `${playerName(g.player) ?? "a player"}'s`;
      const verb = g.needTarget === 1 ? "needs" : "need";
      const n = g.count === 1 ? "" : `${g.needTarget} `;
      return `${head} · ${n}${verb} ${whose} target`;
    })
    .join("; ");
}
