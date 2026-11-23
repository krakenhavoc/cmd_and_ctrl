package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Junktown — Land (Fallout, #5469):
//
//	"{T}: Add {C}.
//	 {4}{R}, {T}, Sacrifice this land: Create three Junk tokens.
//	 (They're artifacts with "{T}, Sacrifice this token: Exile the top
//	 card of your library. You may play that card this turn. Activate
//	 only as a sorcery.")"
//
// A colorless land's mana ability and a sacrifice ability making three
// of the printed Junk tokens (JunkToken, CR 111.10t) in one creation.
// The sacrifice is part of the cost, so the land is gone whatever the
// answer.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c470a802-931f-4b63-92fd-ef9cf2e796dd",
		Name:         "Junktown",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{R}, {T}, Sacrifice this land: Create three Junk tokens.",
			Cost:    effects.Plus(effects.ManaCost("{4}{R}"), effects.TapCost(), effects.SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  effects.Do(effects.CreateToken{Template: effects.JunkToken(), N: 3}),
		}},
	})
}
