package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Princess Takes Flight — Enchantment — Saga {2}{W} (Wilds of Eldraine,
// #4471):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Exile up to one target creature.
//	 II — Target creature you control gets +2/+2 and gains flying until
//	 end of turn.
//	 III — Return the exiled card to the battlefield under its owner's
//	 control."
//
// Chapter I's exile is linked to this Saga object (CR 607.2a), read back
// by chapter III exactly as The Measure of a Man does; a Saga that left
// and came back has no claim on the card. Chapter II is Thwip!'s pump and
// flying.
//
// No simplification.
func init() {
	const name = "The Princess Takes Flight"
	effects.Register(effects.Spec{
		OracleID:     "70d52457-1867-482c-ba7a-1f7dacc4514f",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(name, 1, "exile up to one target creature"),
				effects.TargetCreature("up to one target creature").WithCount(0, 1),
				effects.MeasureOfAManExile),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel(name, 2, "target creature you control gets +2/+2 and gains flying until end of turn"),
				effects.TargetCreature("target creature you control", effects.YouControl()),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					if err := (effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: name + " — +2/+2"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: name + " — flying"}.Apply(ctx)
				}),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "return the exiled card to the battlefield under its owner's control"),
				effects.Do(effects.Instruction(effects.MeasureOfAManReturn))),
		},
	})
}
