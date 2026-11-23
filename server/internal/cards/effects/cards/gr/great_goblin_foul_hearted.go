package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Great Goblin, Foul-Hearted — Legendary Creature — Goblin Noble {3}{B}{R},
// 3/3 (The Hobbit Commander, HOC):
//
//	"Whenever Great Goblin enters or attacks, amass Goblins 3.
//	 Armies you control have trample."
//
// One ability with two conditions (Sun Titan's WhenThisEntersOrAttacks). The
// trample is a layer-6 grant to every Army creature its controller has, the
// Army the amass just made included (Army is a rules subtype here, so an Army
// that is also a Goblin or a Zombie counts).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "080f504f-20c8-4975-837c-d704e879d7f5",
		Name:         "Great Goblin, Foul-Hearted",
		Purpose:      game.Purpose{Amass: 3},
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{game.ArmySubtype}, YoursOnly: true}, "trample"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks("Great Goblin — amass Goblins 3", effects.Do(effects.Amass{Subtype: "Goblin", N: 3})),
		},
	})
}
