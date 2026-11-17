package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Urza, Powerstone Prodigy — Legendary Creature — Human Artificer {2}{U},
// 1/3 (The Brothers' War):
//
//	"Vigilance
//	 {1}, {T}: Draw a card, then discard a card.
//	 Whenever you discard one or more artifact cards, create a tapped
//	 Powerstone token. This ability triggers only once each turn."
//
// The discard trigger is the shared Urza/Mishra one
// (wheneverYouDiscardOneOrMoreArtifactCardsOnce): it fires off the
// loot's own discard too, and off any other discard of yours.
//
// The loot is the shared draw-then-discard (lootOne): the discard is the
// controller's choice, from the hand as it is after the draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8ad7fdd2-bff5-4434-a3f9-1f5ce0ed573b",
		Name:            "Urza, Powerstone Prodigy",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Draws: 1, Discards: 1, Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return lootOne(g, item, 1)
			},
		}},
		Triggered: []game.TriggeredAbility{
			wheneverYouDiscardOneOrMoreArtifactCardsOnce(
				"Urza, Powerstone Prodigy — create a tapped Powerstone token",
				Do(CreateTokenAdvanced{Spec: Token(PowerstoneToken()).EntersTapped(), N: 1})),
		},
	})
}
