package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// zone_layer_statics.go — the zone lists a layer-4 or layer-5 static
// over objects off the battlefield declares in
// game.StaticAbility.AffectsZones (#3862, game/zone_layers.go).
//
// The printed sentences split into two kinds of object, and the "you"
// of each is different:
//
//   - SPELLS (the stack): "creature spells you control" — the target's
//     Controller is the spell's controller.
//   - CARDS (every other zone of the game): "creature cards you own
//     that aren't on the battlefield" — the target's Owner.
//
// A card that says both declares two statics, one per kind, so each
// predicate reads the right "you" and nothing else.

// OffTheBattlefield is every zone of the game but the battlefield:
// Mycosynth Lattice's "All cards that aren't on the battlefield,
// spells".
func OffTheBattlefield() []game.ZoneKind {
	return []game.ZoneKind{game.ZoneStack, game.ZoneHand, game.ZoneLibrary, game.ZoneGraveyard, game.ZoneExile, game.ZoneCommand}
}

// CardsOffTheBattlefield is every zone a CARD can be in other than the
// battlefield and the stack: Maskwood Nexus' "creature cards you own
// that aren't on the battlefield".
func CardsOffTheBattlefield() []game.ZoneKind {
	return []game.ZoneKind{game.ZoneHand, game.ZoneLibrary, game.ZoneGraveyard, game.ZoneExile, game.ZoneCommand}
}

// SpellsOnTheStack is the stack, for a static over spells' types or
// colours: "permanent spells you control" (Secret Arcade).
func SpellsOnTheStack() []game.ZoneKind {
	return []game.ZoneKind{game.ZoneStack}
}
