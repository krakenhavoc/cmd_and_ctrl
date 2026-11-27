package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elvish Berserker — Creature — Elf Berserker {G}, 1/1 (Exodus, S82 OLD b740,
// #5626):
//
//	"Whenever this creature becomes blocked, it gets +1/+1 until end of turn
//	 for each creature blocking it."
//
// Berserk Murlodont's trigger on this creature alone: the pump counts the
// creatures blocking it as the ability resolves (CR 608.2b), so a blocker that
// has left combat by then is not counted, and a creature that has left the
// battlefield is not pumped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c9046281-dedb-4246-8e5d-7786fe0c41c7",
		Name:         "Elvish Berserker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, effects.ThisBecameBlockedEvent,
				"Elvish Berserker — it gets +1/+1 until end of turn for each creature blocking it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if info, ok := ctx.TriggeringPermanent(); !ok || info.Left {
						return nil
					}
					id := ctx.Trigger().Object.Ref().ID
					n := g.CreaturesBlockingForEffect(id)
					return effects.BoostUntilEOT{Target: id, Power: n, Toughness: n, Label: "Elvish Berserker — +1/+1 for each creature blocking it"}.Apply(ctx)
				}),
		},
	})
}
