package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fear of Death — Enchantment — Aura {1}{U} (Innistrad: Crimson Vow):
//
//	"Enchant creature
//	 When this Aura enters, mill two cards.
//	 Enchanted creature gets -X/-0, where X is the number of cards in your
//	 graveyard."
//
// The -X/-0 is a layer 7c modification (CR 613.4c) read on every recompute
// (CR 611.3a): X is the size of the Aura controller's graveyard as it stands
// then, so a card arriving or leaving changes it at once. The mill is
// mandatory and happens as the enters trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5187c0b2-1996-42ea-bd77-3031b3add596",
		Name:         "Fear of Death",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Targets:      TargetCreature("enchant creature"),
		Static: []game.StaticAbility{{
			Label:     "Enchanted creature gets -X/-0, where X is the number of cards in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: AttachedToSource,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				if !slices.Contains(c.Types, "Creature") {
					return
				}
				if p := g.PlayerByIDForEffect(source.Controller); p != nil && p.Graveyard != nil {
					c.Power -= len(p.Graveyard.Cards)
				}
			},
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Fear of Death — mill two cards", Do(MillCards{N: 2})),
		},
	})
}
