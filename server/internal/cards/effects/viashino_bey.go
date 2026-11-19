package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Viashino Bey — Creature — Lizard, {2}{R}{R}, 4/3 (Urza's Legacy):
//
//	"If this creature attacks, all creatures you control attack if able."
//
// #5840. Not a trigger (the 2004-10-04 ruling: "It is a constraint on
// declaring an attack"), so it is a static that puts one CR 508.1d
// requirement on every other creature you control, gated on the Bey
// being in the declaration (IfThisAttacksAllYourCreaturesAttack,
// game.AttackRequirement.IfSourceAttacks). Declaring no attackers, or
// attackers without the Bey, owes nothing; declaring the Bey owes an
// attack from every other creature you control that can attack for
// free, and the active player's pass is refused, naming the Bey, until
// they do. A creature that can't attack, or could only for a cost,
// is not asked (CR 508.1d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d9c9b933-74f4-4898-bbb8-2da76b86ccda",
		Name:         "Viashino Bey",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{IfThisAttacksAllYourCreaturesAttack()},
	})
}
