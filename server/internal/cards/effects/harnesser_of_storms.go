package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Harnesser of Storms — Creature — Otter Wizard {2}{R}, 1/4 (Bloomburrow,
// #4467):
//
//	"Whenever you cast a noncreature or Otter spell, you may exile the top
//	 card of your library. Until end of turn, you may play that card. This
//	 ability triggers only once each turn."
//
// A cast trigger over a noncreature spell or an Otter spell (the type a
// word of the Harnesser's text, CR 612.2), held to once each turn per
// object (TriggersOnlyOnceEachTurn: a trigger that was countered, or one
// you answered "no", still used up the turn). The "you may" is asked as
// it resolves (CR 603.5); a yes is Count on Luck's impulse exile, one
// card you may play this turn, a land included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fb8f7047-280b-464f-aca2-6add3e21df8f",
		Name:         "Harnesser of Storms",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(TriggersOnlyOnceEachTurn(WheneverYouCast(Or(Noncreature(), OfCreatureType("Otter")),
				"Harnesser of Storms — exile the top card of your library; you may play it this turn",
				func(g *game.Game, item *game.StackItem) error {
					_, err := b12ImpulseExileForTurn(g, item, 1)
					return err
				})),
				"Harnesser of Storms — exile the top card of your library? You may play it this turn."),
		},
	})
}
