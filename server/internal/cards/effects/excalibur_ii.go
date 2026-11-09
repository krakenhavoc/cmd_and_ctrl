package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Excalibur II — Legendary Artifact — Equipment {1}:
//
//	"Whenever you gain life, put a charge counter on Excalibur II.
//	 Equipped creature gets +1/+1 for each charge counter on Excalibur II.
//	 Equip {3}"
//
// The counter rides the life-gain event (one per event, not per point:
// the printed rulings), like Ajani's Pridemate. The pump counts the
// counters on the Equipment itself, re-read each recompute.
func init() {
	const charge = "charge"
	Register(Spec{
		OracleID:     "bfb6651c-e0b3-4dbd-b26a-3eeb690985fc",
		Name:         "Excalibur II",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttachedPer(1, 1, func(_ *game.Game, source *game.Card) int { return source.Counters[charge] }),
		},
		Triggered: []game.TriggeredAbility{
			WheneverYouGainLife("Excalibur II — put a charge counter on it", Do(CounterOnThis{Kind: charge, N: 1})),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
