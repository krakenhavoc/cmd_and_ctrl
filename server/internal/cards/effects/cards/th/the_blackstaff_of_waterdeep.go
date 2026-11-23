package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Blackstaff of Waterdeep — Legendary Artifact for {U}:
//
//	"You may choose not to untap The Blackstaff of Waterdeep during your untap step.
//	 Animate Walking Statue — {1}{U}, {T}: Another target nontoken artifact you control becomes a 4/4 artifact creature for as long as The Blackstaff of Waterdeep remains tapped. Activate only as a sorcery."
//
// ADR 0109 owner decision 4. "Becomes a 4/4 artifact creature" is a
// type change (layer 4) and a base power and toughness (layer 7b) from
// a resolving ability, kept for as long as the Blackstaff remains
// tapped (CR 611.2b). The target is already an artifact, so it keeps
// every type it had and adds creature. "Animate Walking Statue" is an
// ability word with no rules meaning (CR 207.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a4faed58-4b1b-4cab-8a12-4fcd20b320ab",
		Name:         "The Blackstaff of Waterdeep",
		Completeness: effects.CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{effects.MayChooseNotToUntapSelf("The Blackstaff of Waterdeep — you may choose not to untap The Blackstaff of Waterdeep")},
		Activated: []effects.ActivatedAbility{{
			Label:        "Animate Walking Statue \u2014 {1}{U}, {T}: Another target nontoken artifact you control becomes a 4/4 artifact creature for as long as The Blackstaff of Waterdeep remains tapped. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{1}{U}"), effects.TapCost()),
			SorcerySpeed: true,
			Targets: effects.Another(effects.TargetPermanent("another target nontoken artifact you control",
				effects.Artifact(), effects.YouControl(), effects.Not(effects.IsTokenPredicate()))),
			Effect: effects.TargetGetsWhileThisRemainsTapped("The Blackstaff of Waterdeep — a 4/4 artifact creature while tapped",
				game.AddTypesMod("Artifact", "Creature"), game.SetBasePowerMod(4), game.SetBaseToughnessMod(4)),
		}},
	})
}
