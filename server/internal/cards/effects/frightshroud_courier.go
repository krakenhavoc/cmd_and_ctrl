package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frightshroud Courier — Creature — Zombie (2/1) for {2}{B}:
//
//	"You may choose not to untap this creature during your untap step.
//	 {2}{B}, {T}: Target Zombie creature gets +2/+2 and has fear for as long as this creature remains tapped. (It can't be blocked except by artifact creatures and/or black creatures.)"
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
		OracleID:     "10e3e284-943a-45aa-accb-bb7d2a420098",
		Name:         "Frightshroud Courier",
		Completeness: CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{mayChooseNotToUntapSelf("Frightshroud Courier — you may choose not to untap this creature")},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}, {T}: Target Zombie creature gets +2/+2 and has fear for as long as this creature remains tapped. (It can't be blocked except by artifact creatures and/or black creatures.)",
			Cost:    Plus(ManaCost("{2}{B}"), TapCost()),
			Targets: TargetCreature("target Zombie creature", OfCreatureType("Zombie")),
			Effect:  TargetGetsWhileThisRemainsTapped("Frightshroud Courier — +2/+2 and fear while tapped", game.ModifyPTMod(2, 2), game.AddKeywordsMod("fear")),
		}},
	})
}
