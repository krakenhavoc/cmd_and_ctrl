package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Massive Raid — Instant {1}{R}{R} (Gatecrash, S82 OLD #5565):
//
//	"Massive Raid deals damage to any target equal to the number of
//	 creatures you control."
//
// Dogpile's body with every creature you control counted instead of the
// attackers, read once as the spell resolves; with no creatures it deals no
// damage (CR 120.8). A target gone by then makes it do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "032367fd-e869-463e-aa0f-65e3c6567e60",
		Name:         "Massive Raid",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return massiveRaidDamage(ctx.Game, item)
		},
	})
}

// massiveRaidDamage is the spell's damage, counted once as it resolves.
var massiveRaidDamage = effects.SourceDealsComputedDamageToEachLegalTarget(func(ctx *effects.Context) int {
	return effects.B04CreaturesControlled(ctx.Game, ctx.Controller())
})
