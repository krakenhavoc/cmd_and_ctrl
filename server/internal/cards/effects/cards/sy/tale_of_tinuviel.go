package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tale of Tinúviel — Enchantment — Saga {3}{W}{W} (The Lord of the Rings:
// Tales of Middle-earth, #5948):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Target creature you control gains indestructible for as long as you
//	 control this Saga.
//	 II — Return target creature card from your graveyard to the battlefield.
//	 III — Up to two target creatures you control each gain lifelink until end
//	 of turn."
//
// I is Jin-Gitaxias' "for as long as you control this Saga" duration
// (CR 611.2b): with the Saga gone or a new object the grant never begins, and
// it ends the moment you stop controlling the Saga. II is the plain targeted
// reanimation. III names up to two creatures as it goes on the stack and
// grants each that is still legal lifelink until end of turn.
//
// No simplification.
func init() {
	const name = "Tale of Tinúviel"
	effects.Register(effects.Spec{
		OracleID:     "bce30bc5-b059-42bb-a9a4-5886f5607160",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(name, 1, "target creature you control gains indestructible for as long as you control this Saga"),
				effects.TargetCreature("target creature you control", effects.YouControl()),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					targets := ctx.LegalTargets()
					if len(targets) == 0 {
						return nil
					}
					d, ok := effects.DurationWhileYouControlSource(ctx, item.SourceCardID, item.Controller)
					if !ok {
						return nil
					}
					return effects.ScopedEffectFor{
						Target:   targets[0].ID,
						Mods:     []game.Mod{game.AddKeywordsMod("indestructible")},
						Duration: d,
						Label:    name + " — indestructible for as long as you control this Saga",
					}.Apply(ctx)
				}),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel(name, 2, "return target creature card from your graveyard to the battlefield"),
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
				effects.SheoldredReanimate),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel(name, 3, "up to two target creatures you control each gain lifelink until end of turn"),
				effects.TargetCreature("up to two target creatures you control", effects.YouControl()).WithCount(0, 2),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"lifelink"}, Label: name + " — lifelink"}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
