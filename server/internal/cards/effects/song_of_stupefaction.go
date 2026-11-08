package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Song of Stupefaction — Enchantment — Aura {1}{U} (The Lost Caverns of
// Ixalan, #4470):
//
//	"Enchant creature or Vehicle
//	 When this Aura enters, you may mill two cards.
//	 Fathomless descent — Enchanted permanent gets -X/-0, where X is the
//	 number of permanent cards in your graveyard."
//
// The enchant clause is Aether Meltdown's: it admits a Vehicle that is not
// a creature, and the CR 704.5m re-check runs the same clause, so the Aura
// stays on an uncrewed Vehicle. The enters trigger is Deathcap
// Marionette's optional mill of two (CR 701.17a).
//
// Fathomless descent is an ability word with no rules meaning (CR 207.2c).
// The -X/-0 is a layer 7c modification (CR 613.4c) of the enchanted
// permanent, read on every recompute (CR 611.3a): X is
// permanentCardsInGraveyard for the Aura's controller, the descend count,
// and a card entering or leaving that graveyard is a zone move, which drops
// the layer cache. A noncreature permanent has no power (CR 208.3), so the
// modification changes something only while the host is a creature, as
// pumpEnchantedCreature does; a crewed Vehicle gets it at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "87300967-1a59-4dce-8876-f008ff3bfb0b",
		Name:         "Song of Stupefaction",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Targets:      TargetPermanent("enchant creature or Vehicle", Or(Creature(), HasSubtype("Vehicle"))),
		Static: []game.StaticAbility{{
			Label:     "Fathomless descent — Enchanted permanent gets -X/-0, where X is the number of permanent cards in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: AttachedToSource,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				if !slices.Contains(c.Types, "Creature") {
					return
				}
				c.Power -= permanentCardsInGraveyard(g, source.Controller)
			},
		}},
		Triggered: []game.TriggeredAbility{
			Optional(WhenThisEnters("Song of Stupefaction — mill two cards", Do(MillCards{N: 2})),
				"Song of Stupefaction — mill two cards?"),
		},
	})
}
