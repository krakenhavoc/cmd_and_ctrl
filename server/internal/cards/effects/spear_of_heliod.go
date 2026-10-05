package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spear of Heliod — Legendary Enchantment Artifact {1}{W}{W}:
//
//	"Creatures you control get +1/+1.
//	 {1}{W}{W}, {T}: Destroy target creature that dealt damage to you
//	 this turn."
//
// The anthem is Glorious Anthem's. The ability's target clause reads
// the per-turn record of which creature objects dealt damage to you
// (#2149) and is checked again as the ability resolves (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:     "fd66aa66-75a2-41b6-8d57-dc4b9c221ccb",
		Name:         "Spear of Heliod",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}{W}, {T}: Destroy target creature that dealt damage to you this turn",
			Cost:    Plus(ManaCost("{1}{W}{W}"), TapCost()),
			Targets: TargetCreature("target creature that dealt damage to you this turn", DealtDamageToYouThisTurn()),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
