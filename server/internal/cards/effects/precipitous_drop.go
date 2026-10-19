package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Precipitous Drop — Enchantment — Aura {2}{B}:
//
//	"Enchant creature
//	 When this Aura enters, venture into the dungeon.
//	 Enchanted creature gets -2/-2. It gets -5/-5 instead as long as you
//	 've completed a dungeon."
//
// One layer 7c modification whose size reads the Aura's controller's
// completed dungeons; the layer pass bumps on EventDungeonCompleted, so
// the venture this Aura's own trigger makes can turn -2/-2 into -5/-5.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d60c8cfe-435c-419e-a44b-b47d49d1afef",
		Name:         "Precipitous Drop",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Triggered:    []game.TriggeredAbility{WhenThisEntersVenture("Precipitous Drop")},
		Static: []game.StaticAbility{{
			Label:     "Enchanted creature gets -2/-2, or -5/-5 as long as you've completed a dungeon.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: AttachedToSource,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 2
				if completedADungeonSource(g, source) {
					n = 5
				}
				c.Power -= n
				c.Toughness -= n
			},
		}},
	})
}
