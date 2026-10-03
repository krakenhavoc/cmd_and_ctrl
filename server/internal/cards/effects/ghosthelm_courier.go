package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ghosthelm Courier — Creature — Human Wizard (2/1) for {2}{U}:
//
//	"You may choose not to untap this creature during your untap step.
//	 {2}{U}, {T}: Target Wizard creature gets +2/+2 and has shroud for as long as this creature remains tapped. (It can't be the target of spells or abilities.)"
//
// ADR 0109 owner decision 4: a continuous effect from a resolving
// ability (CR 611.2), timed by "for as long as this creature remains
// tapped" (CR 611.2b). The effect is a ScopedEffect record, data a
// restore point carries; it ends the moment the creature untaps, and
// for good. "You may choose not to untap" is the untap opt-out, which
// is how its controller keeps the effect going.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e447f7cc-4cdc-4a79-8058-8c1f1224d357",
		Name:         "Ghosthelm Courier",
		Completeness: CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{mayChooseNotToUntapSelf("Ghosthelm Courier — you may choose not to untap this creature")},
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}, {T}: Target Wizard creature gets +2/+2 and has shroud for as long as this creature remains tapped. (It can't be the target of spells or abilities.)",
			Cost:    Plus(ManaCost("{2}{U}"), TapCost()),
			Targets: TargetCreature("target Wizard creature", OfCreatureType("Wizard")),
			Effect:  TargetGetsWhileThisRemainsTapped("Ghosthelm Courier — +2/+2 and shroud while tapped", game.ModifyPTMod(2, 2), game.AddKeywordsMod("shroud")),
		}},
	})
}
