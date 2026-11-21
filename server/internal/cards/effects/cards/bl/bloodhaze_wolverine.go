package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodhaze Wolverine — Creature — Wolverine {1}{R}, 2/1 (Throne of Eldraine, #5516):
//
//	"Whenever you draw your second card each turn, this creature gets +1/+1
//	 and gains first strike until end of turn."
//
// Thopter Fabricator's "second card each turn" condition
// (b41DrewYourSecondCardThisTurn): it fires once a turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "601fd61b-4939-459a-84d1-d947d29c9564",
		Name:         "Bloodhaze Wolverine",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Bloodhaze Wolverine — +1/+1 and first strike until end of turn",
				effects.ThisCreatureUntilEOT("Bloodhaze Wolverine — +1/+1 and first strike", 1, 1, "first strike")),
		},
	})
}
