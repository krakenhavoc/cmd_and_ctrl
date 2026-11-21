package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "82259c84-773d-4670-9b09-d73f5a3be7ea",
		Name:         "Mushroom Watchdogs",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "Sacrifice a Food: Put a +1/+1 counter on this creature. It gains vigilance until end of turn. Activate only as a sorcery.",
			Cost:         game.AbilityCost{SacrificeOther: sacrificeSpec("a Food", HasSubtype("Food"))},
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				if err := plusOneCountersOnThis(1)(g, item); err != nil {
					return err
				}
				return thisGainsKeywordUntilEndOfTurn("vigilance", "Mushroom Watchdogs — vigilance until end of turn")(g, item)
			},
		}},
	})
}
