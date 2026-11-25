package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jelenn Sphinx — Creature — Sphinx {3}{W}{U}, 1/5 (Battlebond):
//
//	"Flying, vigilance
//	 Whenever this creature attacks, other attacking creatures get +1/+1 until
//	 end of turn."
//
// Demolisher Spawn's attack trigger without the condition: the creatures that
// get the bonus are those attacking as it resolves-begins and are locked then
// (CR 611.2c), so a creature put onto the battlefield attacking later gets
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d23b514f-3081-45da-a2cf-7ed3cf49dc24",
		Name:            "Jelenn Sphinx",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.ThisAttacked,
				"Jelenn Sphinx — other attacking creatures get +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{
						Match:     effects.And(effects.AttackingCreature(), effects.OtherThan(item.SourceCardID)),
						Power:     1,
						Toughness: 1,
						Label:     "Jelenn Sphinx",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
