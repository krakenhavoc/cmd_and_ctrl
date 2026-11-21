package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Splitting Headache — Sorcery {3}{B}:
//
//	"Choose one —
//	 • Target player discards two cards.
//	 • Target player reveals their hand. You choose a card from it.
//	   That player discards that card."
//
// The first bullet is the ordinary discard: the player chooses
// (CR 701.9b). The second is the revealed-hand pick (ADR 0116) with no
// filter, so any card may be chosen, a land included. Either bullet can
// target the caster.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3a5937de-5957-47a9-9bd4-0dc7f8b833ef",
		Name:         "Splitting Headache",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Target player discards two cards.",
				effects.TargetPlayer("target player"),
				effects.TheModesTargetDiscards(2)),
			effects.ModeDoing("Target player reveals their hand. You choose a card from it. That player discards that card.",
				effects.TargetPlayer("target player"),
				effects.ModeTargetRevealsYouChooseDiscard(nil, "")),
		),
	})
}
