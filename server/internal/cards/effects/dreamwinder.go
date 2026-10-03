package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dreamwinder — Creature — Serpent {3}{U}, 4/3:
//
//	"This creature can't attack unless defending player controls an
//	 Island.
//	 {U}, Sacrifice an Island: Target land becomes an Island until end of
//	 turn."
//
// The restriction is ADR 0107 §2's (#1879, CR 508.1c), with the defending
// player worked out per target (CR 508.5, 508.5a), so in Commander it may
// attack only the opponents who control an Island. The ability is its own
// answer to it: ADR 0109 §1's (#1881) CR 305.7 type set makes an
// opponent's land an Island until end of turn (its other subtypes stay, CR
// 205.1a; it loses its rules-text abilities and taps for {U}, CR 305.6).
// The sacrificed Island is any permanent you control with that subtype,
// one an effect made an Island included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eb70d548-9769-4569-ae79-cddc0e623aef",
		Name:         "Dreamwinder",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			CantAttackUnlessDefendingPlayerControls(QuerySubtype("Island")),
		},
		Activated: []ActivatedAbility{{
			Label:   "{U}, Sacrifice an Island: Target land becomes an Island until end of turn.",
			Cost:    Plus(ManaCost("{U}"), SacrificeN(1, "an Island", HasSubtype("Island"))),
			Targets: TargetPermanent("target land", Land()),
			Effect:  TargetLandBecomesUntilEOT("Dreamwinder", "Island"),
		}},
	})
}
