package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sibilant Spirit — Creature — Spirit {5}{U}, 5/6 (Ice Age, S82 OLD batch 837,
// #5636):
//
//	"Flying
//	 Whenever this creature attacks, defending player may draw a card."
//
// Mtenda Lion's defending player (the player it attacks, or whoever controls the
// planeswalker or protects the battle it attacks), asked as the trigger resolves.
// The question goes to that player and the draw is theirs (CR 121.3a); a player
// who can't draw because of a draw limit is not asked (MayChoice.Draws).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9798d307-a44a-4f44-94c7-03e499b94962",
		Name:            "Sibilant Spirit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclared(ev, source)
			}, "Sibilant Spirit — defending player may draw a card",
				func(g *game.Game, item *game.StackItem) error {
					defender := g.DefendingPlayerForAttackerForEffect(item.SourceCardID)
					if g.PlayerByIDForEffect(defender) == nil {
						return nil
					}
					return effects.MayChoice{
						Player:   defender,
						Question: "Sibilant Spirit — draw a card?",
						Draws:    1,
						Drawer:   defender,
						Then:     sibilantSpiritDraw,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// sibilantSpiritDraw is the defending player's answer, keyed (#4264): the
// chooser is the defending player, who draws.
var sibilantSpiritDraw = effects.MayAnswer("confirm/sibilant-spirit-draw", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if !r.Accepted {
		return nil
	}
	return effects.DrawCards{Player: r.Chooser, N: 1}.Apply(ctx)
})
