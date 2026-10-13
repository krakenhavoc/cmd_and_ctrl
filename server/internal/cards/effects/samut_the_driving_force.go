package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Samut, the Driving Force — Legendary Creature — Human Warrior Cleric
// {3}{R}{G}{W}, 4/5:
//
//	"First strike, vigilance, haste
//	 Start your engines!
//	 Other creatures you control get +X/+0, where X is your speed.
//	 Noncreature spells you cast cost {X} less to cast, where X is your
//	 speed."
//
// ADR 0138 (#2122, #2711). Both X are the controller's speed, read live:
// the pump is a layer 7c static (every speed change bumps the layer
// version) and the discount is a cost reduction asked at each cast. A
// player with no speed has X = 0 (CR 702.179f). The reduction spends
// against generic mana only, as every reduction does (CR 601.2f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ee89c719-5ae8-4316-bfd0-cfe3684b5859",
		Name:            "Samut, the Driving Force",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "vigilance", "haste", StartYourEngines},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller &&
					target.InstanceID != source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power += YourSpeed(g, source.Controller)
			},
		}},
		CostModifiers: []game.CostModifier{
			CostsLessEach(func(q game.CostQuery) int { return YourSpeed(q.Game, q.Source.Controller) },
				"Noncreature spells you cast cost {X} less to cast, where X is your speed.",
				YourSpell(), NoncreatureSpell()),
		},
	})
}
