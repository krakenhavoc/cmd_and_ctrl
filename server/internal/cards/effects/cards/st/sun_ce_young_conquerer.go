package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sun Ce, Young Conquerer — Legendary Creature — Human Soldier {3}{U}{U}, 3/3
// (Portal Three Kingdoms, #5622):
//
//	"Horsemanship (This creature can't be blocked except by creatures with horsemanship.)"
//	"When Sun Ce enters, you may return target creature to its owner's hand."
//
// Horsemanship is an engine keyword carried as printed data: the block
// check refuses any blocker without horsemanship. It doesn't stop this
// creature blocking.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "93b0185c-b86f-4137-bc6d-96212eee1708",
		Name:            "Sun Ce, Young Conquerer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"horsemanship"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisEnters("Sun Ce — return target creature to its owner's hand", effects.BounceTheTargetToHand),
				effects.TargetCreature("target creature")),
				"Sun Ce — return target creature to its owner's hand?"),
		},
	})
}
