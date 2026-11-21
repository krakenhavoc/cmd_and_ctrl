package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Penumbra Umbra — Enchantment — Aura {1}{G}:
//
//	"Enchant creature you control
//	 When Penumbra Umbra is put into a graveyard from the battlefield,
//	 create a token that's a copy of enchanted creature, except the token
//	 is black.
//	 Umbra armor (If enchanted creature would be destroyed, instead remove
//	 all damage from it and destroy this Aura.)"
//
// The enchanted creature is read as the Aura leaves (CR 608.2h), so the copy
// is of the creature that armour just saved, or of the one that died with
// the Aura (its last-known copiable values). The exception sets the token's
// colour to black alone (CR 707.9b, the copy-exception colour of Kaya,
// Intangible Slayer). The trigger is not optional.
func init() {
	Register(Spec{
		OracleID:     "5cce6146-94a6-442e-9890-6f2914cbcebb",
		Name:         "Penumbra Umbra",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("enchant creature you control", YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Penumbra Umbra — create a token that's a copy of enchanted creature, except the token is black",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					host, ok := enchantedCreatureAsItResolves(ctx)
					if !ok {
						return nil
					}
					return CreateTokenCopy{Copy: host, N: 1, Except: func(t *game.Card) {
						t.SetCopyExceptionColors("B")
					}}.Apply(ctx)
				}),
		},
		Replacements: []game.ReplacementEffect{UmbraArmor()},
	})
}
