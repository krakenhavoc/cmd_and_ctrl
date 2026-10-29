package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Long-Lost Lances — Artifact — Equipment {2} (The Hobbit Commander):
//
//	"Equipped creature gets +2/+0.
//	 During your turn, creatures you control that are equipped have first
//	 strike and vigilance.
//	 Equip {2}"
//
// #4454, S76. PumpAttached for the +2/+0. The second sentence is not about
// this Equipment's own creature: it is every creature its controller
// controls that has any Equipment attached (isEquippedCreature), during that
// controller's turn (Bilbo's Ring's duringSourceControllersTurn), granted as
// two layer-6 keyword grants.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0eaec4ec-0f50-4470-a7bf-acd8c6c52dc2",
		Name:         "Long-Lost Lances",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 0),
			Labelled(KeywordGrant(hobB5EquippedCreatureOnYourTurn, "first strike"),
				"During your turn, creatures you control that are equipped have first strike."),
			Labelled(KeywordGrant(hobB5EquippedCreatureOnYourTurn, "vigilance"),
				"During your turn, creatures you control that are equipped have vigilance."),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}

// hobB5EquippedCreatureOnYourTurn is "during your turn, creatures you control
// that are equipped": a creature the source's controller controls with an
// Equipment attached, while it is that controller's turn.
func hobB5EquippedCreatureOnYourTurn(target *game.Card, g *game.Game, source *game.Card) bool {
	return target.Controller == source.Controller &&
		IsYourTurn(g, source.Controller) &&
		isEquippedCreature(g, *target)
}
