package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Meddling Youths — Creature — Human Detective {3}{R}{W}, 4/5 (Murders at
// Karlov Manor, #4469):
//
//	"Haste
//	 Whenever you attack with three or more creatures, investigate.
//	 (Create a Clue token. It's an artifact with "{2}, Sacrifice this
//	 token: Draw a card.")"
//
// "Attack with three or more creatures" is Eiganjo Dynastorian's count:
// the EventAttack events of one declaration by you, settled as a batch,
// one trigger when there are at least three (ThatManyEventsAtLeast). The
// Youths need not be among the attackers, and an additional combat's
// declaration is its own batch. Investigate is one Clue (CR 701.16a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9acb74a9-f372-43ad-a80d-4f8d52236eb8",
		Name:            "Meddling Youths",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.WithAtBatchEnd(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source != nil && effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Meddling Youths — investigate (you attacked with three or more creatures)",
				effects.Do(effects.Investigate{N: 1})),
				effects.ThatManyEventsAtLeast(3)),
		},
	})
}
