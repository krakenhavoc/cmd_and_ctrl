package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Teysa Karlov — "If a creature dying causes a triggered ability of a
// permanent you control to trigger, that ability triggers an additional
// time. Creature tokens you control have vigilance and lifelink."
func init() {
	doubler := effects.DoublesDying(effects.Creature())
	doubler.Label = "Teysa Karlov"
	doubler.RowLabel = "Abilities that trigger when a creature dies trigger an additional time."
	effects.Register(effects.Spec{
		OracleID:        "644eeefd-e684-4ca8-8aef-a892ca130c07",
		Name:            "Teysa Karlov",
		Completeness:    effects.CompletenessFull,
		TriggerDoublers: []game.TriggerDoubler{doubler},
		Static: []game.StaticAbility{
			effects.B16GrantKeywords(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && effects.IsToken(*target) && target.IsCreature()
			}, "vigilance", "lifelink"),
		},
	})
}
