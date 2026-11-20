package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "601fd61b-4939-459a-84d1-d947d29c9564",
		Name:         "Bloodhaze Wolverine",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Bloodhaze Wolverine — +1/+1 and first strike until end of turn",
				thisCreatureUntilEOT("Bloodhaze Wolverine — +1/+1 and first strike", 1, 1, "first strike")),
		},
	})
}
