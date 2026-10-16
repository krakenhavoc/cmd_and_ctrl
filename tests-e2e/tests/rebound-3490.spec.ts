import { expect, test, type Page } from "@playwright/test";
import { adminLogin, createGame, startGameAs, uploadDeckAs } from "./lobby-api";
import { COMMANDER_NAME } from "./deck-fixture";
import { DOCK_PRIMARY } from "./dock-primary";
import { closeAll, joinAsPlayer, type JoinedPlayer } from "./players";
import {
  adminMoveByName,
  findCardInPlayerGraveyard,
  keepAllHands,
  openAdminClient,
  seedHandWithCard,
  type AdminClient,
  type SnapshotView,
} from "./s19-helpers";

// #3490 / CR 702.88a: rebound. "If you cast this spell from your hand, exile
// it as it resolves. At the beginning of your next upkeep, you may cast this
// card from exile without paying its mana cost."
//
// The player's report: the upkeep prompt came, "Cast it free" was answered,
// and the card stayed in exile. This walks the whole loop through the real
// server and two real browsers, both on the default gameplay settings (smart
// auto-pass on), which is how the report was played. Three things broke it:
//
//   - auto-pass passed 80 ms after the answer, and passing is the decline of a
//     pass-closed free cast (ADR 0099 §4). #3624 holds auto-pass instead.
//   - "Cast it free" opened nothing. The card is in exile before the answer
//     and not castable yet, so the client dropped its cast request as stale
//     on the frame the click was made on (freeCastRequest.ts).
//   - after a turn that ended on a hand-size discard, the upkeep's triggers
//     waited until the draw step to go on the stack (cleanup.go).
//
//   1. Quantum Misalignment is cast from hand at a creature: a token copy, and
//      the card is exiled by rebound instead of going to the graveyard.
//   2. Both players end their turns with the dock's End turn; the second one
//      ends on a discard to hand size.
//   3. The caster's upkeep, still the upkeep: the rebound trigger asks "Cast
//      it free"; the answer opens targeting for the card at once, and nothing
//      passes for the player while it is open.
//   4. Clicking the creature casts it from exile; it resolves into a second
//      token copy and goes to its owner's graveyard, not back to exile.

const MISALIGNMENT = "Quantum Misalignment";
const BEARS = "Grizzly Bears";

function makeDeck(): string {
  return (
    [
      "Commander:",
      `1 ${COMMANDER_NAME}`,
      "",
      "Mainboard:",
      `1 ${MISALIGNMENT}`,
      `1 ${BEARS}`,
      "97 Island",
    ].join("\n") + "\n"
  );
}

function bearsOf(v: SnapshotView, controller: string): number {
  return v.battlefield.cards.filter(
    (c) => c.name === BEARS && c.controller === controller,
  ).length;
}

function onStack(v: SnapshotView, cardID: string): boolean {
  return (v.stack_items ?? []).some((s) => s.source_card_id === cardID);
}

function inExile(v: SnapshotView, cardID: string): boolean {
  return v.exile.cards.some((c) => c.instance_id === cardID);
}

function pageOfPriority(v: SnapshotView, players: JoinedPlayer[]): Page | null {
  const t = v.turn;
  if (!t || t.priority_holder < 0) return null;
  const id = v.seats.find((s) => s.seat === t.priority_holder)?.id;
  return players.find((p) => p.playerID === id)?.page ?? null;
}

// passUntil clicks the dock's primary for whoever holds priority until
// `done` holds. Smart auto-pass passes most of these windows on its own; a
// window a browser holds (its own spell on the stack, say) gets the click a
// player would give it.
async function passUntil(
  admin: AdminClient,
  players: JoinedPlayer[],
  done: (v: SnapshotView) => boolean,
  what: string,
  timeoutMs = 60_000,
): Promise<SnapshotView> {
  const deadline = Date.now() + timeoutMs;
  for (;;) {
    const v = admin.snapshot();
    if (done(v)) return v;
    if (Date.now() > deadline) throw new Error(`passUntil timed out: ${what}`);
    const t = v.turn;
    const page = pageOfPriority(v, players);
    if (page) {
      await page
        .getByRole("button", { name: DOCK_PRIMARY })
        .click({ timeout: 1500 })
        .catch(() => undefined);
    }
    await admin
      .waitFor(
        (n) =>
          done(n) ||
          n.turn?.priority_holder !== t?.priority_holder ||
          n.turn?.step !== t?.step,
        "priority moved",
        2_000,
      )
      .catch(() => undefined);
  }
}

