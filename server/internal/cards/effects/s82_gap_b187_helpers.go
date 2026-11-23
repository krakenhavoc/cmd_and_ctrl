package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap_b187_helpers.go — shared bodies for S82 GAP batch 187 (#5948),
// Warhammer 40,000 Commander creatures.

// GapB187ReturnDiedTappedBody is a delayed trigger's body: "return it to the
// battlefield tapped under your control" for the creature that died (Cryptek).
// "Your" is the delayed ability's controller, not the card's owner. A card
// that has left the graveyard since is a different object and stays put
// (CR 400.7); a token has no card to return.
var GapB187ReturnDiedTappedBody = game.SimpleDelayedBody("gap-b187/return-died-tapped", gapB187ReturnDiedTapped)

func gapB187ReturnDiedTapped(g *game.Game, item *game.StackItem) error {
	for _, id := range NewContext(g, item).PayloadCards() {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if _, err := g.ReturnToBattlefieldForEffect(id, item.Controller, true); err != nil {
			return err
		}
	}
	return nil
}
