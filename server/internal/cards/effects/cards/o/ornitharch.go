package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ornitharch — {3}{W}{W} Creature — Archon 3/3 (Born of the Gods):
//
//	"Flying
//	 Tribute 2 (As this creature enters, an opponent of your choice may put two
//	 +1/+1 counters on it.)
//	 When this creature enters, if tribute wasn't paid, create two 1/1 white Bird
//	 creature tokens with flying."
//
// Tribute (CR 702.104) and its intervening "if" (CR 603.4) are the shared seam.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "19639595-c59b-4233-a167-f5cffcb0e38b",
		Name:            "Ornitharch",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    effects.Tribute("Ornitharch", 2),
		Triggered: []game.TriggeredAbility{
			effects.WhenEntersIfTributeWasntPaid("Ornitharch — create two 1/1 white Bird creature tokens with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Bird with flying"), N: 2})),
		},
	})
}
