package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Roc Hatchling — Creature — Bird {R}, 0/1 (Weatherlight, S82 older sets batch 780, #5629):
//
//	"This creature enters with four shell counters on it.
//	 At the beginning of your upkeep, remove a shell counter from this creature.
//	 As long as this creature has no shell counters on it, it gets +3/+2 and has flying."
//
// The counters come with the entry (a self-replacement, CR 614.1c), one comes off at each of its controller's upkeeps,
// and two layer-read statics (+3/+2 in layer 7c, flying in layer 6) hold while it has none.
//
// No simplification.
func init() {
	noShells := effects.S82OldB780ThisHasNoCounters("shell")
	flying := effects.KeywordGrant(noShells, "flying")
	flying.Label = "As long as this creature has no shell counters on it, it has flying."
	effects.Register(effects.Spec{
		OracleID:     "6764d4d9-1fe5-4b42-830c-a3b4c3b37f1f",
		Name:         "Roc Hatchling",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.B10EntersWithCounters("shell", 4, "Roc Hatchling: enters with four shell counters")},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Roc Hatchling — remove a shell counter from this creature", effects.RemoveACounterFromThis("shell")),
		},
		Static: []game.StaticAbility{
			{
				Label:     "As long as this creature has no shell counters on it, it gets +3/+2.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: noShells,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power += 3
					c.Toughness += 2
				},
			},
			flying,
		},
	})
}
