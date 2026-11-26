package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ratchet Bomb — Artifact {2} (Magic 2014, S82 old sets):
//
//	"{T}: Put a charge counter on this artifact.
//	 {T}, Sacrifice this artifact: Destroy each nonland permanent with mana
//	 value equal to the number of charge counters on this artifact."
//
// Blast Zone's sweep on an artifact. The sacrifice is a cost, so the count is the
// Bomb's last-known information (CR 608.2h), read through the source permanent
// rather than the graveyard card, whose counters are cleared. The sweep spares
// lands; a token's mana value is 0, so a Bomb with no counters destroys every
// nonland token, and an indestructible permanent survives it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "255d752a-8251-4e4c-99d6-4a1648de1474",
		Name:         "Ratchet Bomb",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Put a charge counter on this artifact.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.TapCost(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterCharge, N: 1}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{T}, Sacrifice this artifact: Destroy each nonland permanent with mana value equal to the number of charge counters on this artifact.",
				Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}},
				Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
				Effect:  effects.DestroyNonlandsWithManaValueOfChargeCounters,
			},
		},
	})
}
