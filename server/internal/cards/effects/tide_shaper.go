package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tide Shaper — Creature — Merfolk Wizard {U}, 1/1:
//
//	"Kicker {1} (You may pay an additional {1} as you cast this spell.)
//	 When this creature enters, if it was kicked, target land becomes an
//	 Island for as long as this creature remains on the battlefield.
//	 This creature gets +1/+1 as long as an opponent controls an Island."
//
// "If it was kicked" is an intervening if (CR 603.4), read off the
// permanent when the ability would trigger, as Gatekeeper of Malakir's is:
// an unkicked Tide Shaper puts nothing on the stack. The effect is ADR
// 0109 §1's (#1881) CR 305.7 type set for as long as the Shaper stays on
// the battlefield (CR 611.2b): the land is an Island (other subtypes stay,
// CR 205.1a), loses its rules-text abilities and taps for {U} (CR 305.6).
// Aimed at an opponent's land it switches on the Shaper's own +1/+1, a
// layer-7c static read live.
//
// No simplification.
func init() {
	island := func(c game.Card) bool { return c.HasSubtype("Island") }
	opponentHasIsland := OpponentControlsAtLeast(1, island)
	Register(Spec{
		OracleID:      "cf4be71e-0a9a-47a3-b0ec-43e04c3c07a0",
		Name:          "Tide Shaper",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Kicker("{1}")},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && opponentHasIsland(g, source.Controller, source.InstanceID)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: AllOf(Self, func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return game.CardKickedTimes(*source) > 0
			}),
			Key:     "Tide Shaper — kicked, target land becomes an Island",
			Targets: TargetPermanent("target land", Land()),
			Effect:  TargetLandBecomesWhileSourceRemains("Tide Shaper", "Island"),
		}},
	})
}
