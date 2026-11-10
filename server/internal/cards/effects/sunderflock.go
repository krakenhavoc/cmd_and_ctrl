package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sunderflock — Creature — Elemental {7}{U}{U}, 5/5 (Lorwyn Eclipsed,
// #4458):
//
//	"This spell costs {X} less to cast, where X is the greatest mana value
//	 among Elementals you control.
//	 Flying
//	 When this creature enters, if you cast it, return all non-Elemental
//	 creatures to their owners' hands."
//
// The discount is a self cost modifier like Ghalta's, spending generic
// mana only (CR 601.2f), with X the greatest mana value among the
// caster's Elementals as the cost is determined; an {X} in a permanent's
// cost is 0 on the battlefield, so it counts as 0. "If you cast it" is
// Cresting Mosasaurus's intervening if (CR 603.4), and the bounce is the
// same simultaneous return of every creature that is not an Elemental.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c01b221f-3993-43ab-8cde-b42fb344e2a7",
		Name:            "Sunderflock",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(stdB63GreatestElementalManaValue(),
				"This spell costs {X} less to cast, where X is the greatest mana value among Elementals you control."),
		},
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepBounce, Partial: true}},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(Self, crestingMosasaurusYouCastIt),
				"Sunderflock — return all non-Elemental creatures to their owners' hands",
				Do(BounceAllMatching{Match: And(Creature(), Not(HasSubtype("Elemental")))})),
		},
	})
}
