package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap_b77_helpers.go — shared pieces for the S82 GAP batch 77 cards
// (Ravnica Remastered creatures and enchantments, #5948).

// gapB77AnotherCreatureOrPlaneswalkerYouControlDied is the condition of
// "Whenever another creature or planeswalker you control dies" (Rising
// Populace): a permanent that went to a graveyard from the battlefield,
// other than the source, that its controller (read as it died) matches the
// source's controller, and that was a creature or a planeswalker as it left
// (last-known types, CR 603.10a).
func gapB77AnotherCreatureOrPlaneswalkerYouControlDied(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard || ev.CardID == source.InstanceID {
		return false
	}
	dead, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || dead.Controller != source.Controller {
		return false
	}
	return leftAsType(ev, dead, "creature") || leftAsType(ev, dead, "planeswalker")
}
