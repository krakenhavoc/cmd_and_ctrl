package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soldevi Adnate — Creature — Human Cleric {1}{B}, 1/2 (Alliances, #5634):
//
//	"{T}, Sacrifice a black or artifact creature: Add an amount of {B} equal to
//	 the sacrificed creature's mana value."
//
// A mana ability (CR 605.1a) whose cost taps the Adnate and sacrifices one
// black or artifact creature (any creature of the controller's, the Adnate
// included if it qualifies): the mana is read off the sacrificed creature as
// it last existed on the battlefield (CR 608.2h), so a token or a {0} card adds
// none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e3174070-542f-429e-9149-b572f6614537",
		Name:         "Soldevi Adnate",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{
				Tap:            true,
				SacrificeOther: effects.SacrificeN(1, "a black or artifact creature", effects.And(effects.Creature(), effects.Or(effects.OfColor("B"), effects.Artifact()))).SacrificeOther,
			},
			ProducedForPaid: effects.OldB815BlackPerManaValueOfTheSacrificed,
			Label:           "{T}, Sacrifice a black or artifact creature: Add an amount of {B} equal to the sacrificed creature's mana value.",
			Answers:         game.AnswerSacOutlet,
		}},
	})
}
