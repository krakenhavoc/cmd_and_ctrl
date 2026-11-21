package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Game Plan — Sorcery {5}{U}:
//
//	"Assist (Another player can pay up to {5} of this spell's cost.)
//	 Each player shuffles their hand and graveyard into their library,
//	 then draws seven cards. Exile Game Plan."
//
// Echo of Eons' resolution, with the Assist keyword in front and the
// card exiling itself at the end. The shuffle-then-draw chain is
// echoOfEonsShuffleEach, reused as it stands, so a CR 903.9b prompt
// pauses it exactly as it does for Echo of Eons.
//
// "Exile Game Plan" is done first, as Expropriate does. Game Plan is
// on the stack while it resolves, in neither a hand nor a graveyard,
// so the shuffle cannot see it either way: the card ends in exile
// whatever the chain does.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "02746b0e-3ef9-46c0-9517-6c52c2905d0b",
		Name:            "Game Plan",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordAssist},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.ExileTarget{Target: ctx.Source()}).Apply(ctx); err != nil {
				return err
			}
			players := effects.TablePlayers(ctx)
			return effects.EchoOfEonsShuffleEach(ctx.Game, players, players)
		},
	})
}
