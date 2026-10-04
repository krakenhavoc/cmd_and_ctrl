package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mikaeus, the Unhallowed — Legendary Creature — Zombie Cleric
// {3}{B}{B}{B}, 5/5:
//
//	"Intimidate
//	 Whenever a Human deals damage to you, destroy it.
//	 Other non-Human creatures you control get +1/+1 and have undying."
//
// The +1/+1 is not a counter, so it never stops undying (ruling). A
// non-Human that dies alongside Mikaeus still has the undying it was
// given as it last existed, and returns (ruling; CR 603.10a). The damage
// trigger fires on any Human's damage to you, combat or not, yours
// included, and destroys it if it is still on the battlefield, even
// after Mikaeus has left (ruling).
//
// No simplification.
func init() {
	nonHumans := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return TribeFilter{Others: true, YoursOnly: true}.Matches(target, g, source) && !target.HasSubtype("Human")
	}
	Register(Spec{
		OracleID:        "5d27c63e-d1ef-48af-b51d-01ebc6daeac9",
		Name:            "Mikaeus, the Unhallowed",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"intimidate"},
		Static: []game.StaticAbility{
			{
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: nonHumans,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power++
					c.Toughness++
				},
			},
			KeywordGrant(nonHumans, game.KeywordUndying),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Amount <= 0 || ev.Target != source.Controller {
					return false
				}
				if z := g.FindCardZoneForEffect(ev.Source); z == nil || z.Kind != game.ZoneBattlefield {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.Source)
				return ok && c.HasSubtype("Human")
			}, "Mikaeus, the Unhallowed — destroy the Human that damaged you", destroyTheCreatureThatDealtTheDamage),
		},
	})
}
