package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Weatherlight — Legendary Artifact — Vehicle {4}, 4/5 (Reprint, #5908):
//
//	"Flying
//	 Whenever Weatherlight deals combat damage to a player, look at the top
//	 five cards of your library. You may reveal a historic card from among
//	 them and put it into your hand. Put the rest on the bottom of your
//	 library in a random order. (Artifacts, legendaries, and Sagas are
//	 historic.)
//	 Crew 3"
//
// The trigger is Commune with Spirits' look-and-take over five cards with
// the historic predicate (artifacts, legendaries and Sagas, CR 700.6): at
// most one is taken, revealed, and the rest go to the bottom in a random
// order. Crew 3 is the ordinary CR 702.122 crew ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "73b3682f-396a-4ebe-aaac-6ce3ee36283a",
		Name:            "Weatherlight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Weatherlight — look at the top five cards; you may reveal a historic card and put it into your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := item.Controller
					return effects.TakeFromLibraryToHand{
						Player:   player,
						Cards:    g.LookAtTopOfLibraryForEffect(player, 5),
						Match:    effects.B14Historic(),
						Max:      1,
						Optional: true,
						Label:    "Weatherlight — you may reveal a historic card and put it into your hand",
						Key:      effects.TakeRevealRestBottomRandom,
					}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    effects.CrewCost(3),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Weatherlight"),
		}},
	})
}
