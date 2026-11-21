package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blade Historian — Creature — Human Cleric {R/W}{R/W}{R/W}{R/W}, 2/3
// (EDHREC rank 2711):
//
//	"Attacking creatures you control have double strike."
//
// Berserkers' Onslaught on a body, and the same code: a layer 6 static
// gated on Card.AttackingTarget (AttackingCreaturesHave), declared
// DependsOnAttackingStatus so the layer cache is rebuilt as attackers
// are declared, put onto the battlefield attacking, removed from combat,
// or combat ends. The Historian itself has double strike while it
// attacks. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "314f0a96-5e91-42a3-9d75-3f641f22a9ee",
		Name:         "Blade Historian",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.AttackingCreaturesHave("Attacking creatures you control have double strike.", nil, "double strike"),
		},
	})
}
