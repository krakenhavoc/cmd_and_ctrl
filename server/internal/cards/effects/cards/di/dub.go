package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dub — Enchantment — Aura {2}{W} (Core Set 2021, #5511):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2, has first strike, and is a Knight in
//	 addition to its other types."
//
// Three layered effects of one static ability: the Knight subtype is added in
// layer 4 (a word of the Aura's text, CR 612.2), first strike in layer 6 and
// the bonus in layer 7c.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e3ca6b70-01d2-44e9-937a-a2d50bbf9d10",
		Name:         "Dub",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("first strike"),
			effects.AttachedIsAlso("Knight"),
		},
	})
}
