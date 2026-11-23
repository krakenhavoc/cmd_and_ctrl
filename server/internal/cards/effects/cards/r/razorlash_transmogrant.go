package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "580996d4-db9a-4b57-a0e1-a2dbb524121a",
		Name:         "Razorlash Transmogrant",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{B}{B}: Return this card from your graveyard to the battlefield with a +1/+1 counter on it. This ability costs {4} less to activate if an opponent controls four or more nonbasic lands.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{4}{B}{B}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			CostModifiers: []game.CostModifier{
				effects.CostsLess(4, "This ability costs {4} less to activate if an opponent controls four or more nonbasic lands.",
					effects.CmdrB3AnOpponentControlsNonbasicLands(4)),
			},
			Effect: effects.CmdrB3ReturnThisFromGraveyardWithCounter(game.CounterPlusOne),
		}},
	})
}
