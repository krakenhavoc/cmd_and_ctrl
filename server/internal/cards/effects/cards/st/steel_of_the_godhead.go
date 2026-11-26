package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Steel of the Godhead — Enchantment — Aura {2}{W/U} (Shadowmoor, S82 older sets):
//
//	"Enchant creature
//	 As long as enchanted creature is white, it gets +1/+1 and has lifelink.
//	 As long as enchanted creature is blue, it gets +1/+1 and can't be blocked."
//
// Shield of the Oversoul's shape: each clause is a +1/+1 static and a granted
// keyword, both conditional on the enchanted creature's colour as the layers
// leave it (colour is layer 5, before abilities and P/T), so a creature of both
// colours gets both clauses and a colourless one neither.
//
// No simplification.
func init() {
	one := effects.B43AttachedHasColor("W")
	two := effects.B43AttachedHasColor("U")
	effects.Register(effects.Spec{
		OracleID:     "90d2e092-6eb8-4d13-bdad-19e7325a8b36",
		Name:         "Steel of the Godhead",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.B43PumpAttachedWhen(one, 1, 1),
			effects.B43GrantToAttachedWhen(one, "lifelink"),
			effects.B43PumpAttachedWhen(two, 1, 1),
		},
		// "can't be blocked" is a block rule, read off the attacker's colours as the
		// block is checked.
		BlockRules: []game.BlockRule{effects.CantBeBlockedWhile(effects.OnAttached(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
			for _, col := range c.EffectiveColors() {
				if col == "U" {
					return true
				}
			}
			return false
		})},
	})
}
