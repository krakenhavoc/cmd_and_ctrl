package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Razorlash Transmogrant — Artifact Creature — Zombie {2}, 3/1 (Foundations
// Commander):
//
//	"This creature can't block.
//	 {4}{B}{B}: Return this card from your graveyard to the battlefield with
//	 a +1/+1 counter on it. This ability costs {4} less to activate if an
//	 opponent controls four or more nonbasic lands."
//
// Bloodghast's can't-block. The ability is activated from the graveyard
// (Resurrected Cultist's shape) and puts the card back with the counter as
// it enters. The discount is the ability's own cost modifier, counted as the
// cost is determined (CR 602.2b, 601.2f): {4} off the generic part when some
// opponent controls four or more nonbasic lands, leaving {B}{B}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "580996d4-db9a-4b57-a0e1-a2dbb524121a",
		Name:         "Razorlash Transmogrant",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Activated: []ActivatedAbility{{
			Label:   "{4}{B}{B}: Return this card from your graveyard to the battlefield with a +1/+1 counter on it. This ability costs {4} less to activate if an opponent controls four or more nonbasic lands.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{4}{B}{B}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			CostModifiers: []game.CostModifier{
				CostsLess(4, "This ability costs {4} less to activate if an opponent controls four or more nonbasic lands.",
					cmdrB3AnOpponentControlsNonbasicLands(4)),
			},
			Effect: cmdrB3ReturnThisFromGraveyardWithCounter(game.CounterPlusOne),
		}},
	})
}
