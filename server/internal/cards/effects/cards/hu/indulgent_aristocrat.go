package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Indulgent Aristocrat — Creature — Vampire Noble {B}, 1/1 (EDHREC
// rank 2555):
//
//	"Lifelink
//	 {2}, Sacrifice a creature: Put a +1/+1 counter on each Vampire
//	 you control."
//
// The one-drop Vampire sac outlet. Lifelink rides PrintedKeywords;
// the activation is a CR 602 ability with a mana-and-sacrifice cost
// (the Aristocrat itself is a legal sacrifice — it is a creature —
// and then grows nothing), and its body is Cordial Vampire's
// b17PutCounterOnEachVampireYouControl: every Vampire the controller
// controls, the Aristocrat included, post-layer subtype so a
// changeling counts, snapshotted before the first counter lands.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c7054a2d-2e7b-4487-8fc3-f6a47a716fd3",
		Name:            "Indulgent Aristocrat",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice a creature: Put a +1/+1 counter on each Vampire you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeACreature()),
			Effect:  effects.B17PutCounterOnEachVampireYouControl,
		}},
	})
}
