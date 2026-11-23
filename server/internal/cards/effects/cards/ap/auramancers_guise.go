package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Auramancer's Guise — Enchantment — Aura {2}{U}{U} (Planar Chaos, #5595):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 for each Aura attached to it and has
//	 vigilance."
//
// PumpAttachedPer over b18AurasAttachedTo (Kor Spiritdancer's count),
// re-read every layer pass, so the Guise counts itself and grows and
// shrinks as other Auras come and go; Equipment does not count. Vigilance
// is GrantToAttached.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b761f6d9-1063-40f8-a3f1-25b1f9b46963",
		Name:         "Auramancer's Guise",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(2, 2, func(g *game.Game, source *game.Card) int {
				if source.AttachedTo.Kind != game.TargetCard {
					return 0
				}
				return effects.B18AurasAttachedTo(g, source.AttachedTo.ID)
			}),
			effects.GrantToAttached("vigilance"),
		},
	})
}
