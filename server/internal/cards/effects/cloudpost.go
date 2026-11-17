package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cloudpost — Land — Locus (Mirrodin, #5606):
//
//	"This land enters tapped.
//	 {T}: Add {C} for each Locus on the battlefield."
//
// The tapped entry is the SelfEntersTapped replacement. The mana is
// Priest of Titania's count of the whole battlefield, every player's
// permanents, on the Locus land type: Cloudpost counts itself, and an
// opponent's Glimmerpost or Cloudpost counts too. "Locus" is a word of
// the land's text (CR 612.2) and is read through it. The count is the
// effective subtypes, so a land made a Locus by an effect counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f705c0eb-9c6c-4315-a860-208ed0c5d93e",
		Name:         "Cloudpost",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:  ManaAbilityCost{Tap: true},
			Label: "Add {C} for each Locus on the battlefield",
			ProducedAmount: func(g *game.Game, _, source uuid.UUID) game.ManaAmount {
				locus := g.TextReaderOfForEffect(source).TextSubtype("Locus")
				n := 0
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.HasSubtype(locus) {
						n++
					}
				}
				return game.ManaAmount{Options: []string{"C"}, N: n}
			},
		}},
	})
}
