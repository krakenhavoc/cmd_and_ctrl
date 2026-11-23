package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Three Blind Mice — Enchantment — Saga {2}{W} (Wilds of Eldraine, #4471):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after IV.)
//	 I — Create a 1/1 white Mouse creature token.
//	 II, III — Create a token that's a copy of target token you control.
//	 IV — Creatures you control get +1/+1 and gain vigilance until end of
//	 turn."
//
// "II, III" is two chapter abilities with one text (CR 714.2c). The copy is
// Caretaker's Talent's token copy (CR 707.2) of the target chosen as the
// chapter goes on the stack; a target that left leaves nothing to copy.
// Chapter IV locks its set as it resolves (CR 611.2c).
//
// No simplification.
func init() {
	const name = "Three Blind Mice"
	copyChapter := func(n int) game.TriggeredAbility {
		return effects.ChapterTriggerTargeting(n, effects.SagaChapterLabel(name, n, "create a token that's a copy of target token you control"),
			effects.TargetPermanent("target token you control", effects.IsTokenPredicate(), effects.YouControl()),
			func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.CreateTokenCopy{Controller: ctx.Controller(), Copy: id, N: 1}.Apply(ctx)
			})
	}
	effects.Register(effects.Spec{
		OracleID:     "e5926e14-d30a-4f81-803c-f36542df2db3",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "create a 1/1 white Mouse creature token"),
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Mouse"), N: 1})),
			copyChapter(2),
			copyChapter(3),
			effects.ChapterTrigger(4, effects.SagaChapterLabel(name, 4, "creatures you control get +1/+1 and gain vigilance until end of turn"),
				effects.Do(
					effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1, Label: name + " — +1/+1"},
					effects.GrantKeywordUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Keywords: []string{"vigilance"}, Label: name + " — vigilance"},
				)),
		},
	})
}
