package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Persistent Petitioners — Creature — Human Advisor {1}{U}, 1/3 (Ravnica
// Remastered, #5948):
//
//	"{1}, {T}: Target player mills a card.
//	 Tap four untapped Advisors you control: Target player mills twelve cards.
//	 A deck can have any number of cards named Persistent Petitioners."
//
// The second ability's cost taps four untapped Advisors (a creature type read
// through the layers, so a changeling counts), the Petitioners itself among them
// if it is untapped; tapping a creature is not the {T} symbol (CR 302.6), so
// a creature that arrived this turn pays it. The deckbuilding sentence is CR
// 100.2a's, read from the card's text by the deck validator.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0e488c6c-aae2-450f-b969-7bb5a1b37a66",
		Name:         "Persistent Petitioners",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{1}, {T}: Target player mills a card.",
				Cost:    Plus(ManaCost("{1}"), TapCost()),
				Targets: TargetPlayer("target player"),
				Effect:  b31MillChosenPlayer(1),
			},
			{
				Label: "Tap four untapped Advisors you control: Target player mills twelve cards.",
				Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
					Count:  4,
					Filter: TargetPermanent("four untapped Advisors you control", Creature(), HasSubtype("Advisor")),
					Label:  "four untapped Advisors you control",
				}},
				Targets: TargetPlayer("target player"),
				Effect:  b31MillChosenPlayer(12),
			},
		},
	})
}
