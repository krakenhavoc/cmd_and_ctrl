package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sinew Dancer — Creature — Phyrexian Soldier {W}, 1/1:
//
//	"{3}{W}, {T}: Tap target creature.
//	 Corrupted — {W}, {T}: Tap target creature. Activate only if an opponent
//	 has three or more poison counters."
//
// Two separate abilities. The cheaper one carries the corrupted activation
// condition (CR 602.1b), read as it is announced.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a605864a-f46f-4ebc-8e56-ce7322c918b0",
		Name:         "Sinew Dancer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{3}{W}, {T}: Tap target creature.",
				Cost:    Plus(ManaCost("{3}{W}"), TapCost()),
				Targets: TargetCreature("target creature"),
				Effect:  sinewDancerTap,
			},
			{
				Label:     "Corrupted — {W}, {T}: Tap target creature. Activate only if an opponent has three or more poison counters.",
				Cost:      Plus(ManaCost("{W}"), TapCost()),
				Targets:   TargetCreature("target creature"),
				Condition: corruptedAnd(nil),
				Effect:    sinewDancerTap,
			},
		},
	})
}

func sinewDancerTap(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return TapTarget{Target: t.ID}.Apply(ctx)
		}
	}
	return nil
}
