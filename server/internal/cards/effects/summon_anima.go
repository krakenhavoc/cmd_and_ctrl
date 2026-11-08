package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	pain := Do(DrawCards{N: 1}, GainLife{Amount: -1})
	Register(Spec{
		OracleID:        "0990e57f-279e-4348-a049-e06187cc38d4",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "Pain — draw a card and lose 1 life"), pain),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "Pain — draw a card and lose 1 life"), pain),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "Pain — draw a card and lose 1 life"), pain),
			TriggerEdict(ChapterTrigger(4, SagaChapterLabel(name, 4, "Oblivion — each opponent sacrifices a creature and loses 3 life"),
				Do(
					EachPlayerSacrifices{ExceptController: true, Match: Creature(), Label: "a creature"},
					EachOpponentLosesLife{N: 3},
				))),
		},
	})
}
