package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Krosan Restorer — Creature — Human Druid {2}{G}, 1/2 (Dominaria
// Remastered):
//
//	"{T}: Untap target land.
//	 Threshold — {T}: Untap up to three target lands. Activate only if
//	 seven or more cards are in your graveyard."
//
// Hope Tender's untap (B22UntapEachLegalTarget) with a {T} cost. Threshold
// is the second ability's activation condition (CR 602.1b), checked when it
// is activated: seven or more cards of any kind in your graveyard
// (GraveyardAtLeast). "Up to three" is zero to three targets.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea845951-0158-4646-acd9-3b0b28da40b1",
		Name:         "Krosan Restorer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Untap target land.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetPermanent("target land", effects.Land()),
				Effect:  effects.B22UntapEachLegalTarget,
			},
			{
				Label:     "Threshold — {T}: Untap up to three target lands. Activate only if seven or more cards are in your graveyard.",
				Purpose:   game.Purpose{Answers: game.AnswerValue},
				Cost:      effects.TapCost(),
				Targets:   effects.TargetPermanent("up to three target lands", effects.Land()).WithCount(0, 3),
				Condition: effects.GraveyardAtLeast(7, nil),
				Effect:    effects.B22UntapEachLegalTarget,
			},
		},
	})
}
