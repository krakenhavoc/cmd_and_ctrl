package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Walking Sponge — Creature — Sponge {1}{U}, 1/1 (Urza's Legacy, S82 OLD b716, #5623):
//
//	"{T}: Target creature loses your choice of flying, first strike, or trample until end of turn."
//
// The target is chosen on activation; the keyword is the controller's choice
// as the ability resolves (WalkingSpongeResolve), and the loss is a layer-6
// ability-removing effect until end of turn. A tap ability, so summoning
// sickness applies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "000e5d65-96c3-498b-bd01-72b1a1991850",
		Name:         "Walking Sponge",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature loses your choice of flying, first strike, or trample until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.WalkingSpongeResolve,
		}},
	})
}
