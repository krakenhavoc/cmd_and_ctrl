package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wall of Forgotten Pharaohs — {2} Artifact Creature — Wall 0/4 (Hour of Devastation):
//
//	"Defender
//	 {T}: This creature deals 1 damage to target player or planeswalker.
//	 Activate only if you control a Desert or there is a Desert card in your
//	 graveyard."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7ab15fcd-8371-4b3c-b534-e9ef44249213",
		Name:            "Wall of Forgotten Pharaohs",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:     "{T}: This creature deals 1 damage to target player or planeswalker. Activate only if you control a Desert or there is a Desert card in your graveyard.",
			Purpose:   effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:      effects.TapCost(),
			Targets:   effects.TargetPlayerOrPlaneswalker(),
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool { return effects.OldB240YouHaveADesert(g, controller) },
			Effect:    effects.SourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
