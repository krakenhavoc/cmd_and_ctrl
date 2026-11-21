package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nick Valentine, Private Eye — Legendary Artifact Creature — Synth
// Detective {2}{U}, 2/2 (Fallout):
//
//	"Nick Valentine can't be blocked except by artifact creatures.
//	 Whenever Nick Valentine or another artifact creature you control dies,
//	 you may investigate."
//
// The block rule is Deluxe Dragster's with artifact creatures as the only
// blockers. The death trigger reads the creature as it last existed on the
// battlefield (CR 603.10a): an artifact creature, under my control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5663288a-7300-4e1b-8fb0-43c467709fe5",
		Name:         "Nick Valentine, Private Eye",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedExceptBy(effects.OnSelf(), effects.Artifact(), "artifact creatures"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventLTB, effects.GapArtifactCreatureYouControlDied,
				"Nick Valentine, Private Eye — you may investigate",
				effects.Do(effects.Investigate{N: 1})), "Investigate?"),
		},
	})
}
