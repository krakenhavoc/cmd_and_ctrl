package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Teysa Karlov — "If a creature dying causes a triggered ability of a
// permanent you control to trigger, that ability triggers an additional
// time. Creature tokens you control have vigilance and lifelink."
func init() {
	doubler := DoublesDying(Creature())
	doubler.Label = "Teysa Karlov"
	Register(Spec{
		OracleID:        "644eeefd-e684-4ca8-8aef-a892ca130c07",
		Name:            "Teysa Karlov",
		Completeness:    CompletenessFull,
		TriggerDoublers: []game.TriggerDoubler{doubler},
		Static: []game.StaticAbility{
			b16GrantKeywords(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && IsToken(*target) && target.IsCreature()
			}, "vigilance", "lifelink"),
		},
	})
}
