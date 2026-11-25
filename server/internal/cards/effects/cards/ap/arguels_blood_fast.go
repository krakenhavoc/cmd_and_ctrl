package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arguel's Blood Fast // Temple of Aclazotz — Legendary Enchantment {1}{B}
// (Ixalan, S82 older sets):
//
//	"{1}{B}, Pay 2 life: Draw a card.
//	 At the beginning of your upkeep, if you have 5 or less life, you may
//	 transform Arguel's Blood Fast."
//
//	Temple of Aclazotz — Legendary Land:
//	"{T}: Add {B}.
//	 {T}, Sacrifice a creature: You gain life equal to the sacrificed
//	 creature's toughness."
//
// The upkeep trigger's "if" is intervening (CR 603.4): checked as the upkeep
// begins and again on resolution, and the "may" is asked as it resolves. The
// life gain is Miren's: the sacrificed creature's last-known toughness.
//
// No simplification.
func init() {
	const oracle = "be2a4bc4-8af6-48c5-9421-32d26272e71a"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Arguel's Blood Fast",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, Pay 2 life: Draw a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.PayLife(2)),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.InterveningIfYou("you have 5 or less life",
				func(g *game.Game, you uuid.UUID) bool {
					p := g.PlayerByIDForEffect(you)
					return p != nil && p.Life <= 5
				},
				effects.AtYourUpkeep("Arguel's Blood Fast — transform it", effects.Do(effects.TransformThis{}))),
				"Arguel's Blood Fast — transform it?"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Temple of Aclazotz",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "{T}: Add {B}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice a creature: You gain life equal to the sacrificed creature's toughness.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeACreature()),
			Effect:  effects.GainLifeEqualToSacrificedToughness,
		}},
	})
}
