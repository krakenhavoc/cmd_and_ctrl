package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pick Your Poison — Sorcery {G} (EDHREC rank 3204):
//
//	"Choose one —
//	 • Each opponent sacrifices an artifact of their choice.
//	 • Each opponent sacrifices an enchantment of their choice.
//	 • Each opponent sacrifices a creature with flying of their
//	   choice."
//
// Green's one-mana edict. Three modes over the same primitive —
// EachPlayerSacrifices with the caster excepted — differing only in
// the predicate; "of their choice" is the rules content, so each
// opponent gets their own prompt offering only their own permanents
// and nothing targets. "Creature with flying" reads the creature's
// effective abilities, so a granted flying counts and a lost one
// does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9af4a832-d634-47aa-91ed-79d44fe08864",
		Name:         "Pick Your Poison",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Each opponent sacrifices an artifact of their choice."),
			effects.Mode("Each opponent sacrifices an enchantment of their choice."),
			effects.Mode("Each opponent sacrifices a creature with flying of their choice."),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			switch {
			case ctx.HasMode(0):
				return effects.EachPlayerSacrifices{ExceptController: true, Match: effects.Artifact(), Label: "an artifact"}.Apply(ctx)
			case ctx.HasMode(1):
				return effects.EachPlayerSacrifices{ExceptController: true, Match: effects.Enchantment(), Label: "an enchantment"}.Apply(ctx)
			case ctx.HasMode(2):
				return effects.EachPlayerSacrifices{ExceptController: true, Match: effects.And(effects.Creature(), effects.HasKeyword("flying")), Label: "a creature with flying"}.Apply(ctx)
			}
			return nil
		},
	})
}
