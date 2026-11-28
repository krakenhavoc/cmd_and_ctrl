package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Black Carriage — Creature — Horse {3}{B}{B}, 4/4 (Homelands, S82 OLD b821,
// #5635):
//
//	"Trample
//	This creature doesn't untap during your untap step.
//	Sacrifice a creature: Untap this creature. Activate only during your upkeep."
//
// Phyrexian Colossus's untap restriction and untap, with a sacrifice (Ruthless
// Knave's SacrificeACreature, which may be the Carriage itself) as the cost and
// Shard Phoenix's upkeep window (CR 602.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "6bf5dcf1-5c93-42f4-adaf-a6f2e5807596",
		Name:                  "Black Carriage",
		Completeness:          effects.CompletenessFull,
		PrintedKeywords:       []string{"trample"},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringYourUntapStep()},
		Activated: []effects.ActivatedAbility{{
			Label:     "Sacrifice a creature: Untap this creature. Activate only during your upkeep.",
			Purpose:   game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:      effects.SacrificeACreature(),
			Condition: effects.DuringYourUpkeep(),
			Effect:    effects.B31UntapSelf,
		}},
	})
}
