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

// stampManaEnforcement returns `params` with the enforcement flags the
// setting asks for. A payload that already says `strict` (the cast
// override toast's force_cast, the auto-tap retry) is left alone.
export function stampManaEnforcement(
  type: string,
  params: Record<string, unknown>,
  strictMana: boolean,
): Record<string, unknown> {
  if (params.strict !== undefined) return params;
  if (type === "cast_spell") return { ...params, strict: strictMana };
  // Only the catalog path (an ability_index) pays a real cost; the
  // free-form "label" announce is a sandbox stack item with nothing to
  // charge.
  if (type === "activate_ability" && strictMana && params.ability_index !== undefined) {
    return { ...params, strict: true, auto_tap: true };
  }
  return params;
}
