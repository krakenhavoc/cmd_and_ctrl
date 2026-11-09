package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Towashi — Plane — Kamigawa (#4870):
//
//	"Modified creatures you control have trample and "Whenever this creature
//	 deals combat damage to a player or planeswalker, draw a card." (Equipment,
//	 Auras you control, and counters are modifications.)
//	 Whenever chaos ensues, distribute three +1/+1 counters among one, two, or
//	 three target creatures you control."
//
// "Modified" is CR 700.9 (b07IsModified): a counter, an Equipment, or an Aura
// its controller controls. The static grants a layer-6 trample and a bundle
// carrying the draw trigger to every such creature the planar controller
// controls. The chaos ability announces one to three targets with the three
// counters divided as it goes on the stack, each target getting at least one
// (CR 601.2d); a target that is illegal on resolution loses its share (the
// ruling).
//
// No simplification.
func init() {
	modified := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.Controller == source.Controller && b07IsModified(g, *target)
	}
	Register(Spec{
		OracleID:     "0c1cb8e2-8ed4-4a0e-926b-4df5466c607c",
		Name:         "Towashi",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: towashiDrawGrant,
			Triggered: []game.TriggeredAbility{
				On(game.EventDealDamage, ThisDealtCombatDamageToAPlayerOrPlaneswalker,
					"Towashi — draw a card", Do(DrawCards{N: 1})),
			},
			Text: "Whenever this creature deals combat damage to a player or planeswalker, draw a card.",
		}},
		Static: []game.StaticAbility{
			{
				Layer:     game.Layer6Ability,
				Label:     "Modified creatures you control have trample.",
				AppliesTo: modified,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !eotHasAbility(c.Abilities, "trample") {
						c.Abilities = append(c.Abilities, "trample")
					}
				},
			},
			GrantAbilities(modified, towashiDrawGrant),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverChaosEnsues("Towashi — distribute three +1/+1 counters among one, two, or three target creatures you control",
				func(g *game.Game, item *game.StackItem) error {
					return PutDividedCounters(NewContext(g, item), game.CounterPlusOne)
				}),
				TargetCreature("one, two, or three target creatures you control", YouControl()).WithCount(1, 3).Dividing(Divide(3))),
		},
	})
}

// towashiDrawGrant is the granted bundle's key. An on-disk key.
const towashiDrawGrant = "towashi/draw-on-combat-damage"
