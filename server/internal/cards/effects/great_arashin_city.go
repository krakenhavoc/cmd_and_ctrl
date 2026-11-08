package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Great Arashin City — Land (Tarkir: Dragonstorm, #4463):
//
//	"This land enters tapped unless you control a Forest or a Plains.
//	 {T}: Add {B}.
//	 {1}{B}, {T}, Exile a creature card from your graveyard: Create a 1/1
//	 white Spirit creature token."
//
// The TDM village shape (Mistrise Village): the checkland entry over a
// Forest or a Plains and a plain {B}. The activation is Moorland Haunt's
// cost, a creature card from your own graveyard exiled as it is paid, so a
// graveyard with no creature card offers nothing to pay with. The Spirit
// has no abilities.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f40f374b-acaf-459d-9ccd-b0b22d1a3f28",
		Name:         "Great Arashin City",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTappedUnless(youControlLandTyped("forest", "plains"))},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}, {T}, Exile a creature card from your graveyard: Create a 1/1 white Spirit creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    Plus(ManaCost("{1}{B}"), TapCost(), ExileFromGraveyard(1, "a creature card", MatchCreature)),
			Effect:  Do(CreateToken{Template: TokenCard("1/1 white Spirit"), N: 1}),
		}},
	})
}
