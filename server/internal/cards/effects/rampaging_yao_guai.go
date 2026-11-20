package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rampaging Yao Guai — Creature — Bear Mutant {X}{G}{G}{G}, 2/2 (Fallout,
// #5948):
//
//	"Vigilance, trample
//	 This creature enters with X +1/+1 counters on it.
//	 When this creature enters, destroy any number of target artifacts
//	 and/or enchantments with total mana value X or less."
//
// The counters are CR 614.1c's "enters with X" (XCounters). The trigger
// reads the X the creature was cast with (CR 107.3m, game.XFromCast): its
// clause is "any number" of artifacts and/or enchantments with a total mana
// value ceiling of that X, bound as the trigger is put on the stack and
// summed again as it resolves (TotalManaValueAtMostX, CR 608.2b: when the
// picks still legal no longer fit, none is destroyed). A permanent whose cost
// can't be read fits no total, and one with mana value 0 fits any. Cast
// without X (put onto the battlefield some other way) the ceiling is 0.
//
// No simplification.
func init() {
	targets := TargetPermanent("any number of target artifacts and/or enchantments with total mana value X or less",
		Or(Artifact(), Enchantment())).WithCount(0, 0).WithTotal(TotalManaValueAtMostX())
	Register(Spec{
		OracleID:        "d37e8f75-c7cb-4270-bb2f-0125e972ed15",
		Name:            "Rampaging Yao Guai",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
		XMatters:        true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			XCounters(game.CounterPlusOne),
		},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Key:       "Rampaging Yao Guai — destroy any number of target artifacts and/or enchantments with total mana value X or less",
			XFromCast: true,
			Targets:   targets,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				var ids []uuid.UUID
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						ids = append(ids, t.ID)
					}
				}
				if len(ids) == 0 {
					return nil
				}
				ctx.Game.DestroyPermanentsForEffect(ids)
				return nil
			},
		}},
	})
}
