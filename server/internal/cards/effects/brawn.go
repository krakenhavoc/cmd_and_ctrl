package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brawn — Creature — Incarnation {3}{G}, 3/3:
//
//	"Trample
//	 As long as this card is in your graveyard and you control
//	 a Forest, creatures you control have trample."
//
// One of the four Judgment incarnations, and #1221's proof that a
// static can function from a graveyard (CR 113.6c). The printed
// keyword on the body is ordinary; the anthem is the clause that had
// nowhere to live until StaticAbility.Zones and the layer pass's
// declared-zone gather.
//
// "You" is the card's OWNER (CR 108.4) — a card in a graveyard has no
// controller — so a Brawn milled out of an opponent's library helps
// THEM. And the anthem does not apply while Brawn is on the
// battlefield: the declared zone list is the list, which is why the
// card prints its own keyword as a separate line.
//
// See incarnations.go and game/static_zones.go.
func init() {
	Register(Spec{
		OracleID:     "00876e98-d062-4a12-85e6-86a2b20cf867",
		Name:         "Brawn",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{incarnationAnthem("trample", "Forest")},
	})
}
