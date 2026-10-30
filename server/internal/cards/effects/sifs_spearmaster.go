package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sif's Spearmaster — Creature — God Warrior Hero {2}{R}, 1/4 (Marvel Super
// Heroes Commander, #4455):
//
//	"{T}: This creature deals damage equal to its power to target opponent."
//
// The power is read as the ability resolves, from the last known
// information if the Spearmaster has left.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9ae33bde-ccc9-40e2-9957-05cd9725f8cf",
		Name:         "Sif's Spearmaster",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals damage equal to its power to target opponent.",
			Cost:    TapCost(),
			Targets: TargetPlayer("target opponent", Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					return DealDamage{
						Source: item.SourceCardID,
						Target: t.ID,
						Amount: powerNowOrLastKnown(g, item.SourceCardID),
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
