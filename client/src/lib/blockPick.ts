// blockPick.ts — declaring blockers on the board (#3898, ADR 0105's
// amendment "Blocking on the board").
//
// During declare blockers the defender clicks one of their creatures,
// then the attacker it should block, and nothing opens in between: no
// picker, no list. Each block is staged on the server as it is made
// (declare_blocker), drawn as a block arrow, and the dock's existing
// "Done blocking" confirms the lot.
//
//   - A click on your creature selects it; a second click lets it go.
//   - With a creature selected, the attackers it can block wear the
//     target ring. The ones it can't are greyed and unclickable, with
//     the server's sentence on hover (CardView.block_refusals).
//   - A click on an attacker blocks it. On the attacker the selected
//     creature already blocks, the click takes the block back; on
//     another, it changes the block (the server re-points it), or adds
//     one for a creature that can block several (#1706), which stays
//     selected while it has room.
//   - An attacker that takes more than one blocker (menace, CR
//     509.1b) cannot be blocked by one creature, so its click starts a
//     GROUP: the dock asks for the rest, each further creature clicked
//     joins it, and the whole group is sent as one declare_blockers once
//     it is big enough (min_blockers). Taking one creature out of a
//     group that would leave it too small takes the rest out too.
//   - "Remove block" in the dock, or the list's "No block", takes the
//     selected creature's blocks back (remove_from_combat, then the
//     blocks it keeps re-declared).
//
// Legality is the server's, never re-derived here: the refusals and the
// minimum are stamped on the cards, the candidates come from the
// legal-action digest. A click the server would still refuse comes back
// as the illegal_block toast, as before.
//
// Piles (#1724 / #3046): a pile of identical attackers, or of your own
// creatures, fans out into its cards on click while you declare
// (`fannedPiles`), so you pick the copy itself. The list fallback
// (BlockListSheet, "Show as list" in the dock) lists every creature and
// every attacker one by one, and is the keyboard path.
//
// Pure functions plus two stores, so the state machine is pinned by
// blockPick.test.ts without a component.

import { guardedWritable } from "./guardedStore";
import { attackersDefendedBy, blockedAttackersOf, defendingPlayerOf } from "./attackTargets";
import { isCreature } from "./cardTypes";
import type { LegalActions } from "./legalActions";
import type { CardView, GameView } from "./protocol";

// A group being collected for an attacker that takes more than one
// blocker. Nothing is sent until it has `need` creatures.
export interface BlockGroup {
  attacker: string;
  blockers: readonly string[];
  need: number;
}

export interface BlockPickState {
  // The creature whose block the next attacker click declares.
  selected: string | null;
  group: BlockGroup | null;
}

export const NO_BLOCK_PICK: BlockPickState = Object.freeze({ selected: null, group: null });

// What a click on one incoming attacker would do for the selected
// creature.
export type BlockTarget =
  // Declares the block (or changes it, or adds one).
  | { kind: "block" }
  // The selected creature blocks it already: the click takes it back.
  | { kind: "remove" }
  // It takes `need` blockers: the click starts a group.
  | { kind: "group"; need: number; reason: string }
  // It can't be blocked by this creature; `reason` is the server's.
  | { kind: "refused"; reason: string };

export interface BlockContext {
  view: GameView;
  viewerID: string;
  // The frame's FULL legal-action lookup (legalGate), never the
  // highlight-filtered one: the setting must not open or shut a click.
  legal: LegalActions;
}

export interface BlockPair {
  blocker: string;
  attacker: string;
}

export type BlockSend =
  | { type: "declare_blocker"; params: BlockPair }
  | { type: "declare_blockers"; params: { blocks: BlockPair[] } }
  | { type: "remove_from_combat"; params: { instance_id: string } };

export interface BlockStep {
  state: BlockPickState;
  // Sent in order.
  send: BlockSend[];
  // Why the click did nothing, for a click on a refused attacker.
  refused?: string;
}

