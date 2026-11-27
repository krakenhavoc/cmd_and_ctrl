package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Uncovered Clues — Sorcery {2}{U} (Dragon's Maze, S82 OLD batch 403, #5564):
//
//	"Look at the top four cards of your library. You may reveal up to two instant
//	 and/or sorcery cards from among them and put the revealed cards into your
//	 hand. Put the rest on the bottom of your library in any order."
//
// Arcane Infusion's look at the top four with Duskwatch Recruiter's rest: the look
// is private; the instants and sorceries you may take are revealed as they go to
// hand, zero, one or two of them; the rest go to the bottom in an order you choose.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "942d550d-4823-4db8-8197-cbf15045f4f5",
		Name:         "Uncovered Clues",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player:   player,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 4),
				Match:    effects.Or(effects.Instant(), effects.Sorcery()),
				Max:      2,
				Optional: true,
				Label:    "Uncovered Clues — you may reveal up to two instant and/or sorcery cards and put them into your hand",
				Key:      effects.TakeRevealRestBottomAny,
			}.Apply(ctx)
		},
	})
}
