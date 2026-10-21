package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Norn's Annex — Artifact {3}{W/P}{W/P}:
//
//	"({W/P} can be paid with either {W} or 2 life.)
//	 Creatures can't attack you or planeswalkers you control unless
//	 their controller pays {W/P} for each of those creatures."
//
// Sphere of Safety's clause with a Phyrexian price, and the card that
// needed the attack dock to learn the "or 2 life" half (#4198).
//
// # The price
//
// AttackTax("{W/P}") is the tax for ONE attacking creature; the
// engine concatenates it once per attacker, so three attackers cost
// "{W/P}{W/P}{W/P}". Each symbol is paid with {W} or with 2 life
// (CR 107.4f) at the attacking player's choice. The total cost to
// attack is locked in when it is determined (CR 508.1h) and paid in
// full (CR 508.1j), and life can be paid only down to nothing
// (CR 119.4), so the choice is announced with the declaration:
// game.DeclareAttackersParams.PhyrexianLife, wire `phyrexian_life` on
// declare_attacker and declare_attackers, goes through the one
// strike-and-pay helper a cast and an activation use. Auto-tap never
// pays life for the player; the claim is theirs.
//
// The view says how many symbols one creature's tax prints
// (`attack_targets[].tax_phyrexian`), so the attack dock asks mana or
// life for each symbol without parsing a mana string, and the
// enumerator offers the life payment to a bot seat that cannot find
// {W}.
//
// # "Or planeswalkers you control"
//
// ProtectingPlaneswalkers, as for Sphere of Safety. Battles are not
// covered, and no card in this family says they are.
func init() {
	Register(Spec{
		OracleID:     "9a1fbe72-4a17-42be-8e23-d7d30a5e59c1",
		Name:         "Norn's Annex",
		Completeness: CompletenessFull,
		AttackTaxes: []game.AttackTax{
			ProtectingPlaneswalkers(AttackTax(
				"{W/P}",
				"Creatures can't attack you or planeswalkers you control unless their controller pays {W/P} for each of those creatures.",
			)),
		},
	})
}
