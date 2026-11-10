package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moxite Refinery — Artifact {2} (Edge of Eternities Commander):
//
//	"{2}, {T}, Remove X counters from an artifact or creature you control:
//	 Choose one. Activate only as a sorcery.
//	 • Put X charge counters on target artifact.
//	 • Put X +1/+1 counters on target creature."
//
// A modal activated ability is one ability per bullet with the same cost
// (Goblin Cratermaker's shape). The cost is the new any-kind, one-permanent
// removal (RemoveCountersXFromOne): the activator announces X and names the
// artifact or creature and the counters of each kind it gives, paid at
// announce so a response cannot spend them twice, and X reaches the effect
// as the paid count (Context.CountersRemoved). The permanent may be the
// target of the bullet (counters moved from a creature to itself change
// kinds), and a target gone by resolution gets nothing.
//
// No simplification.
func init() {
	cost := Plus(ManaCost("{2}"), TapCost(), RemoveCountersXFromOne("an artifact or creature you control", Or(Artifact(), Creature())))
	Register(Spec{
		OracleID:     "3fe9367b-def1-4820-9e23-92ce3d7a4afc",
		Name:         "Moxite Refinery",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:        "{2}, {T}, Remove X counters from an artifact or creature you control: Put X charge counters on target artifact. Activate only as a sorcery.",
				Cost:         cost,
				Targets:      TargetPermanent("target artifact", Artifact()),
				SorcerySpeed: true,
				Effect:       cmdrB9XCountersOnTarget(game.CounterCharge),
			},
			{
				Label:        "{2}, {T}, Remove X counters from an artifact or creature you control: Put X +1/+1 counters on target creature. Activate only as a sorcery.",
				Cost:         cost,
				Targets:      TargetCreature("target creature"),
				SorcerySpeed: true,
				Effect:       cmdrB9XCountersOnTarget(game.CounterPlusOne),
			},
		},
	})
}

// cmdrB9XCountersOnTarget puts X counters (the count the cost removed) of
// the kind on the still-legal target.
func cmdrB9XCountersOnTarget(kind string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		x := ctx.CountersRemoved()
		if x <= 0 {
			return nil
		}
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				return AddCounter{Target: t.ID, Kind: kind, N: x}.Apply(ctx)
			}
		}
		return nil
	}
}
