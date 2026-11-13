package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zhalfirin Commander — Creature — Human Knight {2}{W}, 2/2 (Mirage,
// #5633):
//
//	"Flanking (Whenever a creature without flanking blocks this creature,
//	 the blocking creature gets -1/-1 until end of turn.)
//	 {1}{W}{W}: Target Knight creature gets +1/+1 until end of turn."
//
// Flanking is flanking.go's trigger with the declared keyword (#5757,
// CR 702.25a). The ability is Dwarven Lieutenant's shape at +1/+1 over a
// Knight: the target must be a Knight as the ability is activated and
// again as it resolves (CR 608.2b), any player's Knight, the Commander
// itself included. No {T} in the cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1053a778-da6a-4b8e-81fb-537a355d4bfe",
		Name:            "Zhalfirin Commander",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordFlanking},
		Triggered:       []game.TriggeredAbility{Flanking("Zhalfirin Commander")},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}{W}: Target Knight creature gets +1/+1 until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 1, 1)),
			Cost:    ManaCost("{1}{W}{W}"),
			Targets: TargetCreature("target Knight creature", HasSubtype("Knight")),
			Effect:  pumpAndGrantFirstTarget(1, 1, "Zhalfirin Commander — +1/+1 until end of turn"),
		}},
	})
}
