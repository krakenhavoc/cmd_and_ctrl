package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mariposa Military Base — Land:
//
//	"You may have this land enter tapped. If you do, you get two rad
//	 counters.
//	 {T}: Add {C}.
//	 {5}, {T}: Draw a card. This ability costs {1} less to activate for
//	 each rad counter you have."
//
// #2809, #2701. The first line is an optional enters-tapped replacement
// (mayEnterTappedToGetRadCounters): the player who plays the land is
// asked, and "yes" taps it and gives them the counters. A land put
// onto the battlefield by an effect that cannot ask enters untapped
// with no counters, the weaker branch of the "may".
//
// The draw ability's discount is its OWN cost clause
// (ActivatedAbility.CostModifiers): {1} off per rad counter the
// activator has when the cost is determined (CR 602.2b, 601.2f), which
// can only reduce the generic {5}, so with five or more counters the
// ability costs nothing but the tap.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "f1e03d99-024a-430b-9342-ffd2268bd103",
		Name:         "Mariposa Military Base",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			mayEnterTappedToGetRadCounters("Mariposa Military Base", 2),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{5}, {T}: Draw a card. This ability costs {1} less to activate for each rad counter you have.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{5}"), TapCost()),
			CostModifiers: []game.CostModifier{
				CostsLessEach(radCountersYouHave, "This ability costs {1} less to activate for each rad counter you have."),
			},
			Effect: b36DrawOne,
		}},
	})
}
