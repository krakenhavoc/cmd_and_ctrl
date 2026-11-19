package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sengir Vampire — Creature — Vampire {3}{B}{B}, 4/4 (Mystery Booster 2):
//
//	"Flying
//	 Whenever a creature dealt damage by this creature this turn dies, put a
//	 +1/+1 counter on this creature."
//
// Watches every creature dying that this Vampire damaged earlier in the
// turn (the damage record is the turn's, CR 120.1), the Bushi Tenderfoot
// reading; the counter goes on the Vampire if it is still on the
// battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "749141aa-f6c4-4ad8-b146-406e68ae9b0b",
		Name:            "Sengir Vampire",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, died := diedCreature(ev, g)
				return died && ev.CardID != source.InstanceID && dealtDamageByThisTurn(g, source.InstanceID, ev.CardID)
			}, "Sengir Vampire — put a +1/+1 counter on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
