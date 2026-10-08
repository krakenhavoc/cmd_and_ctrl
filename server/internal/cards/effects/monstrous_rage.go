package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Monstrous Rage — Instant {R}:
//
//	"Target creature gets +2/+0 until end of turn. Create a Monster Role
//	 token attached to it. (If you control another Role on it, put that
//	 one into the graveyard. Enchanted creature gets +1/+1 and has
//	 trample.)"
//
// The Role goes on the target only if it is still a creature on the
// battlefield; a target that left makes the spell fizzle first
// (CR 608.2b). No simplifications.
func init() {
	Register(Spec{
		OracleID:     "646a2371-54c0-4492-ac2f-20f109d6108c",
		Name:         "Monstrous Rage",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			host := item.Targets[0].ID
			if err := (BoostUntilEOT{Target: host, Power: 2, Label: "Monstrous Rage — +2/+0"}).Apply(ctx); err != nil {
				return err
			}
			return CreateRoleToken{Role: RoleMonster, Host: host}.Apply(ctx)
		},
	})
}
