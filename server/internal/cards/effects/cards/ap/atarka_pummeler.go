package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Atarka Pummeler — Creature — Ogre Warrior {4}{R}, 4/5 (Dragons of Tarkir, S82
// older sets):
//
//	"Formidable — {3}{R}{R}: Creatures you control gain menace until end of turn.
//	 Activate only if creatures you control have total power 8 or greater."
//
// A group keyword grant until end of turn, the set of creatures fixed as the
// ability resolves (CR 611.2c), behind the formidable activation condition (CR
// 602.5b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4b2831f5-70e3-4f4f-8393-4cd141c9da34",
		Name:         "Atarka Pummeler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{effects.OldB324FormidableRow(effects.ActivatedAbility{
			Purpose: game.Purpose{TeamPump: &game.Pump{Keywords: []string{"menace"}}, Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{3}{R}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.YouControl()),
					Keywords: []string{"menace"},
					Label:    "Atarka Pummeler — menace until end of turn",
				}.Apply(effects.NewContext(g, item))
			},
		}, "{3}{R}{R}: Creatures you control gain menace until end of turn.")},
	})
}
