// tokenGroups — identical tokens share one card on the board (#1724).
//
// A token-heavy board (twenty 1/1 Soldiers, a pile of Treasures) used
// to draw every token as its own full card and push the rest of the
// table off the screen. The rows now fold tokens that LOOK the same
// into a group and draw at most two cards for it: one for the
// untapped members and one for the tapped ones, each with a count.
// Clicking a group opens a list of every member (TokenGroupModal),
// which is where one, N or all of them are picked for an action.
//
// The rules below are the whole contract, and each is here for a
// reason the board would otherwise get wrong:
//
//   - Only tokens group. A printed card is never folded away, even two
//     copies of the same basic creature: the land strip's piles are
//     the one existing exception and they keep their own logic.
//   - The key is what the table SEES: controller, name, type line,
//     colours, the token's printed text, and power / toughness with
//     the P/T counters taken back off. Counters stay in the group (the
//     issue's resolved open question: the compact visual wins) and
//     show as a badge on the member's row in the list.
//   - Combat is visible state, so an attacker only groups with tokens
//     attacking the same thing, and a blocker with tokens blocking the
//     same attackers. That keeps every combat arrow's endpoint on a
//     card that really is in that fight.
//   - A token wearing an Equipment or an Aura keeps its group even
//     when the attachment changed its P/T: it joins the group of an
//     unadorned token that otherwise matches. The attachment itself is
//     still drawn behind the group's card, so it never leaves the
//     board, and its name is a badge on the row.
//   - Face-down and phased-out objects never group: the first shows a
//     card back that must not be merged with anything, the second is
//     "treated as though it does not exist" and is drawn inert.
//   - A group needs two members. A lone token draws as itself.
//
// Nothing here hides a choice: every member is still reachable through
// the list, and every hidden member keeps an anchor with its instance
// ID on the group's card (BattlefieldRow), so an arrow or a picker that
// looks a card up by `[data-instance-id]` still finds where it is.

import { blockedAttackersOf } from "./attackTargets";
import type { CardView } from "./protocol";

// isToken is the test the engine makes (game.Card.IsToken): "Token" is
// a supertype on the type line, and the view's type line is the
// effective one, which keeps the supertype.
export function isToken(c: Pick<CardView, "type_line">): boolean {
  return !!c.type_line && /\btoken\b/i.test(c.type_line);
}

// ptCounterDelta is what the P/T counters on a permanent add, for every
// "+N/+M" kind (CR 122.1a) — the client's copy of game.PTCounterDelta,
// so a group key can take counters back off the wire's P/T.
export function ptCounterDelta(counters: Record<string, number> | undefined): {
  power: number;
  toughness: number;
} {
  let power = 0;
  let toughness = 0;
  for (const [name, n] of Object.entries(counters ?? {})) {
    if (!(n > 0)) continue;
    const m = /^([+-])(\d{1,3})\/([+-])(\d{1,3})$/.exec(name);
    if (!m) continue;
    power += (m[1] === "-" ? -1 : 1) * Number(m[2]) * n;
    toughness += (m[3] === "-" ? -1 : 1) * Number(m[4]) * n;
  }
  return { power, toughness };
}

// basePT is the P/T a token shows with its P/T counters taken off:
// the wire's power is clamped at zero, so the signed value comes from
// negative_power when there is one.
function basePT(c: CardView): string {
  if (c.power === undefined && c.toughness === undefined) return "";
  const d = ptCounterDelta(c.counters);
  const signedPower = (c.negative_power ?? 0) < 0 ? c.negative_power! : (c.power ?? 0);
  return `${signedPower - d.power}/${(c.toughness ?? 0) - d.toughness}`;
}

// canGroup reports whether a card may be folded into a group at all.
export function canGroup(c: CardView): boolean {
  if (!isToken(c)) return false;
  if (c.is_commander || c.face_down || c.phased_out || c.known_by_you === false) return false;
  return true;
}

// looseKey is every part of the group key except P/T — the part an
// attachment cannot change.
function looseKey(c: CardView): string {
  const combat = c.attacking_target
    ? `atk:${c.attacking_target_kind ?? "player"}:${c.attacking_target}`
    : blockedAttackersOf(c).length > 0
      ? `blk:${blockedAttackersOf(c).sort().join(",")}`
      : "";
  return [
    c.controller,
    c.name,
    c.type_line ?? "",
    [...(c.colors ?? [])].sort().join(""),
    c.token_text ?? "",
    combat,
  ].join("\u0001");
}

