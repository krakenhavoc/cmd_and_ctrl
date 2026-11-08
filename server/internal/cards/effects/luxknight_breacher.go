package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Luxknight Breacher — Creature — Human Knight {3}{W}, 2/2 (Edge of
// Eternities, #4461):
//
//	"This creature enters with a +1/+1 counter on it for each other
//	 creature and/or artifact you control."
//
// Crowd-Control Warden's self-entry replacement (CR 614.1c), counted as
// it enters: each other permanent you control that is a creature, an
// artifact or both counts once. The Breacher is not on the battlefield
// yet, and is left out by identity anyway.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e0a508b5-0896-4620-a93e-60b958829703",
		Name:         "Luxknight Breacher",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b19EntersWithCountersCounted(game.CounterPlusOne, luxknightBreacherX,
				"Luxknight Breacher: enters with a +1/+1 counter per other creature and/or artifact you control"),
		},
	})
}

// luxknightBreacherX is "each other creature and/or artifact you control".
func luxknightBreacherX(g *game.Game, src *game.Card) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != src.InstanceID && c.Controller == src.Controller && (c.IsCreature() || c.IsArtifact()) {
			n++
		}
	}
	return n
}
