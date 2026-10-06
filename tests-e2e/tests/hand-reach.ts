import type { Page } from "@playwright/test";

// hand-reach: which cards of the viewer's own hand the pointer cannot
// reach (#2395).
//
// The hand is a fan that overlaps itself and peeks above the bottom
// edge, and the tutorial's coach card, the piles and the commander
// strip sit beside it. A card is reachable when some point of it, on
// screen, is the card itself (or a child of it) rather than something
// laid over it or clipped away. Each card is sampled on a 10×10 grid of
// its on-screen box.
//
// Move the pointer off the board first: a card under the pointer lifts
// the hand over the board, which is a different layout.

/** The names of the hand's cards that show no point the pointer could reach. */
export async function unreachableHandCards(page: Page): Promise<string[]> {
  return page.evaluate(() => {
    const hand = document.querySelector('[aria-label="your hand"]');
    const out: string[] = [];
    for (const card of hand?.querySelectorAll(".hand-slot [data-instance-id]") ?? []) {
      const r = card.getBoundingClientRect();
      let reached = false;
      for (let fy = 0.05; fy < 1 && !reached; fy += 0.1) {
        for (let fx = 0.05; fx < 1 && !reached; fx += 0.1) {
          const x = r.left + r.width * fx;
          const y = r.top + r.height * fy;
          if (x < 0 || y < 0 || x >= innerWidth || y >= innerHeight) continue;
          const hit = document.elementFromPoint(x, y);
          reached = !!hit && card.contains(hit);
        }
      }
      if (!reached) {
        const name = card.getAttribute("aria-label") ?? card.getAttribute("title") ?? "?";
        out.push(`${name} at x ${Math.round(r.left)}..${Math.round(r.right)}`);
      }
    }
    return out;
  });
}

/** How many cards the viewer's own hand draws. */
export async function handCardCount(page: Page): Promise<number> {
  return page.locator('[aria-label="your hand"] .hand-slot [data-instance-id]').count();
}
