// stackShares — #4114: what each target of a divided or distributed
// effect gets, said on the board.
//
// Armament Dragon's "distribute three +1/+1 counters among one, two or
// three target creatures you control" and Fireball's "X damage divided
// as you choose" put several targets on the stack with a different
// amount for each. The server has always sent the division as the
// item's `distribution` ({target id: share}); the stack lane now says
// it on each target's phrase (stackLane.ts), and this module is the
// other end: a small badge on the targeted permanent or player, in the
// ring's colour, so the table reads "+2" on the creature that gets two
// counters instead of only a ring and an arrow.
//
// Same shape as the rings (stackArrows.ts planTargetMarks /
// findTargetMarks): which targets get a badge is a decision made here
// and pinned by stackShares.test.ts; StackTargetRings only draws what
// comes back. The badge is positioned from the target's own rect in
// viewport coordinates, so it needs no layer of the board's.

import type { StackLaneItem } from "./stackLane";
import { findAnchor, hasSize } from "./boardAnchor";
import { STACK_SURFACES, TOP_ARROW_COLOR, targetSelector } from "./stackArrows";

/** A badge, before its target is found in the DOM. */
export interface ShareBadgePlan {
  kind: "permanent" | "player";
  id: string;
  /** "+2", "3 dmg", "2". */
  text: string;
  /** The full words, for the badge's title: "2 counters". */
  title: string;
  color: string;
}

/**
 * Every permanent and player a stack item gives a share of a division
 * to, once each. The top item's share wins over a lower item's (it is
 * the one that resolves next), as its ring colour does; otherwise the
 * first item to reach a target.
 */
export function planShareBadges(items: readonly StackLaneItem[]): ShareBadgePlan[] {
  const out = new Map<string, ShareBadgePlan & { fromTop: boolean }>();
  for (const item of items) {
    for (const t of item.targets) {
      if (t.share === null || t.shareBadge === null || t.shareText === null) continue;
      if (t.kind !== "permanent" && t.kind !== "player") continue;
      const key = `${t.kind}:${t.id}`;
      const prev = out.get(key);
      if (prev && (prev.fromTop || !item.isTop)) continue;
      out.set(key, {
        kind: t.kind,
        id: t.id,
        text: t.shareBadge,
        title: `${t.name} gets ${t.shareText} from ${item.name}`,
        color: item.isTop ? TOP_ARROW_COLOR : item.casterColor,
        fromTop: item.isTop,
      });
    }
  }
  return [...out.values()].map(({ fromTop: _fromTop, ...plan }) => plan);
}

/** A badge placed on the screen: the top-right corner of its target. */
export interface ShareBadge extends ShareBadgePlan {
  /** Stable per target. */
  key: string;
  /** Viewport pixels of the target's top-right corner. */
  right: number;
  top: number;
}

/**
 * planShareBadges found on the board: each target's element through
 * boardAnchor (an expanded overlay's copy first, ADR 0120 §3), passing
 * over an element with no size and anything inside the stack's own
 * surfaces. A target that is not on screen gets no badge.
 */
export function measureShareBadges(
  board: ParentNode,
  items: readonly StackLaneItem[],
): ShareBadge[] {
  const out: ShareBadge[] = [];
  const seen = new Set<HTMLElement>();
  for (const plan of planShareBadges(items)) {
    const el = findAnchor(board, targetSelector(plan.kind, plan.id), {
      accept: (e) => hasSize(e) && e.closest(STACK_SURFACES) === null,
    });
    if (!el || seen.has(el)) continue;
    seen.add(el);
    const r = el.getBoundingClientRect();
    out.push({ ...plan, key: `${plan.kind}:${plan.id}`, right: r.right, top: r.top });
  }
  return out;
}
