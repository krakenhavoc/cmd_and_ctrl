package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sibsig Appraiser — Creature — Zombie Advisor {2}{U}, 2/1 (Tarkir:
// Dragonstorm, #4463):
//
//	"When this creature enters, look at the top two cards of your
//	 library. Put one of them into your hand and the other into your
//	 graveyard."
//
// Visionary's Dance's look-and-take on an enters trigger: a private look
// at the top two, a mandatory pick of one for the hand, and the other put
// into the graveyard (takeRestGraveyard), not discarded or milled. A
// library of one card puts that card into the hand; an empty one does
// nothing.
//
// No simplification.
func init() {
	const label = "Sibsig Appraiser — put one of them into your hand and the other into your graveyard"
	Register(Spec{
		OracleID:     "5b51958d-533a-45e4-aac1-3c157670978b",
		Name:         "Sibsig Appraiser",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sibsig Appraiser — look at the top two cards of your library",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					player := ctx.Controller()
					return TakeFromLibraryToHand{
						Player: player,
						Cards:  g.LookAtTopOfLibraryForEffect(player, 2),
						Max:    1,
						Label:  label,
						Key:    takeRestGraveyard,
					}.Apply(ctx)
				}),
		},
	})
}
