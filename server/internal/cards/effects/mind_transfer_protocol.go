package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mind Transfer Protocol — {2}{U} Instant (Teenage Mutant Ninja Turtles,
// #4457):
//
//	"Until end of turn, target artifact or creature becomes an artifact
//	 creature with base power and toughness 4/5.
//	 Draw a card."
//
// Iron Man, Futurist Paragon's change with a duration: one scoped effect
// that adds the Artifact and Creature types in layer 4 (CR 613.1d) and sets
// base 4/5 in layer 7b (CR 613.4b), so counters and pumps still apply on
// top. Nothing else about the target changes: a creature keeps its
// creature types and abilities. The draw follows. A target that is gone by
// resolution makes the whole spell do nothing (CR 608.2b), the draw
// included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0d80d0f7-a436-431c-a54e-e7dd7dbba8de",
		Name:         "Mind Transfer Protocol",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if id, ok := b16FirstLegalTargetCard(ctx); ok {
				mods := []game.Mod{game.AddTypesMod("Artifact", "Creature")}
				mods = append(mods, game.SetBasePTMods(4, 5)...)
				if err := untilEndOfTurn(ctx, id, nil,
					"Mind Transfer Protocol — an artifact creature with base power and toughness 4/5", mods...); err != nil {
					return err
				}
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
