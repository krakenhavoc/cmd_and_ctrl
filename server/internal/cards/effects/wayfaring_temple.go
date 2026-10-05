package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wayfaring Temple — Creature — Elemental {1}{G}{W}, */*:
//
//	"Wayfaring Temple's power and toughness are each equal to the
//	 number of creatures you control.
//	 Whenever this creature deals combat damage to a player, populate.
//	 (Create a token that's a copy of a creature token you control.)"
//
// The P/T is a Layer 7a characteristic-defining ability (Adeline's
// shape, setting both numbers); the count includes the Temple itself.
// The populate is a combat-damage trigger on the Temple alone, using
// the shared ThisDealtCombatDamageToAPlayer.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c0525645-9b5a-488d-b4d1-e80f16a1e4e1",
		Name:         "Wayfaring Temple",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := b04CreaturesControlled(g, source.Controller)
				c.Power = n
				c.Toughness = n
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, ThisDealtCombatDamageToAPlayer,
				"Wayfaring Temple — populate", Do(Populate{})),
		},
	})
}
