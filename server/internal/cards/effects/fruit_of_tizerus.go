package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fruit of Tizerus — Sorcery {B}:
//
//	"Target player loses 2 life.
//	 Escape—{3}{B}, Exile three other cards from your graveyard."
//
// A one-mana drain that comes back, and the smallest demonstration
// of what escape does to a sorcery: the spell goes to the graveyard
// on resolution like any other, and every escape after the first is
// bought with three cards rather than with a second copy.
//
// A loss, not damage — nothing prevents it and the target's own
// "whenever you lose life" payoffs see it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "8d6ad0a0-3b71-4fab-8874-470285c39299",
		Name:             "Fruit of Tizerus",
		Completeness:     CompletenessFull,
		Targets:          TargetPlayer("target player"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Escape("{3}{B}", 3)},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetPlayer {
					continue
				}
				return ctx.Game.ChangePlayerLifeForEffect(item.SourceCardID, t.ID, -2)
			}
			return nil
		},
	})
}
