package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sentinel's Eyes — Enchantment — Aura {W} (Secrets of Strixhaven
// Commander, #4834):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and has vigilance.
//	 Escape—{W}, Exile two other cards from your graveyard. (You may cast
//	 this card from your graveyard for its escape cost.)"
//
// The Aura is Angelic Destiny's shape: the attached creature's +1/+1 and
// vigilance are statics (PumpAttached, GrantToAttached), so vigilance is
// read through the layered view. Escape is the keyword's alternative cost
// (Escape, CR 702.138a): the graveyard is opened as a cast zone and the
// cost is {W} plus two OTHER cards of the graveyard exiled as a cost. An
// escaped Aura is cast as any Aura is, choosing its enchant target as it
// is cast (CR 303.4a); escape's permission changes nothing about when it
// may be cast (the ruling), so it is cast at sorcery speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "0a7d6489-b239-4ce8-aea3-7b5d0e436f70",
		Name:             "Sentinel's Eyes",
		Completeness:     CompletenessFull,
		Targets:          EnchantCreature(),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Escape("{W}", 2)},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("vigilance"),
		},
	})
}
