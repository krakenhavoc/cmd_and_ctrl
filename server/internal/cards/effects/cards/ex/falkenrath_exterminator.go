package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Falkenrath Exterminator — Creature — Vampire Archer {1}{R}, 1/1:
//
//	"Whenever this creature deals combat damage to a player, put a +1/+1
//	 counter on it.
//	 {2}{R}: This creature deals damage to target creature equal to the
//	 number of +1/+1 counters on this creature."
//
// The counter is Ancient Copper Dragon's shape (a combat-damage trigger
// on the source). The activated ability reads the counters when it
// resolves, as the source now is or, if it has left the battlefield, as
// it last was (CR 608.2h), and deals damage only if the target is still
// legal; a Exterminator with no counters deals none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fd5fc69e-1bba-4f39-88c1-5966fa941f32",
		Name:         "Falkenrath Exterminator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Falkenrath Exterminator — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}: This creature deals damage to target creature equal to the number of +1/+1 counters on this creature.",
			Cost:    effects.ManaCost("{2}{R}"),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				n := effects.TrekB1CountersNowOrLastKnown(g, item.SourceCardID, game.CounterPlusOne)
				if n <= 0 {
					return nil
				}
				for _, t := range ctx.LegalTargets() {
					return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
