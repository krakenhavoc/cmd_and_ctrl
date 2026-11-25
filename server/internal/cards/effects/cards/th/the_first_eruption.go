package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The First Eruption — Enchantment — Saga {2}{R} (Dominaria):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — This Saga deals 1 damage to each creature without flying.
//	 II — Add {R}{R}.
//	 III — Sacrifice a Mountain. If you do, this Saga deals 3 damage to each
//	 creature."
//
// I is Earthquake's damage with the Saga as source, to every creature without
// flying as the chapter resolves, in one damage event. II adds the mana during
// the main phase. III asks you which Mountain to sacrifice (a mandatory pick
// when you control one, none asked when you don't), and only a Mountain that
// really left makes the 3 damage; the Saga is the source even if it is gone
// by then (CR 608.2h).
//
// No simplification.
func init() {
	const name = "The First Eruption"
	effects.Register(effects.Spec{
		OracleID:     "df7a3868-4f03-452c-8307-1493977d42ef",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "this Saga deals 1 damage to each creature without flying"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.DamageEachMatching(effects.NewContext(g, item), effects.And(effects.Creature(), effects.WithoutKeyword("flying")), 1)
				}),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "add {R}{R}"),
				effects.Do(effects.AddMana{Produced: "{R}{R}"})),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "sacrifice a Mountain; if you do, this Saga deals 3 damage to each creature"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.MustSacrificeOne(g, item, name+" — sacrifice a Mountain", effects.HasSubtype("Mountain"), firstEruptionSacrificed)
				}),
		},
	})
}

// firstEruptionSacrificed is "if you do": the 3 damage to each creature. An
// on-disk key: never renamed, never reused.
var firstEruptionSacrificed = effects.RegisterSacrificedOne("prompt-run/the-first-eruption-damage", func(ctx *effects.Context, _ uuid.UUID) error {
	return effects.DamageEachMatching(ctx, effects.Creature(), 3)
})
