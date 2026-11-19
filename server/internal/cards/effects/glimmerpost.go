package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glimmerpost — Land — Locus (Modern Horizons 3 Commander, #5948):
//
//	"When this land enters, you gain 1 life for each Locus on the
//	 battlefield.
//	 {T}: Add {C}."
//
// Cloudpost's count of the whole battlefield, every player's lands, read as
// the trigger resolves: the Glimmerpost itself is a Locus and counts, so a
// lone one gains 1, and a Locus an opponent controls counts too. "Locus" is
// a word of the land's text (CR 612.2) and the count is the effective
// subtypes, so a land made a Locus by an effect counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "92c9aad6-35ec-425d-be7d-393328992820",
		Name:         "Glimmerpost",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Glimmerpost — gain 1 life for each Locus on the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					locus := g.TextReaderOfForEffect(item.SourceCardID).TextSubtype("Locus")
					n := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.HasSubtype(locus) {
							n++
						}
					}
					return GainLife{Amount: n}.Apply(NewContext(g, item))
				}),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
	})
}
