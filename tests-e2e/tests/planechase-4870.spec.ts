import { expect, test, type Page } from "@playwright/test";
import { mkdirSync } from "node:fs";
import path from "node:path";
import { L } from "../../client/src/lib/labels";
import {
  adminLogin,
  joinViaAPI,
  startGameAs,
  uploadDeckAs,
  type GameMeta,
} from "./lobby-api";
import { makeCommanderDeck } from "./deck-fixture";
import { CLIENT_ORIGIN } from "./env";
import { closeAll, joinAsPlayer, type JoinedPlayer } from "./players";
import {
  openAdminClient,
  type AdminClient,
  type SnapshotView,
} from "./s19-helpers";

// #4870, ADR 0216 slice 3: a Planechase table, end to end. A table made
// with the option on, four seats, the first plane face up beside the
// table, the dock's planar die rolled on the viewer's turns until it
// shows the planeswalk face, and the plane changing when the planeswalk
// resolves, with its log lines.
//
// The option is offered only where the server offers it (owner answer
// 3: cmd-dev until the plane batches land): against a server whose
// /config does not list `planechase`, which is every prod-env run, the
// spec skips rather than fails. Run it against a dev server
// (CMDCTRL_ENV=dev) with the card database loaded.
//
// SHOT_DIR writes the 820px and 1440px screenshots ADR 0216 §10 asks
// for (the landscape plane must not cover the dock or the stack).

const SHOT_DIR = process.env.SHOT_DIR ?? "";

// The views the walk reads that SnapshotView does not declare.
interface PlanarView extends SnapshotView {
  planechase?: {
    face_up: { instance_id: string; name: string }[];
    deck_count: number;
  };
  log?: {
    seq: number;
    kind: string;
    planar_face?: string;
    text: string;
    seat: number;
  }[];
}

const planar = (v: SnapshotView): PlanarView => v as PlanarView;

async function shoot(page: Page, name: string): Promise<void> {
  if (!SHOT_DIR) return;
  mkdirSync(SHOT_DIR, { recursive: true });
  for (const width of [820, 1440]) {
    await page.setViewportSize({ width, height: width === 820 ? 1180 : 900 });
    await page.waitForTimeout(400);
    await page.screenshot({
      path: path.join(SHOT_DIR, `${name}-${width}.png`),
    });
  }
  await page.setViewportSize({ width: 1280, height: 720 });
}

// Seats is the table's admin sockets: one AdminClient binds two seats,
// so a four-seat table takes two. `view` is the first client's merged
// view (the turn, the stack, the log and the planar zone are public);
// `act` sends as any seat through the client bound to it.
interface Seats {
  view: AdminClient;
  act: (playerID: string, type: string) => Promise<unknown>;
}

async function keepAll(t: Seats): Promise<void> {
  for (let i = 0; i < 16; i++) {
    const v = await t.view.waitFor(
      (s) =>
        s.mulligans_open !== true ||
        s.seats.some((p) => p.mulligan_turn === true),
      "a seat to decide on its opening hand",
      10_000,
    );
    if (v.mulligans_open !== true) return;
    const who = v.seats.find((p) => p.mulligan_turn === true)!;
    await t.act(who.id, "keep_hand");
    await t.view.waitFor(
      (s) =>
        s.mulligans_open !== true ||
        s.seats.find((p) => p.mulligan_turn === true)?.id !== who.id,
      "the next seat to decide",
      10_000,
    );
  }
  throw new Error("the opening hands were not all kept");
}

// toViewersMain passes every other seat's turn, and every priority
// window on the way, until the viewer holds priority in their own
// precombat main with an empty stack. The admin may act for any seat.
async function toViewersMain(seats: Seats, viewerID: string): Promise<void> {
  const admin = seats.view;
  for (let i = 0; i < 120; i++) {
    const v = admin.snapshot();
    const t = v.turn;
    const viewer = v.seats.find((s) => s.id === viewerID)!;
    const active = v.seats.find((s) => s.seat === t?.active_seat);
    const holder = v.seats.find((s) => s.seat === t?.priority_holder);
    const stackEmpty = (v.stack_items ?? []).length === 0;
    if (
      t?.active_seat === viewer.seat &&
      t.step === "precombat_main" &&
      t.priority_holder === viewer.seat &&
      stackEmpty
    ) {
      return;
    }
    const before = JSON.stringify([
      t?.active_seat,
      t?.step,
      t?.priority_holder,
      v.stack_items?.length,
    ]);
    // The viewer's own client passes its upkeep and draw by itself; acting
    // for it as well would race it past its main phase. A refusal is a
    // stale read: read again.
    if (
      active &&
      active.id !== viewerID &&
      t?.step === "precombat_main" &&
      stackEmpty
    ) {
      await seats.act(active.id, "pass_turn").catch(() => undefined);
    } else if (holder && holder.id !== viewerID) {
      await seats.act(holder.id, "pass_priority").catch(() => undefined);
    }
    await admin
      .waitFor(
        (nv) =>
          JSON.stringify([
            nv.turn?.active_seat,
            nv.turn?.step,
            nv.turn?.priority_holder,
            nv.stack_items?.length,
          ]) !== before,
        "the turn moves on",
        5_000,
      )
      .catch(() => undefined);
  }
  throw new Error("never reached the viewer's precombat main");
}

