package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dwarven Patrol — Creature — Dwarf {2}{R}, 4/2 (Apocalypse, #5614):
//
//	"This creature doesn't untap during your untap step.
//	 Whenever you cast a nonred spell, untap this creature."
//
// Brass Man's untap-step restriction; the trigger watches your own casts and
// is not optional. A colourless spell is nonred.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "4369adc8-d8e0-4d51-b2a9-7e4a8dfd9f8f",
		Name:                  "Dwarven Patrol",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringYourUntapStep()},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && !spell.HasColor(source.TextColor("R"))
			}, "Dwarven Patrol — untap this creature", effects.B31UntapSelf),
		},
	})
}
