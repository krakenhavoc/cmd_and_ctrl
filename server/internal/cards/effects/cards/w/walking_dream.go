package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Walking Dream — Creature — Illusion {3}{U}, 3/3 (Stronghold, S82 OLD b750,
// #5627):
//
//	"This creature can't be blocked.
//	 This creature doesn't untap during your untap step if an opponent controls
//	 two or more creatures."
//
// Plasma Elemental's unblockability, and Steel Dromedary's conditional untap
// restriction (CR 502.3): asked at each of its controller's untap steps, true
// when some opponent controls two or more creatures then (read through the
// layered view, so an animated land counts).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "286e7ef1-d763-4b48-8603-a161dcf7efbc",
		Name:         "Walking Dream",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
		UntapStepRestrictions: []game.UntapStepRestriction{{
			Label: "This creature doesn't untap during your untap step if an opponent controls two or more creatures.",
			Restricts: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && effects.OldB750AnOpponentControlsTwoOrMoreCreatures(g, source.Controller)
			},
		}},
	})
}