// tokenGroupKey is the key two tokens must share to group, or null for
// a card that never groups. Exported for tests; the board goes through
// groupTokens, which also folds in the attachment rule.
export function tokenGroupKey(c: CardView): string | null {
  if (!canGroup(c)) return null;
  return `${looseKey(c)}\u0001${basePT(c)}`;
}

export interface TokenGroup {
  key: string;
  // Every member, in the order the row was given them (battle_x).
  members: CardView[];
}

// groupTokens assigns each card of a row to its group key. Cards that
// do not group, and groups of one, are absent from the result. The
// attachment rule: a token carrying an Equipment or an Aura takes the
// key of the first unadorned token with the same loose key when its
// own full key has no unadorned member (the attachment changed its
// P/T, which the client cannot subtract).
export function groupTokens(
  cards: readonly CardView[],
  attachmentsByHost: Record<string, readonly CardView[]> = {},
): Map<string, TokenGroup> {
  const adorned = (c: CardView) => (attachmentsByHost[c.instance_id] ?? []).length > 0;
  const plainKeys = new Set<string>();
  const plainByLoose = new Map<string, string>();
  for (const c of cards) {
    const k = tokenGroupKey(c);
    if (k === null || adorned(c)) continue;
    plainKeys.add(k);
    const lk = looseKey(c);
    if (!plainByLoose.has(lk)) plainByLoose.set(lk, k);
  }
  const groups = new Map<string, TokenGroup>();
  for (const c of cards) {
    let k = tokenGroupKey(c);
    if (k === null) continue;
    if (adorned(c) && !plainKeys.has(k)) k = plainByLoose.get(looseKey(c)) ?? k;
    const g = groups.get(k);
    if (g) g.members.push(c);
    else groups.set(k, { key: k, members: [c] });
  }
  for (const [k, g] of groups) if (g.members.length < 2) groups.delete(k);
  return groups;
}

// RowEntry is one thing a battlefield row draws: a card on its own, or
// one half (untapped or tapped) of a token group.
export type RowEntry =
  | { kind: "card"; key: string; card: CardView }
  | {
      kind: "group";
      // Unique per half: the group key plus the half.
      key: string;
      groupKey: string;
      tapped: boolean;
      // The member drawn on the board — the plainest one of this half.
      rep: CardView;
      // This half's members, rep included, in row order.
      members: CardView[];
    };

// plainest picks the member drawn on the group's card: the first with
// no counters, no damage and nothing attached, so the card on the
// board shows what the group has in common rather than one member's
// extras. Falls back to the first member.
function plainest(
  members: readonly CardView[],
  attachmentsByHost: Record<string, readonly CardView[]>,
): CardView {
  return (
    members.find(
      (c) =>
        Object.values(c.counters ?? {}).every((n) => !(n > 0)) &&
        !(c.damage_marked && c.damage_marked > 0) &&
        (attachmentsByHost[c.instance_id] ?? []).length === 0,
    ) ?? members[0]
  );
}

// rowEntries folds a row's cards (already in display order) into what
// the row draws. A group is placed where its first member stood, and
// its two halves sit side by side there — untapped, then tapped — so
// the group does not move when one of its tokens taps or untaps. A
// half with a single member draws as a plain card (it is clicked
// directly, and a "×1" badge would say nothing); a half with two or
// more is a group card.
export function rowEntries(
  cards: readonly CardView[],
  attachmentsByHost: Record<string, readonly CardView[]> = {},
): RowEntry[] {
  const groups = groupTokens(cards, attachmentsByHost);
  const groupOf = new Map<string, TokenGroup>();
  for (const g of groups.values()) for (const m of g.members) groupOf.set(m.instance_id, g);
  const out: RowEntry[] = [];
  const placed = new Set<string>();
  for (const c of cards) {
    const g = groupOf.get(c.instance_id);
    if (!g) {
      out.push({ kind: "card", key: c.instance_id, card: c });
      continue;
    }
    if (placed.has(g.key)) continue;
    placed.add(g.key);
    for (const tapped of [false, true]) {
      const half = g.members.filter((m) => !!m.tapped === tapped);
      if (half.length === 0) continue;
      if (half.length === 1) {
        out.push({ kind: "card", key: half[0].instance_id, card: half[0] });
        continue;
      }
      out.push({
        kind: "group",
        key: `group:${tapped ? "t" : "u"}:${g.key}`,
        groupKey: g.key,
        tapped,
        rep: plainest(half, attachmentsByHost),
        members: half,
      });
    }
  }
  return out;
}

