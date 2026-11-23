package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mushroom Watchdogs — Creature — Dog {1}{G}, 2/2 (The Lord of the Rings:
// Tales of Middle-earth Commander):
//
//	"Sacrifice a Food: Put a +1/+1 counter on this creature. It gains
//	 vigilance until end of turn. Activate only as a sorcery."
//
// Bosco's "Sacrifice a Food" cost (any Food artifact you control, token or
// not) on Adaptive Gemguard's sorcery-speed activation. The counter and
// the vigilance go on the Watchdogs only while it is still the permanent
// that activated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "82259c84-773d-4670-9b09-d73f5a3be7ea",
		Name:         "Mushroom Watchdogs",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "Sacrifice a Food: Put a +1/+1 counter on this creature. It gains vigilance until end of turn. Activate only as a sorcery.",
			Cost:         game.AbilityCost{SacrificeOther: effects.SacrificeSpec("a Food", effects.HasSubtype("Food"))},
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				if err := effects.PlusOneCountersOnThis(1)(g, item); err != nil {
					return err
				}
				return effects.ThisGainsKeywordUntilEndOfTurn("vigilance", "Mushroom Watchdogs — vigilance until end of turn")(g, item)
			},
		}},
	})
}
