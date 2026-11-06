package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hook Swords — Artifact — Equipment {2}{R/W} (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"When this Equipment enters, create a 1/1 white Ally creature token,
//	 then attach this Equipment to it.
//	 During your turn, equipped creature gets +1/+1 and has first strike.
//	 Equip {3}"
//
// The enters trigger is Foot Chopper's (createTokenAndAttachSource) with
// the set's 1/1 white Ally. The bonus is two statics gated by Bilbo's
// Ring's duringSourceControllersTurn, read on every recompute against
// the turn: "your" is the Equipment's controller (CR 301.5d), so a
// creature you equip is bigger on your turn only. Equip is CR 702.6a.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "610e5fda-ff6c-408d-9d58-ba2b02c5101f",
		Name:         "Hook Swords",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:    "During your turn, equipped creature gets +1/+1.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return AttachedToSource(target, g, source) && duringSourceControllersTurn(target, g, source)
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power++
					c.Toughness++
				},
			},
			GrantToAttachedWhile(duringSourceControllersTurn, "first strike"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Hook Swords — create a 1/1 white Ally creature token, then attach this Equipment to it",
				func(g *game.Game, item *game.StackItem) error {
					return createTokenAndAttachSource(g, item, "1/1 white Ally", true)
				}),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
