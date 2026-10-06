import { expect, test, type Locator, type Page } from "@playwright/test";
import { L } from "../../client/src/lib/labels";
import { ADMIN_TOKEN } from "./env";
import { unreachableHandCards } from "./hand-reach";

// tutorial-walk: the practice tutorial, walked end to end (ADR 0125 §8,
// ADR 0076 sub-PR 6, #1085).
//
// Signs in with the admin token, opens #/practice, and walks all
// fourteen steps by doing each step's gesture: Start, Roll (and "I go
// first" when this page wins), Keep hand, rest on the hand, play a
// Forest, rest on the lands, tap one, cast a lit creature, press next,
// right-click, rest on the command zone, press next, autopass on, watch
// the bot's turn, attack, Finish. Each step's own title is asserted as
// it comes up.
//
// It is the regression test for the tutorial's anchors: the coach logs
// `tutorial: step <id> has no anchor on the page; advancing` when a step
// points at something the page does not render, and this spec fails on
// any such line. The labels it selects on come from the contract-label
// registry (client/src/lib/labels.ts), so a rename follows on its own.
//
// It tolerates `cannot happen` lines: the decks are shuffled, so a hand
// may have no castable creature (step 7 gives up, and step 8 with it),
// and a creature that cannot attack makes step 13 give up. The walk
// mulligans a few times for a hand with a Forest and a one-drop, as a
// player might, so most nights every step runs. A step that gives up
// must say so on the console; one that moves on any other way (a
// timeout, a hover it could not see) fails the walk.
//
// The bot's pace is not ours: every wait is on what the page shows (the
// coach's step counter and title, the dock's buttons), never a sleep. A
// coach card the walk cannot move within a budget fails it, naming the
// card and the walk's last attempt. `WALK_DEBUG=1` prints each card and
// every retry.

/** Each step's own title, as the coach shows it (tutorialSteps.ts). */
const STEPS: { n: number; id: string; title: string }[] = [
  { n: 1, id: "welcome", title: "A five-minute practice game" },
  { n: 2, id: "opening-roll", title: "Roll for the first turn" },
  { n: 3, id: "read-hand", title: "Read your hand" },
  { n: 4, id: "play-land", title: "Play a land" },
  { n: 5, id: "land-piles", title: "Lands stack into piles" },
  { n: 6, id: "tap-land", title: "Tap a land for mana" },
  { n: 7, id: "cast-creature", title: "Cast a creature" },
  { n: 8, id: "on-the-stack", title: "Your spell is on the stack" },
  { n: 9, id: "right-click", title: "Abilities live on right-click" },
  { n: 10, id: "commander", title: "Your commander" },
  { n: 11, id: "move-along", title: "Move the turn along" },
  { n: 12, id: "watch-bot", title: "Let the bot play" },
  { n: 13, id: "attack", title: "Attack" },
  { n: 14, id: "handoff", title: "That is the whole interface" },
];
const TOTAL = STEPS.length;

/** A hand card's accessible name ends in what makes it ready (legalActions.ts). */
const PLAYABLE_FOREST = /^Forest, .*playable land/;
const CASTABLE = /, castable/;
/**
 * The practice deck's creatures that one Forest pays for, and that can
 * attack (tutorial.go). A hand with one of them and a Forest can do
 * steps 4 to 13.
 */
const ONE_DROPS = new Set([
  "Llanowar Elves",
  "Elvish Mystic",
  "Fyndhorn Elves",
  "Boreal Druid",
  "Birds of Paradise",
  "Essence Warden",
  "Experiment One",
  "Sazh's Chocobo",
  "Virulent Emissary",
  "Dragon Sniper",
  "Gladecover Scout",
  "Memnite",
  "Ornithopter",
  "Phyrexian Walker",
]);
/** Defenders cannot attack, so step 13 would have nothing to send. */
const DEFENDER = /^(Wall of |Crashing Drawbridge)/;

