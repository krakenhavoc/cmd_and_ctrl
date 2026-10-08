package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ferocious Werefox // Guard Change — Creature — Elf Fox Warrior
// {3}{G}, 4/3 // Instant — Adventure {1}{G} (CR 715):
//
//	Ferocious Werefox — "Trample"
//	Guard Change — "Create a Monster Role token attached to target
//	 creature you control. (Enchanted creature gets +1/+1 and has
//	 trample.)"
//
// Both faces register under the Adventure keyspace (bare oracle ID and
// "#1"). Guard Change is an instant while it is on the stack
// (CR 715.3a), so it can be cast at instant speed.
//
// No simplifications.
const ferociousWerefoxOracleID = "fef67393-9b7b-495d-8b03-865537386d42"

func init() {
	Register(Spec{
		OracleID:        ferociousWerefoxOracleID,
		Name:            "Ferocious Werefox",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
	Register(Spec{
		OracleID:     ferociousWerefoxOracleID + "#1",
		Name:         "Guard Change",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if id, ok := b16FirstLegalTargetCard(ctx); ok {
				return CreateRoleToken{Role: RoleMonster, Host: id}.Apply(ctx)
			}
			return nil
		},
	})
}
