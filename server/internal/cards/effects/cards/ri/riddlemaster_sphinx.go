package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Riddlemaster Sphinx — Creature — Sphinx {4}{U}{U}, 5/5 (Core Set 2020):
//
//	"Flying
//	 When this creature enters, you may return target creature an opponent
//	 controls to its owner's hand."
//
// Separatist Voidmage's "you may" bounce, aimed at a creature an opponent
// controls: the target is chosen as the trigger goes on the stack, the
// question is asked as it resolves (CR 603.5) and a target that has gone is
// not returned (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "29f61caf-1aba-40cd-aff6-c3685a0997d7",
		Name:            "Riddlemaster Sphinx",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisEnters("Riddlemaster Sphinx — return target creature an opponent controls to its owner's hand", effects.BounceTheTargetToHand),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
				"Riddlemaster Sphinx — return target creature an opponent controls to its owner's hand?"),
		},
	})
}
