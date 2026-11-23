package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Infernal Rebirth — Instant {3}{B} (Marvel Super Heroes Commander,
// #4455):
//
//	"Return one or two target creature cards from your graveyard to
//	 your hand.
//	 Basic landcycling {2}"
//
// One or two targets (CR 601.2c); a pick that left the graveyard in
// response is skipped and the other still returns (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "05a1e1ee-67f0-4158-a012-01612bc9cc70",
		Name:         "Infernal Rebirth",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("one or two target creature cards from your graveyard", effects.YouOwn(), effects.Creature()).WithCount(1, 2),
		OnResolve:    effects.ResolveSteps(effects.Instruction(effects.ReturnLegalGraveyardTargetsToHand)),
		Activated:    []effects.ActivatedAbility{effects.BasicLandcycling("{2}")},
	})
}
