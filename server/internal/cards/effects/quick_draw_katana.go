package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quick-Draw Katana — Artifact — Equipment {2} (Foundations, #4465):
//
//	"During your turn, equipped creature gets +2/+0 and has first strike.
//	 Equip {2}"
//
// Hook Swords' two turn-gated statics: both are read on every recompute
// against whose turn it is, and "your" is the Equipment's controller
// (CR 301.5d), so the creature is bigger on that player's turn only.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "604ef241-df8f-48f8-8680-f9ed77ff30ae",
		Name:         "Quick-Draw Katana",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:    "During your turn, equipped creature gets +2/+0.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return AttachedToSource(target, g, source) && duringSourceControllersTurn(target, g, source)
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power += 2
				},
			},
			GrantToAttachedWhile(duringSourceControllersTurn, "first strike"),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
