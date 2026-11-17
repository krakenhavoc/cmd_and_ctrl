package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "79ab207f-e7ff-4236-b260-5a109a73fbed",
		Name:            "Zealot's Conviction",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			{
				Label:    "Corrupted — enchanted creature gets an additional +1/+0 as long as an opponent has three or more poison counters.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return AttachedToSource(target, g, source) && seedcoreCorrupted(g, source.Controller, source.InstanceID)
				},
				Apply: pumpBy(1, 0),
			},
			GrantToAttachedWhile(func(_ *game.Card, g *game.Game, source *game.Card) bool {
				return seedcoreCorrupted(g, source.Controller, source.InstanceID)
			}, "first strike"),
		},
	})
}
