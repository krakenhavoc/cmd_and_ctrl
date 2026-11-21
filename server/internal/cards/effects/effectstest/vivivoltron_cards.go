package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// vivivoltron_cards_test.go — the five cards added for the ViviVoltron
// deck request (#1640): Slip Out the Back, Diviner's Portent, Ancient
// Silver Dragon, Return the Favor and March of Swirling Mist.

const (
	SlipOutTheBackOracle      = "82506e51-3fe1-446f-83c9-69e67583cefc"
	DivinersPortentOracle     = "119585c7-ddfa-47ed-b2f8-488ebc156222"
	AncientSilverDragonOracle = "939556c8-e363-4453-a624-ca52b151467e"
	ReturnTheFavorOracle      = "30377bd5-99ab-4888-9ea7-e5a069b527e8"
	MarchOfSwirlingMistOracle = "debc69ea-372a-4720-838a-16856cd50b07"
)

// vvPhasedOutCard is the phased-out copy of `id`, read off the holding
// slice (the card is not on the battlefield while phased out).
func VvPhasedOutCard(g *game.Game, id uuid.UUID) (game.Card, bool) {
	var out game.Card
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.PhasedOutCardsForEffect() {
			if c.InstanceID == id {
				out, found = c, true
			}
		}
	})
	return out, found
}
