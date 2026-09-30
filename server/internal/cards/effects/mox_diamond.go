package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mox Diamond — Artifact {0}:
//
//	"If this artifact would enter, you may discard a land card instead.
//	 If you do, put this artifact onto the battlefield. If you don't,
//	 put it into its owner's graveyard.
//	 {T}: Add one mana of any color."
//
// ADR 0098 (#1744). The first sentence is a CR 614.1a replacement of
// the Mox's own entry, from whatever zone it enters (CR 113.6h): cast,
// put from a hand or a library, reanimated. Its controller is asked
// before it enters (CR 614.12a) which land card to discard; the
// discard is an effect's (Library of Leng applies, and every discard
// payoff sees it), and "if you do" means the land really left the hand.
// Otherwise the Mox goes to its owner's graveyard and never enters, so
// nothing that watches an entry triggers (2008-05-01 ruling). With no
// land card in hand it is not asked: it goes to the graveyard.
//
// "Any color" is the pipe over all five, listed with the commander's
// identity first — the text does not say "in your commander's color
// identity".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f3c5978a-70fa-431f-933b-b954bd0db0ea",
		Name:         "Mox Diamond",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersOnlyIfYouDiscardFromHand("Mox Diamond", "a land card", isLandCard),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
