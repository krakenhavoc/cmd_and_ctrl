package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Entangling Trap — Enchantment {1}{W} (Lorwyn, S82 OLD b557):
//
//	"Whenever you clash, tap target creature an opponent controls. If you
//	 won, that creature doesn't untap during its controller's next untap
//	 step. (This ability triggers after the clash ends.)"
//
// Watches game.EventClashed for its controller, whoever's effect started the
// clash and whether or not they won. The event's Amount is 1 when the Actor
// won (CR 701.30d), which the resolution reads for the freeze.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "229b1e23-8e91-46e8-b0c3-1c5b2c9411da",
		Name:         "Entangling Trap",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventClashed, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Actor == source.Controller
				}, "Entangling Trap — tap target creature an opponent controls; if you won, it doesn't untap during its controller's next untap step",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						if item.Trigger.Event.Amount == 1 {
							return effects.B751TapFreezeTargets(ctx, uuid.Nil, "Entangling Trap — doesn't untap during its controller's next untap step")
						}
						for _, t := range ctx.LegalTargets() {
							if t.Kind != game.TargetCard {
								continue
							}
							if err := g.TapTargetForEffect(t.ID); err != nil {
								return err
							}
						}
						return nil
					}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