export type BlockEvent =
  | { kind: "own"; id: string }
  | { kind: "attacker"; id: string }
  // The dock's Remove block: every block of the selected creature.
  | { kind: "remove" }
  | { kind: "cancel" };

function cardOf(view: GameView, id: string): CardView | undefined {
  return view.battlefield?.cards?.find((c) => c.instance_id === id);
}

// isOwnCreature is a creature the viewer controls on the battlefield.
export function isOwnCreature(ctx: BlockContext, id: string): boolean {
  const c = cardOf(ctx.view, id);
  return !!c && c.controller === ctx.viewerID && isCreature(c);
}

// canSelectBlocker is a creature that takes part in the declaration: it
// may block something (the digest), it blocks something now (a click on
// its attacker takes the block back), or the server listed why it can't
// block an attacker (selecting it shows the reasons).
export function canSelectBlocker(ctx: BlockContext, id: string): boolean {
  const c = cardOf(ctx.view, id);
  if (!c || c.controller !== ctx.viewerID || !isCreature(c)) return false;
  return (
    ctx.legal.blockableAttackers(id).length > 0 ||
    blockedAttackersOf(c).length > 0 ||
    (c.block_refusals?.length ?? 0) > 0
  );
}

// blockTargets is, for the selected creature, what a click on each
// attacker the viewer defends against would do.
export function blockTargets(ctx: BlockContext, blockerID: string): Map<string, BlockTarget> {
  const out = new Map<string, BlockTarget>();
  const b = cardOf(ctx.view, blockerID);
  if (!b) return out;
  const blocking = new Set(blockedAttackersOf(b));
  for (const a of attackersDefendedBy(ctx.view, ctx.viewerID)) {
    const id = a.instance_id;
    if (blocking.has(id)) {
      out.set(id, { kind: "remove" });
      continue;
    }
    const r = b.block_refusals?.find((x) => x.attacker === id);
    if (!r) {
      out.set(id, { kind: "block" });
    } else if (r.reason === "too_few_blockers" && (a.min_blockers ?? 0) >= 2) {
      out.set(id, { kind: "group", need: a.min_blockers!, reason: r.message });
    } else {
      out.set(id, { kind: "refused", reason: r.message });
    }
  }
  return out;
}

// groupTargets is blockTargets while a group is being collected: only
// its attacker is live, and every other attacker waits.
export function groupTargets(ctx: BlockContext, group: BlockGroup): Map<string, BlockTarget> {
  const out = new Map<string, BlockTarget>();
  for (const a of attackersDefendedBy(ctx.view, ctx.viewerID)) {
    if (a.instance_id === group.attacker) {
      out.set(a.instance_id, { kind: "group", need: group.need, reason: "" });
    }
  }
  return out;
}

// canJoinGroup is a creature that may be one of a group's blockers: it
// is a candidate, not already in the group, and nothing but the count
// stops it blocking the group's attacker.
export function canJoinGroup(ctx: BlockContext, group: BlockGroup, id: string): boolean {
  if (group.blockers.includes(id) || !canSelectBlocker(ctx, id)) return false;
  const t = blockTargets(ctx, id).get(group.attacker);
  return t?.kind === "group" || t?.kind === "block";
}

// hasRoomAfter reports whether the blocker could still take one more
// attacker once it also blocks `attackerID`: the server's capacity
// stamps (#1706), so a High Ground creature stays selected for its
// second block and an ordinary one lets go.
function hasRoomAfter(b: CardView, attackerID: string): boolean {
  if (b.blocks_any_number) return true;
  const after = new Set([...blockedAttackersOf(b), attackerID]);
  return after.size < (b.block_capacity ?? 1);
}

