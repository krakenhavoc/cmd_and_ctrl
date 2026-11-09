package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kulrath Zealot — Creature — Elemental Warrior {5}{R}, 6/5 (Lorwyn
// Eclipsed, #4458):
//
//	"When this creature enters, exile the top card of your library. Until
//	 the end of your next turn, you may play that card.
//	 Basic landcycling {1}{R}"
//
// The exile is the shared impulse exile of one card (ExileTopNUntilYourNextTurn):
// a play permission, so a land can be played and a spell cast, with the
// card's own costs and the usual timing (the ruling), until the end of the
// controller's next turn. Basic landcycling is the cycling family's
// BasicLandcycling.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d18ca75c-3572-4851-9580-27bafe10edd6",
		Name:         "Kulrath Zealot",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Kulrath Zealot — exile the top card of your library; you may play it until the end of your next turn",
				func(g *game.Game, item *game.StackItem) error {
					return ExileTopNUntilYourNextTurn(NewContext(g, item), 1)
				}),
		},
		Activated: []ActivatedAbility{BasicLandcycling("{1}{R}")},
	})
}
