package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Malamet Brawler — Creature — Cat Warrior {1}{G}, 2/2 (The Lost Caverns
// of Ixalan, #4470):
//
//	"Whenever this creature attacks, target attacking creature gains
//	 trample until end of turn."
//
// The target is chosen as the trigger goes on the stack (CR 603.3d) from
// the creatures attacking then, the Brawler included, and re-checked as
// it resolves (CR 608.2b): a creature removed from combat in response is
// no longer an attacking creature and gains nothing. Trample is CR
// 702.19.
//
// No simplification.
func init() {
	const label = "Malamet Brawler — target attacking creature gains trample until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "06905dfb-b2df-4849-8848-fd3c964aa8c9",
		Name:         "Malamet Brawler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks(label, effects.JmpB7FirstLegalTargetGains(label, "trample")),
				effects.TargetCreature("target attacking creature", effects.AttackingCreature())),
		},
	})
}
