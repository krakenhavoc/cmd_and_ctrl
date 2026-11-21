package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Timin, Youthful Geist — Legendary Creature — Spirit {4}{U}, 3/4
// (Innistrad: Crimson Vow Commander):
//
//	"Partner with Rhoda, Geist Avenger (When this creature enters, target
//	 player may put Rhoda into their hand from their library, then
//	 shuffle.)
//	 Flying
//	 At the beginning of each combat, tap up to one target creature."
//
// Partner with is the shared entry trigger; flying is printed data. The
// combat trigger fires at the beginning of every combat, on every player's
// turn, with up to one target (zero taps nothing).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0a836332-5905-4afa-a442-47d0bcf8acea",
		Name:            "Timin, Youthful Geist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.PartnerWith("Timin, Youthful Geist", "Rhoda, Geist Avenger"),
			effects.Targeting(
				effects.AtBeginningOfEachCombat("Timin — tap up to one target creature", effects.Do(effects.Instruction(effects.VowB12TapUpToOneTarget))),
				effects.VowB12UpToOneCreature()),
		},
	})
}
