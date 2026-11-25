package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arcades, the Strategist — Legendary Creature — Elder Dragon {1}{G}{W}{U},
// 3/5 (Core Set 2019, #5527):
//
//	"Flying, vigilance
//	 Whenever a creature you control with defender enters, draw a card.
//	 Each creature you control with defender assigns combat damage equal to
//	 its toughness rather than its power and can attack as though it didn't
//	 have defender."
//
// Felothar the Steadfast's two statics, narrowed to creatures you control
// with defender (read through the layered view, CR 613.1f): the toughness
// rule is DamageByToughnessAlways (CR 510.1c), and the waiver is a
// DefenderWaiver (CR 702.3b), which leaves the keyword on the creature so
// the first static still finds it. The draw reads the entering creature's
// keywords as it enters, Arcades itself having none.
//
// No simplification.
func init() {
	yourDefenders := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller && target.IsCreature() && game.HasKeyword(target, "defender")
	}
	effects.Register(effects.Spec{
		OracleID:        "e7fc5ad6-f2f0-4b06-a61b-05022db9d93b",
		Name:            "Arcades, the Strategist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Static: []game.StaticAbility{{
			Layer:     game.Layer6Ability,
			Label:     "Each creature you control with defender assigns combat damage equal to its toughness rather than its power.",
			AppliesTo: yourDefenders,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.DamageByToughnessAlways = true
			},
		},
			effects.DefenderWaiverStatic("Each creature you control with defender can attack as though it didn't have defender.",
				yourDefenders),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && game.HasKeyword(&c, "defender")
			}, "Arcades, the Strategist — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
