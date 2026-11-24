package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Invasion of the Giants — Enchantment — Saga {U}{R} (Kaldheim, #5505):
//
//	"I — Scry 2.
//	 II — Draw a card. Then you may reveal a Giant card from your hand. When
//	 you do, this Saga deals 2 damage to target opponent or planeswalker.
//	 III — The next Giant spell you cast this turn costs {2} less to cast."
//
// Chapter II draws first, then offers every Giant card in the controller's
// hand (a changeling is a Giant too); the pick may be none. A card chosen is
// revealed and creates the CR 603.12 reflexive trigger, whose target is
// chosen as it goes on the stack and rechecked as it resolves. The damage is
// the Saga's (its last-known information if it has left). Chapter III is a
// next-spell promise over Giant spells, a CR 601.2f reduction of generic mana.
//
// No simplification.
func init() {
	const name = "Invasion of the Giants"
	effects.Register(effects.Spec{
		OracleID:     "807e35c0-8ea1-44a5-abac-51fd516aad03",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "scry 2"),
				effects.Do(effects.Scry{N: 2})),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "draw a card; then you may reveal a Giant card from your hand; when you do, this Saga deals 2 damage to target opponent or planeswalker"),
				invasionOfTheGiantsDraw),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "the next Giant spell you cast this turn costs {2} less to cast"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.GrantNextSpellPromise{From: name, Promise: game.NextSpellPromise{
						Filter: game.PermissionFilter{CreatureType: ctx.TextSubtype("Giant")},
						Reduce: 2,
						Text:   "The next Giant spell you cast this turn costs {2} less to cast.",
					}}.Apply(ctx)
				}),
		},
	})
}

// invasionOfTheGiantsDraw draws, then asks for a Giant card to reveal.
func invasionOfTheGiantsDraw(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := (effects.DrawCards{N: 1}).Apply(ctx); err != nil {
		return err
	}
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil {
		return nil
	}
	giant := ctx.TextSubtype("Giant")
	var giants []uuid.UUID
	for _, c := range p.Hand.Cards {
		if c.HasSubtype(giant) {
			giants = append(giants, c.InstanceID)
		}
	}
	if len(giants) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   ctx.Source(),
		Question: "Invasion of the Giants — you may reveal a Giant card from your hand (then it deals 2 damage to target opponent or planeswalker)",
		Cards:    giants,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneHand,
		ThenKey:  invasionOfTheGiantsRevealed,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// invasionOfTheGiantsRevealed reveals the chosen Giant and creates the
// reflexive trigger. An on-disk key: never renamed, never reused.
var invasionOfTheGiantsRevealed = effects.ItemChoice("choose-cards/invasion-of-the-giants-reveal", func(ctx *effects.Context, r game.CardsChosen) error {
	if !r.Answered || len(r.Picked) == 0 {
		return nil
	}
	if err := (effects.RevealCards{Cards: r.Picked[:1], Reason: "Invasion of the Giants — reveal a Giant card"}).Apply(ctx); err != nil {
		return err
	}
	return effects.WhenYouDo("Invasion of the Giants — deals 2 damage to target opponent or planeswalker", invasionOfTheGiantsDamage).Apply(ctx)
})

// invasionOfTheGiantsDamage is the reflexive trigger. An on-disk key.
var invasionOfTheGiantsDamage = game.ReflexiveBody("invasion-of-the-giants/damage", effects.SimpleBody(
	func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 2}.Apply(ctx)
		}
		return nil
	}),
	effects.ConstTargets(effects.TargetOpponentOrPlaneswalker))
