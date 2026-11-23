package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slogurk, the Overslime — Legendary Creature — Ooze {1}{G}{U}, 3/3:
//
//	"Trample
//	 Whenever a land card is put into your graveyard from anywhere, put a
//	 +1/+1 counter on Slogurk.
//	 Remove three +1/+1 counters from Slogurk: Return it to its owner's
//	 hand.
//	 When Slogurk leaves the battlefield, return up to three target land
//	 cards from your graveyard to your hand."
//
// The land trigger watches every way a land card reaches your graveyard (a
// mill, a discard, a destroy, a sacrifice), once per card,
// with the counter skipped once Slogurk has left. The bounce ability's
// whole cost is the three counters (paid at announcement, CR 602.2b), and
// the leaves trigger targets up to three land cards in your graveyard as it
// is put on the stack (CR 603.3d), so it works for any exit, the bounce
// included; a land that left the graveyard in response is an illegal
// target and stays put (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cee836f1-87a4-43db-a6ea-55575a4ce781",
		Name:            "Slogurk, the Overslime",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.S82MidB15LandCardToYourGraveyard(
				"Slogurk, the Overslime — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			effects.Targeting(
				effects.WhenThisLeaves("Slogurk, the Overslime — return up to three target land cards from your graveyard to your hand",
					func(g *game.Game, item *game.StackItem) error {
						return effects.ReturnLegalGraveyardTargetsToHand(effects.NewContext(g, item))
					}),
				effects.TargetCardInGraveyard("up to three target land cards from your graveyard", effects.Land(), effects.YouOwn()).WithCount(0, 3)),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove three +1/+1 counters from Slogurk: Return it to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.RemoveCountersFromThis(game.CounterPlusOne, 3),
			Effect:  effects.ReturnThisPermanentToOwnersHand,
		}},
	})
}
