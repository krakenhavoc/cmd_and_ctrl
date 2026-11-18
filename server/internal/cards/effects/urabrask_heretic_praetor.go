package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Urabrask, Heretic Praetor — Legendary Creature — Phyrexian Praetor
// {3}{R}{R}, 4/4 (Streets of New Capenna, #5875):
//
//	"Haste
//	 At the beginning of your upkeep, exile the top card of your
//	 library. You may play it this turn.
//	 At the beginning of each opponent's upkeep, the next time they
//	 would draw a card this turn, instead they exile the top card of
//	 their library. They may play it this turn."
//
// The first trigger is an impulse exile for Urabrask's controller.
//
// The second sets up a replacement effect on the opponent whose upkeep
// it is (CR 614.1a), lasting until end of turn and spent by the first
// draw it replaces (game.NextDrawExileTopPlayableForEffect): that draw
// is not made, and that opponent exiles the top card of their library
// and may play it this turn. The permission is theirs, not Urabrask's
// controller's. A second draw that turn is an ordinary draw, and a turn
// in which they draw nothing ends with the replacement unused. The
// replacement does not depend on Urabrask once it is made, so it still
// applies if Urabrask has left the battlefield by the draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        urabraskHereticPraetorOracleID,
		Name:            "Urabrask, Heretic Praetor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Urabrask, Heretic Praetor — exile the top card of your library; you may play it this turn",
				func(g *game.Game, item *game.StackItem) error {
					return ExileTopWithPermission{From: item.Controller, GrantTo: item.Controller, N: 1}.Apply(NewContext(g, item))
				}),
			AtEachOpponentsUpkeep("Urabrask, Heretic Praetor — the next time that player would draw a card this turn, they exile the top card of their library instead and may play it this turn",
				urabraskNextDrawExiles),
		},
	})
}

const urabraskHereticPraetorOracleID = "3fe61a65-b1c0-4d15-b379-e4ce22c96941"

// urabraskNextDrawExiles is the opponent's-upkeep trigger: the player
// whose upkeep it is (the trigger's event Actor) gets the replacement.
func urabraskNextDrawExiles(g *game.Game, item *game.StackItem) error {
	player := NewContext(g, item).Trigger().Event.Actor
	if player == uuid.Nil {
		return nil
	}
	g.NextDrawExileTopPlayableForEffect(item.SourceCardID, item.Controller, player,
		"Urabrask, Heretic Praetor — the next draw this turn exiles the top card instead")
	return nil
}
