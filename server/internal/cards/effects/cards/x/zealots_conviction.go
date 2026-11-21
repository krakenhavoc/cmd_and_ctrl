package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zealot's Conviction — Enchantment — Aura {W}:
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets +1/+1.
//	 Corrupted — As long as an opponent has three or more poison counters,
//	 enchanted creature gets an additional +1/+0 and has first strike."
//
// The +1/+1 is unconditional. The corrupted rider is a second layer 7c
// pump and a layer 6 grant, each re-answered on every recompute against the
// Aura's controller's opponents.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "79ab207f-e7ff-4236-b260-5a109a73fbed",
		Name:            "Zealot's Conviction",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			{
				Label:    "Corrupted — enchanted creature gets an additional +1/+0 as long as an opponent has three or more poison counters.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return effects.AttachedToSource(target, g, source) && effects.SeedcoreCorrupted(g, source.Controller, source.InstanceID)
				},
				Apply: effects.PumpBy(1, 0),
			},
			effects.GrantToAttachedWhile(func(_ *game.Card, g *game.Game, source *game.Card) bool {
				return effects.SeedcoreCorrupted(g, source.Controller, source.InstanceID)
			}, "first strike"),
		},
	})
}
