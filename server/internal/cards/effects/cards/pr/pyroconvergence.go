package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pyroconvergence — Enchantment {4}{R} (Return to Ravnica, S82 OLD):
//
//	"Whenever you cast a multicolored spell, this enchantment deals 2
//	 damage to any target."
//
// A cast trigger on your own spells whose colours, read as the spell sits on
// the stack, number two or more (CR 105.2b; Hero of Precinct One's read). The
// target is chosen as the trigger goes on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8da5c156-3688-4d6e-b5c4-95c082d1b426",
		Name:         "Pyroconvergence",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventCast && ev.Actor == source.Controller && effects.B25CastIsMulticolored(ev, g)
			}, "Pyroconvergence — 2 damage to any target", effects.DealDamageToTheTarget(2)), effects.TargetAny()),
		},
	})
}
