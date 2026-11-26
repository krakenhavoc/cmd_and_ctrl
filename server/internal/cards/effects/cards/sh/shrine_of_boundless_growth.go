package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shrine of Boundless Growth — Artifact {3} (New Phyrexia):
//
//	"At the beginning of your upkeep and whenever you cast a green spell,
//	 put a charge counter on this artifact.
//	 {T}, Sacrifice this artifact: Add {C} for each charge counter on this artifact."
//
// The counter clause is ONE triggered ability with two trigger conditions.
// The second ability sacrifices the Shrine as a cost, so the number of
// charge counters is its last-known information (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6047a38c-ec03-428d-9f4f-15d767f8176e",
		Name:         "Shrine of Boundless Growth",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB460ShrineCounterTrigger("Shrine of Boundless Growth", "G"),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:  effects.ManaAbilityCost{Tap: true, Sacrifice: true},
			Label: "{T}, Sacrifice this artifact: Add {C} for each charge counter on this artifact.",
			ProducedForPaid: func(g *game.Game, _, source uuid.UUID, _ game.PaidCost) game.ManaOutput {
				return game.ManaOutput{Amount: game.ManaAmount{Options: []string{"C"}, N: effects.OldB460ChargeCountersOnSource(g, source)}}
			},
		}},
	})
}
