package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Griffin Guide — Enchantment — Aura {2}{W} (Dominaria Remastered, #5948):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and has flying.
//	 When enchanted creature dies, create a 2/2 white Griffin creature token
//	 with flying."
//
// The pump and the flying are the shared Aura statics. The death trigger is
// the Aura's (it watches the creature it was attached to, by last-known
// information), so the Aura's controller, not the creature's, makes the
// Griffin.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c5323a43-82de-4340-8578-b3ffcc66f8fa",
		Name:         "Griffin Guide",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Tokens: 1, CreatureTokens: 1},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("flying"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Griffin Guide — create a 2/2 white Griffin creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 white Griffin with flying"), N: 1})),
		},
	})
}
