package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Evil Eye of Orms-by-Gore — Creature — Eye {4}{B}, 3/6 (Dominaria
// Remastered):
//
//	"Non-Eye creatures you control can't attack.
//	 This creature can't be blocked except by Walls."
//
// The first half is Bast's CantAttack layer 6 static widened to every
// creature its controller controls that is not an Eye, read through the
// layered subtypes (so a changeling is an Eye and may attack); the Eye
// itself is free to. The evasion is Prowler's Helm's block rule on the
// creature itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3f79780a-accd-4782-94a8-a6e71fb3ada7",
		Name:         "Evil Eye of Orms-by-Gore",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "Non-Eye creatures you control can't attack.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller &&
					!HasSubtype("Eye")(g, source.Controller, *target)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack
			},
		}},
		BlockRules: []game.BlockRule{
			CantBeBlockedExceptBy(OnSelf(), OfCreatureType("Wall"), "Walls"),
		},
	})
}
