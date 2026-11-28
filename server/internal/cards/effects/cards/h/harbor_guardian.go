package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Harbor Guardian — Creature — Gargoyle {2}{W}{U}, 3/4 (Mirage):
//
//	"Reach
//	 Whenever this creature attacks, defending player may draw a card."
//
// The defending player is read as the attack is declared and is the one
// asked; the draw is theirs (CR 121.2b stops a player who can't draw from
// choosing to). An attack on a planeswalker or battle names that
// permanent's controller (CR 506.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d72b69aa-fb61-441e-9b62-4534d7a32a69",
		Name:            "Harbor Guardian",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventAttack},
			AppliesTo: effects.ThisAttacked,
			Key:       "Harbor Guardian — defending player may draw a card",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Harbor Guardian — defending player may draw a card")
				item.Params.Player = effects.B17DefendingPlayer(g, ev)
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				defender := item.Params.Player
				if defender == uuid.Nil || !effects.DefendingPlayerStillIn(g, defender) {
					return nil
				}
				return effects.MayChoice{
					Player:   defender,
					Question: "Harbor Guardian — draw a card?",
					YesLabel: "Draw a card",
					NoLabel:  "Don't draw",
					Draws:    1,
					Then:     harborGuardianDraw,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

// harborGuardianDraw is the defending player's "yes": they draw. An on-disk key.
var harborGuardianDraw = effects.MayYes("harbor-guardian/draw", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	return effects.DrawCards{Player: r.Chooser, N: 1}.Apply(ctx)
})
