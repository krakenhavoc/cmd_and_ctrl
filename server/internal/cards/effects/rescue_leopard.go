package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rescue Leopard — Creature — Cat {2}{R}, 4/2 (Tarkir: Dragonstorm, #4463):
//
//	"Whenever this creature becomes tapped, you may discard a card. If you
//	 do, draw a card."
//
// "Becomes tapped" watches both the tap event and the attack, since an
// attacker is tapped without a tap event (ThisBecameTapped, Chrome
// Companion's); a vigilance attacker did not become tapped. The discard is
// chosen as the trigger resolves, up to one card, and a card is drawn only
// if one was discarded.
//
// No simplification.
func init() {
	const label = "Rescue Leopard — you may discard a card; if you do, draw a card"
	Register(Spec{
		OracleID:     "d70ad04a-b940-4bb1-bcdc-408166e91723",
		Name:         "Rescue Leopard",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
				label, tdmMayDiscardThenDrawOne(label)),
		},
	})
}