test.describe("Planechase table (#4870)", () => {
  test.setTimeout(240_000);

  test("create a Planechase table, roll the planar die, see a planeswalk", async ({
    browser,
    request,
    playwright,
  }) => {
    const cfg = (await (await request.get("/config")).json()) as {
      features?: { planechase?: boolean };
    };
    test.skip(
      cfg.features?.planechase !== true,
      "this server does not offer Planechase",
    );

    const adminToken = await adminLogin(request);
    const res = await request.post("/games", {
      data: { name: "Planechase 4870", planechase: true },
      headers: { Authorization: `Bearer ${adminToken}` },
    });
    expect(res.ok(), await res.text()).toBe(true);
    const meta = (await res.json()) as GameMeta;
    const invite = meta.invite_token!;

    let alice: JoinedPlayer | null = null;
    let admin: AdminClient | null = null;
    let admin2: AdminClient | null = null;
    try {
      alice = await joinAsPlayer(browser, meta.id, invite, "Alice");
      const others = [];
      for (const name of ["Bob", "Carol", "Dave"]) {
        // A context of its own: the join's session cookie would
        // otherwise outrank the admin's bearer token on `request`.
        const ctx = await playwright.request.newContext({
          baseURL: CLIENT_ORIGIN,
        });
        others.push(await joinViaAPI(ctx, meta.id, invite, name));
        await ctx.dispose();
      }
      const deck = makeCommanderDeck();
      for (const id of [alice.playerID, ...others.map((o) => o.player_id)]) {
        await uploadDeckAs(request, adminToken, meta.id, id, deck);
      }
      const settings = await request.get(`/games/${meta.id}/settings`, {
        headers: { Authorization: `Bearer ${adminToken}` },
      });
      expect(
        ((await settings.json()) as { settings: { planechase?: boolean } })
          .settings.planechase,
      ).toBe(true);

      await startGameAs(request, adminToken, meta.id);
      admin = await openAdminClient(
        adminToken,
        meta.id,
        alice.playerID,
        others[0].player_id,
      );
      admin2 = await openAdminClient(
        adminToken,
        meta.id,
        others[1].player_id,
        others[2].player_id,
      );
      const first2 = new Set([alice.playerID, others[0].player_id]);
      const a1 = admin;
      const a2 = admin2;
      const seats: Seats = {
        view: a1,
        act: (id, type) =>
          (first2.has(id) ? a1 : a2).sendActionAsPlayer(id, type, {}),
      };
      await keepAll(seats);
      const page = alice.page;

      // CR 901.5: the first plane is face up once every hand is kept.
      const zone = page.getByRole("group", { name: L.planarZone, exact: true });
      const first = await admin.waitFor(
        (v) => (planar(v).planechase?.face_up.length ?? 0) > 0,
        "the first plane face up",
        15_000,
      );
      const firstPlane = planar(first).planechase!.face_up[0].name;
      await expect(zone).toBeVisible();
      await expect(zone).toContainText(firstPlane);
      await expect(
        zone.getByRole("status", { name: L.planarDeck.any.stem, exact: false }),
      ).toBeVisible();

      const roll = page.getByRole("button", {
        name: L.rollPlanarDie,
        exact: true,
      });
      await expect(roll).toBeVisible();

      // Roll on each of Alice's turns until the die shows the planeswalk
      // face: 1 in 6 a roll, so 40 turns miss with p < 0.001.
      let shot = false;
      for (let turn = 0; turn < 40; turn++) {
        await toViewersMain(seats, alice.playerID);
        await expect(roll).toBeEnabled({ timeout: 10_000 });
        const before = Math.max(
          0,
          ...(planar(admin.snapshot()).log ?? []).map((e) => e.seq),
        );
        await roll.click();
        const rolled = await admin.waitFor(
          (v) =>
            (planar(v).log ?? []).some(
              (e) => e.seq > before && e.kind === "roll" && e.planar_face,
            ),
          "the planar die's log line",
          10_000,
        );
        const line = (planar(rolled).log ?? []).find(
          (e) => e.seq > before && e.kind === "roll" && e.planar_face,
        )!;
        if (!shot) {
          // The die is in the air beside Alice's seat: the table at both widths.
          await shoot(page, "planechase-roll");
          shot = true;
        }
        if (line.planar_face !== "planeswalk") {
          await seats.act(alice.playerID, "pass_turn");
          continue;
        }
        // The planeswalking ability is on the stack (CR 901.8): the table
        // with the stack pile up, then resolve it.
        await admin.waitFor(
          (v) => (v.stack_items ?? []).length > 0,
          "the planeswalk trigger",
          10_000,
        );
        await page.waitForTimeout(600);
        await shoot(page, "planechase-stack");
        const walked = await (async () => {
          for (let i = 0; i < 20; i++) {
            const v = planar(admin!.snapshot());
            const left = v.planechase?.face_up[0]?.name;
            if (
              left &&
              left !== firstPlane &&
              (v.stack_items ?? []).length === 0
            )
              return v;
            const holder = v.seats.find(
              (s) => s.seat === v.turn?.priority_holder,
            );
            if (holder) await seats.act(holder.id, "pass_priority");
            await page.waitForTimeout(150);
          }
          throw new Error("the planeswalk did not resolve");
        })();
        const next = walked.planechase!.face_up[0].name;
        await expect(zone).toContainText(next);
        await expect(zone).not.toContainText(firstPlane);
        await shoot(page, "planechase-planeswalked");
        // The log names both: away from the old plane, to the new one.
        expect(
          (walked.log ?? []).some(
            (e) => e.kind === "planeswalk" && e.text.includes(next),
          ),
        ).toBe(true);
        return;
      }
      throw new Error(
        "the planar die never showed the planeswalk face in 40 turns",
      );
    } finally {
      admin?.close();
      admin2?.close();
      await closeAll(alice);
    }
  });
});
