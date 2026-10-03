package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flamestick Courier — Creature — Goblin (2/1) for {2}{R}:
//
//	"You may choose not to untap this creature during your untap step.
//	 {2}{R}, {T}: Target Goblin creature gets +2/+2 and has haste for as long as this creature remains tapped."
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
		OracleID:     "ea56ca6e-721d-43c7-9a30-4be7c85c67d3",
		Name:         "Flamestick Courier",
		Completeness: CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{mayChooseNotToUntapSelf("Flamestick Courier — you may choose not to untap this creature")},
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}, {T}: Target Goblin creature gets +2/+2 and has haste for as long as this creature remains tapped.",
			Cost:    Plus(ManaCost("{2}{R}"), TapCost()),
			Targets: TargetCreature("target Goblin creature", OfCreatureType("Goblin")),
			Effect:  TargetGetsWhileThisRemainsTapped("Flamestick Courier — +2/+2 and haste while tapped", game.ModifyPTMod(2, 2), game.AddKeywordsMod("haste")),
		}},
	})
}
