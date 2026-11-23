package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fall of Gil-galad — Enchantment — Saga {1}{G} (The Lord of the Rings:
// Tales of Middle-earth, #5948):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Scry 2.
//	 II — Put two +1/+1 counters on target creature you control.
//	 III — Until end of turn, target creature you control gains "When this
//	 creature dies, draw two cards." Then that creature fights up to one
//	 other target creature."
//
// Chapter III has two target clauses chosen as it goes on the stack: the
// creature you control (required) and up to one other creature. The grant is
// Perigee Beckoner's duration grant of a dies trigger, pinned to the target
// as it resolves (CR 611.2c). The fight is Amok Time's slot-zero-fights-
// slot-one (CR 701.14b): with the second clause empty, or either creature no
// longer a legal target, nothing fights, but the grant is still made if the
// first target is legal.
//
// No simplification.
func init() {
	const name = "Fall of Gil-galad"
	effects.Register(effects.Spec{
		OracleID:     "c248537f-e28f-47cb-9e47-a6eb05e1af63",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Grants: []effects.AbilityGrant{{
			Key: fallOfGilGaladDraw,
			Triggered: []game.TriggeredAbility{
				effects.WhenThisDies("Fall of Gil-galad — draw two cards", effects.Do(effects.DrawCards{N: 2})),
			},
			Text: "When this creature dies, draw two cards.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "scry 2"), effects.Do(effects.Scry{N: 2})),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel(name, 2, "put two +1/+1 counters on target creature you control"),
				effects.TargetCreature("target creature you control", effects.YouControl()),
				effects.GapB129PutTwoCountersOnTarget),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel(name, 3, "target creature you control gains \"When this creature dies, draw two cards\" until end of turn; then it fights up to one other target creature"),
				effects.Clauses(
					effects.TargetCreature("target creature you control", effects.YouControl()),
					effects.TargetCreature("up to one other target creature").WithCount(0, 1),
				),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					mine, ok := ctx.ClauseTarget(0)
					if !ok || mine.Kind != game.TargetCard || !ctx.IsTargetLegal(mine) {
						return nil
					}
					if err := (effects.GrantAbilitiesFor{Target: mine.ID, Keys: []string{fallOfGilGaladDraw},
						Label: "Fall of Gil-galad — dies: draw two cards"}).Apply(ctx); err != nil {
						return err
					}
					return effects.SlotZeroFightsSlotOne(g, item)
				}),
		},
	})
}

const fallOfGilGaladDraw = "fall-of-gil-galad/dies-draw-two"
