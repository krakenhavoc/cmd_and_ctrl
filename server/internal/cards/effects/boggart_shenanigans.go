package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boggart Shenanigans — Kindred Enchantment — Goblin {2}{R} (Lorwyn,
// #5593):
//
//	"Whenever another Goblin you control is put into a graveyard from the
//	 battlefield, you may have this enchantment deal 1 damage to target
//	 player or planeswalker."
//
// "Goblin" is any permanent of the type you controlled as it left, a Kindred
// one or a token as well as a creature card, read as it last existed on the
// battlefield (CR 603.10a), so a changeling counts. Any graveyard counts.
// The target is chosen as the trigger goes on the stack (CR 603.3d) and the
// damage is the "you may" (CR 603.5); its source is the enchantment.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "997d2f38-5a50-451e-bd48-e36dcb24d967",
		Name:         "Boggart Shenanigans",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(
				Targeting(
					On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						c, ok := lrwB1PutIntoAGraveyard(ev, g, source.TextSubtype("Goblin"))
						return ok && c.InstanceID != source.InstanceID && leftUnderControlOf(ev, c) == source.Controller
					}, "Boggart Shenanigans — 1 damage to target player or planeswalker", DamageToFirstTarget(1)),
					targetPlayerOrPlaneswalker()),
				"Boggart Shenanigans — deal 1 damage to target player or planeswalker?"),
		},
	})
}
