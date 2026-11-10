package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ripjaw Raptor — Creature — Dinosaur {2}{G}{G}, 4/5:
//
//	"Enrage — Whenever this creature is dealt damage, draw a card."
//
// Strong, the Brutish Thespian's enrage shape: one trigger per damage
// instance, not per source (CR 120.3, CR 510.2), waiting for the damage
// to settle (AtBatchEnd). Lethal damage still triggers it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "80bb9ce4-1f3c-42ca-b75f-9a6b2f42073b",
		Name:         "Ripjaw Raptor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b35SelfWasDealtDamage(ev, source)
			},
			Key:        "Ripjaw Raptor — Enrage: draw a card",
			AtBatchEnd: ThatMuchDamage(),
			Effect:     Do(DrawCards{N: 1}),
		}},
	})
}
