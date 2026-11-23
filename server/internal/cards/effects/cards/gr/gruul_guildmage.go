package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gruul Guildmage — Creature — Human Shaman {R/G}{R/G}, 2/2 (Ravnica
// Remastered, #5948):
//
//	"({R/G} can be paid with either {R} or {G}.)
//	 {3}{R}, Sacrifice a land: This creature deals 2 damage to target player
//	 or planeswalker.
//	 {3}{G}: Target creature gets +2/+2 until end of turn."
//
// The first row is Aethertorch Renegade's player-or-planeswalker ping with a
// land sacrifice paid as the ability is activated (CR 602.2b); the damage is
// dealt by the Guildmage, as it last existed if it has gone (CR 608.2h). The
// second is Alaborn Veteran's +2/+2 until end of turn on any creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75645b1f-0e12-4a05-b47f-0d2e2616c3ca",
		Name:         "Gruul Guildmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{3}{R}, Sacrifice a land: This creature deals 2 damage to target player or planeswalker.",
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
				Cost:    effects.Plus(effects.ManaCost("{3}{R}"), effects.SacrificeN(1, "a land", effects.Land())),
				Targets: effects.TargetPlayerOrPlaneswalker(),
				Effect:  effects.SourceDealsDamageToEachLegalTarget(2),
			},
			{
				Label:   "{3}{G}: Target creature gets +2/+2 until end of turn.",
				Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)), game.AnswerPump),
				Cost:    effects.ManaCost("{3}{G}"),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.BoostTargetUntilEOT("Gruul Guildmage — +2/+2", 2, 2),
			},
		},
	})
}
