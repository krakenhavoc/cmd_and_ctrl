package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Commune with Beavers — Sorcery {G} (Final Fantasy, #4462):
//
//	"Look at the top three cards of your library. You may reveal an
//	 artifact, creature, or land card from among them and put it into
//	 your hand. Put the rest on the bottom of your library in any order."
//
// Commune with Dinosaurs' dig (TakeFromLibraryToHand with
// takeRevealRestBottomAny) over three cards and a wider filter: the pick
// is optional, at most one card, revealed, and the rest go to the bottom
// in an order the player chooses.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9108fffb-90f4-444e-8d3e-46a3d4ff3681",
		Name:         "Commune with Beavers",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			player := ctx.Controller()
			return TakeFromLibraryToHand{
				Player:   player,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 3),
				Match:    Or(Artifact(), Creature(), Land()),
				Max:      1,
				Optional: true,
				Label:    "Commune with Beavers — reveal an artifact, creature, or land card and put it into your hand",
				Key:      takeRevealRestBottomAny,
			}.Apply(ctx)
		},
	})
}
