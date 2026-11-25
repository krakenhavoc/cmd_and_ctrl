package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ajani's Comrade — {1}{G} Creature — Elf Soldier 2/2 (Aether Revolt, #5536):
//
//	"Trample
//	 At the beginning of combat on your turn, if you control an Ajani
//	 planeswalker, put a +1/+1 counter on this creature."
//
// An intervening "if" (CR 603.4): an Ajani planeswalker (the subtype, not
// the name) must be under your control as combat begins and as the
// ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "15b86440-2040-47ae-ab4e-8445b3e24462",
		Name:            "Ajani's Comrade",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you control an Ajani planeswalker",
				func(g *game.Game, you uuid.UUID) bool { return effects.B42ControlsPlaneswalkerNamed(g, you, "Ajani") },
				effects.On(game.EventStepBegan, effects.StepBegan(game.StepBeginCombat, true),
					"Ajani's Comrade — put a +1/+1 counter on this creature",
					effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
		},
	})
}
