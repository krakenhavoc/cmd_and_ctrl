package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "1053a778-da6a-4b8e-81fb-537a355d4bfe",
		Name:            "Zhalfirin Commander",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordFlanking},
		Triggered:       []game.TriggeredAbility{effects.RetiredFlankingRow("Zhalfirin Commander")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}{W}: Target Knight creature gets +1/+1 until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)),
			Cost:    effects.ManaCost("{1}{W}{W}"),
			Targets: effects.TargetCreature("target Knight creature", effects.HasSubtype("Knight")),
			Effect:  effects.PumpAndGrantFirstTarget(1, 1, "Zhalfirin Commander — +1/+1 until end of turn"),
		}},
	})
}
