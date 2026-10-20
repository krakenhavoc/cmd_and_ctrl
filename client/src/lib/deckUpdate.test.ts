import { describe, expect, it } from "vitest";
import {
  commanderLine,
  coverageLine,
  defaultChoice,
  diffCardLabel,
  diffChangeLabel,
  diffSummary,
  requestOptionText,
  requestResultMessage,
  updateChoices,
  updateDoneMessage,
  updateInput,
  type DeckUpdateDiff,
} from "./deckUpdate";
import type { MyDeckInfo } from "./myDecks";

const diff = (over: Partial<DeckUpdateDiff> = {}): DeckUpdateDiff => ({
  added: [],
  removed: [],
  changed: [],
  unchanged: 0,
  commanders_before: ["A"],
  commanders_after: ["A"],
  commander_changed: false,
  ...over,
});

describe("updateChoices", () => {
  it("puts re-pull first, only for a deck with a link", () => {
    expect(
      updateChoices({ source_url: "https://archidekt.com/decks/1" }).map((c) => c.label),
    ).toEqual(["Re-pull from archidekt.com", "New link", "Paste list"]);
    expect(updateChoices({}).map((c) => c.key)).toEqual(["url", "list"]);
    expect(defaultChoice({ source_url: "https://archidekt.com/decks/1" })).toBe("relink");
    expect(defaultChoice({})).toBe("list");
  });
});

describe("updateInput", () => {
  it("builds the body for each choice and refuses an empty field", () => {
    expect(updateInput("relink", "", "")).toEqual({ source: "relink" });
    expect(updateInput("url", "  https://x/y ", "")).toEqual({ url: "https://x/y" });
    expect(updateInput("url", "  ", "")).toBeNull();
    expect(updateInput("list", "", "1 Forest")).toEqual({ list: "1 Forest" });
    expect(updateInput("list", "", " \n")).toBeNull();
  });
});

describe("diff wording", () => {
  it("summarises only what changed", () => {
    expect(diffSummary(diff())).toBe("The cards are the same");
    expect(diffSummary(diff(), true)).toBe("No changes");
    expect(
      diffSummary(
        diff({
          added: [{ name: "a", count: 1 }],
          removed: [
            { name: "b", count: 1 },
            { name: "c", count: 2 },
          ],
          changed: [{ name: "d", from: 1, to: 2 }],
          commander_changed: true,
        }),
      ),
    ).toBe("1 added, 2 removed, 1 changed, commander changed");
  });

  it("labels rows", () => {
    expect(diffCardLabel({ name: "Forest", count: 3 })).toBe("3× Forest");
    expect(diffCardLabel({ name: "Sol Ring", count: 1, sideboard: true })).toBe(
      "Sol Ring (sideboard)",
    );
    expect(diffChangeLabel({ name: "Forest", from: 10, to: 12 })).toBe("Forest 10 → 12");
    expect(commanderLine(diff({ commanders_before: [], commanders_after: ["X", "Y"] }))).toBe(
      "none → X / Y",
    );
  });
});

describe("coverage and request wording", () => {
  it("reads the headline, and nothing when there is none", () => {
    expect(coverageLine(undefined)).toBe("");
    expect(
      coverageLine({ as_printed_copies: 95, deck_size: 100 } as unknown as MyDeckInfo["coverage"]),
    ).toBe("95 of 100 cards play as printed");
  });

  it("words the checkbox and each outcome", () => {
    expect(requestOptionText({ available: true, new_cards: ["a"], already_requested: 0 })).toBe(
      "Request the 1 new missing card",
    );
    expect(
      requestOptionText({ available: true, new_cards: ["a", "b"], already_requested: 3 }),
    ).toBe("Request the 2 new missing cards (3 already requested)");
    expect(requestResultMessage({ status: "filed", new_cards: ["a", "b"] })).toBe(
      "Filed a new issue for 2 new cards.",
    );
    expect(requestResultMessage({ status: "joined", new_cards: ["a"] })).toBe(
      "Added 1 new card to the deck's open issue.",
    );
    expect(requestResultMessage({ status: "rate_limited", retry_after: 7200 })).toContain(
      "2 hours",
    );
    expect(requestResultMessage({ status: "unavailable", message: "No Discord." })).toBe(
      "No Discord.",
    );
    expect(requestResultMessage({ status: "nothing_new" })).toContain("already lists");
  });

  it("puts the request after the diff in the page's line", () => {
    const deck = { id: "d", name: "Atraxa" } as MyDeckInfo;
    expect(updateDoneMessage({ deck, diff: diff({ added: [{ name: "a", count: 1 }] }) })).toBe(
      "Updated Atraxa: 1 added.",
    );
    expect(
      updateDoneMessage({
        deck,
        diff: diff({ added: [{ name: "a", count: 1 }] }),
        request: { status: "filed", new_cards: ["a"] },
      }),
    ).toBe("Updated Atraxa: 1 added. Filed a new issue for 1 new card.");
  });
});
