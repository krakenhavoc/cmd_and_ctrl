import { describe, it, expect, beforeEach } from "vitest";
import { get } from "svelte/store";

import {
  botDeckNames,
  botDeckLabel,
  ensureBotDeckNamesLoaded,
  indexBotDecks,
  resetBotDeckNamesForTests,
  type BotDeckNameMap,
} from "./botDeckNames";
import type { BotDeckInfo, BotOptions } from "./api";

function deck(over: Partial<BotDeckInfo> = {}): BotDeckInfo {
  return { id: "esper-control", name: "Esper Control", ...over };
}

function options(decks: BotDeckInfo[]): BotOptions {
  return { enabled: true, tiers: [], decks };
}

describe("indexBotDecks", () => {
  it("indexes decks by id", () => {
    const map = indexBotDecks([deck(), deck({ id: "gruul-aggro", name: "Gruul Aggro" })]);
    expect(map.get("esper-control")?.name).toBe("Esper Control");
    expect(map.get("gruul-aggro")?.name).toBe("Gruul Aggro");
  });

  it("is empty for undefined or an empty list", () => {
    expect(indexBotDecks(undefined).size).toBe(0);
    expect(indexBotDecks([]).size).toBe(0);
  });

  it("skips an entry with no id rather than clobbering the map at key ''", () => {
    const map = indexBotDecks([deck({ id: "" }), deck()]);
    expect(map.has("")).toBe(false);
    expect(map.size).toBe(1);
  });
});

describe("botDeckLabel", () => {
  const names: BotDeckNameMap = indexBotDecks([deck()]);

  it("resolves a known id to its name", () => {
    expect(botDeckLabel("esper-control", names)).toBe("Esper Control");
  });

  // The fallback #688 asks for vitest coverage on: an id the catalog
  // does not (or does not yet) recognize must still render as
  // something readable, never blank.
  it("falls back to the raw id when the catalog doesn't resolve it", () => {
    expect(botDeckLabel("some-pasted-decklist-id", names)).toBe("some-pasted-decklist-id");
  });

  it("falls back to the raw id before the catalog has loaded at all", () => {
    expect(botDeckLabel("izzet-aggro", new Map())).toBe("izzet-aggro");
  });

  it("falls back to the raw id when the catalog has an entry with a blank name", () => {
    const blankName = indexBotDecks([deck({ name: "   " })]);
    expect(botDeckLabel("esper-control", blankName)).toBe("esper-control");
  });

  it("is empty for a seat with no deck id", () => {
    expect(botDeckLabel(undefined, names)).toBe("");
    expect(botDeckLabel(null, names)).toBe("");
    expect(botDeckLabel("", names)).toBe("");
  });
});

describe("ensureBotDeckNamesLoaded", () => {
  beforeEach(() => resetBotDeckNamesForTests());

  it("populates the store from the fetched options", async () => {
    await ensureBotDeckNamesLoaded(async () =>
      options([deck(), deck({ id: "gruul-aggro", name: "Gruul Aggro" })]),
    );
    const names = get(botDeckNames);
    expect(names.get("esper-control")?.name).toBe("Esper Control");
    expect(names.get("gruul-aggro")?.name).toBe("Gruul Aggro");
  });

  it("fetches once even when called concurrently", async () => {
    let calls = 0;
    const fetcher = async () => {
      calls++;
      return options([deck()]);
    };
    await Promise.all([
      ensureBotDeckNamesLoaded(fetcher),
      ensureBotDeckNamesLoaded(fetcher),
      ensureBotDeckNamesLoaded(fetcher),
    ]);
    expect(calls).toBe(1);
  });

  it("does not re-fetch once the first call has resolved", async () => {
    let calls = 0;
    const fetcher = async () => {
      calls++;
      return options([deck()]);
    };
    await ensureBotDeckNamesLoaded(fetcher);
    await ensureBotDeckNamesLoaded(fetcher);
    expect(calls).toBe(1);
  });

  it("leaves the store empty (and the fallback intact) when the fetch fails", async () => {
    await ensureBotDeckNamesLoaded(async () => {
      throw new Error("network error");
    });
    const names = get(botDeckNames);
    expect(names.size).toBe(0);
    expect(botDeckLabel("esper-control", names)).toBe("esper-control");
  });
});
