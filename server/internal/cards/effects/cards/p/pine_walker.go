package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pine Walker — Khans of Tarkir (S82 OLD b345, #5554):
//
//	Creature — Elemental {3}{G}{G}, 5/5: Morph {4}{G}. Whenever this creature or
//	another creature you control is turned face up, untap that creature.
//
// The creature is read from the event, so the untap hits the one that turned up.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "4c92b862-34c7-4195-bee3-bcc0cafad140",
		Name:             "Pine Walker",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{4}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventTurnedFaceUp, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.Controller == source.Controller && c.IsCreature()
			}, "Pine Walker — untap that creature", effects.OldB345UntapEventCreature),
		},
	})
}
