package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dauntless Escort — Creature — Rhino Soldier, {1}{G}{W}, 3/3:
//
//	"Sacrifice this creature: Creatures you control gain
//	 indestructible until end of turn."
//
// Selfless Spirit's older, bigger, slower cousin — no flying, one
// more mana, two more toughness. Kept as a separate catalog entry
// rather than folded into a shared helper because a singleton deck
// plays both and one file per card is the rule; see
// selfless_spirit.go for the reasoning on why the cost being a
// SACRIFICE (and therefore not destruction, CR 701.21a) is the
// interesting part of this ability.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c0516c2c-d7c9-4763-9180-980f60205a28",
		Name:         "Dauntless Escort",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Creatures you control gain indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive, Saves: &game.Saves{Matches: game.SweepCreatures}},
			Cost:    game.AbilityCost{SacrificeSelf: true},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.YouControl()),
					Keywords: []string{"indestructible"},
					Label:    "Dauntless Escort — indestructible",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
