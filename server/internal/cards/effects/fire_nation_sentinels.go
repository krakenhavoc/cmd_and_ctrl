package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fire Nation Sentinels — Creature — Human Soldier {3}{B}{B}, 4/4 (Avatar:
// The Last Airbender Eternal, #5075):
//
//	"Whenever a nontoken creature an opponent controls dies, put a +1/+1
//	 counter on each creature you control."
//
// Overseer of the Damned's death watch (b33OpponentsNontokenCreatureDied,
// the dying creature's last-known controller and token status) with the
// Lieutenant body (b11PutCounterOnEachCreatureYouControl): every creature
// its controller controls as the trigger resolves, one placement each.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4ba60296-5612-41ed-bf53-a061783ed3b9",
		Name:         "Fire Nation Sentinels",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b33OpponentsNontokenCreatureDied(ev, source, g)
			}, "Fire Nation Sentinels — put a +1/+1 counter on each creature you control",
				b11PutCounterOnEachCreatureYouControl),
		},
	})
}
