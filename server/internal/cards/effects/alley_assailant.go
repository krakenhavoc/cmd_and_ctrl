package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alley Assailant — Creature — Vampire Rogue {2}{B}, 3/3 (Murders at
// Karlov Manor, #4469):
//
//	"This creature enters tapped.
//	 Disguise {4}{B}{B} (You may cast this card face down for {3} as a
//	 2/2 creature with ward {2}. Turn it face up any time for its
//	 disguise cost.)
//	 When this creature is turned face up, target opponent loses 3 life
//	 and you gain 3 life."
//
// The enters-tapped replacement is the card's own, so a face-down
// Assailant, which has no abilities (CR 708.2a), enters untapped: the
// engine runs no self replacement for a face-down entry (#1322). The
// face-up trigger is Bloodborn Scoundrels' drain.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "57eba48e-87bd-4b54-8635-7348020f2ba4",
		Name:             "Alley Assailant",
		Completeness:     CompletenessFull,
		Replacements:     []game.ReplacementEffect{SelfEntersTapped()},
		AlternativeCosts: []game.AlternativeCost{Disguise("{4}{B}{B}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisIsTurnedFaceUp("Alley Assailant — target opponent loses 3 life and you gain 3 life",
					func(g *game.Game, item *game.StackItem) error {
						return targetOpponentLosesAndYouGain(g, item, 3)
					}),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}
