package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Summon: Magus Sisters — Enchantment Creature — Saga Faerie {4}{G}, 5/5:
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I, II, III — Choose one at random —
//	 • Combine Powers! — Put three +1/+1 counters on target creature.
//	 • Defense! — Put a shield counter on target creature. You gain 3
//	   life.
//	 • Fight! — This creature fights up to one target creature an
//	   opponent controls.
//	 Haste"
//
// A Saga creature whose three chapters are the same ability, each a
// ChooseOneAtRandom trigger (CR 700.2, #3399): the engine draws the
// bullet as the chapter goes on the stack, then asks the drawn bullet's
// target. Defense!'s shield counter is CR 122.1c's (#2000). Its target
// is the ability's only one, so with it gone the ability does nothing,
// the life included (CR 608.2b). Fight! with no creature picked does
// nothing, and a Saga that has left deals and takes no damage (CR
// 701.14b).
//
// No simplification.
func init() {
	const name = "Summon: Magus Sisters"
	chapter := func(n int) game.TriggeredAbility {
		t := effects.ChapterTrigger(n, effects.SagaChapterLabel(name, n, "choose one at random: three +1/+1 counters, a shield counter and 3 life, or a fight"),
			func(*game.Game, *game.StackItem) error { return nil })
		t.Modes = effects.ChooseOneAtRandom(
			effects.ModeDoing("Combine Powers! — Put three +1/+1 counters on target creature.", effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					tgt, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.AddCounter{Target: tgt.ID, Kind: game.CounterPlusOne, N: 3}.Apply(ctx)
				}),
			// The life waits for a prompt the counter asks (CR 608.2c, #4155).
			effects.ModeSteps("Defense! — Put a shield counter on target creature. You gain 3 life.", effects.TargetCreature("target creature"),
				effects.Instruction(func(ctx *effects.Context) error {
					tgt, ok := effects.ModeTarget(ctx, ctx.Occurrence())
					if !ok {
						return nil
					}
					return effects.AddCounter{Target: tgt.ID, Kind: game.CounterShield, N: 1}.Apply(ctx)
				}),
				effects.GainLife{Amount: 3}),
			effects.ModeDoing("Fight! — This creature fights up to one target creature an opponent controls.",
				effects.TargetCreature("up to one target creature an opponent controls", effects.OpponentControls()).WithCount(0, 1),
				func(item *game.StackItem, ctx *effects.Context, occ int) error {
					tgt, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.B10Fight(ctx, item.SourceCardID, tgt.ID)
				}),
		)
		return t
	}
	effects.Register(effects.Spec{
		OracleID:        "06f11ff8-cc99-411d-b00b-464ba4b7b87d",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered:       []game.TriggeredAbility{chapter(1), chapter(2), chapter(3)},
	})
}
