// #519: what the client says when the socket is down.
//
// These are prose assertions, which is unusual for a unit test and
// deliberate here. The bug being fixed is not a wrong value — it is
// that the client said nothing a player could see, and the last time
// it did say something it said it in the vocabulary of a different
// bug. So the copy itself is the behaviour under test.

import { describe, it, expect } from "vitest";

import {
  FREEZE_WORDS,
  OFFLINE_ERROR_CODE,
  STALE_BOARD_SENTENCE,
  actionsDisabled,
  boardIsStale,
  connectionAnnouncement,
  connectionBanner,
  offlineSendMessage,
  usesFreezeVocabulary,
} from "./connectionBanner";
import type { ConnectionStatus } from "./ws";

const ALL_STATUSES: ConnectionStatus[] = [
  "connecting",
  "connected",
  "reconnecting",
  "disconnected",
  "session_ended",
];

describe("boardIsStale", () => {
  it("is true while the client is retrying", () => {
    expect(boardIsStale("reconnecting")).toBe(true);
  });

  it("is true once the client has given up", () => {
    expect(boardIsStale("disconnected")).toBe(true);
  });

  it("is false on a live connection", () => {
    expect(boardIsStale("connected")).toBe(false);
  });

  // The first dial of a fresh mount has no snapshot to be stale about,
  // and the route already renders "waiting for snapshot…" for it. A
  // banner here would fire on every normal page load.
  it("is false on the first connect, which has nothing to be stale about", () => {
    expect(boardIsStale("connecting")).toBe(false);
  });

  // #1475: a dead session is the one stale state that never clears
  // itself, but it is still stale — the board on screen is not live.
  it("is true once the session has ended", () => {
    expect(boardIsStale("session_ended")).toBe(true);
  });
});

describe("actionsDisabled", () => {
  it("is false only on a live connection", () => {
    const live = ALL_STATUSES.filter((s) => !actionsDisabled(s));
    expect(live).toEqual(["connected"]);
  });

  // Broader than boardIsStale: during the opening dial there is no
  // socket to send on either, so a control that looks live would lie
  // even though the board is not yet stale.
  it("covers the opening dial, which boardIsStale does not", () => {
    expect(actionsDisabled("connecting")).toBe(true);
    expect(boardIsStale("connecting")).toBe(false);
  });
});

describe("connectionBanner", () => {
  it("says nothing on a live connection", () => {
    expect(connectionBanner("connected", 0)).toBeNull();
  });

  it("says nothing on the opening dial", () => {
    expect(connectionBanner("connecting", 0)).toBeNull();
  });

  it("names the connection, not the board, while reconnecting", () => {
    const banner = connectionBanner("reconnecting", 1);
    expect(banner?.tone).toBe("retrying");
    expect(banner?.headline).toContain("Connection lost");
  });

  it("carries the attempt count #518 exposes", () => {
    expect(connectionBanner("reconnecting", 3)?.detail).toContain("attempt 3");
  });

  // The count is read straight off a store, and a banner that says
  // "attempt NaN" to a player mid-game is worse than one that says
  // nothing about attempts at all.
  it("drops the attempt count rather than rendering a non-number", () => {
    for (const bad of [0, -1, Number.NaN, Number.POSITIVE_INFINITY]) {
      const detail = connectionBanner("reconnecting", bad)?.detail ?? "";
      expect(detail).toContain("Retrying automatically.");
      expect(detail).not.toContain("attempt");
    }
  });

  it("offers a manual retry in both stale states", () => {
    expect(connectionBanner("reconnecting", 1)?.retryLabel).toBeTruthy();
    expect(connectionBanner("disconnected", 0)?.retryLabel).toBeTruthy();
  });

  it("stops promising a retry once the client has given up", () => {
    const banner = connectionBanner("disconnected", 0);
    expect(banner?.tone).toBe("lost");
    expect(banner?.detail).not.toContain("Retrying automatically");
  });

  it("tells the player what they are looking at in every stale state", () => {
    for (const status of ["reconnecting", "disconnected", "session_ended"] as const) {
      expect(connectionBanner(status, 1)?.detail).toContain(STALE_BOARD_SENTENCE);
    }
  });
});

