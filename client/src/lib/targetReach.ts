// targetReach.ts — whether a legal player target can be clicked on the
// board (#4435).
//
// Players are picked by clicking their avatar. The action dock, an open
// sheet, a popover and the hand all float over the board, and any of them
// can land on a legal avatar, most easily at the narrow quadrant table
// where the viewer's own avatar stands just above the dock and rises as it
// grows. A legal target nobody can click is a prompt nobody can answer.
//
// So the targeting request offers the legal players as buttons (the
// `playerRow` of lib/targetingDock.ts), and Game.svelte asks this module
// whether it must. The row is not always there because it costs the dock
// a couple of lines of height, and the dock's height is what pushes the
// viewer's own avatar up. It is there from the moment any legal avatar is
// covered until the prompt ends.

// What the board draws for a player who may be picked: the avatar disc
// PlayerIdentity ringed as `cast-targetable` (or `cast-picked`, once in a
// pick list). The same seat can be drawn twice, on the table and in an
// expanded board, and either copy is a way to click them.
export const TARGET_AVATAR_SELECTOR = [
  ".identity.cast-targetable .avatar-wrap[data-seat-id]",
  ".identity.cast-picked .avatar-wrap[data-seat-id]",
].join(", ");

// Which element is on top at a point, or null where nothing is (off the
// screen). `document.elementFromPoint` in a browser.
export type HitTest = (x: number, y: number) => Element | null;

// coveredPlayerTargets lists the seats whose every legal avatar on the
// board is covered at its centre: the top element there is neither the
// avatar nor anything inside it. A seat with one clear copy is reachable,
// and is not listed. An avatar that is not laid out (no size) is skipped:
// it is not on the board at all, as in a jsdom render.
export function coveredPlayerTargets(root: ParentNode, hit: HitTest): string[] {
  const clear = new Set<string>();
  const covered = new Set<string>();
  for (const wrap of root.querySelectorAll<HTMLElement>(TARGET_AVATAR_SELECTOR)) {
    const id = wrap.dataset.seatId;
    if (!id) continue;
    const r = wrap.getBoundingClientRect();
    if (r.width <= 0 || r.height <= 0) continue;
    const top = hit(r.left + r.width / 2, r.top + r.height / 2);
    if (top && wrap.contains(top)) clear.add(id);
    else covered.add(id);
  }
  return [...covered].filter((id) => !clear.has(id));
}
