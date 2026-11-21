package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spontaneous Mutation — Enchantment — Aura {U} (Innistrad Remastered):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets -X/-0, where X is the number of cards in your
//	 graveyard."
//
// X is the Aura's controller's graveyard, read on every layer pass, so it
// follows the graveyard up and down.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ab78fa13-f599-4d08-bd27-d3d8740ae58f",
		Name:            "Spontaneous Mutation",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(-1, 0, func(g *game.Game, source *game.Card) int {
				p := g.PlayerByIDForEffect(source.Controller)
				if p == nil || p.Graveyard == nil {
					return 0
				}
				return p.Graveyard.Size()
			}),
		},
	})
}
