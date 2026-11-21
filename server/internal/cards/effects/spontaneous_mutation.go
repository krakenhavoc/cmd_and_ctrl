package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "ab78fa13-f599-4d08-bd27-d3d8740ae58f",
		Name:            "Spontaneous Mutation",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			PumpAttachedPer(-1, 0, func(g *game.Game, source *game.Card) int {
				p := g.PlayerByIDForEffect(source.Controller)
				if p == nil || p.Graveyard == nil {
					return 0
				}
				return p.Graveyard.Size()
			}),
		},
	})
}
