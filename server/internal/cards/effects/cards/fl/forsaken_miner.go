package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forsaken Miner — Creature — Skeleton Rogue {B}, 2/2:
//
//	"This creature can't block.
//	 Whenever you commit a crime, you may pay {B}. If you do, return
//	 this card from your graveyard to the battlefield. (Targeting
//	 opponents, anything they control, and/or cards in their graveyards
//	 is a crime.)"
//
// The graveyard trigger and crime classification are handled by the
// shared event engine. The optional black mana is asked at resolution;
// both before offering and after payment, the card must still be the
// same graveyard object that triggered (CR 400.7).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d37e0127-6f68-438e-8b9e-a71bb3c34733",
		Name:         "Forsaken Miner",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.On(game.EventBecomesTarget,
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.YouCommittedCrime(ev, source)
				},
				"Forsaken Miner — you may pay {B} to return it to the battlefield",
				forsakenMinerMayPay)),
		},
	})
}

func forsakenMinerMayPay(g *game.Game, item *game.StackItem) error {
	return effects.MayPayForGraveyardReturn(g, item, "{B}", "Forsaken Miner — pay {B} to return it to the battlefield?")
}