test("the practice tutorial walks all fourteen steps", async ({ browser }) => {
  // The bot plays a whole turn in the middle of this, at the table's
  // normal pace.
  test.setTimeout(300_000);
  const context = await browser.newContext({
    viewport: { width: 1440, height: 900 },
  });
  // No single gesture waits long; the bot's turn has its own budget.
  context.setDefaultTimeout(20_000);
  const page = await context.newPage();

  const noAnchor: string[] = [];
  const gaveUp: string[] = [];
  const otherwise: string[] = [];
  page.on("console", (msg) => {
    const text = msg.text();
    if (text.includes("has no anchor")) noAnchor.push(text);
    const m = /^tutorial: step (\S+) cannot happen/.exec(text);
    if (m) gaveUp.push(m[1]);
    // A step that timed out, or could not be rested on, moved on without
    // the gesture the walk made for it.
    else if (
      /^tutorial: step .*; advancing$/.test(text) &&
      !text.includes("has no anchor")
    ) {
      otherwise.push(text);
    }
    if (process.env.WALK_DEBUG && /^(tutorial|hint):/.test(text))
      console.log(`[console] ${text}`);
  });

  try {
    await page.goto("/");
    await page.evaluate(() => localStorage.removeItem("cmdctrl.session"));
    // The token form lives on #/admin only (ADR 0112 §2 item 8).
    await page.goto("/#/admin");
    await page.getByPlaceholder("admin token").fill(ADMIN_TOKEN);
    await page.getByRole("button", { name: "log in" }).click();
    await expect(page).toHaveURL(/#\/lobby$/);

    await page.goto("/#/practice");
    await expect(page).toHaveURL(/#\/games\//, { timeout: 20_000 });

    const walk = new Walk(page);
    await walk.run();

    expect(noAnchor, "a tutorial step or a hint pointed at nothing").toEqual(
      [],
    );
    expect(otherwise, "a step moved on without its gesture").toEqual([]);
    // Every step either showed its own title or said why it could not
    // happen; none advanced itself for any other reason.
    for (const s of STEPS) {
      if (walk.shown.has(s.n)) continue;
      expect(
        gaveUp,
        `step ${s.n} (${s.id}) neither showed nor gave up`,
      ).toContain(s.id);
    }
    expect(walk.shown.has(1)).toBe(true);
    expect(walk.shown.has(2)).toBe(true);
    expect(walk.shown.has(TOTAL)).toBe(true);
  } finally {
    await context.close();
  }
});

/** How long the walk lets one coach card stand before it calls the walk stuck. */
function budgetFor(c: Coach): number {
  // The bot's turn, and a turn pressed through, run at the bot's pace.
  if (/bot|turn|Attack/.test(c.title)) return 150_000;
  return 60_000;
}

/** Where visiblePoint looks first, as fractions of what shows: the middle, then outwards. */
const MIDDLE_FIRST = [0.5, 0.35, 0.65, 0.2, 0.8, 0.1, 0.9];
/** Where restOn looks first: the bottom of what shows, then upwards. */
const BOTTOM_FIRST = [0.95, 0.9, 0.85, 0.8, 0.75, 0.7, 0.6, 0.5, 0.4, 0.3, 0.2, 0.1];

/**
 * visiblePoint finds a point of `el`, relative to its box, where the
 * page shows `el` itself (or a child of it) rather than something laid
 * over it, trying the rows of its on-screen part in `rows` order and
 * each row from the middle outwards. Null when none of it shows. Runs
 * in the page.
 */
function visiblePoint(
  el: Element,
  rows: number[],
): { x: number; y: number } | null {
  const r = el.getBoundingClientRect();
  const top = Math.max(r.top, 0);
  const bottom = Math.min(r.bottom, window.innerHeight);
  const left = Math.max(r.left, 0);
  const right = Math.min(r.right, window.innerWidth);
  if (bottom - top < 4 || right - left < 4) return null;
  const steps = [0.5, 0.35, 0.65, 0.2, 0.8, 0.1, 0.9];
  for (const fy of rows) {
    for (const fx of steps) {
      const x = left + (right - left) * fx;
      const y = top + (bottom - top) * fy;
      const hit = document.elementFromPoint(x, y);
      if (hit && el.contains(hit)) return { x: x - r.left, y: y - r.top };
    }
  }
  return null;
}

/**
 * A hand card the pointer cannot reach (#2395). Not a race with the
 * game for the walk to retry: it fails the walk where it happens.
 */
class HiddenCard extends Error {}

function escapeRegExp(s: string): string {
  return s.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

/** What the coach says right now. */
interface Coach {
  /** The step number, or 0 while the card is not up. */
  n: number;
  title: string;
}

class Walk {
  readonly shown = new Set<number>();
  private mulligans = 0;
  private readonly coach: Locator;
  private readonly dock: Locator;
  private readonly me: Locator;
  private readonly hand: Locator;

  constructor(private readonly page: Page) {
    this.coach = page.getByRole("complementary", { name: "tutorial coach" });
    this.dock = page.getByRole("region", { name: L.actions, exact: true });
    this.me = page.getByRole("region", { name: L.yourBoard, exact: true });
    // The hand is a labelled generic element, not a landmark.
    this.hand = page.getByLabel(L.yourHand, { exact: true });
  }

  private async read(): Promise<Coach> {
    const count =
      (await this.coach.locator(".coach-count").textContent()) ?? "";
    const m = /^(\d+) \/ (\d+)$/.exec(count.trim());
    if (!m) return { n: 0, title: "" };
    expect(Number(m[2]), "the coach counts fourteen steps").toBe(TOTAL);
    const title =
      (await this.coach.getByRole("heading", { level: 2 }).textContent()) ?? "";
    return { n: Number(m[1]), title: title.trim() };
  }

  /** Wait until the coach says something other than `c`. */
  private async moved(c: Coach, timeout = 15_000): Promise<void> {
    await expect
      .poll(
        async () => {
          const now = await this.read();
          return now.n !== c.n || now.title !== c.title;
        },
        { timeout },
      )
      .toBe(true);
  }

  private lands(): Locator {
    return this.me.getByRole("list", { name: L.lands, exact: true });
  }

  /**
   * Move the pointer off the board. A card in the hand lifts while the
   * pointer rests on it and covers the rows above, as it would for a
   * player, who also moves the pointer away first.
   */
  private async park(): Promise<void> {
    const size = this.page.viewportSize();
    await this.page.mouse.move((size?.width ?? 1280) / 2, 4);
  }

  /**
   * Play a ready card from the hand: the first of `names`, in order of
   * preference. The hand is a fan that overlaps itself and peeks above
   * the bottom edge, so a card's centre is often not on the card: click
   * a point of it that is.
   *
   * Every card of the hand must show such a point, coach card up or not
   * (#2395: the fan fits its row, so none runs under the coach card, the
   * piles or the commander). One that shows none fails the walk outright,
   * naming it; there is no keyboard fallback to hide it behind.
   */
  private async clickInHand(names: RegExp[]): Promise<void> {
    await this.park();
    const any = this.hand.getByRole("button", { name: names[0] });
    // A card is a button only while this page may play it: wait for one.
    await expect(any.first()).toBeVisible({ timeout: 10_000 });
    // A card dealt a moment ago may still be sliding in: give the fan a
    // moment to settle before calling a card hidden.
    let hidden: string[] = [];
    await expect
      .poll(async () => (hidden = await unreachableHandCards(this.page)), {
        timeout: 5_000,
      })
      .toEqual([])
      .catch(() => {
        throw new HiddenCard(
          `hand cards the pointer cannot reach (#2395): ${hidden.join("; ")}`,
        );
      });
    for (const name of names) {
      const cards = this.hand.getByRole("button", { name });
      for (let i = 0; i < (await cards.count()); i++) {
        const card = cards.nth(i);
        const at = await card.evaluate(visiblePoint, MIDDLE_FIRST);
        if (at) {
          await card.click({ position: at });
          return;
        }
      }
    }
    throw new HiddenCard(
      `no ready card in the hand shows a point to click (#2395): ${names.join(", ")}`,
    );
  }

  /**
   * Rest the pointer on a card of `zone`, the one nearest the middle,
   * on the lowest part of it that shows. A card lifts while the pointer
   * is on it, and its hover holds wherever the pointer rests (#2396):
   * the slot it rose out of keeps the pointer. Near the bottom edge is
   * where it used to lift away from under the pointer, drop back and
   * never be hovered long, so that is where the walk rests.
   */
  private async restOn(zone: Locator): Promise<void> {
    await this.park();
    const cards = zone.locator("[data-instance-id]");
    await expect(cards.first()).toBeVisible();
    const n = await cards.count();
    const order = Array.from({ length: n }, (_, i) => i).sort(
      (a, b) => Math.abs(a - (n - 1) / 2) - Math.abs(b - (n - 1) / 2),
    );
    for (const i of order) {
      const card = cards.nth(i);
      const at = await card.evaluate(visiblePoint, BOTTOM_FIRST);
      if (at) {
        await card.hover({ position: at });
        return;
      }
    }
    throw new Error("no card there shows a point to rest on");
  }

  private next(): Locator {
    return this.dock.getByRole("button", { name: L.next, exact: true });
  }

  /** Press next once, when this page holds priority, and wait for the game to move. */
  private async pressNext(c: Coach): Promise<void> {
    const phase = this.page.getByLabel(L.turnPhase, { exact: true });
    const before = (await phase.textContent()) ?? "";
    const next = this.next();
    await next.click();
    // Done when the coach moves, or when the game has moved on and come
    // back to this page: a changed step label alone can be read before
    // the coach has caught up with it, and a second press then passes a
    // main phase the coach was about to say was here.
    await expect
      .poll(
        async () => {
          const now = await this.read();
          if (now.n !== c.n || now.title !== c.title) return true;
          return (
            ((await phase.textContent()) ?? "") !== before &&
            (await next.isEnabled())
          );
        },
        { timeout: 30_000 },
      )
      .toBe(true);
  }

  async run(): Promise<void> {
    await expect(this.coach).toBeVisible({ timeout: 20_000 });
    let key = "";
    let since = Date.now();
    let lastError = "";
    for (;;) {
      const c = await this.read();
      const step = STEPS.find((s) => s.n === c.n);
      if (step && c.title === step.title && !this.shown.has(c.n)) {
        this.shown.add(c.n);
        test
          .info()
          .annotations.push({ type: "step", description: `${c.n} ${c.title}` });
      }
      if (c.n === TOTAL) {
        await expect(this.coach.getByRole("heading", { level: 2 })).toHaveText(
          STEPS[TOTAL - 1].title,
        );
        await this.coach.getByRole("button", { name: "Finish" }).click();
        await expect(this.coach).toHaveCount(0);
        return;
      }
      // One card that the walk's gestures cannot move is a failure, with
      // what the last attempt said. The bot's turn gets longer.
      const now = `${c.n} / ${TOTAL}: ${c.title}`;
      if (now !== key) {
        key = now;
        since = Date.now();
      } else if (Date.now() - since > budgetFor(c)) {
        throw new Error(
          `the tutorial stayed on "${now}"; last attempt: ${lastError || "none"}`,
        );
      }
      // A gesture can lose a race with the game (a card that moved, a
      // button the bot's pass disabled): the loop reads the coach again
      // and tries what it says then.
      try {
        await this.act(c);
      } catch (err) {
        if (err instanceof HiddenCard) throw err;
        lastError = String(err).split("\n")[0];
        if (process.env.WALK_DEBUG) await this.debugRetry(err);
      }
    }
  }

  /** WALK_DEBUG: what a failed gesture met, the dock and the hand. */
  private async debugRetry(err: unknown): Promise<void> {
    console.log(`[walk]   retry: ${String(err).slice(0, 1500)}`);
    const dock = (await this.dock.innerText().catch(() => "")) ?? "";
    console.log("[dbg] dock:", dock.replace(/\s+/g, " ").slice(0, 300));
    const hand = await this.hand
      .locator("[data-instance-id]")
      .evaluateAll((els) =>
        els.map((e) => {
          const r = e.getBoundingClientRect();
          const hit = document.elementFromPoint(
            r.left + r.width / 2,
            Math.min(r.bottom, innerHeight) - 30,
          );
          const over = hit?.closest("[aria-label]")?.getAttribute("aria-label");
          return `${e.getAttribute("role")}:${e.getAttribute("aria-label")} at ${Math.round(r.left)},${Math.round(r.top)} under ${over ?? "nothing"}`;
        }),
      )
      .catch(() => []);
    console.log("[dbg] hand:", hand);
  }

  /** Do what the coach asks for, once, and wait for the page to move. */
  private async act(c: Coach): Promise<void> {
    if (process.env.WALK_DEBUG)
      console.log(`[walk] ${c.n} / ${TOTAL}: ${c.title}`);
    switch (c.title) {
      case "A five-minute practice game":
        await this.coach.getByRole("button", { name: "Start" }).click();
        return this.moved(c);

      case "Roll for the first turn":
        return this.roll(c);
      case "You won the roll":
        await this.page
          .getByRole("dialog", { name: L.chooseFirstTurn, exact: true })
          .getByRole("button", { name: L.iGoFirst, exact: true })
          .click();
        return this.moved(c);

      case "First, keep your hand":
        return this.keepOrMulligan(c);

      case "Read your hand":
        // Rest on a card in the middle of the hand (only the cards take
        // the pointer, not the strip's empty ends); it lifts, and the
        // coach reads the hand's hover.
        await this.restOn(this.hand);
        return this.moved(c);

      case "Play a land":
        await this.clickInHand([PLAYABLE_FOREST]);
        return this.moved(c);

      case "Lands stack into piles":
        await this.restOn(this.lands());
        return this.moved(c);

      case "Tap a land for mana":
        await this.park();
        await this.lands()
          .getByRole("button", { name: /^Forest/ })
          .first()
          .click();
        return this.moved(c);

      case "Cast a creature":
        return this.cast(c);

      case "Your spell is on the stack":
      case "Move the turn along":
      case "First, your main phase":
      case "First, your turn":
      case "First, leave your main phase":
      case "First, go to combat":
        return this.pressNext(c);

      case "Abilities live on right-click":
        return this.rightClick(c);

      case "Your commander":
        // "<name> command zone, N cards": the stem is the fixed part.
        await this.restOn(this.me.getByLabel(L.commandZone.any.stem).first());
        return this.moved(c);

      case "Let the bot play":
        await this.dock
          .getByRole("button", { name: L.autopass, exact: true })
          .click();
        return this.moved(c);
      case "First, autopass off":
        await this.dock
          .getByRole("button", { name: L.autopass, exact: true })
          .click();
        return this.moved(c);

      // The bot's turn, and any turn this page must press through.
      case "Watch the bot play":
      case "Attacks wait for your turn":
      case "Attack next turn":
        return this.waitOut(c);

      case "Attack":
        return this.attack(c);

      default:
        // A card the walk does not know yet (or the step count with the
        // card between states): wait for it to move rather than guess.
        return this.moved(c);
    }
  }

  /** Step 2: Roll, again after a tie, until the roll is over or this page chooses. */
  private async roll(c: Coach): Promise<void> {
    const roll = this.page
      .getByRole("dialog", { name: L.rollForFirstTurn, exact: true })
      .getByRole("button", { name: L.roll, exact: true });
    await expect
      .poll(
        async () => {
          const now = await this.read();
          if (now.n !== c.n || now.title !== c.title) return true;
          if ((await roll.isVisible()) && (await roll.isEnabled())) {
            await roll.click().catch(() => undefined);
          }
          return false;
        },
        { timeout: 30_000 },
      )
      .toBe(true);
  }

  /**
   * The opening hand. The walk wants a hand that can do every step: a
   * Forest and a creature it can cast off one Forest, so it mulligans
   * (a free redraw at this table) a few times for one, as a player
   * might, then keeps whatever it has. The steps that still cannot
   * happen say so, and the walk tolerates that.
   */
  private async keepOrMulligan(c: Coach): Promise<void> {
    const dialog = this.page.getByRole("dialog", {
      name: L.mulligan,
      exact: true,
    });
    const cards = dialog
      .getByRole("list", { name: "your opening hand" })
      .getByRole("listitem");
    await expect(cards.first()).toBeVisible();
    const names = await cards.evaluateAll((els) =>
      els.map((e) => e.getAttribute("title") ?? ""),
    );
    const good =
      names.some((n) => n.startsWith("Forest")) &&
      names.some((n) => ONE_DROPS.has(n));
    if (!good && this.mulligans < 4) {
      this.mulligans++;
      const before = names.join("|");
      await dialog
        .getByRole("button", { name: "Mulligan", exact: true })
        .click();
      await expect
        .poll(async () =>
          (
            await cards.evaluateAll((els) =>
              els.map((e) => e.getAttribute("title") ?? ""),
            )
          ).join("|"),
        )
        .not.toBe(before);
      return;
    }
    await dialog.getByRole("button", { name: L.keepHand, exact: true }).click();
    return this.moved(c);
  }

  /** Step 7: click a lit creature in the hand, one that can attack later if there is one. */
  private async cast(c: Coach): Promise<void> {
    const lit = this.hand.getByRole("button", { name: CASTABLE });
    await expect
      .poll(async () => {
        const now = await this.read();
        if (now.n !== c.n || now.title !== c.title) return "moved";
        return (await lit.count()) > 0 ? "lit" : "none";
      })
      .not.toBe("none");
    const names = await lit.evaluateAll((els) =>
      els.map((e) => e.getAttribute("aria-label") ?? ""),
    );
    if (names.length === 0) return this.moved(c);
    // One that can attack first, so step 13 has an attacker.
    const order = [
      ...names.filter((n) => !DEFENDER.test(n)),
      ...names.filter((n) => DEFENDER.test(n)),
    ];
    await this.clickInHand(
      [...new Set(order)].map((n) => new RegExp(`^${escapeRegExp(n)}$`)),
    );
    return this.moved(c);
  }

  /** Step 9: right-click the newest creature, else a Forest. */
  private async rightClick(c: Coach): Promise<void> {
    await this.park();
    const creatures = this.me
      .getByRole("list", { name: L.creatures, exact: true })
      .locator("[data-instance-id]");
    const target =
      (await creatures.count()) > 0
        ? creatures.last()
        : this.lands().locator("[data-instance-id]").first();
    await target.click({ button: "right" });
    await this.moved(c);
    // Close the menu it opened, so the next step starts with nothing up.
    await this.page.keyboard.press("Escape");
  }

  /**
   * The bot's turn, or a turn this page has to press through: answer
   * what the game asks of this page (no blocks; next when it holds
   * priority) until the coach moves.
   */
  private async waitOut(c: Coach): Promise<void> {
    const noBlocks = this.dock.getByRole("button", {
      name: "No blocks",
      exact: true,
    });
    const next = this.next();
    await expect
      .poll(
        async () => {
          const now = await this.read();
          if (now.n !== c.n || now.title !== c.title) return true;
          if (await noBlocks.isVisible()) {
            await noBlocks.click({ timeout: 2_000 }).catch(() => undefined);
          } else if ((await next.isVisible()) && (await next.isEnabled())) {
            await next.click({ timeout: 2_000 }).catch(() => undefined);
          }
          return false;
        },
        { timeout: 120_000 },
      )
      .toBe(true);
  }

  /** Step 13: send every creature that can attack at the bot. */
  private async attack(c: Coach): Promise<void> {
    const all = this.dock
      .getByRole("group", { name: L.declareAttackers, exact: true })
      .getByRole("button", { name: /^Attack .* with all/ });
    await all.click();
    return this.moved(c);
  }
}
