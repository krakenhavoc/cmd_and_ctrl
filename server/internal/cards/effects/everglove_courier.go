package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Everglove Courier — Creature — Elf (2/1) for {2}{G}:
//
//	"You may choose not to untap this creature during your untap step.
//	 {2}{G}, {T}: Target Elf creature gets +2/+2 and has trample for as long as this creature remains tapped."
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
		OracleID:     "31da3df3-5b79-42d0-9018-76bc73411954",
		Name:         "Everglove Courier",
		Completeness: CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{mayChooseNotToUntapSelf("Everglove Courier — you may choose not to untap this creature")},
		Activated: []ActivatedAbility{{
			Label:   "{2}{G}, {T}: Target Elf creature gets +2/+2 and has trample for as long as this creature remains tapped.",
			Cost:    Plus(ManaCost("{2}{G}"), TapCost()),
			Targets: TargetCreature("target Elf creature", OfCreatureType("Elf")),
			Effect:  TargetGetsWhileThisRemainsTapped("Everglove Courier — +2/+2 and trample while tapped", game.ModifyPTMod(2, 2), game.AddKeywordsMod("trample")),
		}},
	})
}
