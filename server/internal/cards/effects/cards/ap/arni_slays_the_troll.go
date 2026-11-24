package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arni Slays the Troll — Enchantment — Saga {R}{G} (Kaldheim, #5505):
//
//	"I — Target creature you control fights up to one target creature you
//	 don't control.
//	 II — Add {R}. Put two +1/+1 counters on up to one target creature you
//	 control.
//	 III — You gain life equal to the greatest power among creatures you
//	 control."
//
// Chapter I is Bridgeworks Battle's two clauses: your creature, then "up to
// one" creature you don't control, and the fight only happens while both are
// still legal (CR 701.14b). Chapter II adds the mana first (it is added when
// the chapter resolves, whatever the target does, so a chapter with no
// creature to name still makes {R}), then the counters on the target if it is
// still legal. Chapter III reads the greatest layered power among the
// creatures you control as it resolves.
//
// No simplification.
func init() {
	const name = "Arni Slays the Troll"
	effects.Register(effects.Spec{
		OracleID:     "7cb47a7e-fe60-45a2-9c07-bd5c6ddca82d",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(name, 1, "target creature you control fights up to one target creature you don't control"),
				effects.Clauses(
					effects.TargetCreature("target creature you control", effects.YouControl()),
					effects.TargetCreature("up to one target creature you don't control", effects.OpponentControls()).WithCount(0, 1),
				),
				effects.SlotZeroFightsSlotOne),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel(name, 2, "add {R}; put two +1/+1 counters on up to one target creature you control"),
				effects.TargetCreature("up to one target creature you control", effects.YouControl()).WithCount(0, 1),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.AddMana{Produced: "{R}"}).Apply(ctx); err != nil {
						return err
					}
					t, ok := ctx.ClauseTarget(0)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
				}),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "you gain life equal to the greatest power among creatures you control"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.GainLife{Amount: effects.B42GreatestPowerControlledBy(g, item.Controller)}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
