package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Daniels, Temporal Agent — Legendary Creature — Human Advisor
// {2}{B/G}{B/G}, 3/1 (Star Trek, #4973):
//
//	"When Daniels enters, return target creature card from your
//	 graveyard to your hand.
//	 {2}, Put this card from your graveyard on the bottom of your
//	 library: You gain 2 life."
//
// The enters trigger targets a creature card in its controller's own
// graveyard (CR 115.1); with none it is removed from the stack, and a
// card that leaves in response is not returned (CR 608.2b). The second
// ability functions only from the graveyard (CR 113.6): putting the card
// on the bottom is part of its cost (game.AbilityCost.BottomSelf, paid
// from the graveyard), so the card is in the library before anyone can
// respond, and the ability gains the life on its own (CR 113.7a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2e44a02f-af94-4696-b979-cdf68285c87a",
		Name:         "Daniels, Temporal Agent",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Daniels, Temporal Agent — return target creature card from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					return returnLegalGraveyardTargetsToHand(NewContext(g, item))
				}),
				TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature())),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, Put this card from your graveyard on the bottom of your library: You gain 2 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), PutThisOnTheBottomOfItsOwnersLibrary()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GainLife{Amount: 2}.Apply(NewContext(g, item))
			},
		}},
	})
}
