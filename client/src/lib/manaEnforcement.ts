// manaEnforcement.ts — the viewer's `gameplay.strictMana` setting,
// stamped onto the actions that spend mana (S15, widened by #1296).
//
// The server gates a cast or an activation's mana on two payload flags
// (game.CastSpellParams / ActivateAbilityParams Strict + AutoTap).
// With NEITHER set it takes the sandbox posture: spend what the pool
// covers and, if it does not cover the cost, waive the charge and mark
// it paid "on paper". That posture is the whole of strictMana = off.
//
// #1296 ("Equip costs were not paid"): the setting was only ever
// stamped on cast_spell, so every activated ability a player clicked —
// an equip, a Channel, a sac outlet — took the paper path whatever the
// setting said, and an empty pool made it free. An activation is
// stamped now too, and with auto_tap as well as strict: casts have an
// "Auto-tap & cast" override toast to fall back on, activations have no
// such retry, so the engine taps the lands itself (as it already does
// for every bot activation, special action and attack tax) and refuses
// only when the board cannot pay.
//
// ADR 0118 §1 (#2188): strict is the default, and a clicked cast is
// stamped `auto_tap` too, as an activation and a drag already are. A
// click on a card the board can pay for taps the lands and casts it in
// one click; the server spends what is floating first and plans only
// what the pool is missing (ADR 0118 PR 3). With the setting off, a
// cast is stamped `strict: false` exactly as before.

// stampManaEnforcement returns `params` with the enforcement flags the
// setting asks for. A payload that already says `strict` (Cast anyway's
// force_cast, the auto-tap retry, a drag) is left alone.
export function stampManaEnforcement(
  type: string,
  params: Record<string, unknown>,
  strictMana: boolean,
): Record<string, unknown> {
  if (params.strict !== undefined) return params;
  if (type === "cast_spell") {
    return strictMana ? { ...params, strict: true, auto_tap: true } : { ...params, strict: false };
  }
  // Only the catalog path (an ability_index) pays a real cost; the
  // free-form "label" announce is a sandbox stack item with nothing to
  // charge.
  if (type === "activate_ability" && strictMana && params.ability_index !== undefined) {
    return { ...params, strict: true, auto_tap: true };
  }
  // #2215: a mana ability whose cost has a mana component (Crystal
  // Quarry's "{5}, {T}", a Signet's "{1}, {T}") pays it for real in
  // every mode — there is no paper path — so without auto_tap the only
  // answer to an empty pool is an error. With it the server taps the
  // player's other sources for what the pool is missing. Stamped
  // whatever the setting says, and inert on an ability with no mana
  // component.
  if (type === "activate_mana_ability") {
    return { ...params, auto_tap: true };
  }
  return params;
}
