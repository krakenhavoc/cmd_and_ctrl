package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "e447f7cc-4cdc-4a79-8058-8c1f1224d357",
		Name:         "Ghosthelm Courier",
		Completeness: effects.CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{effects.MayChooseNotToUntapSelf("Ghosthelm Courier — you may choose not to untap this creature")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}, {T}: Target Wizard creature gets +2/+2 and has shroud for as long as this creature remains tapped. (It can't be the target of spells or abilities.)",
			Cost:    effects.Plus(effects.ManaCost("{2}{U}"), effects.TapCost()),
			Targets: effects.TargetCreature("target Wizard creature", effects.OfCreatureType("Wizard")),
			Effect:  effects.TargetGetsWhileThisRemainsTapped("Ghosthelm Courier — +2/+2 and shroud while tapped", game.ModifyPTMod(2, 2), game.AddKeywordsMod("shroud")),
		}},
	})
}
