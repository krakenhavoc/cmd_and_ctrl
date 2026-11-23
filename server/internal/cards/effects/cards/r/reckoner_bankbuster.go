package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reckoner Bankbuster — Artifact — Vehicle {2}, 4/4 (Kamigawa: Neon
// Dynasty, S82):
//
//	"This Vehicle enters with three charge counters on it.
//	 {2}, {T}, Remove a charge counter from this Vehicle: Draw a card. Then
//	 if there are no charge counters on this Vehicle, create a Treasure
//	 token and a 1/1 colorless Pilot creature token with "This token crews
//	 Vehicles as though its power were 2 greater."
//	 Crew 3"
//
// The counter is a cost, so the count the "if" reads is the Bankbuster's as
// the ability resolves (CR 608.2h: its last-known information if it has left,
// which is the count after the cost). The Pilot is the shared Pilot token of
// Back on Track.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b6be94dc-7162-4887-b959-3bda709724eb",
		Name:         "Reckoner Bankbuster",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterCharge, 3, "Reckoner Bankbuster: enters with three charge counters"),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}, {T}, Remove a charge counter from this Vehicle: Draw a card. Then if there are no charge counters on this Vehicle, create a Treasure token and a 1/1 colorless Pilot creature token.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.RemoveCountersFromThis(game.CounterCharge, 1)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
						return err
					}
					info, ok := ctx.SourcePermanent()
					if !ok || info.Counters[game.CounterCharge] != 0 {
						return nil
					}
					if err := (effects.CreateToken{Template: effects.TreasureToken(), N: 1}).Apply(ctx); err != nil {
						return err
					}
					return effects.CreateToken{Template: effects.PilotToken(), N: 1}.Apply(ctx)
				},
			},
			{
				Label:   "Crew 3",
				Cost:    effects.CrewCost(3),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  effects.CrewEffect("Reckoner Bankbuster"),
			},
		},
	})
}
