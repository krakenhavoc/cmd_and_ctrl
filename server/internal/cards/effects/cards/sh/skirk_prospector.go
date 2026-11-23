package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skirk Prospector — Creature — Goblin {R}, 1/1 (EDHREC rank 1335):
//
//	"Sacrifice a Goblin: Add {R}."
//
// The Goblin deck's free sacrifice outlet and ritual: every Krenko
// token is one red mana. A mana ability — no stack, no priority
// window, any number of activations — with a sacrifice-another clause
// over the Goblin subtype (effective subtypes, so a changeling
// counts). No tap component, so summoning sickness never applies,
// and the Prospector is itself a Goblin it can sacrifice, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c18013e4-0b99-44e3-a2b2-027ace68723a",
		Name:         "Skirk Prospector",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{
				SacrificeOther: effects.B12SacrificeAGoblin(),
			},
			Produced: "{R}",
			Label:    "Sacrifice a Goblin: Add {R}",
			Answers:  game.AnswerSacOutlet,
		}},
	})
}
