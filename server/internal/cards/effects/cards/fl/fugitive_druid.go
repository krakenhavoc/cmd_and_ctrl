package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fugitive Druid — Creature — Human Druid {3}{G}, 3/2 (Tempest, S82 OLD batch
// 766, #5628):
//
//	"Whenever this creature becomes the target of an Aura spell, you draw a
//	 card."
//
// Brine Comber's reading of "becomes the target of an Aura spell": the
// becomes-target event from a spell that is an Aura, read off the stack as the
// targets are chosen (CR 601.2c). An ability that targets it, or any spell that
// is not an Aura, does nothing. The controller of the Druid draws, whoever cast
// the Aura.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e05f3236-dc82-4353-87ab-d26ae7caa0e7",
		Name:         "Fugitive Druid",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.TargetedByAnAuraSpell(ev, g, source.InstanceID)
			}, "Fugitive Druid — you draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
