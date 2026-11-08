package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Formation Breaker — Creature — Beast {1}{G}, 2/1 (Tarkir: Dragonstorm,
// #4463):
//
//	"Creatures with power less than this creature's power can't block
//	 it.
//	 As long as you control a creature with a counter on it, this
//	 creature gets +1/+2."
//
// The block rule is Elusive Otter's, judged against the Breaker's power
// as blockers are declared. The +1/+2 is a layer-7c static (Magitek
// Infantry's shape) that is on while you control a creature with a
// counter of any kind on it, the Breaker included; it raises the power
// the block rule compares against.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "82f59ac4-bf94-4fa2-9457-f0bec8360bd9",
		Name:         "Formation Breaker",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBlockAttackers(PowerLessThanSource(), OnSelf(),
				"creatures with power less than Formation Breaker's can't block it"),
		},
		Static: []game.StaticAbility{{
			Label:    "This creature gets +1/+2 as long as you control a creature with a counter on it.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return selfCreature(target, g, source) && tdmControlsCreatureWithACounter(g, source.Controller)
			},
			Apply: pumpBy(1, 2),
		}},
	})
}
