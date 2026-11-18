package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Organ Hoarder — Creature — Zombie {3}{U}, 3/2 (Innistrad: Midnight Hunt,
// S82):
//
//	"When this creature enters, look at the top three cards of your
//	 library, then put one of them into your hand and the rest into your
//	 graveyard."
//
// Commune with Evil's look-and-take over three cards: a private look, a
// mandatory pick of one for the hand, and the other two put into the
// graveyard (not milled or discarded).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a6f0dcbc-ca45-4ae8-961e-e42ef8d8a975",
		Name:         "Organ Hoarder",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Organ Hoarder — look at the top three cards, put one into your hand and the rest into your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					player := ctx.Controller()
					return TakeFromLibraryToHand{
						Player: player,
						Cards:  g.LookAtTopOfLibraryForEffect(player, 3),
						Max:    1,
						Label:  "Organ Hoarder — put one of them into your hand and the rest into your graveyard",
						Key:    organHoarderTake,
					}.Apply(ctx)
				}),
		},
	})
}

// organHoarderTake is the look's continuation: the rest into the graveyard.
// An on-disk key: never renamed, never reused.
var organHoarderTake = RegisterTake("choose-cards/organ-hoarder-take",
	TakeDef{Then: func(g *game.Game, res TakeFromLibraryResult) error {
		return TakeRestIntoGraveyard(g, res)
	}})
