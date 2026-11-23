package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "bfb6651c-e0b3-4dbd-b26a-3eeb690985fc",
		Name:         "Excalibur II",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 1, func(_ *game.Game, source *game.Card) int { return source.Counters[charge] }),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouGainLife("Excalibur II — put a charge counter on it", effects.Do(effects.CounterOnThis{Kind: charge, N: 1})),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
