import { expect, test } from "@playwright/test";
import { adminLogin, createGame, startGameAs, uploadDeckAs } from "./lobby-api";
import { COMMANDER_NAME } from "./deck-fixture";
import { closeAll, joinAsPlayer, type JoinedPlayer } from "./players";
import {
  adminMoveByName,
  findCardOnBattlefield,
  openAdminClient,
  type AdminClient,
} from "./s19-helpers";

// #4711: the "new" ring of a permanent that arrives on an opponent's
// board (FieldMotion) is drawn in a layer over the board, at a rect
// measured on a snapshot frame, once more after the entrance pop and on
// a board resize. Two things put its tag over a group's label:
//   - the rows re-laid themselves out after the last measure (fitRow
//     settling over a few frames, a row wrapping, the back row moving
//     down as it grew), so the ring stayed where the card had been
//     while the label moved under its tag;
//   - the card was partly or wholly outside what its row shows (scrolled
//     aside, or past the edge of a full land strip), and the ring was
//     drawn whole outside the row, over the neighbouring groups.
// Three seats, so each opponent's panel is a narrow one, seen at
// 1366x768, 1280x720 and 1920x1080. On one opponent
// the type-groups board of #4087, then a land arrives next to the Lands
// group, and while its ring is live:
//   - the ring frames what is on screen of its card: the card's box cut
//     by every ancestor that clips it, within 1px, and no ring at all
//     for a card wholly out of view;
//   - after the opponent's back row moves 20px down with no snapshot and
//     no board resize (as a row that re-lays itself out does), the ring
//     has followed its card;
//   - with each row around the card that scrolls sideways scrolled to
//     either end, the ring again frames what is left of the card on
//     screen;
//   - every time, the tag meets no group label on that panel.

const WALKERS = [
  "Sorin, Vengeful Bloodlord",
  "Sorin, Imperious Bloodlord",
  "Sorin, Solemn Visitor",
  "Sorin the Mirthless",
];
const BOARD = [
  ...WALKERS,
  "Invasion of Ikoria",
  "Sol Ring",
  "Mind Stone",
  "Arcane Signet",
  "Burnished Hart",
  "Rhystic Study",
  "Propaganda",
  "Glorious Anthem",
  "Llanowar Elves",
  "Dryad Arbor",
  "Darksteel Citadel",
];
// One fresh land per size, so each size has a ring of its own.
const SIZES = [
  { width: 1366, height: 768, land: "Forest" },
  { width: 1280, height: 720, land: "Island" },
  { width: 1920, height: 1080, land: "Swamp" },
];

function makeDeck(): string {
  const named = [...BOARD, ...SIZES.map((s) => s.land)];
  return (
    [
      "Commander:",
      `1 ${COMMANDER_NAME}`,
      "",
      "Mainboard:",
      ...named.map((n) => `1 ${n}`),
      `${99 - named.length} Plains`,
    ].join("\n") + "\n"
  );
}

type Box = { x: number; y: number; width: number; height: number };
const overlap = (a: Box, b: Box): number => {
  const w = Math.min(a.x + a.width, b.x + b.width) - Math.max(a.x, b.x);
  const h = Math.min(a.y + a.height, b.y + b.height) - Math.max(a.y, b.y);
  return w > 0 && h > 0 ? w * h : 0;
};

// FieldMotion's ARRIVAL_MS.
const RING_LIFE_MS = 2600;

interface Shot {
  when: string;
  /** ms since the card was first drawn. */
  age: number;
  ring: Box | null;
  tag: Box | null;
  vis: Box;
  labels: { text: string; box: Box }[];
}

