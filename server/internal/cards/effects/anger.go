package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Anger — Creature — Incarnation {3}{R}, 2/2:
//
//	"Haste
//	 As long as this card is in your graveyard and you control
//	 a Mountain, creatures you control have haste."
//
// One of the four Judgment incarnations, and #1221's proof that a
// static can function from a graveyard (CR 113.6c). The printed
// keyword on the body is ordinary; the anthem is the clause that had
// nowhere to live until StaticAbility.Zones and the layer pass's
// declared-zone gather.
//
// "You" is the card's OWNER (CR 108.4) — a card in a graveyard has no
// controller — so an Anger milled out of an opponent's library helps
// THEM. And the anthem does not apply while Anger is on the
// battlefield: the declared zone list is the list, which is why the
// card prints its own keyword as a separate line.
//
// See incarnations.go and game/static_zones.go.
func init() {
	Register(Spec{
		OracleID:     "eaabd151-2160-4bff-82c0-3fa88659be98",
		Name:         "Anger",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{incarnationAnthem("haste", "Mountain")},
	})
}
