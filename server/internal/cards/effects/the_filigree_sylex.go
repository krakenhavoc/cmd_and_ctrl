package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "38347ab3-ae43-4d10-b665-8520ace0eb76",
		Name:         "The Filigree Sylex",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: Put an oil counter on The Filigree Sylex.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    TapCost(),
				Effect:  putACounterOnThis(s82OneB12Oil),
			},
			{
				Label:   "{T}, Sacrifice The Filigree Sylex: Destroy each nonland permanent with mana value equal to the number of oil counters on The Filigree Sylex.",
				Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}},
				Cost:    Plus(TapCost(), SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if _, ok := ctx.SourcePermanent(); !ok {
						return nil
					}
					return DestroyAllMatching{Match: And(Nonland(), b03ManaValueIs(s82OneB12SourceOil(ctx)))}.Apply(ctx)
				},
			},
			{
				Label:   "{T}, Remove ten oil counters from among permanents you control and sacrifice The Filigree Sylex: It deals 10 damage to any target.",
				Purpose: ForTargets(DamageToTarget(0, 10)),
				Cost: Plus(TapCost(),
					RemoveCountersAmong(s82OneB12Oil, 10, "permanents you control", YouControl()),
					SacrificeThis()),
				Targets: TargetAny(),
				Effect:  DealDamageToTheTarget(10),
			},
		},
	})
}
