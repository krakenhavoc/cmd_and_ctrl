package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eye of Yawgmoth — Artifact {3} (Nemesis, S82 OLD, #5618):
//
//	"{3}, {T}, Sacrifice a creature: Reveal a number of cards from the top of
//	 your library equal to the sacrificed creature's power. Put one into your
//	 hand and exile the rest."
//
// The number is the sacrificed creature's power as it last existed on the
// battlefield, floored at zero (CR 107.1b). The controller picks the card that
// goes to hand from those revealed; the rest are exiled once it has landed. A
// library shorter than the number reveals what it has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e0af613a-05ba-438d-ba74-3f6a9958576c",
		Name:         "Eye of Yawgmoth",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice a creature: Reveal a number of cards from the top of your library equal to the sacrificed creature's power. Put one into your hand and exile the rest.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.SacrificeACreature()),
			Effect:  effects.S82OldB677EyeOfYawgmoth,
		}},
	})
}
