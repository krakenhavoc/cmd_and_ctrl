// Cross-language drift test for the error-frame tokens (#1533).
//
// protocol.ts promises "the tokens the server sends today" for the
// error codes and the refusal reasons, and nothing held it to that:
// #1507 added `illegal_attack` / `attack_limit` and the block reason
// `declaration_limit` on the server, and the mirror silently went
// without them. Nothing broke — the toast shows `message` verbatim —
// but a client that switches on `code` or `reason` (the attack-tax
// picker, and now the attack-limit picker, both do) would miss a
// refusal it was meant to answer.
//
// Same approach as logKind.test.ts: parse the Go const declarations
// and diff them against the client's runtime lists.

import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

import { describe, expect, it } from "vitest";

import { ATTACK_REFUSAL_REASONS, BLOCK_REFUSAL_REASONS, ErrorCode } from "./protocol";

function goSource(rel: string): string {
  return readFileSync(fileURLToPath(new URL(`../../../server/${rel}`, import.meta.url)), "utf8");
}

const protocolGo = goSource("internal/protocol/protocol.go");
const blockLegalityGo = goSource("internal/game/block_legality.go");

// constValues collects every `<prefix>Name [Type] = "value"` constant.
// The optional type is how `BlockReasonFlying BlockReason = "flying"`
// and `CodeInternal = "internal"` both match one pattern; the prefix
// keeps a quoted string in a doc comment out of it, and the optional
// `const` covers a one-line declaration outside a const block.
function constValues(source: string, prefix: string, typeName = ""): string[] {
  const re = new RegExp(
    `^\\s*(?:const\\s+)?${prefix}\\w+\\s*${typeName}\\s*=\\s*"([a-z_]+)"`,
    "gm",
  );
  const out = new Set<string>();
  let m: RegExpExecArray | null;
  while ((m = re.exec(source)) !== null) out.add(m[1]);
  return [...out].sort();
}

const sorted = (xs: readonly string[]): string[] => [...xs].sort();

describe("error-frame tokens: protocol.ts agrees with the server", () => {
  it("ErrorCode lists exactly the server's Code* constants", () => {
    const server = constValues(protocolGo, "Code");
    // A parse that finds almost nothing is a broken regex, not a
    // server with two codes.
    expect(server.length).toBeGreaterThan(5);
    expect(sorted(Object.values(ErrorCode))).toEqual(server);
  });

  it("carries illegal_attack, the #1507 code", () => {
    expect(ErrorCode.IllegalAttack).toBe("illegal_attack");
  });

  it("BLOCK_REFUSAL_REASONS lists exactly game.BlockReason's values", () => {
    const server = constValues(blockLegalityGo, "BlockReason", "BlockReason");
    expect(server.length).toBeGreaterThan(10);
    expect(sorted(BLOCK_REFUSAL_REASONS)).toEqual(server);
    expect(BLOCK_REFUSAL_REASONS).toContain("declaration_limit");
  });

  it("ATTACK_REFUSAL_REASONS lists exactly the server's AttackRefusal* constants", () => {
    const server = constValues(protocolGo, "AttackRefusal");
    expect(server).toContain("attack_limit");
    expect(sorted(ATTACK_REFUSAL_REASONS)).toEqual(server);
  });
});
