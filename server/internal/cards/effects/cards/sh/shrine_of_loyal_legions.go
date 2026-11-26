package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shrine of Loyal Legions — Artifact {2} (New Phyrexia):
//
//	"At the beginning of your upkeep and whenever you cast a white spell,
//	 put a charge counter on this artifact.
//	 {3}, {T}, Sacrifice this artifact: Create a 1/1 colorless Phyrexian Myr artifact creature token for each charge counter on this artifact."
//
// The counter clause is ONE triggered ability with two trigger conditions.
// The second ability sacrifices the Shrine as a cost, so the number of
// charge counters is its last-known information (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9696ad3d-2e17-4ab7-88ee-3c56a632b7a8",
		Name:         "Shrine of Loyal Legions",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB460ShrineCounterTrigger("Shrine of Loyal Legions", "W"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice this artifact: Create a 1/1 colorless Phyrexian Myr artifact creature token for each charge counter on this artifact.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.SacrificeThis()),
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				n := effects.OldB460ChargeCountersOnSource(g, item.SourceCardID)
				if n <= 0 {
					return nil
				}
				return effects.CreateTokenAdvanced{
					Spec: effects.Token(effects.TokenCard("1/1 colorless Phyrexian Myr artifact")),
					N:    n,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