// groupMembersOf resolves a group key back to its live members, for
// the list modal: it is opened from a snapshot and must follow the
// board as tokens arrive, leave or change.
export function groupMembersOf(
  cards: readonly CardView[],
  attachmentsByHost: Record<string, readonly CardView[]>,
  groupKey: string,
): CardView[] {
  return groupTokens(cards, attachmentsByHost).get(groupKey)?.members ?? [];
}

// ---- The list's selection (TokenGroupModal) ----

// GroupListMode is what a selection in the list is FOR, which decides
// the action buttons and which members "select N" and "all" reach for.
//   - target: a targeting prompt is live and some member is legal.
//   - attack: the viewer is declaring attackers with their own group.
//   - block:  the viewer is declaring blockers with their own group.
//   - tap:    no prompt; the viewer may tap and untap these.
//   - look:   none of the above (an opponent's group, a spectator).
export type GroupListMode = "target" | "attack" | "block" | "tap" | "look";

export interface GroupListContext {
  // Whether the targeting prompt accepts this member.
  isLegalTarget?: (c: CardView) => boolean;
  // Why a member can't attack right now, or null when it can.
  attackBlocker?: (c: CardView) => string | null;
}

// bulkCandidates is what "select N" and "all" pick from, in the order
// they pick it. For an attack it is exactly the members that can
// attack now — untapped, not summoning sick, not already attacking —
// so "attack with 5" never spends a slot on a token the server would
// skip. A block reaches for the untapped members; a target prompt for
// the legal ones. Tapping takes untapped first, so "select 3, tap"
// taps three that were untapped.
export function bulkCandidates(
  members: readonly CardView[],
  mode: GroupListMode,
  ctx: GroupListContext = {},
): CardView[] {
  switch (mode) {
    case "target":
      return members.filter((c) => ctx.isLegalTarget?.(c) ?? false);
    case "attack":
      return members.filter(
        (c) => !c.attacking_target && (ctx.attackBlocker?.(c) ?? null) === null,
      );
    case "block":
      return members.filter((c) => !c.tapped);
    case "tap":
      return [...members.filter((c) => !c.tapped), ...members.filter((c) => !!c.tapped)];
    default:
      return [...members];
  }
}

// selectFirstN is "select N": the first n candidates' IDs, clamped to
// what there is.
export function selectFirstN(candidates: readonly CardView[], n: number): string[] {
  const k = Math.max(0, Math.min(Math.floor(n), candidates.length));
  return candidates.slice(0, k).map((c) => c.instance_id);
}

// toggleSelection flips one ID in or out.
export function toggleSelection(selected: readonly string[], id: string): string[] {
  return selected.includes(id) ? selected.filter((x) => x !== id) : [...selected, id];
}

// liveSelection drops IDs that are no longer members (a token died or
// left the group while the list was open), so an action never sends a
// stale one. Returns the same array when nothing changed.
export function liveSelection(selected: string[], members: readonly CardView[]): string[] {
  const ids = new Set(members.map((c) => c.instance_id));
  const kept = selected.filter((id) => ids.has(id));
  return kept.length === selected.length ? selected : kept;
}

// RowBadge is one small label on a member's row: what makes this token
// different from the card drawn on the board.
export interface RowBadge {
  kind: "counter" | "attachment" | "damage" | "sick" | "tapped" | "combat" | "reason";
  text: string;
}

// rowBadges lists a member's differences, in a fixed order: counters,
// attachments, damage, summoning sickness, tapped, combat, and last
// the reason the current action can't use it.
export function rowBadges(
  c: CardView,
  attachments: readonly CardView[] = [],
  reason: string | null = null,
): RowBadge[] {
  const out: RowBadge[] = [];
  for (const [name, n] of Object.entries(c.counters ?? {})) {
    if (!(n > 0)) continue;
    out.push({ kind: "counter", text: n === 1 ? name : `${name} ×${n}` });
  }
  for (const a of attachments) out.push({ kind: "attachment", text: a.name || "attachment" });
  if (c.damage_marked && c.damage_marked > 0) {
    out.push({ kind: "damage", text: `${c.damage_marked} damage` });
  }
  if (c.summoning_sick) out.push({ kind: "sick", text: "summoning sick" });
  if (c.tapped) out.push({ kind: "tapped", text: "tapped" });
  if (c.attacking_target) out.push({ kind: "combat", text: "attacking" });
  else if (blockedAttackersOf(c).length > 0) out.push({ kind: "combat", text: "blocking" });
  if (reason) out.push({ kind: "reason", text: reason });
  return out;
}
