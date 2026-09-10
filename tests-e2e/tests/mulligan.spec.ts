import { expect, test } from "@playwright/test";
import { adminLogin, createGame, uploadDeckAs, startGameAs } from "./lobby-api";
import { makeCommanderDeck } from "./deck-fixture";
import { joinAsPlayer, closeAll, type JoinedPlayer } from "./players";

// mulligan: the keep/mulligan dialog and the roll-call strip beside
// it. full-game.spec.ts only ever clicks "Keep hand"; the mulligan
// half — redraw, the taken-count copy, and the cross-client roll call
// that tells you who is still deciding — was untested, and it is the
// first interactive surface every game puts in front of a player.
//
// Prereqs match full-game.spec.ts: server + Scryfall bulk dump.

test.describe("opening hand", () => {
  let alice: JoinedPlayer | null = null;
  let bob: JoinedPlayer | null = null;

  test.afterEach(async () => {
    await closeAll(alice, bob);
    alice = null;
    bob = null;
  });

  test("mulligan redraws, roll call tracks both seats, keep closes the dialog", async ({
    browser,
    request,
  }) => {
    test.slow();

    const adminToken = await adminLogin(request);
    const game = await createGame(request, adminToken, `Mulligan ${Date.now()}`);
    expect(game.invite_token).toBeTruthy();

    alice = await joinAsPlayer(browser, game.id, game.invite_token!, "Alice");
    bob = await joinAsPlayer(browser, game.id, game.invite_token!, "Bob");

    const deck = makeCommanderDeck();
    await uploadDeckAs(request, adminToken, game.id, alice.playerID, deck);
    await uploadDeckAs(request, adminToken, game.id, bob.playerID, deck);
    await startGameAs(request, adminToken, game.id);

    const aliceDialog = alice.page.getByRole("dialog", { name: /keep or mulligan/i });
    const bobDialog = bob.page.getByRole("dialog", { name: /keep or mulligan/i });

    await expect(aliceDialog).toBeVisible();
    await expect(bobDialog).toBeVisible();

    // The opening hand is shown, not just counted.
    await expect(aliceDialog.getByRole("list", { name: "your opening hand" })).toBeVisible();
    await expect(aliceDialog.getByRole("listitem")).toHaveCount(7);
    await expect(aliceDialog).toContainText("Hand size: 7");

    // The roll call is visible to everyone while the window is open,
    // and starts with both seats undecided.
    const rollCall = alice.page.getByLabel("opening hand decisions");
    await expect(rollCall).toBeVisible();
    await expect(rollCall).toContainText("Alice");
    await expect(rollCall).toContainText("Bob");
    await expect(rollCall.getByText("deciding…")).toHaveCount(2);

    // ---- Alice mulligans. Simplified London: redraw to 7, no
    //      bottom-N penalty yet, and the dialog stays open. ----
    await aliceDialog.getByRole("button", { name: "Mulligan" }).click();
    await expect(aliceDialog).toContainText("Mulligans taken: 1");
    await expect(aliceDialog.getByRole("listitem")).toHaveCount(7);
    // Still undecided — a mulligan is not a keep.
    await expect(rollCall.getByText("deciding…")).toHaveCount(2);

    // ---- Bob keeps. His dialog tears down; Alice's roll call sees
    //      it over the wire without a reload. ----
    await bobDialog.getByRole("button", { name: "Keep hand" }).click();
    await expect(bobDialog).toHaveCount(0);
    await expect(rollCall.getByText("kept")).toHaveCount(1);
    await expect(rollCall.getByText("deciding…")).toHaveCount(1);
    // Alice is still holding the table up, so her dialog stays.
    await expect(aliceDialog).toBeVisible();

    // ---- Alice keeps. The window closes for the table and the
    //      roll call goes away with it. ----
    await aliceDialog.getByRole("button", { name: "Keep hand" }).click();
    await expect(aliceDialog).toHaveCount(0);
    await expect(alice.page.getByLabel("opening hand decisions")).toHaveCount(0);
    await expect(bob.page.getByLabel("opening hand decisions")).toHaveCount(0);

    // The table is live: the active seat can act.
    await expect(alice.page.getByRole("button", { name: "pass turn" })).toBeEnabled();
  });
});
