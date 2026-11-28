package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Savage Gorilla — Creature — Ape {4}{G}, 3/3 (Apocalypse, S82 OLD, #5614):
//
//	"{U}{B}, {T}, Sacrifice this creature: Target creature gets -3/-3 until
//	 end of turn. Draw a card."
//
// Phyrexian Denouncer's sacrifice-to-shrink row at -3/-3 with a mana cost,
// then the controller draws. The sacrifice is a cost, so the Gorilla is gone
// before the ability resolves; a target that left in response makes the whole
// ability fizzle (CR 608.2b).
//
// No simplification.
func init() {
	purpose := effects.ForTargets(effects.HarmUntilEOT(0, -3, -3))
	purpose.Draws = 1
	shrink := effects.BoostTargetUntilEOT("Savage Gorilla — -3/-3", -3, -3)
	effects.Register(effects.Spec{
		OracleID:     "6c876a57-c567-4d26-8ab6-a9c051e83ace",
		Name:         "Savage Gorilla",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}{B}, {T}, Sacrifice this creature: Target creature gets -3/-3 until end of turn. Draw a card.",
			Purpose: purpose,
			Cost:    effects.Plus(effects.ManaCost("{U}{B}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := shrink(g, item); err != nil {
					return err
				}
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
