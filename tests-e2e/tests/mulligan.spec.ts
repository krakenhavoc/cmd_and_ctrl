import { expect, test, type Page } from "@playwright/test";
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
    const game = await createGame(
      request,
      adminToken,
      `Mulligan ${Date.now()}`,
    );
    expect(game.invite_token).toBeTruthy();

    alice = await joinAsPlayer(browser, game.id, game.invite_token!, "Alice");
    bob = await joinAsPlayer(browser, game.id, game.invite_token!, "Bob");

    const deck = makeCommanderDeck();
    await uploadDeckAs(request, adminToken, game.id, alice.playerID, deck);
    await uploadDeckAs(request, adminToken, game.id, bob.playerID, deck);
    await startGameAs(request, adminToken, game.id);

    // ADR 0111 PR 6: the opening hand is a sheet that grows up out of
    // the action dock (region "actions"), and Keep hand / Mulligan are
    // its action bar. The sheet and the bar are one non-modal dialog,
    // still named "keep or mulligan your hand", so every button below
    // is found inside it.
    const dockOf = (page: Page) =>
      page.getByRole("region", { name: "actions", exact: true });
    const aliceDialog = dockOf(alice.page).getByRole("dialog", {
      name: /keep or mulligan/i,
    });
    const bobDialog = dockOf(bob.page).getByRole("dialog", {
      name: /keep or mulligan/i,
    });

    await expect(aliceDialog).toBeVisible();
    await expect(bobDialog).toBeVisible();
    // Not a modal: the table behind it is not blocked.
    await expect(aliceDialog).not.toHaveAttribute("aria-modal", "true");

    // The opening hand is shown, not just counted.
    await expect(
      aliceDialog.getByRole("list", { name: "your opening hand" }),
    ).toBeVisible();
    await expect(aliceDialog.getByRole("listitem")).toHaveCount(7);
    await expect(aliceDialog).toContainText("Hand size: 7");

    // The roll call is visible to everyone while the window is open,
    // and starts with both seats undecided.
    const rollCall = alice.page.getByLabel("opening hand decisions");
    await expect(rollCall).toBeVisible();
    await expect(rollCall).toContainText("Alice");
    await expect(rollCall).toContainText("Bob");
    // CR 103.5: decisions go in turn order, starting seat first, and the
    // starting seat is rolled for, so either player may be first. Only
    // the seat whose turn it is can press; the other sees who it waits
    // for.
    const keepOf = (d: typeof aliceDialog) =>
      d.getByRole("button", { name: "Keep hand" });
    await expect
      .poll(
        async () =>
          (await keepOf(aliceDialog).isEnabled()) ||
          (await keepOf(bobDialog).isEnabled()),
        { timeout: 10_000 },
      )
      .toBe(true);
    const aliceFirst = await keepOf(aliceDialog).isEnabled();
    const [firstName, secondName] = aliceFirst
      ? ["Alice", "Bob"]
      : ["Bob", "Alice"];
    const [firstDialog, secondDialog] = aliceFirst
      ? [aliceDialog, bobDialog]
      : [bobDialog, aliceDialog];
    await expect(keepOf(secondDialog)).toBeDisabled();
    await expect(secondDialog).toContainText(
      `Waiting for ${firstName} to decide`,
    );
    await expect(rollCall.getByText("deciding…")).toHaveCount(1);
    await expect(rollCall.getByText("waiting")).toHaveCount(1);

    // ---- The first seat mulligans. Simplified London: redraw to 7, no
    //      bottom-N penalty yet, and the dialog stays open. ----
    await firstDialog.getByRole("button", { name: "Mulligan" }).click();
    await expect(firstDialog).toContainText("Mulligans taken: 1");
    await expect(firstDialog.getByRole("listitem")).toHaveCount(7);
    // A mulligan is not a keep: the turn passes to the other seat, and
    // the first seat waits for it.
    await expect(keepOf(secondDialog)).toBeEnabled();
    await expect(keepOf(firstDialog)).toBeDisabled();
    await expect(firstDialog).toContainText(
      `Waiting for ${secondName} to decide`,
    );

    // ---- The second seat keeps. Its dialog tears down; the roll call
    //      sees it over the wire without a reload. ----
    await keepOf(secondDialog).click();
    await expect(secondDialog).toHaveCount(0);
    await expect(rollCall.getByText("kept")).toHaveCount(1);
    await expect(rollCall.getByText("deciding…")).toHaveCount(1);
    // Round two is the seat that mulliganed alone, and its dialog stays.
    await expect(firstDialog).toBeVisible();
    await expect(keepOf(firstDialog)).toBeEnabled();

    // ---- The first seat keeps. The window closes for the table and
    //      the roll call goes away with it. ----
    await keepOf(firstDialog).click();
    await expect(firstDialog).toHaveCount(0);
    await expect(alice.page.getByLabel("opening hand decisions")).toHaveCount(
      0,
    );
    await expect(bob.page.getByLabel("opening hand decisions")).toHaveCount(0);

    // The table is live: the active seat can act. The game rolls for
    // the starting player (#1486), so it is either seat's button.
    const passButtons = [alice, bob].map((p) =>
      p.page.getByRole("button", { name: "pass turn" }),
    );
    await expect
      .poll(
        async () =>
          (await passButtons[0].isEnabled()) ||
          (await passButtons[1].isEnabled()),
        {
          timeout: 10_000,
        },
      )
      .toBe(true);
  });
});
