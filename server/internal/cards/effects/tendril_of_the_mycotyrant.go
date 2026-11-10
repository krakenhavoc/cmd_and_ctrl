package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tendril of the Mycotyrant — Creature — Fungus Wizard {1}{G}, 2/2 (The Lost
// Caverns of Ixalan, #4470):
//
//	"{5}{G}{G}: Put seven +1/+1 counters on target noncreature land you
//	 control. It becomes a 0/0 Fungus creature with haste. It's still a
//	 land."
//
// Awaken's counters-then-animate (CR 702.113a) with the Fungus named
// (game.AnimateLandWithCountersForEffect): the target is re-checked as the
// ability resolves (CR 608.2b), the seven counters go on through the
// counter pipeline (so Hardened Scales and Doubling Season apply), and the
// land becomes a creature permanently, a land still, with haste (CR 305.7),
// only if it is still the object that was targeted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d2824464-963b-48db-a482-cc08a0a97332",
		Name:         "Tendril of the Mycotyrant",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{5}{G}{G}: Put seven +1/+1 counters on target noncreature land you control. It becomes a 0/0 Fungus creature with haste. It's still a land.",
			Cost:    ManaCost("{5}{G}{G}"),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Targets: TargetPermanent("target noncreature land you control", Land(), Noncreature(), YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return animateLandWithCounters(ctx, t.ID, 7, "Fungus")
			},
		}},
	})
}
