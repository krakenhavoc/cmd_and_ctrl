package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vendilion Clique — Legendary Creature — Faerie Wizard {1}{U}{U}, 3/1 (Jumpstart 2022, #5948):
//
//	"Flash
//	 Flying
//	 When Vendilion Clique enters, look at target player's hand. You may
//	 choose a nonland card from it. If you do, that player reveals the
//	 chosen card, puts it on the bottom of their library, then draws a
//	 card."
//
// Oildeep Gearhulk's look-at-a-hand pick, narrowed to nonland cards and
// with the chosen card kept in the hand (Keep) so the continuation can do
// the rest in the printed order: the player reveals it (everyone sees it),
// puts it on the bottom of their library (the shared tuck, so a commander
// may go to the command zone), then draws a card, through a registered
// tail. Choosing nothing does nothing (CR 608.2c): the hand was only looked
// at.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "244d4807-0802-41bc-9460-55ac38a28a72",
		Name:            "Vendilion Clique",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Vendilion Clique — look at target player's hand; you may choose a nonland card from it; that player reveals it, puts it on the bottom of their library, then draws a card",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						return effects.ChooseFromRevealedHand{
							Player:   effects.TargetedPlayer(ctx),
							Filter:   effects.Nonland(),
							Label:    "nonland card",
							Look:     true,
							Optional: true,
							Keep:     true,
							Then:     vendilionCliqueTuck,
						}.Apply(ctx)
					}),
				effects.TargetPlayer("target player")),
		},
	})
}

// vendilionCliqueDraw is "then draws a card", for the tail's Player (the
// player whose card was put on the bottom). An on-disk key.
var vendilionCliqueDraw = effects.CardTail("then/vendilion-clique-draw", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.DrawCards{Player: o.Tail.Player, N: 1}.Apply(ctx)
})

// vendilionCliqueTuck is the chosen card's reveal, tuck and draw. An
// on-disk key.
var vendilionCliqueTuck = effects.RevealedPickThen("revealed-pick/vendilion-clique-tuck", func(ctx *effects.Context, pick game.RevealedPick) error {
	if len(pick.Chosen) == 0 {
		return nil
	}
	id := pick.Chosen[0].InstanceID
	if err := (effects.RevealCards{Player: pick.FromPlayer, Cards: []uuid.UUID{id}, Reason: "Vendilion Clique — the chosen card"}).Apply(ctx); err != nil {
		return err
	}
	tail := ctx.Tail(vendilionCliqueDraw)
	tail.Player = pick.FromPlayer
	return effects.PutIntoLibrary{Card: id, ToBottom: true, ThenTail: tail}.Apply(ctx)
})
