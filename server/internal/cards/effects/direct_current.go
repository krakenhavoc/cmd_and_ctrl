package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Direct Current — Sorcery {1}{R}{R} (Ravnica: Clue Edition):
//
//	"Direct Current deals 2 damage to any target.
//	 Jump-start (You may cast this card from your graveyard by discarding a
//	 card in addition to paying its other costs. Then exile this card.)"
//
// Jump-start is the shared cmdr2S39JumpStart (CR 702.133a).
func init() {
	Register(Spec{
		OracleID:         "54e34440-c890-4a07-83c6-63a2f7a65639",
		Name:             "Direct Current",
		Completeness:     CompletenessFull,
		Targets:          TargetAny(),
		Purpose:          game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 2)), Answers: game.AnswerRemove},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{cmdr2S39JumpStart("{1}{R}{R}")},
		OnResolve:        damageToFirstTarget(2),
	})
}
