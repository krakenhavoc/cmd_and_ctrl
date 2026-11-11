package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kulrath Knight — Creature — Elemental Knight {3}{B/R}{B/R}, 3/3 (Lorwyn
// Eclipsed Commander, #5368):
//
//	"Flying
//	 Wither
//	 Creatures your opponents control with counters on them can't attack
//	 or block."
//
// Wither is the engine's damage tail (game/infect_wither_toxic.go). The
// static is a layer 6 restriction re-read on every recompute, so a creature
// is held while it has any counter of any kind and freed the moment the
// last one goes (CR 508.1c, 509.1b). "Opponents" is read from the Knight's
// controller, so a creature the Knight's controller steals is theirs.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f3d1b2e9-6ede-416d-937b-2063ea278d2f",
		Name:            "Kulrath Knight",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordWither},
		Static: []game.StaticAbility{{
			Label: "Creatures your opponents control with counters on them can't attack or block.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				if !target.IsCreature() || target.Controller == source.Controller {
					return false
				}
				for _, n := range target.Counters {
					if n > 0 {
						return true
					}
				}
				return false
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttackOrBlock
			},
		}},
	})
}
