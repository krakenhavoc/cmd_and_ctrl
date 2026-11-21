package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fleshtaker — Creature — Human Assassin {W}{B}, 2/2 (Midnight Hunt, #5497):
//
//	"Whenever you sacrifice another creature, you gain 1 life and scry 1.
//	 {1}, Sacrifice another creature: This creature gets +2/+2 until end of
//	 turn."
//
// Body Dropper's shape: the trigger reads EventSacrifice for "another
// creature you sacrificed" (a token, or a creature sacrificed to pay its own
// ability, counts), and the ability's sacrifice is a cost, so the pump lands
// only if Fleshtaker is still on the battlefield when it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "14d5fdb3-14a3-461d-8515-a520c2c9e666",
		Name:         "Fleshtaker",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 1}, // the sacrifice trigger's scry
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventSacrifice, effects.SncB13YouSacrificedAnotherCreature, "Fleshtaker — gain 1 life and scry 1",
				effects.Do(effects.GainLife{Amount: 1}, effects.Scry{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{1}, Sacrifice another creature: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{
				Answers: game.AnswerSacOutlet | game.AnswerPump,
				Pump:    &game.Pump{Power: 2, Toughness: 2},
			},
			Cost:   effects.Plus(effects.ManaCost("{1}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Effect: effects.ThisCreatureUntilEOT("Fleshtaker — +2/+2", 2, 2),
		}},
	})
}
