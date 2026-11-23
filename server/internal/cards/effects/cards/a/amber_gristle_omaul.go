package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Amber Gristle O'Maul — Legendary Creature — Dwarf Cleric {3}{R}, 3/3
// (Commander Legends: Battle for Baldur's Gate):
//
//	"Haste
//	 Whenever Amber Gristle O'Maul attacks, you may discard your hand. If you do,
//	 draw a card for each player being attacked.
//	 Choose a Background (You can have a Background as a second commander.)"
//
// Astonishing Spider-Man's keyed yes/no on the attack trigger: on a yes the whole
// hand is discarded, and a player with no hand discards nothing, so "if you do"
// is not met. The draw counts the distinct players being attacked as it resolves
// (CR 506.3d), a player whose planeswalker or battle is attacked included.
// Choose a Background is a deck-building rule the deck validator applies
// (deck/partner.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d1435f7b-a694-4c7f-a0f2-34c4c646e04b",
		Name:            "Amber Gristle O'Maul",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Amber Gristle O'Maul — you may discard your hand; if you do, draw a card for each player being attacked",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Amber Gristle O'Maul — discard your hand? If you do, draw a card for each player being attacked.",
						Then:     amberGristleDiscardHand,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var amberGristleDiscardHand = effects.MayYes("confirm/gap-b218-amber-gristle-discard-hand", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	n, err := effects.DiscardWholeHand(ctx.Game, r.Chooser)
	if err != nil || n == 0 {
		return err
	}
	return effects.DrawCards{Player: r.Chooser, N: effects.GapB218PlayersBeingAttacked(ctx.Game)}.Apply(ctx)
})
