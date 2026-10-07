package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leyline Axe — Artifact — Equipment for {4}:
//
//	"If this card is in your opening hand, you may begin the game with
//	 it on the battlefield.
//	 Equipped creature gets +1/+1 and has double strike and trample.
//	 Equip {3} ({3}: Attach to target creature you control. Equip only
//	 as a sorcery.)"
//
// The pump, the double strike and trample grant, and the plain Equip
// {3} are all ordinary shapes — PumpAttached and GrantToAttached
// composed the way every other keyword-granting Equipment in the
// catalog is.
//
// The "leyline" opening-hand clause (CR 103.6a) is Spec.OpeningHand
// (ADR 0133): the seat holding it is asked as the mulligan window
// closes, and a yes puts the Equipment onto the battlefield unattached.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4f597675-0f6d-438c-990c-337171927a5e",
		Name:         "Leyline Axe",
		Completeness: CompletenessFull,
		OpeningHand:  BeginTheGameOnTheBattlefield(),
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("double strike", "trample"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{3}"),
		},
	})
}
