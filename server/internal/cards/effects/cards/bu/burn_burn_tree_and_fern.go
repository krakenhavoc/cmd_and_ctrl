package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Burn, Burn, Tree and Fern — Enchantment — Saga {3}{R} (The Hobbit, HOB):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after IV.)
//	 I — This Saga deals 6 damage to target creature an opponent controls.
//	 II — Destroy target artifact an opponent controls.
//	 III, IV — Add {R}."
//
// Chapters I and II announce their target as they go on the stack (CR
// 603.3d, ChapterTriggerTargeting); a chapter with nothing to target is
// dropped and the Saga still advances. The damage is dealt by the Saga
// itself. III and IV are two entries with the same text, as printed, so a
// Saga that skips a chapter fires exactly the ones it crossed. The mana is
// added as the chapter resolves, in the precombat main phase.
//
// No simplification.
func init() {
	const name = "Burn, Burn, Tree and Fern"
	const addR = "add {R}"
	effects.Register(effects.Spec{
		OracleID:     "a97d6c5c-1cff-442d-b535-fc8389160b0b",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(name, 1, "6 damage to target creature an opponent controls"),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
				effects.SourceDealsDamageToEachLegalTarget(6)),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel(name, 2, "destroy target artifact an opponent controls"),
				effects.TargetPermanent("target artifact an opponent controls", effects.Artifact(), effects.OpponentControls()),
				effects.DestroyFirstLegalCardTarget),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, addR), effects.Do(effects.AddMana{Produced: "{R}"})),
			effects.ChapterTrigger(4, effects.SagaChapterLabel(name, 4, addR), effects.Do(effects.AddMana{Produced: "{R}"})),
		},
	})
}
