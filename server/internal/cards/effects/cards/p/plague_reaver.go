package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Plague Reaver — Creature — Beast {2}{B}, 6/5 (Commander Legends):
//
//	"At the beginning of your end step, sacrifice each other creature you
//	 control.
//	 Discard two cards, Sacrifice this creature: Choose target opponent.
//	 Return this creature to the battlefield under that player's control at the
//	 beginning of their next upkeep."
//
// The end-step trigger sacrifices every other creature you control as one
// simultaneous exit (All Is Dust's). The ability's costs are two discards and the
// sacrifice of the Reaver; its target is the opponent chosen as it is put on
// the stack. As it resolves, a delayed trigger is set for that player's next
// upkeep (the step of their next turn; an upkeep already begun is not "next"),
// and when it fires the Reaver returns from the graveyard under that player's
// control, if it is still the card that was sacrificed (CR 400.7). It returns as a new
// object and "your end step" trigger then belongs to its new controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eef50ef8-958b-4f5c-b261-16926d9fe672",
		Name:         "Plague Reaver",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Plague Reaver — sacrifice each other creature you control", func(g *game.Game, item *game.StackItem) error {
				self := item.SourceCardID
				return effects.B08SacrificeAllMatching(effects.NewContext(g, item), func(_ *game.Game, caster uuid.UUID, c game.Card) bool {
					return c.IsCreature() && c.Controller == caster && c.InstanceID != self
				})
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard two cards, Sacrifice this creature: Choose target opponent. Return this creature to the battlefield under that player's control at the beginning of their next upkeep.",
			Cost:    effects.Plus(effects.DiscardN(2, "two cards"), game.AbilityCost{SacrificeSelf: true}),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetPlayer {
						continue
					}
					c, ok := g.LookupCardForEffect(item.SourceCardID)
					if !ok {
						return nil
					}
					if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
						return nil
					}
					return effects.ScheduleDelayedTrigger{
						At:     game.StepUpkeep,
						TurnOf: t.ID,
						Label:  "Plague Reaver — return it to the battlefield under that player's control",
						Cards:  []uuid.UUID{item.SourceCardID},
						Body:   effects.PlagueReaverReturns,
						Params: game.EffectParams{Player: t.ID, Object: game.ObjectRef{ID: item.SourceCardID, Epoch: c.ObjectEpoch}},
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
