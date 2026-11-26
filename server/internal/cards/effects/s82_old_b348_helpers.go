package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// OldB348YourCreatureWithToughnessFourOrMoreDied is "whenever a creature you
// control with toughness 4 or greater dies" (Sultai Flayer): a creature the
// source's controller controlled as it last existed, whose last-known
// toughness (counters and pumps included, CR 603.10a) was 4 or more.
func OldB348YourCreatureWithToughnessFourOrMoreDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := DiedCreature(ev, g)
	if !ok || LeftUnderControlOf(ev, dead) != source.Controller {
		return false
	}
	info, ok := g.LastKnownPermanentForEffect(dead.InstanceID)
	return ok && info.Toughness >= 4
}