// #1475: the terminal state a dead session lands in — no retry button,
// because dialling harder cannot fix a revoked or expired credential.
describe("connectionBanner — session_ended", () => {
  it("says the session ended and offers no retry", () => {
    const banner = connectionBanner("session_ended", 0);
    expect(banner?.tone).toBe("ended");
    expect(banner?.headline).toContain("session ended");
    expect(banner?.retryLabel).toBeFalsy();
  });

  it("links a guest to Login", () => {
    const banner = connectionBanner("session_ended", 0, false);
    expect(banner?.link).toEqual({ href: "#/login", label: "Sign in" });
  });

  it("links a signed-in identity to My games instead", () => {
    const banner = connectionBanner("session_ended", 0, true);
    expect(banner?.link).toEqual({ href: "#/my-games", label: "My games" });
  });

  it("defaults to the guest link when signedIn is omitted", () => {
    expect(connectionBanner("session_ended", 0)?.link?.href).toBe("#/login");
  });
});

describe("offlineSendMessage", () => {
  it("names the thing that did not happen", () => {
    expect(offlineSendMessage('action "pass_priority"')).toContain('action "pass_priority"');
  });

  it("says the send failed, not that it is pending", () => {
    expect(offlineSendMessage("chat message")).toContain("was not sent");
  });

  it("blames the connection", () => {
    expect(offlineSendMessage("ping")).toContain("Not connected");
  });

  it("has a code the toast can key off", () => {
    expect(OFFLINE_ERROR_CODE).toBe("not_connected");
  });
});

describe("connectionAnnouncement", () => {
  it("announces every state, including recovery", () => {
    for (const status of ALL_STATUSES) {
      expect(connectionAnnouncement(status, 1).length).toBeGreaterThan(0);
    }
  });

  // The banner announces its own arrival by appearing. Its departure
  // is silent, so recovery is the one transition a screen-reader user
  // would otherwise never be told about.
  it("says the table came back", () => {
    expect(connectionAnnouncement("connected", 0)).toContain("Reconnected");
  });

  it("distinguishes all five states from each other", () => {
    const lines = ALL_STATUSES.map((s) => connectionAnnouncement(s, 1));
    expect(new Set(lines).size).toBe(ALL_STATUSES.length);
  });
});

// #519 acceptance: "deliberately distinct in wording from the #266
// freeze case, so a real freeze is still reportable as one". The two
// failures are indistinguishable to a player and unrelated underneath;
// if this banner borrows #266's words, every genuine freeze arrives
// filed as a network problem.
describe("distinctness from the #266 freeze surface", () => {
  const everything = [
    STALE_BOARD_SENTENCE,
    offlineSendMessage('action "play_land"'),
    ...ALL_STATUSES.map((s) => connectionAnnouncement(s, 2)),
    ...ALL_STATUSES.flatMap((s) => {
      const banner = connectionBanner(s, 2);
      return banner ? [banner.headline, banner.detail, banner.retryLabel] : [];
    }),
  ];

  it("never reaches for #266's vocabulary", () => {
    for (const line of everything) {
      expect({ line, freezeWords: usesFreezeVocabulary(line) }).toEqual({
        line,
        freezeWords: false,
      });
    }
  });

  it("detects that vocabulary when it is there", () => {
    expect(usesFreezeVocabulary("The board is frozen")).toBe(true);
    expect(usesFreezeVocabulary("IT CRASHED")).toBe(true);
    expect(usesFreezeVocabulary("Connection lost — reconnecting")).toBe(false);
  });

  it("guards a non-empty word list", () => {
    expect(FREEZE_WORDS.length).toBeGreaterThan(0);
  });
});
