// Cross-language drift test (#1256).
//
// protocol.ts's `LogKind` union is hand-maintained against
// server/internal/protocol/log.go's `LogKind` const block, and
// nothing enforced the two staying in sync: a kind missing from the
// client union is also missing from gameLog.ts's `LOG_TONE` map
// without the typechecker noticing, because a `Record<LogKind, string>`
// is only exhaustive over the union it's given — one that is too
// small compiles fine. #1256 found four kinds (`transform`,
// `phase_out`, `phase_in`, `turn_face_down`) plus `choose_name`
// missing this way; each rendered with the "tone-quiet" fallback
// instead of failing a build.
//
// This test parses the server's own const block and diffs it against
// `ALL_LOG_KINDS` — gameLog.ts's runtime proxy for the client union,
// see the comment there — so a future kind added to one side and not
// the other fails here instead of rendering silently wrong.

import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

import { describe, expect, it } from "vitest";

import { ALL_LOG_KINDS } from "./gameLog";

const logGoURL = new URL("../../../server/internal/protocol/log.go", import.meta.url);
const logGoSource = readFileSync(fileURLToPath(logGoURL), "utf8");

// serverLogKinds extracts every `SomeName LogKind = "kind"` constant
// declared inside the LogKind const block. Scoped to the block
// bounded by `type LogKind string` and its closing `)` so a quoted
// string sitting in a doc comment elsewhere in the file (there are
// several — "Foretell {2}", card names, etc.) can't be picked up by
// accident.
function serverLogKinds(source: string): string[] {
  const typeIdx = source.indexOf("type LogKind string");
  if (typeIdx === -1) {
    throw new Error("could not find `type LogKind string` in log.go — has it moved or renamed?");
  }
  const constStart = source.indexOf("const (", typeIdx);
  if (constStart === -1) {
    throw new Error("could not find the `const (` block following `type LogKind string`");
  }
  const rest = source.slice(constStart);
  const end = rest.match(/\n\)\n/);
  if (!end || end.index === undefined) {
    throw new Error("could not find the closing `)` of the LogKind const block");
  }
  const block = rest.slice(0, end.index);

  const kinds = new Set<string>();
  const re = /\bLog\w+\s+LogKind\s*=\s*"([a-z_]+)"/g;
  let m: RegExpExecArray | null;
  while ((m = re.exec(block)) !== null) {
    kinds.add(m[1]);
  }
  return [...kinds].sort();
}

describe("LogKind: client union agrees with server/internal/protocol/log.go", () => {
  it("declares exactly the kinds the server does today", () => {
    const serverKinds = serverLogKinds(logGoSource);
    // A parse that finds almost nothing means the regex broke against
    // a log.go reformat, not that the server suddenly has three
    // kinds — fail loudly rather than reporting a false "missing 33".
    expect(serverKinds.length).toBeGreaterThan(20);

    // Widened to string[]: the comparison below is a plain set diff
    // against server-parsed strings, and LogKind's whole point is
    // that not every string belongs to it.
    const clientKinds: string[] = [...ALL_LOG_KINDS].sort();

    const missingFromClient = serverKinds.filter((k) => !clientKinds.includes(k));
    const extraInClient = clientKinds.filter((k) => !serverKinds.includes(k));

    expect(missingFromClient, "server LogKind values missing from client's LogKind union").toEqual(
      [],
    );
    expect(
      extraInClient,
      "client LogKind values the server no longer declares (renamed or removed?)",
    ).toEqual([]);
  });
});
