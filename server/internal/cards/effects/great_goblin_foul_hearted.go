package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "080f504f-20c8-4975-837c-d704e879d7f5",
		Name:         "Great Goblin, Foul-Hearted",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{game.ArmySubtype}, YoursOnly: true}, "trample"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Great Goblin — amass Goblins 3", Do(Amass{Subtype: "Goblin", N: 3})),
		},
	})
}
