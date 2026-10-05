package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shambling Shell — Creature — Plant Zombie {1}{B}{G}, 3/1:
//
//	"Sacrifice this creature: Put a +1/+1 counter on target creature.
//	 Dredge 3 (If you would draw a card, you may mill three cards
//	 instead. If you do, return this card from your graveyard to your
//	 hand.)"
//
// The sacrifice is the cost, so the Shell is in the graveyard when the
// counter is placed, and dredge can bring it back the same turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9c50f6b1-8850-4b08-a994-3ccc17f25b83",
		Name:         "Shambling Shell",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{Dredge(3)},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this creature: Put a +1/+1 counter on target creature.",
			Cost:    SacrificeThis(),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					return AddCounter{Target: t.ID, Kind: "+1/+1", N: 1}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
