package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guiding Spirit — Creature — Angel Spirit {1}{W}{U} 1/2 (Visions, S82 OLD,
// #5631):
//
//	"Flying
//	 {T}: If the top card of target player's graveyard is a creature card,
//	 put that card on top of that player's library."
//
// The "if" is not an intervening one on an activated ability: the top card
// is read as the ability resolves (CR 608.2), and the card goes to its
// owner's library, which is the graveyard's owner.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c474b52f-2f50-416e-b6c2-07043c7b339d",
		Name:            "Guiding Spirit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: If the top card of target player's graveyard is a creature card, put that card on top of that player's library.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.S82OldB788GuidingSpiritEffect,
		}},
	})
}
