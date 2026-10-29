package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a97d6c5c-1cff-442d-b535-fc8389160b0b",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel(name, 1, "6 damage to target creature an opponent controls"),
				TargetCreature("target creature an opponent controls", OpponentControls()),
				sourceDealsDamageToEachLegalTarget(6)),
			ChapterTriggerTargeting(2, SagaChapterLabel(name, 2, "destroy target artifact an opponent controls"),
				TargetPermanent("target artifact an opponent controls", Artifact(), OpponentControls()),
				destroyFirstLegalCardTarget),
			ChapterTrigger(3, SagaChapterLabel(name, 3, addR), Do(AddMana{Produced: "{R}"})),
			ChapterTrigger(4, SagaChapterLabel(name, 4, addR), Do(AddMana{Produced: "{R}"})),
		},
	})
}
