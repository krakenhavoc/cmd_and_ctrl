package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Royal Treatment — Instant {G}:
//
//	"Target creature you control gains hexproof until end of turn.
//	 Create a Royal Role token attached to that creature. (If you
//	 control another Role on it, put that one into the graveyard.
//	 Enchanted creature gets +1/+1 and has ward {1}.)"
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fd0f8fdc-4065-41f4-b8fb-ecb8f185774d",
		Name:         "Royal Treatment",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "hexproof")),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			host := item.Targets[0].ID
			if err := (effects.GrantKeywordUntilEOT{Target: host, Keywords: []string{"hexproof"}}).Apply(ctx); err != nil {
				return err
			}
			return effects.CreateRoleToken{Role: effects.RoleRoyal, Host: host}.Apply(ctx)
		},
	})
}
