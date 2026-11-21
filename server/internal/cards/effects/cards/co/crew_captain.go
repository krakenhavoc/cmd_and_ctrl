package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crew Captain — Creature — Human Warrior {B}{R}{G}, 4/2 (Streets of New
// Capenna, S82):
//
//	"Haste
//	 This creature has indestructible as long as it entered this turn."
//
// The static is on while the engine's entry record says the Captain entered
// this turn (SourceEnteredThisTurn), not summoning sickness, so it ends at
// the turn's end and a blinked Captain is a new object that starts again.
//
// No simplification.
func init() {
	grant := effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return effects.SelfCreature(target, g, source) && effects.SourceEnteredThisTurn(g, *source)
	}, "indestructible")
	grant.Label = "Has indestructible as long as it entered this turn."
	effects.Register(effects.Spec{
		OracleID:        "0cca60c3-f0be-494d-8c2c-9b135ef83165",
		Name:            "Crew Captain",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static:          []game.StaticAbility{grant},
	})
}
