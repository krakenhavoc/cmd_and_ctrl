package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// patronOfTheNezumiLabel is the trigger's row key.
const patronOfTheNezumiLabel = "Patron of the Nezumi — that player loses 1 life"

// Patron of the Nezumi — Legendary Creature — Spirit {5}{B}{B}, 6/6
// (Betrayers of Kamigawa, #5602):
//
//	"Rat offering (You may cast this spell any time you could cast an
//	 instant by sacrificing a Rat and paying the difference in mana costs
//	 between this and the sacrificed Rat. Mana cost includes color.)
//	 Whenever a permanent is put into an opponent's graveyard, that player
//	 loses 1 life."
//
// Rat offering is Offering (offering.go, CR 702.48). The trigger watches
// every permanent that leaves the battlefield for a graveyard whose owner
// is one of its controller's opponents (bokB1PermanentPutIntoAnOpponentsGraveyard):
// a card goes to its owner's graveyard whoever controlled it, and a token
// counts too (CR 111.7). "That player" is the owner, carried on the item
// as it is built, so one trigger per permanent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "038a8b6b-4529-4e06-bdff-9ee7ba291776",
		Name:          "Patron of the Nezumi",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Offering("Rat", Subtype("Rat"))},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			Key:     patronOfTheNezumiLabel,
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := bokB1PermanentPutIntoAnOpponentsGraveyard(ev, source, g)
				return ok
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				owner, _ := bokB1PermanentPutIntoAnOpponentsGraveyard(ev, source, g)
				item := game.NewTriggeredItem(source, patronOfTheNezumiLabel)
				item.Params.Player = owner
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GainLife{Player: item.Params.Player, Amount: -1}.Apply(NewContext(g, item))
			},
		}},
	})
}
