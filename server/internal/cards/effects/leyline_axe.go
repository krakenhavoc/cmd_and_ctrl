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
// The "leyline" opening-hand clause does NOT ship. There is no
// opening-hand / mulligan-time hook anywhere in the engine — deck
// setup puts every card in the library, full stop — so a card that
// starts a game already on the battlefield is not a thing any spec
// can ask for today. Leaving the clause off is the correct direction
// per #259: it makes the card strictly weaker (it is drawn and cast
// like any other Equipment, never free), never stronger.
func init() {
	Register(Spec{
		OracleID:     "4f597675-0f6d-438c-990c-337171927a5e",
		Name:         "Leyline Axe",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Starting the game with this on the battlefield for free isn't implemented — you always draw and cast it like an ordinary Equipment.",
		},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("double strike", "trample"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{3}"),
		},
	})
}
