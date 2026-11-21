package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Summon: Anima — Enchantment Creature — Saga Horror {4}{B}{B}, 4/4
// (Final Fantasy, #4462):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after IV.)
//	 I, II, III — Pain — You draw a card and you lose 1 life.
//	 IV — Oblivion — Each opponent sacrifices a creature of their choice
//	 and loses 3 life.
//	 Menace"
//
// A Saga creature (Summon: Ixion's shape). Chapters I–III are Greta's
// draw-then-lose body. Chapter IV is Kaya's Guile's edict over the
// opponents, then each opponent loses 3 life once every sacrifice has
// been chosen (Do waits on the edict's prompts, CR 608.2c). An opponent
// with no creature still loses the life.
//
// No simplification.
func init() {
	const name = "Summon: Anima"
	pain := effects.Do(effects.DrawCards{N: 1}, effects.GainLife{Amount: -1})
	effects.Register(effects.Spec{
		OracleID:        "0990e57f-279e-4348-a049-e06187cc38d4",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "Pain — draw a card and lose 1 life"), pain),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "Pain — draw a card and lose 1 life"), pain),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "Pain — draw a card and lose 1 life"), pain),
			effects.TriggerEdict(effects.ChapterTrigger(4, effects.SagaChapterLabel(name, 4, "Oblivion — each opponent sacrifices a creature and loses 3 life"),
				effects.Do(
					effects.EachPlayerSacrifices{ExceptController: true, Match: effects.Creature(), Label: "a creature"},
					effects.EachOpponentLosesLife{N: 3},
				))),
		},
	})
}
