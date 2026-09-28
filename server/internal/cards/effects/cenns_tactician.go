package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cenn's Tactician — Creature — Kithkin Soldier {W}, 1/1:
//
//	"{W}, {T}: Put a +1/+1 counter on target Soldier creature.
//	 Each creature you control with a +1/+1 counter on it can block an
//	 additional creature each combat."
//
// The activated ability is an ordinary counter-placing ability over a
// subtyped target clause. The static is CanBlockAdditional (#1706)
// over a custom predicate — "you control" AND "has a +1/+1 counter" —
// read live off Card.Counters on every recompute, so a creature that
// loses its last +1/+1 counter (to a -1/-1, to Contagion Engine's
// choice of counter kind) loses the extra block in the same beat.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9c4a3ca6-dcf9-4986-81da-bcfd46414bee",
		Name:         "Cenn's Tactician",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{W}, {T}: Put a +1/+1 counter on target Soldier creature.",
			Cost:    Plus(ManaCost("{W}"), TapCost()),
			Targets: TargetCreature("target Soldier creature", HasSubtype("Soldier")),
			Effect:  cennsTacticianPutCounter,
		}},
		Static: []game.StaticAbility{
			CanBlockAdditional(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller &&
					target.IsCreature() &&
					target.Counters[game.CounterPlusOne] > 0
			}, 1),
		},
	})
}

func cennsTacticianPutCounter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	return AddCounter{Target: targets[0].ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}
