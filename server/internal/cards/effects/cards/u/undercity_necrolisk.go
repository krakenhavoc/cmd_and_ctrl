package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Undercity Necrolisk — Creature — Zombie Lizard {3}{B}, 3/3 (Guilds of
// Ravnica):
//
//	"{1}, Sacrifice another creature: Put a +1/+1 counter on this creature.
//	 It gains menace until end of turn. Activate only as a sorcery."
//
// Blood Host's sacrifice-another-creature row with a sorcery-speed gate, and
// Endling's keyword grant until end of turn on the source. Both halves happen
// only while the Necrolisk is still the permanent that activated (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b3418d92-81de-4575-b8da-563be6b31b91",
		Name:         "Undercity Necrolisk",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}, Sacrifice another creature: Put a +1/+1 counter on this creature. It gains menace until end of turn. Activate only as a sorcery.",
			Purpose:      game.Purpose{SourceCounters: 1},
			Cost:         effects.Plus(effects.ManaCost("{1}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := effects.PutACounterOnThis(game.CounterPlusOne)(g, item); err != nil {
					return err
				}
				if !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
					return nil
				}
				return effects.ThisGainsKeywordUntilEndOfTurn("menace", "Undercity Necrolisk — menace")(g, item)
			},
		}},
	})
}
