package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        conceitedWitchOracleID,
		Name:            "Conceited Witch",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
	})
	Register(Spec{
		OracleID:     conceitedWitchOracleID + "#1",
		Name:         "Price of Beauty",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if id, ok := b16FirstLegalTargetCard(ctx); ok {
				return CreateRoleToken{Role: RoleWicked, Host: id}.Apply(ctx)
			}
			return nil
		},
	})
}
