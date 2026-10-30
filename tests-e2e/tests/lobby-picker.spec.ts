import { expect, test, type Page } from "@playwright/test";
import { makeCommanderDeck } from "./deck-fixture";
import { adminLogin, createGame, type GameMeta } from "./lobby-api";

// ADR 0212 §3 and owner decisions 7 and 8: the deck picker under a lobby
// table's seats is a grid behind tabs, and the bot picker takes its
// place for Add bot. Choosing a tile previews it in your seat and sends
// nothing; the footer's one primary seats it (or adds the bot). A guest
// has no library, so no My decks tab, and opens on Pre-built.

async function joinInBrowser(
  page: Page,
  game: GameMeta,
  name: string,
): Promise<void> {
  await page.goto(
    `/#/games/${game.id}/join?t=${encodeURIComponent(game.invite_token!)}`,
  );
  await page.getByPlaceholder("your name").fill(name);
  await page.getByRole("button", { name: "join" }).click();
  await expect(page).toHaveURL(/#\/lobby$/);
}

function cardOf(page: Page, name: string) {
  return page.locator("ul.games > li").filter({ hasText: name });
}

test.describe("the deck picker (ADR 0212 §3)", () => {
  test.beforeEach(async ({ page }) => {
    // Tips off, as in lobby.spec: a tip's card can sit over a control.
    await page.context().addInitScript(() => {
      try {
        if (localStorage.getItem("cmdctrl.settings.v1") === null) {
          localStorage.setItem(
            "cmdctrl.settings.v1",
            JSON.stringify({ help: { tipsOff: true } }),
          );
        }
      } catch {
        // storage unavailable: fall back to the defaults
      }
    });
  });

  test("a guest previews a pre-built deck in their seat, then sits down from the footer", async ({
    page,
    request,
  }) => {
    const admin = await adminLogin(request);
    const game = await createGame(request, admin, `Picker Guest ${Date.now()}`);
    await joinInBrowser(page, game, "Gil");
    const card = cardOf(page, game.name);
    const mine = card
      .getByRole("list", { name: "seats" })
      .getByRole("listitem", { name: "seat 1: Gil" });
    const primary = card.locator(".tfoot .btn.primary");

    // No library, so no My decks tab: Pre-built, then Paste a list.
    const picker = card.getByRole("region", { name: "choose your deck" });
    await expect(picker.getByRole("tab")).toHaveText([
      /^Pre-built/,
      "Paste a list",
    ]);
    await expect(
      picker.getByRole("tab", { name: /^Pre-built/ }),
    ).toHaveAttribute("aria-selected", "true");

    // The selected tile is previewed in the seat, not seated.
    const grid = picker.getByRole("radiogroup", { name: "pre-built decks" });
    const radios = grid.locator('input[name="prebuilt-deck"]');
    await expect(radios.first()).toBeChecked();
    await expect(mine).toContainText("Preview");
    await expect(mine).toContainText("Not seated yet");
    await expect(
      card.getByRole("group", { name: "ready 0 of 1" }),
    ).toBeVisible();

    // The arrow keys move the pick, and the seat follows it.
    await radios.first().focus();
    await page.keyboard.press("ArrowRight");
    await expect(radios.nth(1)).toBeChecked();
    const second = (
      await grid.locator("label.dtile").nth(1).locator(".dname").innerText()
    ).trim();
    await expect(mine).toContainText(`Pre-built · ${second}`);

    await expect(primary).toHaveText("Sit down with this deck");
    await primary.click();
    await expect(card.getByRole("group", { name: "ready 1 of 1" })).toBeVisible(
      { timeout: 15_000 },
    );
    await expect(mine).toContainText(second);
    await expect(mine).toContainText("Ready");
    await expect(mine).not.toContainText("Preview");
    await expect(picker).toHaveCount(0);

    // The browser remembers a guest's pick: changing the deck opens the
    // picker on it.
    await page.reload();
    await mine.getByRole("button", { name: "change your deck" }).click();
    await expect(picker).toBeVisible();
    await expect(radios.nth(1)).toBeChecked();
  });

  test("Paste a list uploads a decklist from the footer", async ({
    page,
    request,
  }) => {
    const admin = await adminLogin(request);
    const game = await createGame(request, admin, `Picker Paste ${Date.now()}`);
    await joinInBrowser(page, game, "Pat");
    const card = cardOf(page, game.name);
    const picker = card.getByRole("region", { name: "choose your deck" });
    const primary = card.locator(".tfoot .btn.primary");

    await picker.getByRole("tab", { name: "Paste a list" }).click();
    await expect(
      picker.getByText("Sign in with Discord to keep the decks you paste."),
    ).toBeVisible();
    await expect(primary).toHaveText("Upload deck");
    await expect(primary).toBeDisabled();

    await picker
      .getByRole("textbox", { name: "decklist source" })
      .fill(makeCommanderDeck());
    await expect(primary).toBeEnabled();
    await primary.click();
    await expect(card.getByRole("group", { name: "ready 1 of 1" })).toBeVisible(
      { timeout: 15_000 },
    );
    await expect(
      card
        .getByRole("list", { name: "seats" })
        .getByRole("listitem", { name: "seat 1: Pat" }),
    ).toContainText("Ready");
    // The box is empty again, so it no longer leads: one seated player
    // waits for another.
    await expect(primary).toHaveText("Waiting for players");
  });

  test("Add bot opens the bot picker in the picker's place, and the footer adds the bot", async ({
    page,
    request,
  }) => {
    const admin = await adminLogin(request);
    const game = await createGame(request, admin, `Picker Bot ${Date.now()}`);
    await joinInBrowser(page, game, "Bea");
    const card = cardOf(page, game.name);
    const seats = card.getByRole("list", { name: "seats" });
    const primary = card.locator(".tfoot .btn.primary");
    await expect(
      card.getByRole("region", { name: "choose your deck" }),
    ).toBeVisible();

    await seats
      .getByRole("listitem", { name: "seat 3: open" })
      .getByRole("button", { name: "Add bot" })
      .click();
    // The bot takes the first open seat, and the picker's place.
    const bots = card.getByRole("region", { name: "Add a bot to seat 2" });
    await expect(bots).toBeVisible();
    await expect(
      card.getByRole("region", { name: "choose your deck" }),
    ).toHaveCount(0);
    await expect(
      bots.getByRole("radiogroup", { name: "bot tier" }),
    ).toBeVisible();
    await expect(
      bots.locator('[role="radiogroup"][aria-label="bot tier"] input:checked'),
    ).toHaveCount(1);
    await expect(bots.locator("select")).toHaveCount(0);
    const grid = bots.getByRole("radiogroup", { name: "pre-built decks" });
    await grid.locator('input[name="bot-deck"]').nth(1).check();

    await expect(primary).toHaveText("Add bot");
    await primary.click();
    await expect(bots).toHaveCount(0, { timeout: 15_000 });
    const botSeat = seats.getByRole("listitem", { name: /^seat 2: / });
    await expect(botSeat).not.toHaveAccessibleName("seat 2: open");
    await expect(botSeat).toContainText("Ready");
    await expect(
      card.getByRole("group", { name: "ready 1 of 2" }),
    ).toContainText("waiting on your deck");
    // Your deck is still to choose.
    await expect(primary).toHaveText("Choose your deck");
  });
});
