package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Madame Masque — Legendary Creature — Human Villain {4}{B}, 3/2 (Marvel
// Super Heroes, #4455):
//
//	"When Madame Masque enters, she connives.
//	 Whenever you draw your second card each turn, create a 2/1 black
//	 Villain creature token with menace."
//
// Connive is ConniveEach over the one creature (Baron Strucker's shape):
// draw, discard, and a +1/+1 counter for a nonland card. The second-draw
// trigger is Atlantean Cavalry's word (b41DrewYourSecondCardThisTurn), and
// the 2/1 menace Villain row is the one Construct a Cosmic Cube uses. The
// connive draw is itself a draw, so it can be the second card of a turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b5b0e85f-6651-4667-829c-2253ff8b130d",
		Name:         "Madame Masque",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Madame Masque — she connives", func(g *game.Game, item *game.StackItem) error {
				return ConniveEach{Controller: item.Controller, Creatures: []uuid.UUID{item.SourceCardID}}.Apply(NewContext(g, item))
			}),
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Madame Masque — create a 2/1 black Villain creature token with menace",
				Do(CreateToken{Template: TokenCard("2/1 black Villain with menace"), N: 1})),
		},
	})
}
