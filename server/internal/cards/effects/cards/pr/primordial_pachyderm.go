package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Primordial Pachyderm — {3}{G} Creature — Elephant Avatar 4/4 (Teenage
// Mutant Ninja Turtles, #4457):
//
//	"Reach, trample
//	 When this creature enters, you gain 2 life."
//
// Two printed keywords and the ordinary enters trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5ebdc1a5-e678-444b-abb8-8753558a39de",
		Name:            "Primordial Pachyderm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Primordial Pachyderm — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
