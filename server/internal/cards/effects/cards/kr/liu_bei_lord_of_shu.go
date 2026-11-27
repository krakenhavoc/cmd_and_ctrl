package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Liu Bei, Lord of Shu — Legendary Creature — Human Soldier {3}{W}{W}, 2/4
// (Portal Three Kingdoms, #5622):
//
//	"Horsemanship (This creature can't be blocked except by creatures with horsemanship.)"
//	"Liu Bei gets +2/+2 as long as you control a permanent named Guan Yu, Sainted Warrior or a permanent named Zhang Fei, Fierce Warrior."
//
// Horsemanship is an engine keyword carried as printed data: the block
// check refuses any blocker without horsemanship. It doesn't stop this
// creature blocking.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f46416cc-0c3d-4d90-bdcb-910c5d4a2014",
		Name:            "Liu Bei, Lord of Shu",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"horsemanship"},
		Static: []game.StaticAbility{{
			Label:     "Liu Bei gets +2/+2 as long as you control a permanent named Guan Yu, Sainted Warrior or a permanent named Zhang Fei, Fierce Warrior.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfWhileYouControlA(effects.Or(effects.GapB219NamedCard("Guan Yu, Sainted Warrior"), effects.GapB219NamedCard("Zhang Fei, Fierce Warrior"))),
			Apply:     effects.PumpBy(2, 2),
		}},
	})
}
