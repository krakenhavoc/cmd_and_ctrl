package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Necrobloom — Legendary Creature — Plant {1}{W}{B}{G}, 2/7
// (EDHREC rank 3703):
//
//	"Landfall — Whenever a land you control enters, create a 0/1
//	 green Plant creature token. If you control seven or more lands
//	 with different names, create a 2/2 black Zombie creature token
//	 instead.
//	 Land cards in your graveyard have dredge 2. (You may return a
//	 land card from your graveyard to your hand and mill two cards
//	 instead of drawing a card.)"
//
// The lands commander. Landfall is Maja's condition
// (b33LandYouControlEntered — a land entered under the controller's
// control, played or fetched or tokened), and the body reads the
// land count at resolution, so the land that entered is among the
// seven: a Plant, or a Zombie once seven differently-named lands are
// in play (Field of the Dead's count, b04LandNamesControlled — two
// copies of the same land count once, as printed).
//
// SANDBOX SIMPLIFICATION — dredge is NOT implemented. "Land cards
// in your graveyard have dredge 2" is a draw replacement the
// controller may choose, per land card, that also asks which land
// card — and the engine gathers replacement effects from battlefield
// permanents with no way for a replacement to open a pick among
// graveyard cards, nor any dredge keyword for the cards to carry.
// The Necrobloom's own draws are ordinary draws. Weaker than
// printed, never stronger: the landfall half, which is the card, is
// whole.
func init() {
	Register(Spec{
		OracleID:     "b981af39-4ee6-4fbc-9a89-618dcad9dfbf",
		Name:         "The Necrobloom",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Dredge isn't implemented — land cards in your graveyard can't be returned in place of a draw."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b33LandYouControlEntered(ev, source, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "The Necrobloom — create a 0/1 green Plant, or a 2/2 black Zombie with seven differently-named lands",
					b35NecrobloomLandfall)
			},
		}},
	})
}
