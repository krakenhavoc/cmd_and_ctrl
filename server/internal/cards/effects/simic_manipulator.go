package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Simic Manipulator — Creature — Mutant Wizard {1}{U}{U}, 0/1:
//
//	"Evolve (Whenever a creature you control enters, if that creature
//	 has greater power or toughness than this creature, put a +1/+1
//	 counter on this creature.)
//	 {T}, Remove one or more +1/+1 counters from this creature: Gain
//	 control of target creature with power less than or equal to the
//	 number of +1/+1 counters removed this way."
//
// ADR 0109 §9 (#1842). The number of counters removed is announced
// with the activation (CR 602.2b, 601.2b) — before the target is chosen
// (CR 601.2c) — so the target clause is bounded by it at announce: a
// power bound whose input is the counters removed
// (WithPowerAtMostX().BoundByTheCountersRemoved()), which the stack
// item carries for the CR 608.2b re-check. A creature pumped past the
// count in response is an illegal target, and the ability does nothing.
//
// The theft has no stated duration, so it lasts for the rest of the
// game (CR 611.2a) — Simic Manipulator leaving the battlefield does
// not give the creature back. Evolve is the engine's keyword trigger
// (game/evolve.go, #1805).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8f06fcc9-9018-4c55-af63-c44350a6cfeb",
		Name:            "Simic Manipulator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordEvolve},
		Activated: []ActivatedAbility{{
			Label: "{T}, Remove one or more +1/+1 counters from this creature: Gain control of target creature with power less than or equal to the number of +1/+1 counters removed this way.",
			Cost:  Plus(TapCost(), RemoveCountersXFromThis(game.CounterPlusOne, 1)),
			Targets: TargetCreature("target creature with power less than or equal to the number of +1/+1 counters removed this way").
				WithPowerAtMostX().BoundByTheCountersRemoved(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return GainControl{
							Target:   t.ID,
							Duration: game.IndefiniteDuration(),
							Label:    "Simic Manipulator — gain control (no stated duration)",
						}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
