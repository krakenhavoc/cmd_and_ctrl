package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hisoka's Guard — Creature — Human Wizard (1/1) for {1}{U}:
//
//	"You may choose not to untap this creature during your untap step.
//	 {1}{U}, {T}: Target creature you control other than this creature has shroud for as long as this creature remains tapped. (It can't be the target of spells or abilities.)"
//
// ADR 0109 owner decision 4: a continuous effect from a resolving
// ability (CR 611.2), timed by "for as long as this creature remains
// tapped" (CR 611.2b). The effect is a ScopedEffect record, data a
// restore point carries; it ends the moment the creature untaps, and
// for good. "You may choose not to untap" is the untap opt-out, which
// is how its controller keeps the effect going.
//
// "Other than this creature" is the clause's ExcludeSource
// (effects.Another).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6136f019-da8f-471e-9ac5-41e535aa8a21",
		Name:         "Hisoka's Guard",
		Completeness: CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{mayChooseNotToUntapSelf("Hisoka's Guard — you may choose not to untap this creature")},
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}, {T}: Target creature you control other than this creature has shroud for as long as this creature remains tapped. (It can't be the target of spells or abilities.)",
			Cost:    Plus(ManaCost("{1}{U}"), TapCost()),
			Targets: Another(TargetCreature("target creature you control other than this creature", YouControl())),
			Effect:  TargetGetsWhileThisRemainsTapped("Hisoka's Guard — shroud while tapped", game.AddKeywordsMod("shroud")),
		}},
	})
}