// endTurn presses the active player's End turn and answers a cleanup discard
// if the turn asks for one (CR 514.1), until the next seat's turn begins.
async function endTurn(admin: AdminClient, p: JoinedPlayer): Promise<void> {
  const dock = p.page.getByRole("region", { name: "actions", exact: true });
  const turnOf = (v: SnapshotView) => v.seats[v.turn?.active_seat ?? -1]?.id;
  await expect(dock.getByRole("button", { name: "end turn" })).toBeEnabled({
    timeout: 20_000,
  });
  await dock.getByRole("button", { name: "end turn" }).click();
  const discard = dock.getByRole("dialog", { name: /discard \d+ card/i });
  await expect
    .poll(
      async () => {
        if (await discard.isVisible().catch(() => false)) {
          await discard
            .locator("button.card-pick:not([disabled])")
            .first()
            .click();
          await discard
            .getByRole("button", { name: "Discard", exact: true })
            .click();
        }
        return turnOf(admin.snapshot()) !== p.playerID;
      },
      { timeout: 30_000 },
    )
    .toBe(true);
}

test.describe("#3490 rebound casts its card from exile at the next upkeep", () => {
  test("Quantum Misalignment: cast, exiled, upkeep offer, targeting, cast free, graveyard", async ({
    browser,
    request,
  }, testInfo) => {
    test.setTimeout(240_000);
    let first: JoinedPlayer | undefined;
    let second: JoinedPlayer | undefined;
    let admin: AdminClient | undefined;
    try {
      const adminToken = await adminLogin(request);
      const game = await createGame(
        request,
        adminToken,
        `Rebound 3490 ${Date.now()}`,
      );
      if (!game.invite_token)
        throw new Error("invite token missing on fresh game");
      // The default gameplay settings: smart auto-pass on, as in the report.
      first = await joinAsPlayer(
        browser,
        game.id,
        game.invite_token,
        "Seat One",
      );
      second = await joinAsPlayer(
        browser,
        game.id,
        game.invite_token,
        "Seat Two",
      );
      for (const p of [first, second])
        await p.page.setViewportSize({ width: 1600, height: 900 });
      await uploadDeckAs(
        request,
        adminToken,
        game.id,
        first.playerID,
        makeDeck(),
      );
      await uploadDeckAs(
        request,
        adminToken,
        game.id,
        second.playerID,
        makeDeck(),
      );
      await startGameAs(request, adminToken, game.id);

      admin = await openAdminClient(
        adminToken,
        game.id,
        first.playerID,
        second.playerID,
      );
      const a = admin;
      await keepAllHands(a);
      await a.waitFor((v) => v.state === "active", "game state active");
      await a.waitFor(
        (v) =>
          v.turn?.step === "precombat_main" &&
          v.turn?.priority_holder === v.turn?.active_seat,
        "cursor settled on the caster's precombat main",
        20_000,
      );
      const activeSeat = a.snapshot().turn?.active_seat ?? 0;
      const caster =
        a.snapshot().seats[activeSeat]?.id === first.playerID ? first : second;
      const other = caster === first ? second : first;
      const players = [caster, other];
      const page = caster.page;

      // ---- 1. The cast from hand. ----
      await adminMoveByName(
        a,
        caster.playerID,
        BEARS,
        "library",
        "battlefield",
      );
      const bear = (
        await a.waitFor(
          (v) => bearsOf(v, caster.playerID) === 1,
          "the Bears are out",
        )
      ).battlefield.cards.find((c) => c.name === BEARS)!;
      const card = await seedHandWithCard(a, caster.playerID, MISALIGNMENT);
      // An admin cast (no strict payment): the hand cast is not what #3490
      // is about, and it still counts as cast from a hand (CR 702.88a).
      await a.sendActionAsPlayer(caster.playerID, "cast_spell", {
        instance_id: card.instance_id,
        from_zone: "hand",
        targets: [{ kind: "card", id: bear.instance_id }],
      });
      await a.waitFor(
        (v) => onStack(v, card.instance_id),
        "Quantum Misalignment on the stack",
        10_000,
      );
      await passUntil(
        a,
        players,
        (v) => !onStack(v, card.instance_id),
        "the hand cast resolves",
      );
      const resolved = await a.waitFor(
        (v) =>
          inExile(v, card.instance_id) && bearsOf(v, caster.playerID) === 2,
        "rebound exiled the card and the copy is out",
        10_000,
      );
      expect(
        findCardInPlayerGraveyard(resolved, caster.playerID, MISALIGNMENT),
      ).toBeNull();

      // ---- 2. Round the table to the caster's next upkeep. ----
      await endTurn(a, caster);
      await a.waitFor(
        (v) =>
          v.seats[v.turn?.active_seat ?? -1]?.id === other.playerID &&
          v.turn?.step === "precombat_main",
        "the other seat's main phase",
        60_000,
      );
      await endTurn(a, other);

      // ---- 3. The upkeep offer. ----
      const offerOpen = (v: SnapshotView) =>
        (v.pending_choices ?? []).some(
          (c) => c.kind === "may_cast" && c.chooser === caster.playerID,
        );
      await passUntil(
        a,
        players,
        offerOpen,
        "the rebound offer at the caster's upkeep",
      );
      expect(a.snapshot().turn?.step).toBe("upkeep");
      expect(a.snapshot().seats[a.snapshot().turn?.active_seat ?? -1]?.id).toBe(
        caster.playerID,
      );

      const prompt = page.getByRole("dialog").filter({ hasText: /Rebound/ });
      await expect(prompt).toBeVisible({ timeout: 20_000 });
      await testInfo.attach("the upkeep offer", {
        body: await page.screenshot(),
        contentType: "image/png",
      });
      await prompt.getByRole("button", { name: "Cast it free" }).click();

      // ---- 4. Targeting opens on its own, and nothing passes under it. ----
      const banner = page.getByRole("dialog", {
        name: new RegExp(`Select target for ${MISALIGNMENT}`, "i"),
      });
      await expect(banner).toBeVisible({ timeout: 20_000 });
      // Auto-pass used to pass here within 100 ms. Give it far longer.
      await page.waitForTimeout(2_000);
      const held = a.snapshot();
      expect(held.turn?.step).toBe("upkeep");
      expect(inExile(held, card.instance_id)).toBe(true);
      await testInfo.attach("targeting the rebound cast", {
        body: await page.screenshot(),
        contentType: "image/png",
      });

      const bearTile = page
        .locator(`.card[data-instance-id="${bear.instance_id}"]`)
        .first();
      await bearTile.click();
      await a.waitFor(
        (v) => onStack(v, card.instance_id),
        "the free cast is on the stack",
        20_000,
      );
      expect(a.snapshot().turn?.step).toBe("upkeep");

      await passUntil(
        a,
        players,
        (v) => !onStack(v, card.instance_id),
        "the rebound cast resolves",
      );
      const done = await a.waitFor(
        (v) =>
          findCardInPlayerGraveyard(v, caster.playerID, MISALIGNMENT) !==
            null && bearsOf(v, caster.playerID) === 3,
        "the second copy is out and the card is in the graveyard",
        20_000,
      );
      expect(inExile(done, card.instance_id)).toBe(false);
      await testInfo.attach("after the rebound cast", {
        body: await page.screenshot(),
        contentType: "image/png",
      });
    } finally {
      admin?.close();
      await closeAll(first ?? null, second ?? null);
    }
  });
});
