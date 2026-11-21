package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ludevic, Necro-Alchemist — Legendary Creature — Human Wizard {1}{U}{R},
// 1/4 (Commander 2016, #5537):
//
//	"At the beginning of each player's end step, that player may draw a
//	 card if a player other than you lost life this turn.
//	 Partner (You can have two commanders if both have partner.)"
//
// Mana Cache's every-end-step trigger; "that player" is the player whose
// end step it is. The condition is not an intervening "if": the ability
// triggers every end step and checks, as it resolves, whether any player
// other than Ludevic's controller lost life this turn (the turn tally
// Wound Reflection reads). If so the end step's player is asked whether
// to draw (Heartwood Storyteller's shared "may draw" answer); the
// controller may be that player, and draws only off another player's
// loss. Partner is a deck-building rule (CR 702.124) that internal/deck
// reads off the oracle text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b9d587c-39d2-44a2-9802-e4217ed389c0",
		Name:         "Ludevic, Necro-Alchemist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, effects.AnyPlayer,
				"Ludevic, Necro-Alchemist — that player may draw a card if a player other than you lost life this turn",
				ludevicMayDraw),
		},
	})
}

func ludevicMayDraw(g *game.Game, item *game.StackItem) error {
	player := effects.TriggeringActor(item)
	if player == uuid.Nil || !ludevicAnotherPlayerLostLife(g, item.Controller) {
		return nil
	}
	return effects.MayChoice{
		Player:   player,
		Question: "Ludevic, Necro-Alchemist — draw a card?",
		Draws:    1,
		Then:     effects.MayDrawCards,
		Word:     "1",
	}.Apply(effects.NewContext(g, item))
}

// ludevicAnotherPlayerLostLife is "a player other than you lost life this
// turn".
func ludevicAnotherPlayerLostLife(g *game.Game, you uuid.UUID) bool {
	for _, p := range g.Seats {
		if p != nil && p.ID != you && effects.B18LifeLostThisTurn(g, p.ID) > 0 {
			return true
		}
	}
	return false
}
