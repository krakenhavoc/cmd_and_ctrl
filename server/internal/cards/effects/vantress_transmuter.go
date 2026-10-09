package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     vantressTransmuterOracleID,
		Name:         "Vantress Transmuter",
		Completeness: CompletenessFull,
	})
	Register(Spec{
		OracleID:     vantressTransmuterOracleID + "#1",
		Name:         "Croaking Curse",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return CreateRoleToken{Role: RoleCursed, Host: id}.Apply(ctx)
		},
	})
}
