package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zelyon Sword — Artifact for {3}:
//
//	"You may choose not to untap this artifact during your untap step.
//	 {3}, {T}: Target creature gets +2/+0 for as long as this artifact remains tapped."
//
// ADR 0109 owner decision 4: a continuous effect from a resolving
// ability (CR 611.2), timed by "for as long as this artifact remains
// tapped" (CR 611.2b). The effect is a ScopedEffect record, data a
// restore point carries; it ends the moment the artifact untaps, and
// for good. "You may choose not to untap" is the untap opt-out, which
// is how its controller keeps the effect going.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "991038e0-f8d6-42e6-b8a2-3d77aae7a256",
		Name:         "Zelyon Sword",
		Completeness: CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{mayChooseNotToUntapSelf("Zelyon Sword — you may choose not to untap this artifact")},
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}: Target creature gets +2/+0 for as long as this artifact remains tapped.",
			Cost:    Plus(ManaCost("{3}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  TargetGetsWhileThisRemainsTapped("Zelyon Sword — +2/+0 while tapped", game.ModifyPTMod(2, 0)),
		}},
	})
}
