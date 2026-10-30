package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alien Symbiosis — Enchantment — Aura {1}{B} (Marvel Super Heroes
// Commander, #4455):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1, has menace, and is a Symbiote in
//	 addition to its other types.
//	 You may cast this card from your graveyard by discarding a card in
//	 addition to paying its other costs."
//
// The graveyard cast is Retrace's model (ADR 0066, 2026-10-07 amendment)
// with any card as the discard: msc17CastFromGraveyardByDiscarding charges
// the printed mana cost and one card from hand discarded as a cost.
// "Symbiote" is a word of the Aura's text (CR 612.2, #3984), read through
// the source's text reader.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "19504e19-2886-4d28-a813-b798a9a446a2",
		Name:             "Alien Symbiosis",
		Completeness:     CompletenessFull,
		Targets:          EnchantCreature(),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{msc17CastFromGraveyardByDiscarding("{1}{B}")},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("menace"),
			{
				Layer:     game.Layer4Type,
				AppliesTo: AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					symbiote := source.TextSubtype("Symbiote")
					if !eotHasType(c.Subtypes, symbiote) {
						c.Subtypes = append(c.Subtypes, symbiote)
					}
				},
			},
		},
	})
}
