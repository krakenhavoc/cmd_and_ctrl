package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Word of Seizing — Instant {3}{R}{R} (Time Spiral, #5596):
//
//	"Split second (As long as this spell is on the stack, players can't
//	 cast spells or activate abilities that aren't mana abilities.)
//	 Untap target permanent and gain control of it until end of turn.
//	 It gains haste until end of turn."
//
// Act of Treason for any permanent, behind split second (CR 702.61a). The
// permanent is untapped first, then taken until end of turn, then given
// haste until end of turn, in printed order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b80d2841-03b0-42d3-ac9b-51e793b1dc24",
		Name:            "Word of Seizing",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordSplitSecond},
		Targets:         effects.TargetPermanent("target permanent"),
		OnResolve:       effects.TspB1UntapGainControlAndHaste("Word of Seizing"),
	})
}
