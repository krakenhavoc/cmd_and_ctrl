package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spawnbroker — Creature — Human Wizard {2}{U}, 1/1 (Ravnica: City of Guilds,
// S82 OLD batch 576, #5600):
//
//	"When this creature enters, you may exchange control of target creature you
//	 control and target creature with power less than or equal to that
//	 creature's power an opponent controls."
//
// A CR 603.5 "you may" trigger with two clauses: the creature you control, and an
// opponent's creature whose power is no greater than the first's
// (PowerNoGreaterThanTarget, judged as the trigger is put on the stack and
// again as it resolves, CR 601.2c, 608.2b). The exchange (CR 701.12) is all or
// nothing: if either creature is gone or no longer legal, control changes
// hands for neither. Neither half has a stated duration. The Spawnbroker may
// be the creature you control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae050bc8-f289-4484-a07f-dfe0f0a962dc",
		Name:         "Spawnbroker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WhenThisEnters("Spawnbroker — you may exchange control of target creature you control and target creature with power less than or equal to that creature's power an opponent controls",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					a, okA := ctx.ClauseTarget(0)
					b, okB := ctx.ClauseTarget(1)
					if !okA || !okB || a.Kind != game.TargetCard || b.Kind != game.TargetCard {
						return nil
					}
					return effects.ExchangeControl{A: a.ID, B: b.ID, Label: "Spawnbroker — exchange control"}.Apply(ctx)
				}),
				effects.Clauses(
					effects.TargetCreature("target creature you control", effects.YouControl()),
					effects.PowerNoGreaterThanTarget(effects.TargetCreature("target creature with power less than or equal to that creature's power an opponent controls", effects.OpponentControls()), 0))),
				"Spawnbroker — exchange control of the two creatures?"),
		},
	})
}
