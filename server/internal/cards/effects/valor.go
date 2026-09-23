package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Valor — Creature — Incarnation {3}{W}, 2/2:
//
//	"First strike
//	 As long as this card is in your graveyard and you control
//	 a Plains, creatures you control have first strike."
//
// One of the four Judgment incarnations, and #1221's proof that a
// static can function from a graveyard (CR 113.6c). The printed
// keyword on the body is ordinary; the anthem is the clause that had
// nowhere to live until StaticAbility.Zones and the layer pass's
// declared-zone gather.
//
// "You" is the card's OWNER (CR 108.4) — a card in a graveyard has no
// controller — so a Valor milled out of an opponent's library helps
// THEM. And the anthem does not apply while Valor is on the
// battlefield: the declared zone list is the list, which is why the
// card prints its own keyword as a separate line.
//
// See incarnations.go and game/static_zones.go.
func init() {
	Register(Spec{
		OracleID:     "b2ac84e3-cc3c-49c6-918b-a407ef1ee06c",
		Name:         "Valor",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{incarnationAnthem("first strike", "Plains")},
	})
}
