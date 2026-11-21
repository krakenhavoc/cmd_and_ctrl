package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Witness the Future — Sorcery {2}{U} (Innistrad: Crimson Vow):
//
//	"Target player shuffles up to four target cards from their graveyard into
//	 their library. You look at the top four cards of your library, then put
//	 one of those cards into your hand and the rest on the bottom of your
//	 library in a random order."
//
// Krosan Reclamation's first sentence at four cards: a player clause and a
// clause for up to four cards from that player's graveyard, with the shared
// TargetPlayerShufflesTheirCards body. The second sentence is the caster's,
// not the target's: the look happens after the shuffle, so a caster who
// targeted themselves may find a shuffled card on top. The look is Memory
// Deluge's: one of the four goes to the hand, the rest to the bottom in a
// random order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fc7f00ad-7cc1-4be4-9924-16531e4fb829",
		Name:         "Witness the Future",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue, Draws: 1},
		Targets: effects.Clauses(
			effects.TargetPlayer("target player"),
			effects.FromTargetPlayersGraveyard(
				effects.TargetCardInGraveyard("up to four target cards from their graveyard").WithCount(0, 4), 0),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.TargetPlayerShufflesTheirCards(ctx, ctx.Item.Targets, 0, 1)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				player := ctx.Controller()
				return effects.TakeFromLibraryToHand{
					Player: player,
					Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 4),
					Max:    1,
					Label:  "Witness the Future — put one of them into your hand and the rest on the bottom of your library in a random order",
					Key:    effects.TakeRestBottomRandom,
				}.Apply(ctx)
			}),
		),
	})
}
