package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mishra's Domination — Enchantment — Aura {1}{R} (The Brothers' War):
//
//	"Enchant creature
//	 As long as you control enchanted creature, it gets +2/+2. Otherwise,
//	 it can't block."
//
// "You" is the Aura's controller; the condition is read live, so a creature
// that changes hands flips from one half to the other.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ede2354a-64b0-402c-8fbe-9da90f811a9b",
		Name:         "Mishra's Domination",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			b43PumpAttachedWhen(func(host *game.Card, g *game.Game, source *game.Card) bool {
				return AttachedToSource(host, g, source) && host.Controller == source.Controller
			}, 2, 2),
			RestrictAttachedWhile(func(host *game.Card, _ *game.Game, source *game.Card) bool {
				return host.Controller != source.Controller
			}, game.CantBlock),
		},
	})
}
