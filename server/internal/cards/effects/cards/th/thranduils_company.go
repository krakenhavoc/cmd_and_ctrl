package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thranduil's Company — Creature — Elf Soldier {2}{G}{U}, 3/4 (The Hobbit,
// HOB):
//
//	"As long as you control another Elf, you may play an additional land
//	 on each of your turns.
//	 Landfall — Whenever a land you control enters, put two +1/+1 counters
//	 on target creature you control. It gains vigilance until end of turn."
//
// The extra land is Exploration's derived allowance (AdditionalLandPlays)
// made conditional: AdditionalLandPlaysNeedsAnother names the Elf, and
// EffectiveLandDropsLocked counts it only while the controller controls a
// second Elf (the Company is an Elf itself and does not count), read live on
// every land play, so it comes and goes with the other Elf and a changeling
// counts. The landfall trigger targets as it goes on the stack; the counters
// may ask for a replacement order first (CR 616.1), so the vigilance is a
// second step that waits for that prompt (CR 608.2c). A target that has left
// by resolution is skipped (CR 608.2b).
//
// No simplification. The Hobbit (#4454).
func init() {
	effects.Register(effects.Spec{
		OracleID:                        "2c195485-83f9-424f-854a-50c4f5b04182",
		Name:                            "Thranduil's Company",
		Completeness:                    effects.CompletenessFull,
		AdditionalLandPlays:             1,
		AdditionalLandPlaysNeedsAnother: "Elf",
		Purpose:                         game.Purpose{ExtraLandDrops: 1}, // #2678: the bot reads the extra drop
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Landfall("Thranduil's Company — put two +1/+1 counters on target creature you control; it gains vigilance until end of turn",
				effects.Do(
					effects.Instruction(func(ctx *effects.Context) error {
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
					}),
					effects.Instruction(func(ctx *effects.Context) error {
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"vigilance"},
							Label: "Thranduil's Company — vigilance"}.Apply(ctx)
					}))),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
