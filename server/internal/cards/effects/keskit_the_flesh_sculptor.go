package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Keskit, the Flesh Sculptor — Legendary Creature — Phyrexian Human Artificer
// {2}{B}, 1/3 (Magic Commander Collection, #5908):
//
//	"{T}, Sacrifice three other artifacts and/or creatures: Look at the top
//	 three cards of your library. Put two of them into your hand and the
//	 other into your graveyard.
//	 Partner (You can have two commanders if both have partner.)"
//
// The look is mandatory for two of the three cards (a shorter library puts
// what it has); the one left over is put, not milled, into the graveyard
// (takeRestGraveyard). Partner is read from the oracle text by the deck
// builder.
//
// No simplification.
func init() {
	const label = "Keskit — put two of those cards into your hand and the other into your graveyard"
	Register(Spec{
		OracleID:     "e665acb1-a839-46c6-b5ae-ad189213bee4",
		Name:         "Keskit, the Flesh Sculptor",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice three other artifacts and/or creatures: Look at the top three cards of your library. Put two of them into your hand and the other into your graveyard.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost: Plus(TapCost(),
				SacrificeAnotherN(3, "three other artifacts and/or creatures", Or(Artifact(), Creature()))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				player := ctx.Controller()
				return TakeFromLibraryToHand{
					Player: player,
					Cards:  g.LookAtTopOfLibraryForEffect(player, 3),
					Max:    2,
					Label:  label,
					Key:    takeRestGraveyard,
				}.Apply(ctx)
			},
		}},
	})
}
