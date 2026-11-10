package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
		return ChapterTriggerTargeting(n, SagaChapterLabel(name, n, "create a token that's a copy of target token you control"),
			TargetPermanent("target token you control", IsTokenPredicate(), YouControl()),
			func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return CreateTokenCopy{Controller: ctx.Controller(), Copy: id, N: 1}.Apply(ctx)
			})
	}
	Register(Spec{
		OracleID:     "e5926e14-d30a-4f81-803c-f36542df2db3",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "create a 1/1 white Mouse creature token"),
				Do(CreateToken{Template: TokenCard("1/1 white Mouse"), N: 1})),
			copyChapter(2),
			copyChapter(3),
			ChapterTrigger(4, SagaChapterLabel(name, 4, "creatures you control get +1/+1 and gain vigilance until end of turn"),
				Do(
					BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Toughness: 1, Label: name + " — +1/+1"},
					GrantKeywordUntilEOT{Match: And(Creature(), YouControl()), Keywords: []string{"vigilance"}, Label: name + " — vigilance"},
				)),
		},
	})
}
