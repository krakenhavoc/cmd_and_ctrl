package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wonder — Creature — Incarnation {3}{U}, 2/2:
//
//	"Flying
//	 As long as this card is in your graveyard and you control
//	 an Island, creatures you control have flying."
//
// One of the four Judgment incarnations, and #1221's proof that a
// static can function from a graveyard (CR 113.6c). The printed
// keyword on the body is ordinary; the anthem is the clause that had
// nowhere to live until StaticAbility.Zones and the layer pass's
// declared-zone gather.
//
// "You" is the card's OWNER (CR 108.4) — a card in a graveyard has no
// controller — so a Wonder milled out of an opponent's library helps
// THEM. And the anthem does not apply while Wonder is on the
// battlefield: the declared zone list is the list, which is why the
// card prints its own keyword as a separate line.
//
// See incarnations.go and game/static_zones.go.
func init() {
	Register(Spec{
		OracleID:     "232284f7-c623-4895-9ab9-8b1a39926830",
		Name:         "Wonder",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{incarnationAnthem("flying", "Island")},
	})
}
