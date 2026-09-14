package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "9af4a832-d634-47aa-91ed-79d44fe08864",
		Name:         "Pick Your Poison",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Each opponent sacrifices an artifact of their choice."),
			Mode("Each opponent sacrifices an enchantment of their choice."),
			Mode("Each opponent sacrifices a creature with flying of their choice."),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			switch {
			case ctx.HasMode(0):
				return EachPlayerSacrifices{ExceptController: true, Match: Artifact(), Label: "an artifact"}.Apply(ctx)
			case ctx.HasMode(1):
				return EachPlayerSacrifices{ExceptController: true, Match: Enchantment(), Label: "an enchantment"}.Apply(ctx)
			case ctx.HasMode(2):
				return EachPlayerSacrifices{ExceptController: true, Match: And(Creature(), HasKeyword("flying")), Label: "a creature with flying"}.Apply(ctx)
			}
			return nil
		},
	})
}
