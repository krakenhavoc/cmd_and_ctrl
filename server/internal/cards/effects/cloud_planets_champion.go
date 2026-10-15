package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cloud, Planet's Champion — Legendary Creature — Human Soldier
// Mercenary {3}{R}{W}, 4/4 (#3778):
//
//	"During your turn, as long as Cloud is equipped, it has double
//	 strike and indestructible.
//	 Equip abilities you activate that target Cloud cost {2} less to
//	 activate."
//
// The keywords are a layer 6 self grant read on every recompute: your
// turn (IsYourTurn) and an Equipment attached to Cloud. The discount is
// the equip-bit cost modifier.
//
// No simplification.
func init() {
	whileEquippedOnYourTurn := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return selfOnly(target, g, source) && IsYourTurn(g, source.Controller) && isEquippedCreature(g, *target)
	}
	ds := KeywordGrant(whileEquippedOnYourTurn, "double strike")
	ds.Label = "During your turn, as long as Cloud is equipped, it has double strike."
	ind := KeywordGrant(whileEquippedOnYourTurn, "indestructible")
	ind.Label = "During your turn, as long as Cloud is equipped, it has indestructible."
	Register(Spec{
		OracleID:     "8512bd08-ecba-4e0e-b65c-8209764a65e5",
		Name:         "Cloud, Planet's Champion",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{ds, ind},
		CostModifiers: []game.CostModifier{
			EquipAbilitiesThatTargetThisCostLess(2,
				"Equip abilities you activate that target Cloud cost {2} less to activate."),
		},
	})
}
