package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hydra Trainer — Creature — Human Warrior {1}{G}, 1/1:
//
//	"You may exert this creature as it attacks. When you do, target
//	 creature gets +X/+X until end of turn, where X is the number of
//	 counters on permanents you control. (An exerted creature won't
//	 untap during your next untap step.)
//	 {2}{G}: Adapt 2. (If this creature has no +1/+1 counters on it,
//	 put two +1/+1 counters on it.)"
//
// ADR 0130 §11: exert as it attacks (CR 701.43d) with a targeted linked
// trigger (CR 607.2h). X is every counter of every kind on every
// permanent its controller controls, counted as the trigger resolves
// (CR 608.2h), so a counter put on a permanent in response counts. The
// target is any creature. Adapt checks for counters as it resolves
// (CR 701.46a), as Evolution Witness's does.
//
// No simplification.
func init() {
	const label = "Hydra Trainer — target creature gets +X/+X until end of turn, where X is the number of counters on permanents you control"
	Register(Spec{
		OracleID:      "c428cbe2-17fd-4bd2-9810-de2561519f14",
		Name:          "Hydra Trainer",
		Completeness:  CompletenessFull,
		ExertOnAttack: ExertAsItAttacks(),
		Activated: []ActivatedAbility{{
			Label:  "{2}{G}: Adapt 2",
			Cost:   ManaCost("{2}{G}"),
			Effect: adaptTwo,
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenExerted(label, func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					x := countersOnPermanentsControlledBy(g, item.Controller)
					return BoostUntilEOT{Target: id, Power: x, Toughness: x, Label: label}.Apply(ctx)
				}),
				TargetCreature("target creature")),
		},
	})
}

// countersOnPermanentsControlledBy sums every counter of every kind on
// the permanents `controller` controls: Hydra Trainer's X.
func countersOnPermanentsControlledBy(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller {
			continue
		}
		for _, v := range c.Counters {
			if v > 0 {
				n += v
			}
		}
	}
	return n
}

// adaptTwo is "Adapt 2" (CR 701.46a): if the source has no +1/+1
// counters on it as the ability resolves, put two on it.
func adaptTwo(g *game.Game, item *game.StackItem) error {
	src, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || src.Counters[game.CounterPlusOne] > 0 {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 2}.Apply(NewContext(g, item))
}
