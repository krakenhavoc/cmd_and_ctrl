package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quintorius, History Chaser — Legendary Planeswalker — Quintorius
// {2}{R}{W}, loyalty 5 (Secrets of Strixhaven Commander, #4834):
//
//	"Whenever one or more cards leave your graveyard, create a 3/2 red
//	 and white Spirit creature token.
//	 +1: You may discard a card. If you do, draw two cards, then mill a
//	 card.
//	 −4: Spirits you control gain double strike and vigilance until end
//	 of turn.
//	 Quintorius, History Chaser can be your commander."
//
// The static trigger is the batch trigger Quintorius, Field Historian
// shares. The +1 is the discard-then-draw prompt (Witch's Mark's shape) with
// "then mill a card" as the draw's own continuation, so the mill waits for
// any prompt the draws pause on. The −4 gives the keywords to the Spirits
// you control as it resolves (CR 611.2c). "Can be your commander" is read
// from the oracle text by deck validation.
//
// No simplification.
func init() {
	const plusOne = "Quintorius, History Chaser — you may discard a card; if you do, draw two cards, then mill a card"
	Register(Spec{
		OracleID:        "5c40a8d4-6969-4d71-aba5-5c7e4109b7d5",
		Name:            "Quintorius, History Chaser",
		Completeness:    CompletenessFull,
		StartingLoyalty: 5,
		Triggered: []game.TriggeredAbility{
			socB11SpiritWhenCardsLeaveYourGraveyard("Quintorius, History Chaser"),
		},
		Activated: []ActivatedAbility{
			{
				Label: "+1: You may discard a card. If you do, draw two cards, then mill a card.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return b39MayDiscardThenDraw(1, false, plusOne, socB11QuintoriusDiscardDraw)(NewContext(g, item))
				},
			},
			{
				Label: "−4: Spirits you control gain double strike and vigilance until end of turn.",
				Cost:  LoyaltyCost(-4),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return GrantKeywordUntilEOT{
						Match:    And(YouControl(), HasSubtype("Spirit")),
						Keywords: []string{"double strike", "vigilance"},
						Label:    "Quintorius, History Chaser — double strike and vigilance",
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
