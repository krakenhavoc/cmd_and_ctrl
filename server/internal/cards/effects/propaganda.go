package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Propaganda — Enchantment {2}{U}:
//
//	"Creatures can't attack you unless their controller pays {2} for
//	 each creature they control that's attacking you."
//
// The card #1063 was opened for, and the one worth reading first in
// the attack-tax family: one flat price, no predicate, no count.
//
// "Can't attack YOU", and only you. An attack on a planeswalker its
// controller controls is FREE — Propaganda does not print the "or
// planeswalkers you control" clause Sphere of Safety and Norn's Annex
// do, so the tax is left at the default AttackTaxOnPlayer scope. That
// is the one line in this family an implementation can get silently
// wrong in the direction of playing stronger than printed, which is
// why the narrow reading is the zero value rather than a flag.
//
// "For each creature THEY CONTROL that's attacking you" is what makes
// the price scale with the swing rather than being a one-off toll, and
// the engine gets it for free: the tax is asked once per attacking
// creature and the prices are concatenated, so three attackers into
// this is "{2}{2}{2}" — six generic, charged as one payment on the
// declaration (CR 508.1a, ADR 0080).
//
// Nothing about the tax is a RESTRICTION. A creature under Propaganda
// still has no CantAttack bit, is still offered by the enumerator when
// its controller can pay, and still attacks the moment the {2} is
// paid. What the clause buys the defender is a price, and the engine
// charges it at declaration and never again — a Propaganda that enters
// AFTER attackers are declared charges nothing, because CR 508.1a is a
// turn-based action that has already happened.
func init() {
	Register(Spec{
		OracleID:     "ea9709b6-4c37-4d5a-b04d-cd4c42e4f9dd",
		Name:         "Propaganda",
		Completeness: CompletenessFull,
		AttackTaxes: []game.AttackTax{
			AttackTax("{2}", "Creatures can't attack you unless their controller pays {2} for each creature they control that's attacking you."),
		},
	})
}
