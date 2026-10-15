package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Plasma Jockey — Creature — Lizard Warrior {3}{R}, 3/1:
//
//	"Whenever this creature attacks, target creature an opponent controls
//	 can't block this turn.
//	 Blitz {2}{R}"
//
// Blitz (CR 702.152, #3772). The restriction is the shared
// RestrictUntilEOT CantBlock (Breeches, Eager Pillager's mode).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "dedf0d3a-090a-4cc3-b5ee-65b6cb990b72",
		Name:             "Plasma Jockey",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Blitz("{2}{R}")},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("Plasma Jockey — target creature can't block this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBlock}.Apply(ctx)
						}
					}
					return nil
				}), TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
