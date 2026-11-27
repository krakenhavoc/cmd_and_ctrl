package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Angel of Fury — Creature — Angel {4}{W}{W}, 3/5 (Portal Second Age, S82 older sets batch 734, #5625):
//
//	"Flying
//	 When this creature dies, you may shuffle it into its owner's library."
//
// Flying rides PrintedKeywords. The "may" is asked as the trigger resolves (CR 603.5); the card is shuffled in only if it is still in the graveyard as the card that died (CR 400.7), and the library is shuffled either way (CR 701.24c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "07eb70cf-e6af-4924-a00d-bdf2c6ba78e1",
		Name:            "Angel of Fury",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WhenThisDies("Angel of Fury — you may shuffle it into its owner's library", effects.S82OldB734ShuffleDiedCardIntoLibrary),
				"Angel of Fury — shuffle it into its owner's library?"),
		},
	})
}
