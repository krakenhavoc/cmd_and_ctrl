package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rabble-Rouser — Creature — Goblin Shaman {3}{R}, 1/1 (Guildpact):
//
//	"Bloodthirst 1 (If an opponent was dealt damage this turn, this creature
//	 enters with a +1/+1 counter on it.)
//	 {R}, {T}: Attacking creatures get +X/+0 until end of turn, where X is
//	 this creature's power."
//
// Bloodthirst is the shared entry replacement (bloodthirst.go). X is the
// Rabble-Rouser's power as the ability resolves, or as it last was if it has
// left the battlefield (CR 608.2h); the set of attackers, any player's, is
// locked then (CR 611.2c). It can be activated outside combat, where it
// pumps nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1d591c00-7086-4337-bdb8-a5f35568f04e",
		Name:         "Rabble-Rouser",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.Bloodthirst(1, "Rabble-Rouser")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, {T}: Attacking creatures get +X/+0 until end of turn, where X is this creature's power.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				x := effects.B43PowerNowOrLastKnown(g, item.SourceCardID)
				return effects.BoostUntilEOT{
					Match: effects.AttackingCreature(),
					Power: x,
					Label: "Rabble-Rouser — attacking creatures get +X/+0",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
