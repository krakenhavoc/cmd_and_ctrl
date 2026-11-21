package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vantress Transmuter // Croaking Curse — Creature — Human Wizard
// {3}{U}, 3/4 // Sorcery — Adventure {1}{U} (CR 715):
//
//	Vantress Transmuter — no rules text.
//	Croaking Curse — "Tap target creature. Create a Cursed Role token
//	 attached to it. (Enchanted creature is 1/1.)"
//
// Both faces register under the Adventure keyspace (bare oracle ID and
// "#1"). The Role goes on the creature just tapped, whoever controls
// it; the Role is the caster's.
//
// No simplifications.
const vantressTransmuterOracleID = "89ae3475-9063-4700-a7bd-47d5dcd43a3b"

func init() {
	effects.Register(effects.Spec{
		OracleID:     vantressTransmuterOracleID,
		Name:         "Vantress Transmuter",
		Completeness: effects.CompletenessFull,
	})
	effects.Register(effects.Spec{
		OracleID:     vantressTransmuterOracleID + "#1",
		Name:         "Croaking Curse",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.TapTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return effects.CreateRoleToken{Role: effects.RoleCursed, Host: id}.Apply(ctx)
		},
	})
}
