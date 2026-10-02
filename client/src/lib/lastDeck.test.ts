// @vitest-environment jsdom
//
// ADR 0110 §5 items 5 and 6 (Delivery PR 7): the last deck, preselected
// from the account for a signed-in person and from the browser for a
// guest, and the guest's remembered name.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { GUEST_NAME_KEY, loadGuestName, rememberGuestName } from "./guestName";
import {
  GUEST_LAST_DECK_KEY,
  loadLastDeck,
  preselectFor,
  readGuestLastDeck,
  rememberGuestLastDeck,
} from "./lastDeck";
import { setSession, type Session } from "./session";

const USER = "6f9619ff-8b86-d011-b42d-00c04fc964ff";
const NIL = "00000000-0000-0000-0000-000000000000";

function sess(userID: string): Session {
  return {
    token: "tok",
    expiresAt: "2026-12-01T00:00:00Z",
    principal: {
      role: "player",
      user_id: userID,
      issued_at: "2026-10-01T00:00:00Z",
      expires_at: "2026-12-01T00:00:00Z",
    },
    gameID: "g",
    playerID: "p",
  };
}

beforeEach(() => localStorage.clear());
afterEach(() => {
  vi.unstubAllGlobals();
  setSession(null);
  localStorage.clear();
});

describe("preselectFor", () => {
  it("picks the last deck only in a picker of its kind that still lists it", () => {
    const last = { kind: "prebuilt" as const, id: "raid" };
    expect(preselectFor("prebuilt", ["deep", "raid"], last)).toBe("raid");
    expect(preselectFor("library", ["raid"], last)).toBeNull();
    expect(preselectFor("prebuilt", ["deep"], last)).toBeNull();
    expect(preselectFor("prebuilt", ["raid"], null)).toBeNull();
  });
});

describe("a guest's last deck", () => {
  it("is remembered in the browser, pre-built decks only", () => {
    expect(readGuestLastDeck()).toBeNull();
    rememberGuestLastDeck("raid");
    expect(readGuestLastDeck()).toEqual({ kind: "prebuilt", id: "raid" });
    localStorage.setItem(GUEST_LAST_DECK_KEY, JSON.stringify({ kind: "library", id: USER }));
    expect(readGuestLastDeck()).toBeNull();
    localStorage.setItem(GUEST_LAST_DECK_KEY, "{not json");
    expect(readGuestLastDeck()).toBeNull();
  });

  it("is what a guest session loads, without asking the server", async () => {
    const fetchSpy = vi.fn();
    vi.stubGlobal("fetch", fetchSpy);
    rememberGuestLastDeck("deep");
    expect(await loadLastDeck(sess(NIL))).toEqual({ kind: "prebuilt", id: "deep" });
    expect(fetchSpy).not.toHaveBeenCalled();
  });
});

describe("a signed-in person's last deck", () => {
  it("comes from the account, and a failure is no preselection", async () => {
    setSession(sess(USER));
    rememberGuestLastDeck("browser-copy");
    let status = 200;
    vi.stubGlobal(
      "fetch",
      vi.fn(async (url: string) => {
        expect(url).toBe("/me/last-deck");
        return {
          ok: status === 200,
          status,
          statusText: "",
          json: async () => ({ last_deck: { kind: "library", id: USER } }),
          clone() {
            return this;
          },
        } as unknown as Response;
      }),
    );
    expect(await loadLastDeck(sess(USER))).toEqual({ kind: "library", id: USER });
    status = 500;
    expect(await loadLastDeck(sess(USER))).toBeNull();
  });
});

describe("a guest's name", () => {
  it("is remembered, trimmed, and blank is ignored", () => {
    expect(loadGuestName()).toBe("");
    rememberGuestName("  Sam  ");
    expect(loadGuestName()).toBe("Sam");
    rememberGuestName("   ");
    expect(localStorage.getItem(GUEST_NAME_KEY)).toBe("Sam");
    rememberGuestName("x".repeat(60));
    expect(loadGuestName()).toHaveLength(40);
  });
});
