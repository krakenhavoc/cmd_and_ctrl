package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lambholt Pacifist // Lambholt Butcher — {1}{G} Creature — Human Shaman
// Werewolf 3/3 // Creature — Werewolf 4/4 (Shadows over Innistrad, S82 OLD
// batch 290, #5543):
//
//	Front: "This creature can't attack unless you control a creature with
//	        power 4 or greater.
//	        At the beginning of each upkeep, if no spells were cast last
//	        turn, transform this creature."
//	Back:  "At the beginning of each upkeep, if a player cast two or more
//	        spells last turn, transform this creature."
//
// The restriction is a layer-6 CantAttack on the source while its controller
// controls no creature with power 4 or greater (the Pacifist itself is 3/3
// and does not count, but a pumped one does), read through the layers and
// checked as attackers are declared (CR 508.1c). The upkeep pair is the
// original-Innistrad werewolf triggers (innistrad_werewolves.go).
//
// No simplification.
func init() {
	const oracle = "e7d6ef32-616f-4c21-bcdd-467ce8296c31"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Lambholt Pacifist",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "This creature can't attack unless you control a creature with power 4 or greater.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.SelfOnly(target, g, source) && !effects.B24ControlsCreatureWithPowerAtLeast(g, source.Controller, 4)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfNoSpellsLastTurn("Lambholt Pacifist"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Lambholt Butcher",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Lambholt Butcher"),
		},
	})
}
