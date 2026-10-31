package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// N'Yami-Class Mother Ship — {6} Artifact — Vehicle 5/7 (Marvel Super Heroes
// Commander, #4455):
//
//	"Flying, vigilance, haste
//	 Whenever this Vehicle deals combat damage to a player, look at the top
//	 card of your library. If it's a permanent card, you may put it onto the
//	 battlefield. If you don't put it onto the battlefield, put it into your
//	 hand.
//	 Crew 3"
//
// Risen Reef's look: only the controller sees the card, the put is
// optional and limited to a permanent card (PutFromLibraryOntoBattlefield
// drops instants and sorceries), and everything not put, whether a
// nonpermanent card, a declined permanent or one a replacement kept off the
// battlefield, goes to the hand (risenReefRestToHand). Crew is Dependable
// Quinjet's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5dec1b09-4df0-4e9b-9dca-b80134be3343",
		Name:            "N'Yami-Class Mother Ship",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", "haste"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer(
				"N'Yami-Class Mother Ship — look at the top card; you may put a permanent onto the battlefield, otherwise put it into your hand",
				func(g *game.Game, item *game.StackItem) error {
					looked := g.LookAtTopOfLibraryForEffect(item.Controller, 1)
					return PutFromLibraryOntoBattlefield{
						Player:   item.Controller,
						Cards:    looked,
						Max:      1,
						Optional: true,
						Label:    "N'Yami-Class Mother Ship — you may put the permanent card onto the battlefield (otherwise it goes into your hand)",
						ThenTail: risenReefRestToHand,
					}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    CrewCost(3),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("N'Yami-Class Mother Ship"),
		}},
	})
}