// releaseSends takes `pairs` out of the declaration. remove_from_combat
// clears a creature's every block, so a creature that keeps some (a
// multi-blocker) gets them back in one declare_blockers after. A pair
// whose attacker would be left with fewer blockers than its minimum
// takes that attacker's other blockers out with it, so the declaration
// is never left holding a block the server would refuse.
export function releaseSends(ctx: BlockContext, pairs: readonly BlockPair[]): BlockSend[] {
  const own = (ctx.view.battlefield?.cards ?? []).filter(
    (c) => c.controller === ctx.viewerID && blockedAttackersOf(c).length > 0,
  );
  const key = (p: BlockPair) => `${p.blocker}>${p.attacker}`;
  const dropped = new Set(pairs.map(key));
  const touched = new Set(pairs.map((p) => p.attacker));
  for (const a of touched) {
    const atk = cardOf(ctx.view, a);
    const min = atk?.min_blockers ?? 0;
    if (min < 2) continue;
    const left = own.filter(
      (c) =>
        blockedAttackersOf(c).includes(a) &&
        !dropped.has(key({ blocker: c.instance_id, attacker: a })),
    );
    if (left.length > 0 && left.length < min) {
      for (const c of left) dropped.add(key({ blocker: c.instance_id, attacker: a }));
    }
  }
  const send: BlockSend[] = [];
  const kept: BlockPair[] = [];
  for (const c of own) {
    const atks = blockedAttackersOf(c);
    if (!atks.some((a) => dropped.has(key({ blocker: c.instance_id, attacker: a })))) continue;
    send.push({ type: "remove_from_combat", params: { instance_id: c.instance_id } });
    for (const a of atks) {
      const p = { blocker: c.instance_id, attacker: a };
      if (!dropped.has(key(p))) kept.push(p);
    }
  }
  if (kept.length > 0) send.push({ type: "declare_blockers", params: { blocks: kept } });
  return send;
}

function selectOwn(ctx: BlockContext, state: BlockPickState, id: string): BlockStep {
  if (state.selected === id) return { state: { selected: null, group: null }, send: [] };
  if (!canSelectBlocker(ctx, id)) return { state, send: [] };
  return { state: { selected: id, group: null }, send: [] };
}

// reduceBlockPick is the whole interaction: a state and an event in, the
// next state and what to send out.
export function reduceBlockPick(
  state: BlockPickState,
  ev: BlockEvent,
  ctx: BlockContext,
): BlockStep {
  switch (ev.kind) {
    case "cancel":
      return { state: NO_BLOCK_PICK, send: [] };
    case "own": {
      const g = state.group;
      if (g) {
        if (g.blockers.includes(ev.id)) {
          const blockers = g.blockers.filter((x) => x !== ev.id);
          return {
            state: { selected: null, group: blockers.length > 0 ? { ...g, blockers } : null },
            send: [],
          };
        }
        if (canJoinGroup(ctx, g, ev.id)) {
          const blockers = [...g.blockers, ev.id];
          if (blockers.length >= g.need) {
            return {
              state: NO_BLOCK_PICK,
              send: [
                {
                  type: "declare_blockers",
                  params: { blocks: blockers.map((b) => ({ blocker: b, attacker: g.attacker })) },
                },
              ],
            };
          }
          return { state: { selected: null, group: { ...g, blockers } }, send: [] };
        }
        // A creature that can't join: the group is dropped for it.
        return selectOwn(ctx, NO_BLOCK_PICK, ev.id);
      }
      return selectOwn(ctx, state, ev.id);
    }
    case "attacker": {
      if (!state.selected) return { state, send: [] };
      const blocker = state.selected;
      const t = blockTargets(ctx, blocker).get(ev.id);
      if (!t) return { state, send: [] };
      switch (t.kind) {
        case "remove":
          return {
            state: NO_BLOCK_PICK,
            send: releaseSends(ctx, [{ blocker, attacker: ev.id }]),
          };
        case "block": {
          const b = cardOf(ctx.view, blocker)!;
          return {
            state: hasRoomAfter(b, ev.id) ? state : NO_BLOCK_PICK,
            send: [{ type: "declare_blocker", params: { blocker, attacker: ev.id } }],
          };
        }
        case "group":
          return {
            state: {
              selected: null,
              group: { attacker: ev.id, blockers: [blocker], need: t.need },
            },
            send: [],
          };
        case "refused":
          return { state, send: [], refused: t.reason };
      }
      return { state, send: [] };
    }
    case "remove": {
      const id = state.selected;
      const b = id ? cardOf(ctx.view, id) : undefined;
      if (!id || !b) return { state, send: [] };
      const pairs = blockedAttackersOf(b).map((a) => ({ blocker: id, attacker: a }));
      return { state: NO_BLOCK_PICK, send: pairs.length > 0 ? releaseSends(ctx, pairs) : [] };
    }
  }
}

