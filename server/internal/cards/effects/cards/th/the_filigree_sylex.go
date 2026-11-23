package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Filigree Sylex — Legendary Artifact {2}:
//
//	"{T}: Put an oil counter on The Filigree Sylex.
//	 {T}, Sacrifice The Filigree Sylex: Destroy each nonland permanent with
//	 mana value equal to the number of oil counters on The Filigree Sylex.
//	 {T}, Remove ten oil counters from among permanents you control and
//	 sacrifice The Filigree Sylex: It deals 10 damage to any target."
//
// Blast Zone's sweep keyed on oil: the sacrifice is a cost, so the number is
// the Sylex's last-known information (CR 608.2h); a token's mana value is 0.
// The third ability removes the ten counters from among permanents you
// control (the Sylex's own included) as a cost, chosen not targeted, and the
// damage is the Sylex's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "38347ab3-ae43-4d10-b665-8520ace0eb76",
		Name:         "The Filigree Sylex",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Put an oil counter on The Filigree Sylex.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.TapCost(),
				Effect:  effects.PutACounterOnThis(effects.S82OneB12Oil),
			},
			{
				Label:   "{T}, Sacrifice The Filigree Sylex: Destroy each nonland permanent with mana value equal to the number of oil counters on The Filigree Sylex.",
				Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}},
				Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if _, ok := ctx.SourcePermanent(); !ok {
						return nil
					}
					return effects.DestroyAllMatching{Match: effects.And(effects.Nonland(), effects.B03ManaValueIs(effects.S82OneB12SourceOil(ctx)))}.Apply(ctx)
				},
			},
			{
				Label:   "{T}, Remove ten oil counters from among permanents you control and sacrifice The Filigree Sylex: It deals 10 damage to any target.",
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 10)),
				Cost: effects.Plus(effects.TapCost(),
					effects.RemoveCountersAmong(effects.S82OneB12Oil, 10, "permanents you control", effects.YouControl()),
					effects.SacrificeThis()),
				Targets: effects.TargetAny(),
				Effect:  effects.DealDamageToTheTarget(10),
			},
		},
	})
}
