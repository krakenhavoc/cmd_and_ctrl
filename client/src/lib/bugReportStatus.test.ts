// #2979: POST /bugreport answers one of three successes. A queued
// report (202) has no issue and the modal must not pretend it does, so
// the client has to carry the server's status and note through intact.
// A server that predates the throttle sends no status and only ever
// filed.

import { afterEach, describe, expect, it, vi } from "vitest";

import { submitBugReport } from "./api";

function reply(status: number, body: unknown): void {
  vi.stubGlobal(
    "fetch",
    vi.fn(async () => new Response(JSON.stringify(body), { status })),
  );
}

afterEach(() => {
  vi.unstubAllGlobals();
});

const draft = { title: "x", description: "" };

describe("submitBugReport status", () => {
  it("passes a queued report through with its note and no issue", async () => {
    reply(202, { status: "queued", report_id: "r1", note: "saved, not filed" });
    const res = await submitBugReport(draft);
    expect(res.status).toBe("queued");
    expect(res.note).toBe("saved, not filed");
    expect(res.url).toBeUndefined();
    expect(res.number).toBeUndefined();
    expect(res.reportID).toBe("r1");
  });

  it("passes a +1 through with the issue it joined", async () => {
    reply(200, { status: "commented", url: "https://example.test/7", number: 7, note: "added" });
    const res = await submitBugReport(draft);
    expect(res.status).toBe("commented");
    expect(res.number).toBe(7);
    expect(res.url).toBe("https://example.test/7");
  });

  it("treats a body with no status as filed", async () => {
    reply(201, { url: "https://example.test/8", number: 8, label: "bug" });
    const res = await submitBugReport(draft);
    expect(res.status).toBe("filed");
    expect(res.label).toBe("bug");
  });
});
