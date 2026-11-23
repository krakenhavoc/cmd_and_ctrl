package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kitesail Cleric — Creature — Kor Cleric {W}, 1/1 (Zendikar Rising,
// #5508):
//
//	"Kicker {2}{W} (You may pay an additional {2}{W} as you cast this
//	 spell.)
//	 Flying
//	 When this creature enters, if it was kicked, tap up to two target
//	 creatures."
//
// Flying is printed. The kicked enters trigger is Heartstabber
// Mosquito's intervening if (CR 603.4); the tap is Eddymurk Crab's over
// up to two creatures, a target gone by resolution skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0b9bb0d7-4139-46b6-a058-2f79be64aef2",
		Name:            "Kitesail Cleric",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		OptionalCosts:   []game.AdditionalCost{effects.Kicker("{2}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
					"Kitesail Cleric — kicked, tap up to two target creatures",
					effects.TapEachLegalTarget),
				effects.TargetCreature("up to two target creatures").WithCount(0, 2),
			),
		},
	})
}
