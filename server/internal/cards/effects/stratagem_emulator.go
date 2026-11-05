package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stratagem Emulator — Artifact Creature — Borg {2}{W}, 2/1 (Star Trek
// Commander, #4974):
//
//	"When this creature dies, create X 1/1 colorless Borg artifact
//	 creature tokens, where X is its power."
//
// X is the Emulator's power as it last existed on the battlefield (CR
// 603.10a, CR 608.2h), so a +1/+1 counter or a pump it died with counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "726cd303-f034-4dbe-8139-550feb59f8f3",
		Name:         "Stratagem Emulator",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Stratagem Emulator — create X 1/1 Borg artifact creature tokens, where X is its power",
				trekB33TokensEqualToItsPower(func() game.Card { return TokenCard("1/1 colorless Borg artifact") })),
		},
	})
}
