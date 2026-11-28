package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Isao, Enlightened Bushi — Legendary Creature — Human Samurai {2}{G}, 2/1
// (Betrayers of Kamigawa, #5602):
//
//	"This spell can't be countered.
//	 Bushido 2
//	 {2}: Regenerate target Samurai."
//
// "Can't be countered" is the spec's CantBeCountered, working on the stack
// (CR 113.6g); Bushido 2 is the shared keyword trigger (CR 702.45a). The
// regeneration is Mad Auntie's row, a shield on the target Samurai as the
// ability resolves (CR 701.19a); "Samurai" is a word of the text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "15ec1bfc-b0a0-4058-87bb-8dbaf17b7f3b",
		Name:            "Isao, Enlightened Bushi",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		Triggered:       []game.TriggeredAbility{effects.Bushido(2, "Isao, Enlightened Bushi")},
		Activated: []effects.ActivatedAbility{
			effects.LrwB1RegenerateTargetRow("{2}: Regenerate target Samurai.", effects.ManaCost("{2}"), effects.TargetCreature("target Samurai", effects.Subtype("Samurai"))),
		},
	})
}
