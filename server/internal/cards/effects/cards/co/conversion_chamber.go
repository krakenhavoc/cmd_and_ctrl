package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Conversion Chamber — Artifact {3} (New Phyrexia):
//
//	"{2}, {T}: Exile target artifact card from a graveyard. Put a charge
//	 counter on this artifact.
//	 {2}, {T}, Remove a charge counter from this artifact: Create a 3/3
//	 colorless Phyrexian Golem artifact creature token."
//
// The counter is a second instruction, not a reward for the exile: it goes
// on whether or not the card was still there to be exiled, so long as the
// ability resolves with a legal target (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "46af64c9-2ab5-4e63-98db-cfb885a96109",
		Name:         "Conversion Chamber",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}, {T}: Exile target artifact card from a graveyard. Put a charge counter on this artifact.",
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
				Targets: effects.TargetCardInGraveyard("target artifact card from a graveyard", effects.Artifact()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if err := effects.ExileTargetCards(g, item); err != nil {
						return err
					}
					return effects.PutChargeCounterOnThis(g, item)
				},
			},
			{
				Label:   "{2}, {T}, Remove a charge counter from this artifact: Create a 3/3 colorless Phyrexian Golem artifact creature token.",
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.RemoveCountersFromThis(game.CounterCharge, 1)),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Effect: effects.Do(effects.CreateTokenAdvanced{
					Spec: effects.Token(effects.TokenCard("3/3 colorless Phyrexian Golem artifact")),
					N:    1,
				}),
			},
		},
	})
}
