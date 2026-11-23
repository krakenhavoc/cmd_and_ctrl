package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frostpeak Yeti — Snow Creature — Yeti {3}{U}, 3/3 (Jumpstart 2022, #5948):
//
//	"{1}{S}: This creature can't be blocked this turn. ({S} can be paid with
//	 one mana from a snow source.)"
//
// Agent of Horizons' unblockable row with a snow cost: it applies to the
// object the Yeti was when the ability was activated (CR 400.7), until end of
// turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0363363b-a538-4cf6-ae6b-6b45cdabd99a",
		Name:         "Frostpeak Yeti",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{S}: This creature can't be blocked this turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{1}{S}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked,
					Label: "Frostpeak Yeti — can't be blocked this turn"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
