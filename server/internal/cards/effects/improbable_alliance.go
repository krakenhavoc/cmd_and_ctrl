package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Improbable Alliance — Enchantment {U}{R} (Magic Commander Collection,
// #5908):
//
//	"Whenever you draw your second card each turn, create a 1/1 blue Faerie
//	 creature token with flying.
//	 {4}{U}{R}: Draw a card, then discard a card."
//
// The draw tally is read after the draw, so the trigger fires on the second
// draw and no other (Alandra's read); the loot is Agna Qela's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b4ca6a20-b1ba-4dee-ae8a-83f8edf360c4",
		Name:         "Improbable Alliance",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Improbable Alliance — create a 1/1 blue Faerie creature token with flying",
				Do(CreateToken{Template: TokenCard("1/1 blue Faerie with flying"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}{U}{R}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{4}{U}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return lootOne(g, item, 1)
			},
		}},
	})
}
