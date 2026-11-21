package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Territorial Boar — Creature — Boar {1}{G}, 2/2 (Ravnica: Clue Edition, #5929): "Whenever a creature you control with power 4 or greater enters, this creature gets +1/+1 and gains vigilance until end of turn."
//
// The entering creature's power is read through the layers as it enters (Garruk, Curse Breaker).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "848fc2b8-7208-4415-8be1-add84c780f6e",
		Name:         "Territorial Boar",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, territorialBoarBigEntered,
				"Territorial Boar — gets +1/+1 and gains vigilance until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					const label = "Territorial Boar — +1/+1 and vigilance"
					if err := effects.ThisGetsUntilEndOfTurn(1, 1, label)(g, item); err != nil {
						return err
					}
					return effects.ThisGainsKeywordUntilEndOfTurn("vigilance", label)(g, item)
				}),
		},
	})
}

func territorialBoarBigEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
	return ok && c.IsCreature() && c.CurrentPower() >= 4
}
