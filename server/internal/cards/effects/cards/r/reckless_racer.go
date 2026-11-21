package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reckless Racer — Creature — Human Pilot {2}{R}, 2/3 (Aether Revolt,
// #5536):
//
//	"First strike
//	 Whenever this creature becomes tapped, you may discard a card. If you
//	 do, draw a card."
//
// Rescue Leopard's becomes-tapped loot: tapped to attack, to crew or by
// an effect, each counts.
//
// No simplification.
func init() {
	const label = "Reckless Racer — you may discard a card; if you do, draw a card"
	effects.Register(effects.Spec{
		OracleID:        "d9a09e7f-6bef-4e75-b5c7-cea20a69cf02",
		Name:            "Reckless Racer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
				label, effects.TdmMayDiscardThenDrawOne(label)),
		},
	})
}
