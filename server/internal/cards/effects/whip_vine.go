package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Whip Vine — Creature — Plant Wall {2}{G}, 1/4:
//
//	"Defender; reach
//	 You may choose not to untap this creature during your untap step.
//	 {T}: Tap target creature with flying blocked by this creature. That creature doesn't untap during its controller's untap step for as long as this creature remains tapped."
//
// #1863 and ADR 0109 §3: "blocked by this creature" is BlockedBySource,
// with the flying test beside it in the same clause, so a target that
// loses flying or leaves combat before resolution is illegal (CR 608.2b).
// The hold is Rust Tick's untap hold with the WhileSourceRemainsTapped
// duration, as on Mole Worms: decline to untap the Vine and the
// creature stays tapped; untap it and the hold ends for good.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4e90ddce-1765-44e4-9b3c-e576f3a1bcd4",
		Name:            "Whip Vine",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender", "reach"},
		UntapOptOuts:    []game.UntapOptOut{mayChooseNotToUntapSelf("Whip Vine — you may choose not to untap this creature")},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Tap target creature with flying blocked by this creature. That creature doesn't untap during its controller's untap step for as long as this creature remains tapped.",
			Cost:    TapCost(),
			Targets: BlockedBySource(TargetCreature("target creature with flying blocked by this creature", HasKeyword("flying"))),
			Effect:  TapTargetsAndHoldWhileThisRemainsTapped,
		}},
	})
}
