package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// trek_b26_helpers.go — shared pieces of Star Trek, batch 26 (#4973).

// trekB26CommunicatorOnAnotherCreature is Combadge's "you control another
// creature with a Communicator attached to it": a creature other than
// `self`, controlled by self's controller, that a permanent with the
// Communicator subtype is attached to. Both are read through the layered
// view, so a creature that stopped being one, or an Equipment that lost
// the subtype, does not count.
func trekB26CommunicatorOnAnotherCreature(g *game.Game, self *game.Card) bool {
	if self == nil {
		return false
	}
	cards := g.BattlefieldCardsForEffect()
	for _, gear := range cards {
		if gear.AttachedTo.Kind != game.TargetCard || gear.AttachedTo.ID == self.InstanceID || !gear.HasSubtype("Communicator") {
			continue
		}
		for _, c := range cards {
			if c.InstanceID == gear.AttachedTo.ID && c.Controller == self.Controller && c.IsCreature() {
				return true
			}
		}
	}
	return false
}
