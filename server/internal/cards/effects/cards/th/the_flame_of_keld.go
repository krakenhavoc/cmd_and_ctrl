package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Flame of Keld — Enchantment — Saga {1}{R} (Dominaria):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Discard your hand.
//	 II — Draw two cards.
//	 III — If a red source you control would deal damage to a permanent or
//	 player this turn, it deals that much damage plus 2 to that permanent or
//	 player instead."
//
// I is Blood for the Blood God's whole-hand discard. III is Taii Wakeen's
// turn-long addition (a scoped damage record, ModAddDamage, CR 614) over a
// narrower source test: a source its controller controls that is red as the
// damage is dealt (read as it last existed, CR 608.2h), to any permanent or
// player, combat or not. Two Flames add 4.
//
// No simplification.
func init() {
	const name = "The Flame of Keld"
	effects.Register(effects.Spec{
		OracleID:     "26b41c78-c5ee-4c50-93be-e3acc35ab355",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "discard your hand"),
				func(g *game.Game, item *game.StackItem) error {
					_, err := effects.DiscardWholeHand(g, item.Controller)
					return err
				}),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "draw two cards"), effects.Do(effects.DrawCards{N: 2})),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "if a red source you control would deal damage to a permanent or player this turn, it deals that much damage plus 2 instead"),
				func(g *game.Game, item *game.StackItem) error {
					g.AddDamageForEffect(game.DamageAddition{
						EffectSource: item.SourceCardID,
						Controller:   item.Controller,
						Amount:       2,
						Sources:      game.DamageSourcesYoursRed,
						Recipients:   game.DamageRecipientsAny,
						Label:        name,
					})
					return nil
				}),
		},
	})
}
