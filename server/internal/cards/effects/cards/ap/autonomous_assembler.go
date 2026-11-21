package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Autonomous Assembler — Artifact Creature — Assembly-Worker {5}, 4/5
// (The Brothers' War, #5656): "Prototype {1}{W} — 2/2", "Vigilance",
// "{1}, {T}: Put a +1/+1 counter on target Assembly-Worker you control."
//
// Prototype is the card's second face (CR 718, game/prototype.go), built
// by the deck importer from the inset; nothing here spells it out.
// Vigilance is a printed keyword. The Assembler is an Assembly-Worker and
// may target itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3afdc98e-aeeb-43e9-bdcd-0c51da44d2f8",
		Name:            "Autonomous Assembler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Put a +1/+1 counter on target Assembly-Worker you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target Assembly-Worker you control", effects.Subtype("Assembly-Worker"), effects.YouControl()),
			Effect:  effects.PlusOneCounterOnChosenTarget,
		}},
	})
}
