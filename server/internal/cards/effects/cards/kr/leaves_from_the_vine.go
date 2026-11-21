package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Leaves from the Vine — Enchantment — Saga {1}{G}:
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Mill three cards, then create a Food token.
//	 II — Put a +1/+1 counter on each of up to two target creatures you
//	 control.
//	 III — Draw a card if there's a creature or Lesson card in your
//	 graveyard."
//
// Chapter I is the mill and The Witch's Vanity's Food token. Chapter II
// takes up to two targets as it goes on the stack (CR 603.3d) and puts a
// counter on each one still legal as it resolves (CR 608.2b). Chapter
// III reads the graveyard as it resolves, so the cards chapter I milled
// count. Lesson is the spell subtype on the printed type line
// (isLessonCard, learn.go); a creature card is read off the card's own
// type line in the graveyard.
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	const name = "Leaves from the Vine"
	effects.Register(effects.Spec{
		OracleID:     "7c4abd72-29b0-4c82-9e67-b481cf7e8775",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "mill three cards, then create a Food token"),
				effects.Do(effects.MillCards{N: 3}, effects.CreateToken{Template: effects.FoodToken(), N: 1})),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel(name, 2, "put a +1/+1 counter on each of up to two target creatures you control"),
				effects.TargetCreature("up to two target creatures you control", effects.YouControl()).WithCount(0, 2),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					var ids []uuid.UUID
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							ids = append(ids, t.ID)
						}
					}
					return effects.B13PutCounterOnEach(ctx, ids)
				}),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "draw a card if there's a creature or Lesson card in your graveyard"),
				func(g *game.Game, item *game.StackItem) error {
					if !leavesFromTheVineGraveyardHasOne(g, item.Controller) {
						return nil
					}
					return effects.DrawCards{N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// leavesFromTheVineGraveyardHasOne is "there's a creature or Lesson card
// in your graveyard".
func leavesFromTheVineGraveyardHasOne(g *game.Game, player uuid.UUID) bool {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return false
	}
	for _, c := range p.Graveyard.Cards {
		if c.IsCreature() || effects.IsLessonCard(c) {
			return true
		}
	}
	return false
}
