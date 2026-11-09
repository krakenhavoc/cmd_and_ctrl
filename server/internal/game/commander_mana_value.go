package game

import "github.com/google/uuid"

// GreatestCommanderManaValueForEffect is "the greatest mana value among
// your commanders" (Tangleweave Armor): the largest mana value of any card
// the player owns that carries the commander designation, wherever it is
// (CR 903.3, the designation stays with the card through every zone
// change), read live so a commander that is a copy on the battlefield
// counts at its current mana value. A token is never a commander (CR 707.2).
// Zero for a player with no commander.
//
// Like every ...ForEffect accessor it does not lock: the caller holds the
// game lock (a layer pass or a resolving effect). CommandersOf is the
// locking, lobby-facing sibling and must not be called from here.
func (g *Game) GreatestCommanderManaValueForEffect(player uuid.UUID) int {
	best := 0
	g.eachIndexedZoneLocked(func(_ cardSlot, _ int, z *Zone) bool {
		for _, c := range z.Cards {
			if c.IsCommander && c.Owner == player && !c.IsToken() {
				if mv := c.ManaValue(); mv > best {
					best = mv
				}
			}
		}
		return false
	})
	return best
}
