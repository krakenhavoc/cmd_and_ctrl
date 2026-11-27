package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thoughts of Ruin — Sorcery {2}{R}{R} (Saviors of Kamigawa, #5601):
//
//	"Each player sacrifices a land of their choice for each card in your
//	 hand."
//
// The count is the caster's hand as the spell resolves, so the spell itself
// is not in it (CR 608.2h). Every player, the caster included, makes one
// choice of that many lands in APNAP order (a player with fewer sacrifices
// what they have; one with none is not asked), and the choices are sacrificed
// together (CR 701.21a, 101.4). Nothing is targeted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bb9e1d12-ebcc-4a8a-baef-9f69638a51cd",
		Name:         "Thoughts of Ruin",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := effects.SokB1CardsInYourHand(ctx)
			return effects.OldB590EachPlayerSacrificesN(n, "a land", effects.Land(), "Thoughts of Ruin — choose lands to sacrifice, one for each card in the caster's hand")(ctx.Game, item)
		},
	})
}
