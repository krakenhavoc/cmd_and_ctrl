package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pearlspear Courier — Creature — Human Soldier (2/1) for {2}{W}:
//
//	"You may choose not to untap this creature during your untap step.
//	 {2}{W}, {T}: Target Soldier creature gets +2/+2 and has vigilance for as long as this creature remains tapped."
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
		OracleID:     "45c48bd8-9150-4699-a1eb-f4883165a4a0",
		Name:         "Pearlspear Courier",
		Completeness: CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{mayChooseNotToUntapSelf("Pearlspear Courier — you may choose not to untap this creature")},
		Activated: []ActivatedAbility{{
			Label:   "{2}{W}, {T}: Target Soldier creature gets +2/+2 and has vigilance for as long as this creature remains tapped.",
			Cost:    Plus(ManaCost("{2}{W}"), TapCost()),
			Targets: TargetCreature("target Soldier creature", OfCreatureType("Soldier")),
			Effect:  TargetGetsWhileThisRemainsTapped("Pearlspear Courier — +2/+2 and vigilance while tapped", game.ModifyPTMod(2, 2), game.AddKeywordsMod("vigilance")),
		}},
	})
}
