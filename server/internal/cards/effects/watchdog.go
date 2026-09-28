package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Watchdog — Artifact Creature — Dog, {3}, 1/2:
//
//	"This creature blocks each combat if able.
//	 As long as this creature is untapped, all creatures attacking you
//	 get -1/-0."
//
// #1684: BlocksEachCombat (#1597) plus a layer 7c static over every
// creature whose attack target is the Watchdog's controller — "you",
// not a planeswalker or battle they control. The static reads attacking
// status and tap state, so it declares DependsOnAttackingStatus (the
// layer cache is dropped when attackers are declared and when combat
// ends) and rides the tap/untap invalidation for "as long as this
// creature is untapped". Blocking does not tap the Watchdog, so the
// -1/-0 stays on for the combat it blocks in.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6a951d2d-fc5a-4f41-bd77-3e301b76cf47",
		Name:         "Watchdog",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			BlocksEachCombat(),
			{
				Layer:                    game.Layer7PT,
				SubLayer:                 game.SubLayer7C_Modify,
				DependsOnAttackingStatus: true,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return !source.Tapped && target.IsCreature() && target.AttackingTarget == source.Controller
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power--
				},
			},
		},
	})
}