// liveBlockPick drops whatever a new frame has made stale: a selected
// creature that left or can no longer take part, a group whose attacker
// is gone or whose blockers left. Returns the same object when nothing
// changed.
export function liveBlockPick(state: BlockPickState, ctx: BlockContext): BlockPickState {
  let { selected, group } = state;
  if (selected && !canSelectBlocker(ctx, selected)) selected = null;
  if (group) {
    const atk = cardOf(ctx.view, group.attacker);
    if (!atk || defendingPlayerOf(atk) !== ctx.viewerID) {
      group = null;
    } else {
      const blockers = group.blockers.filter((b) => canSelectBlocker(ctx, b));
      if (blockers.length === 0) group = null;
      else if (blockers.length !== group.blockers.length) group = { ...group, blockers };
    }
  }
  return selected === state.selected && group === state.group ? state : { selected, group };
}

// ---- What the board reads ----

// BlockBoard is what Card, BattlefieldRow and CombatArrows draw while a
// defender declares: the selected creature, what each attacker would do
// on a click, and the group being collected. Game.svelte publishes it;
// null outside a pending block declaration.
export interface BlockBoard {
  viewerID: string;
  selected: string | null;
  targets: ReadonlyMap<string, BlockTarget>;
  group: BlockGroup | null;
}

export const blockBoard = guardedWritable<BlockBoard | null>(null, "blockBoard");

// The piles fanned out into their cards while blockers are declared, by
// group key. BattlefieldRow adds to it; Game.svelte empties it when the
// declaration ends.
export const fannedPiles = guardedWritable<ReadonlySet<string>>(new Set(), "fannedPiles");

export function fanPile(set: ReadonlySet<string>, key: string): ReadonlySet<string> {
  if (set.has(key)) return set;
  const next = new Set(set);
  next.add(key);
  return next;
}

// fansOnClick says whether a click on a pile fans it out instead of
// opening its member list: while blockers are declared, for a pile of
// the viewer's own creatures (one of them may block) or of attackers
// coming at them (one of them may be blocked).
export function fansOnClick(board: BlockBoard | null, members: readonly CardView[]): boolean {
  const m = members[0];
  if (!board || !m) return false;
  if (m.controller === board.viewerID && isCreature(m)) return true;
  return defendingPlayerOf(m) === board.viewerID;
}

// blockTargetFor is the click an incoming attacker (or any member of its
// pile) would make, for Card.
export function blockTargetFor(
  board: BlockBoard | null,
  ids: readonly string[],
): BlockTarget | undefined {
  if (!board) return undefined;
  for (const id of ids) {
    const t = board.targets.get(id);
    if (t) return t;
  }
  return undefined;
}

// inBlockGroup reports whether one of `ids` is in the group being
// collected.
export function inBlockGroup(board: BlockBoard | null, ids: readonly string[]): boolean {
  const g = board?.group;
  return !!g && ids.some((id) => g.blockers.includes(id));
}

// blockHighlight is the part of a card's board highlight that comes from
// the block declaration, so a pile splits where its copies would be
// drawn differently (BattlefieldRow's highlightOf).
export function blockHighlight(board: BlockBoard | null, id: string): string {
  if (!board) return "";
  const t = board.targets.get(id);
  return `${t?.kind ?? ""}${inBlockGroup(board, [id]) ? "g" : ""}`;
}
