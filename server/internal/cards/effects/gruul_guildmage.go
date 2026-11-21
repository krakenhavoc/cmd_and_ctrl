package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "75645b1f-0e12-4a05-b47f-0d2e2616c3ca",
		Name:         "Gruul Guildmage",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{3}{R}, Sacrifice a land: This creature deals 2 damage to target player or planeswalker.",
				Purpose: ForTargets(DamageToTarget(0, 2)),
				Cost:    Plus(ManaCost("{3}{R}"), SacrificeN(1, "a land", Land())),
				Targets: targetPlayerOrPlaneswalker(),
				Effect:  sourceDealsDamageToEachLegalTarget(2),
			},
			{
				Label:   "{3}{G}: Target creature gets +2/+2 until end of turn.",
				Purpose: withAnswers(ForTargets(BenefitUntilEOT(0, 2, 2)), game.AnswerPump),
				Cost:    ManaCost("{3}{G}"),
				Targets: TargetCreature("target creature"),
				Effect:  boostTargetUntilEOT("Gruul Guildmage — +2/+2", 2, 2),
			},
		},
	})
}
