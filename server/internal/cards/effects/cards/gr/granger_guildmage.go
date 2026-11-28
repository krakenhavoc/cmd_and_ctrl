package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Granger Guildmage — Creature — Human Wizard {G}, 1/1 (Mirage, #5633):
//
//	"{R}, {T}: This creature deals 1 damage to any target and 1 damage to you.
//	 {W}, {T}: Target creature gains first strike until end of turn."
//
// The ping and the damage to its controller are one simultaneous event, and an
// illegal target at resolution does nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0230fdda-d916-457e-9dfd-61783fcf4652",
		Name:         "Granger Guildmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{R}, {T}: This creature deals 1 damage to any target and 1 damage to you.",
				Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost()),
				Targets: effects.TargetAny(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Effect: effects.OldB793DealsOneToTargetAndTo(func(ctx *effects.Context) uuid.UUID {
					return ctx.Controller()
				}),
			},
			{
				Label:   "{W}, {T}: Target creature gains first strike until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "first strike"))},
				Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.EbTargetCreatureGainsUntilEOT("first strike", "Granger Guildmage — first strike until end of turn"),
			},
		},
	})
}
