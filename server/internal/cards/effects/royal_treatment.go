package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Royal Treatment — Instant {G}:
//
//	"Target creature you control gains hexproof until end of turn.
//	 Create a Royal Role token attached to that creature. (If you
//	 control another Role on it, put that one into the graveyard.
//	 Enchanted creature gets +1/+1 and has ward {1}.)"
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "fd0f8fdc-4065-41f4-b8fb-ecb8f185774d",
		Name:         "Royal Treatment",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			host := item.Targets[0].ID
			if err := (GrantKeywordUntilEOT{Target: host, Keywords: []string{"hexproof"}}).Apply(ctx); err != nil {
				return err
			}
			return CreateRoleToken{Role: RoleRoyal, Host: host}.Apply(ctx)
		},
	})
}
