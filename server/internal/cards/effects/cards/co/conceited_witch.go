package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Conceited Witch // Price of Beauty — Creature — Human Warlock {2}{B},
// 2/3 // Sorcery — Adventure {B} (CR 715):
//
//	Conceited Witch — "Menace"
//	Price of Beauty — "Create a Wicked Role token attached to target
//	 creature you control. (Enchanted creature gets +1/+0. When this
//	 token is put into a graveyard, each opponent loses 1 life.)"
//
// Both faces register under the Adventure keyspace (bare oracle ID and
// "#1"). Menace is an enforced keyword; the creature entry declares it.
//
// No simplifications.
const conceitedWitchOracleID = "1c751201-24ba-4e5c-bf44-71c3e4693a92"

func init() {
	effects.Register(effects.Spec{
		OracleID:        conceitedWitchOracleID,
		Name:            "Conceited Witch",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
	})
	effects.Register(effects.Spec{
		OracleID:     conceitedWitchOracleID + "#1",
		Name:         "Price of Beauty",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
				return effects.CreateRoleToken{Role: effects.RoleWicked, Host: id}.Apply(ctx)
			}
			return nil
		},
	})
}
