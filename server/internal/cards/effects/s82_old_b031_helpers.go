package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b031_helpers.go — shared pieces of OLD batch 31 (#5501, Modern
// Horizons 2).

// OldB031SelfEntersWithCounters is "this permanent enters with N <kind>
// counters on it" as an always-on self-replacement (CR 614.1c): Baloth
// Prime's entry without the tapped half. The counters ride the entry event,
// so every counter replacement sees them.
func OldB031SelfEntersWithCounters(kind string, n int) game.ReplacementEffect {
	return game.ReplacementEffect{
		SelfEntryCounters: &game.EntryCounterAmount{Kind: kind, Amount: n},
		RowLabel:          "Enters with " + counterCountPhrase(n, kind) + " on it.",
		EntryOnly:         true,
		Watches:           []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield &&
				src != nil && ev.CardID == src.InstanceID
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.AddCounterAtETB(kind, n)
			return nil
		},
	}
}

// OldB031EnteredFromOrCastFromAGraveyard is Breathless Knight's condition:
// a creature under the source's controller's control (the source itself
// included) entered from a graveyard, or was cast from a graveyard by its
// controller (the permanent's cast provenance, CR 400.7d).
func OldB031EnteredFromOrCastFromAGraveyard(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	c, ok := EnteredUnderYourControl(ev, source, g, false)
	if !ok || !c.IsCreature() {
		return false
	}
	if EnteredFromAGraveyard(ev, source, lki, g) {
		return true
	}
	cast, known := c.CastByItsController()
	return known && cast && c.Provenance.FromZone == game.ZoneGraveyard
}