test("#4711 an arrival ring keeps to its card and its tag off the group labels", async ({
  browser,
  request,
}) => {
  test.setTimeout(240_000);
  const players: JoinedPlayer[] = [];
  let admin: AdminClient | null = null;
  try {
    const adminToken = await adminLogin(request);
    const game = await createGame(request, adminToken, `Arrival ring 4711 ${Date.now()}`);
    if (!game.invite_token) throw new Error("invite token missing");
    for (let i = 0; i < 3; i++) {
      players.push(await joinAsPlayer(browser, game.id, game.invite_token, `Seat ${i + 1}`));
    }
    for (const p of players) {
      await uploadDeckAs(request, adminToken, game.id, p.playerID, makeDeck());
    }
    await startGameAs(request, adminToken, game.id);
    const [viewer, opp] = players;
    // Each seat keeps its own hand: an admin connection acts only for
    // the seats it is bound to.
    const seatAdmins = await Promise.all(
      players.map((p) => openAdminClient(adminToken, game.id, p.playerID, p.playerID)),
    );
    try {
      for (let i = 0; i < 20; i++) {
        const v = await seatAdmins[0].waitFor(
          (s) => s.mulligans_open !== true || s.seats.some((q) => q.mulligan_turn === true),
          "a seat to decide",
          10_000,
        );
        if (v.mulligans_open !== true) break;
        const who = v.seats.find((q) => q.mulligan_turn === true)!;
        const idx = players.findIndex((p) => p.playerID === who.id);
        await seatAdmins[idx].sendActionAsPlayer(who.id, "keep_hand", {});
        await seatAdmins[0].waitFor(
          (s) =>
            s.mulligans_open !== true ||
            s.seats.find((q) => q.mulligan_turn === true)?.id !== who.id,
          "the next seat",
          10_000,
        );
      }
    } finally {
      seatAdmins.forEach((a) => a.close());
    }
    admin = await openAdminClient(adminToken, game.id, viewer.playerID, opp.playerID);
    await admin.waitFor((v) => v.state === "active", "game state active");
    for (const name of BOARD) {
      await adminMoveByName(admin, opp.playerID, name, "library", "battlefield");
    }
    await admin.waitFor(
      (v) => BOARD.every((n) => findCardOnBattlefield(v, n)),
      "the opponent's board is out",
    );

    const page = viewer.page;
    await expect(
      page.locator(".panel.opponent .row[data-zone] .card[data-instance-id]"),
    ).toHaveCount(BOARD.length, { timeout: 60_000 });

    for (const size of SIZES) {
      const where = `${size.width}x${size.height}`;
      await page.setViewportSize({ width: size.width, height: size.height });
      // The board's own rings have run out and the rows have settled at
      // this size before the land arrives. Rings never come back without
      // a new arrival, so none left is a state that holds.
      await expect(page.locator("[data-field-arrival]")).toHaveCount(0, { timeout: 15_000 });
      await page.waitForTimeout(500);

      await adminMoveByName(admin, opp.playerID, size.land, "library", "battlefield");
      const v = await admin.waitFor(
        (s) => findCardOnBattlefield(s, size.land) !== null,
        `${size.land} is on the battlefield`,
      );
      const id = findCardOnBattlefield(v, size.land)!.instance_id;

      // In the page, so the ring is read well inside its 2.6s life: wait
      // for the card, past its entrance pop (320ms) and the measure
      // after it, read; move the back row down; read again.
      const shots = await page.evaluate(async (cardID: string): Promise<Shot[]> => {
        const frames = (n: number) =>
          new Promise<void>((resolve) => {
            let k = n;
            const step = () => (--k <= 0 ? resolve() : requestAnimationFrame(step));
            requestAnimationFrame(step);
          });
        const card = () =>
          document.querySelector<HTMLElement>(
            `.panel.opponent .row [data-instance-id="${cardID}"]`,
          );
        const t0 = performance.now();
        while (!card() && performance.now() - t0 < 20_000) await frames(1);
        const el = card();
        if (!el) throw new Error("the arriving card was never drawn");
        const t1 = performance.now();
        while (performance.now() - t1 < 450) await frames(1);
        await frames(2);

        const r = (b: DOMRect) => ({ x: b.x, y: b.y, width: b.width, height: b.height });
        const panel = el.closest<HTMLElement>(".panel.opponent")!;
        const read = (): Omit<Shot, "when"> => {
          const age = performance.now() - t1;
          const cb = el.getBoundingClientRect();
          let x0 = cb.left;
          let y0 = cb.top;
          let x1 = cb.right;
          let y1 = cb.bottom;
          for (let a = el.parentElement; a && a !== document.body; a = a.parentElement) {
            const cs = getComputedStyle(a);
            if (cs.overflowX === "visible" && cs.overflowY === "visible") continue;
            const k = a.getBoundingClientRect();
            x0 = Math.max(x0, k.left);
            y0 = Math.max(y0, k.top);
            x1 = Math.min(x1, k.right);
            y1 = Math.min(y1, k.bottom);
          }
          const ring = document.querySelector<HTMLElement>(`[data-field-arrival="${cardID}"]`);
          const tag = ring?.querySelector<HTMLElement>(".tag");
          return {
            age,
            ring: ring ? r(ring.getBoundingClientRect()) : null,
            tag: tag ? r(tag.getBoundingClientRect()) : null,
            vis: { x: x0, y: y0, width: Math.max(0, x1 - x0), height: Math.max(0, y1 - y0) },
            labels: [...panel.querySelectorAll<HTMLElement>(".row-label")].map((l) => ({
              text: l.textContent?.trim() ?? "",
              box: r(l.getBoundingClientRect()),
            })),
          };
        };
        const out: Shot[] = [{ ...read(), when: "as drawn" }];
        const back = panel.querySelector<HTMLElement>(".grid-middle")!;
        back.style.position = "relative";
        back.style.top = "20px";
        await frames(3);
        out.push({ ...read(), when: "after the back row moved" });
        back.style.removeProperty("position");
        back.style.removeProperty("top");
        await frames(2);
        // A row that scrolls sideways, scrolled to each end as a player
        // would with the wheel: the card goes partly or wholly out of view.
        for (let a = el.parentElement; a && a !== panel; a = a.parentElement) {
          if (a.scrollWidth - a.clientWidth < 2) continue;
          const was = a.scrollLeft;
          for (const end of [0, a.scrollWidth]) {
            a.scrollLeft = end;
            await frames(3);
            const edge = end ? "its end" : "its start";
            out.push({ ...read(), when: `with ${a.className.split(" ")[0]} scrolled to ${edge}` });
          }
          a.scrollLeft = was;
          await frames(2);
        }
        return out;
      }, id);

      for (const s of shots) {
        const when = `${where} ${s.when}`;
        if (s.vis.width < 1 || s.vis.height < 1) {
          expect.soft(s.ring, `${when}: a ring for a card out of view`).toBeNull();
          continue;
        }
        // On a slow runner the ring may have run its 2.6s course (counted
        // from the frame, a little before the card was drawn); then there
        // is nothing left to check.
        if (!s.ring && s.age > RING_LIFE_MS - 400) continue;
        expect(s.ring, `${when}: the arriving land is ringed`).not.toBeNull();
        const ring = s.ring!;
        expect.soft(Math.abs(ring.x - s.vis.x), `${when}: ring left`).toBeLessThanOrEqual(1);
        expect.soft(Math.abs(ring.y - s.vis.y), `${when}: ring top`).toBeLessThanOrEqual(1);
        expect
          .soft(Math.abs(ring.x + ring.width - (s.vis.x + s.vis.width)), `${when}: ring right`)
          .toBeLessThanOrEqual(1);
        expect
          .soft(Math.abs(ring.y + ring.height - (s.vis.y + s.vis.height)), `${when}: ring bottom`)
          .toBeLessThanOrEqual(1);
        if (s.tag) {
          // What of the tag paints: the ring clips it to the card.
          const x = Math.max(s.tag.x, ring.x);
          const y = Math.max(s.tag.y, ring.y);
          const painted = {
            x,
            y,
            width: Math.max(0, Math.min(s.tag.x + s.tag.width, ring.x + ring.width) - x),
            height: Math.max(0, Math.min(s.tag.y + s.tag.height, ring.y + ring.height) - y),
          };
          for (const l of s.labels) {
            expect
              .soft(overlap(painted, l.box), `${when}: the tag meets the ${l.text} label`)
              .toBeLessThan(1);
          }
        }
      }
    }
  } finally {
    admin?.close();
    await closeAll(...players);
  }
});
