package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tataru Taru — Legendary Creature — Dwarf Advisor {1}{W}:
//
//	"When Tataru Taru enters, you draw a card and target opponent may
//	 draw a card.
//	 Scions' Secretary — Whenever an opponent draws a card, if it
//	 isn't that player's turn, create a tapped Treasure token. This
//	 ability triggers only once each turn."
//
// The enter trigger targets the opponent, so an opponent with hexproof
// can't be picked, and it is that opponent who answers the "may" as
// the trigger resolves. The draw is an ordinary draw, so their own
// draw-watching abilities, and Tataru's, see it — though it happens on
// the controller's turn, which is exactly when Scions' Secretary pays
// out.
//
// "If it isn't that player's turn" is checked when the draw happens
// (CR 603.4's intervening if, so nothing goes on the stack for a draw
// in the drawer's own turn), and "only once each turn" is the same
// per-object tally every such card uses. The Treasure is created under
// Tataru's controller, tapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "70dd5013-f18f-4501-882f-70590c424e20",
		Name:         "Tataru Taru",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters(tataruTaruEnterLabel, tataruTaruEnter),
				TargetPlayer("target opponent", Opponent())),
			On(game.EventDrawCard, tataruTaruOpponentDrewOffTurn, tataruTaruSecretaryLabel,
				func(g *game.Game, item *game.StackItem) error {
					return b13CreateTappedTreasures(NewContext(g, item), item.Controller, 1)
				}),
		},
	})
}

const (
	tataruTaruEnterLabel     = "Tataru Taru — you draw a card and target opponent may draw a card"
	tataruTaruSecretaryLabel = "Tataru Taru — Scions' Secretary: create a tapped Treasure token"
)

// tataruTaruEnter draws for the controller, then asks the chosen
// opponent whether they draw.
func tataruTaruEnter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
		return err
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		opp := t.ID
		return MayChoice{
			Player:   opp,
			Question: "Tataru Taru — draw a card?",
			YesLabel: "Draw a card",
			NoLabel:  "Don't draw",
			OnYes: func(ctx *Context) error {
				return DrawCards{Player: opp, N: 1}.Apply(ctx)
			},
		}.Apply(ctx)
	}
	return nil
}

// tataruTaruOpponentDrewOffTurn is Scions' Secretary's condition: an
// opponent drew a card, it isn't that player's turn, and this ability
// has not triggered yet this turn.
func tataruTaruOpponentDrewOffTurn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Actor == uuid.Nil || ev.Actor == source.Controller {
		return false
	}
	return !IsYourTurn(g, ev.Actor) && !b11TriggeredThisTurn(g, source.InstanceID, tataruTaruSecretaryLabel)
}
