package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "fd5fc69e-1bba-4f39-88c1-5966fa941f32",
		Name:         "Falkenrath Exterminator",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Falkenrath Exterminator — put a +1/+1 counter on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}: This creature deals damage to target creature equal to the number of +1/+1 counters on this creature.",
			Cost:    ManaCost("{2}{R}"),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				n := trekB1CountersNowOrLastKnown(g, item.SourceCardID, game.CounterPlusOne)
				if n <= 0 {
					return nil
				}
				for _, t := range ctx.LegalTargets() {
					return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
