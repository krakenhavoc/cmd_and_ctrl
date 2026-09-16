import { describe, expect, it } from "vitest";

import { REDACTED, redactSecrets, redactURL } from "./redact";

// A realistic session token: 32 random bytes, base64url (the shape
// server/internal/util/token mints).
const TOKEN = "Zk3q9_Rb-2xVn8LmPq4tYw7cHs1dJe6uKo0aBf5gTiA";
const INVITE = "iNv1te-T0ken_abcdefghijklmnop";

describe("redactSecrets — query parameters", () => {
  it("redacts ?token= in a ws URL and keeps game and player", () => {
    const url = `wss://cmd.example/ws?game=g-1&token=${TOKEN}&player=p-9`;
    expect(redactSecrets(url)).toBe(`wss://cmd.example/ws?game=g-1&token=${REDACTED}&player=p-9`);
  });

  it("redacts token as the first and the last parameter", () => {
    expect(redactSecrets(`/games/1/replay?token=${TOKEN}`)).toBe(
      `/games/1/replay?token=${REDACTED}`,
    );
    expect(redactSecrets(`/ws?token=${TOKEN}#frag`)).toBe(`/ws?token=${REDACTED}#frag`);
  });

  it("redacts the invite and reclaim ?t= in a hash route", () => {
    expect(redactSecrets(`https://h/#/games/abc/join?t=${INVITE}&spectator=1`)).toBe(
      `https://h/#/games/abc/join?t=${REDACTED}&spectator=1`,
    );
    expect(redactSecrets(`#/games/abc/reclaim?t=${INVITE}`)).toBe(
      `#/games/abc/reclaim?t=${REDACTED}`,
    );
  });

  it("redacts access_token, refresh_token, id_token and client_secret", () => {
    const got = redactSecrets(
      `cb?access_token=${TOKEN}&refresh_token=r1r1r1&id_token=i2&client_secret=s3&scope=identify`,
    );
    expect(got).toBe(
      `cb?access_token=${REDACTED}&refresh_token=${REDACTED}&id_token=${REDACTED}&client_secret=${REDACTED}&scope=identify`,
    );
  });

  it("redacts the oauth-complete fragment token", () => {
    expect(redactSecrets(`#/oauth-complete?token=${TOKEN}&expires_at=2026-09-17`)).toBe(
      `#/oauth-complete?token=${REDACTED}&expires_at=2026-09-17`,
    );
  });

  it("redacts a bare token= in prose", () => {
    expect(redactSecrets(`console.error: failed with token=${TOKEN} after retry`)).toBe(
      `console.error: failed with token=${REDACTED} after retry`,
    );
  });

  it("is case-insensitive in the key", () => {
    expect(redactSecrets(`?Token=${TOKEN}&ACCESS_TOKEN=x`)).toBe(
      `?Token=${REDACTED}&ACCESS_TOKEN=${REDACTED}`,
    );
  });

  it("redacts an HTML-escaped &amp;t=", () => {
    expect(redactSecrets(`join?game=1&amp;t=${INVITE}`)).toBe(`join?game=1&amp;t=${REDACTED}`);
  });
});

describe("redactSecrets — URL-encoded variants", () => {
  it("redacts a percent-encoded URL nested in another URL and keeps what follows", () => {
    const nested = encodeURIComponent(`/ws?token=${TOKEN}&game=g-1`);
    expect(nested).toContain("%3Ftoken%3D");
    const got = redactSecrets(`https://h/login?next=${nested}`);
    expect(got).not.toContain(TOKEN);
    expect(got).toBe(`https://h/login?next=%2Fws%3Ftoken%3D${REDACTED}%26game%3Dg-1`);
  });

  it("redacts an encoded invite ?t=", () => {
    const nested = encodeURIComponent(`#/games/abc/join?t=${INVITE}`);
    const got = redactSecrets(`redirect=${nested}`);
    expect(got).not.toContain(INVITE);
    expect(got).toContain(`t%3D${REDACTED}`);
  });

  it("redacts a token value that itself carries percent-escapes", () => {
    expect(redactSecrets("?token=ab%2Bcd%2Fef&game=1")).toBe(`?token=${REDACTED}&game=1`);
  });
});

describe("redactSecrets — headers and JSON", () => {
  it("redacts Authorization: Bearer", () => {
    expect(redactSecrets(`Authorization: Bearer ${TOKEN}`)).toBe(
      `Authorization: Bearer ${REDACTED}`,
    );
    expect(redactSecrets(`{"Authorization":"Bearer ${TOKEN}"}`)).toBe(
      `{"Authorization":"Bearer ${REDACTED}"}`,
    );
    expect(redactSecrets(`Authorization%3A%20Bearer%20${TOKEN}`)).toBe(
      `Authorization%3A%20Bearer%20${REDACTED}`,
    );
  });

  it("redacts JSON token fields, plain and escaped", () => {
    expect(redactSecrets(`{"token":"${TOKEN}","gameID":"g"}`)).toBe(
      `{"token":"${REDACTED}","gameID":"g"}`,
    );
    expect(redactSecrets(`"{\\"token\\": \\"${TOKEN}\\"}"`)).toBe(
      `"{\\"token\\": \\"${REDACTED}\\"}"`,
    );
  });

  it("leaves JSON booleans and card text alone", () => {
    const s = `{"token":true,"name":"Bearer of the Heavens","type_line":"Token Creature"}`;
    expect(redactSecrets(s)).toBe(s);
  });
});

describe("redactSecrets — false positives", () => {
  it("leaves ordinary log lines untouched", () => {
    for (const line of [
      "server error code=bad_request message=not your priority",
      "action create_token id=1234abcd",
      "snapshot seq=12 turn=3 step=draw",
      'chat from=Ada text="at t=3 I attacked"',
      "reconnecting in 500ms (attempt 2)",
      "card art failed to load: id=abc face=0 size=small",
      "wss://cmd.example/ws?game=g-1&player=p-9",
    ]) {
      expect(redactSecrets(line)).toBe(line);
    }
  });

  it("is idempotent", () => {
    const once = redactSecrets(`/ws?token=${TOKEN}&t=${INVITE}`);
    expect(redactSecrets(once)).toBe(once);
  });

  it("copes with empty input", () => {
    expect(redactSecrets("")).toBe("");
  });
});

describe("redactURL", () => {
  it("keeps host, path, game and player; redacts the token", () => {
    expect(redactURL(`ws://localhost:8080/ws?game=g&token=${TOKEN}&player=p`)).toBe(
      `ws://localhost:8080/ws?game=g&token=${REDACTED}&player=p`,
    );
  });

  it("does not throw on a malformed URL", () => {
    expect(redactURL(`::not a url?token=${TOKEN}`)).toBe(`::not a url?token=${REDACTED}`);
  });
});
