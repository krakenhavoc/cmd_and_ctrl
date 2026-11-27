package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kjeldoran Javelineer — Creature — Human Soldier {W}, 1/2 (Coldsnap, S82 OLD b558):
//
//	"Cumulative upkeep {1}
//	 {T}: This creature deals damage equal to the number of age counters on it
//	 to target attacking or blocking creature."
//
// The damage is the Javelineer's age counters as the ability resolves (live,
// or last-known if it has left, CR 608.2h); with none it deals no damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "95921d1e-0487-4274-b797-8652e94fc4bc",
		Name:         "Kjeldoran Javelineer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Kjeldoran Javelineer — cumulative upkeep {1}", "{1}"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals damage equal to the number of age counters on it to target attacking or blocking creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				n := effects.AgeCountersAsItLeft(g, item)
				targets := ctx.LegalTargets()
				if n <= 0 || len(targets) == 0 {
					return nil
				}
				return effects.DealDamage{Source: ctx.Source(), Target: targets[0].ID, Amount: n}.Apply(ctx)
			},
		}},
	})
}
