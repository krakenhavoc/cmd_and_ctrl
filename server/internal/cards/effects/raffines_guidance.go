package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raffine's Guidance — Enchantment — Aura {W} (Secrets of Strixhaven
// Commander, #4834):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1.
//	 You may cast this card from your graveyard by paying {2}{W} rather
//	 than paying its mana cost."
//
// The graveyard cast is a graveyard-bound alternative cost (CR 118.9) with
// no exile attached: the Aura goes back to the graveyard when it dies, and
// can be cast again. It opens the graveyard like Flashback does
// (Spec.CastableZones) and is priced under its own key; it is cast as any
// Aura is, with a creature to enchant.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "ad3bf726-e0c2-4038-a66e-9a9dff439563",
		Name:          "Raffine's Guidance",
		Completeness:  CompletenessFull,
		Targets:       EnchantCreature(),
		Static:        []game.StaticAbility{PumpAttached(1, 1)},
		CastableZones: []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{{
			Key:      "raffines-guidance-graveyard",
			Label:    "Cast from your graveyard for {2}{W}",
			ManaCost: "{2}{W}",
			FromZone: game.ZoneGraveyard,
		}},
	})
}
