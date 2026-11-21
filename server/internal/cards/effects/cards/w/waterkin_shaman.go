package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Waterkin Shaman — Creature — Elemental Shaman {1}{U}, 2/1 (Foundations
// Jumpstart, S82, #5929):
//
//	"Whenever a creature you control with flying enters, this creature
//	 gets +1/+1 until end of turn."
//
// Dragon Tempest's flier-entered trigger (judged as it entered, through
// the layered view, the source itself included) with Momo's pump on this
// creature.
//
// No simplification.
func init() {
	const label = "Waterkin Shaman — gets +1/+1 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "64e0c580-0e55-4f42-8594-11ccb584d633",
		Name:         "Waterkin Shaman",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && game.HasKeyword(&c, "flying")
			}, label, effects.ThisGetsUntilEndOfTurn(1, 1, label)),
		},
	})
}
