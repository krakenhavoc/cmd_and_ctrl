package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Satyr Wayfinder — Creature — Satyr {1}{G}, 1/1:
//
//	"When this creature enters, reveal the top four cards of your
//	 library. You may put a land card from among them into your hand.
//	 Put the rest into your graveyard."
//
// The four are revealed to every player first, then one optional pick
// of a land card; everything not picked goes to the graveyard
// (takeRestGraveyard — put, not milled).
//
// No simplification. Star Trek Commander (#4974).
func init() {
	const label = "Satyr Wayfinder — reveal the top four cards, you may put a land card into your hand and the rest into your graveyard"
	Register(Spec{
		OracleID:     "ead4c01a-e2ae-49c9-95d9-651d81005907",
		Name:         "Satyr Wayfinder",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				player := ctx.Controller()
				return TakeFromLibraryToHand{
					Player:   player,
					Cards:    g.RevealTopOfLibraryForEffect(player, ctx.Source(), 4, label),
					Match:    Land(),
					Max:      1,
					Optional: true,
					Label:    label,
					Key:      takeRestGraveyard,
				}.Apply(ctx)
			}),
		},
	})
}
