package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shredder, Unrelenting — Legendary Creature — Human Ninja {4}{B}, 6/4
// (Teenage Mutant Ninja Turtles, #4457):
//
//	"Sneak {3}{B} (You may cast this spell for {3}{B} if you also return an
//	 unblocked attacker you control to hand during the declare blockers
//	 step. He enters tapped and attacking.)
//	 Deathtouch
//	 Whenever Shredder enters or attacks, another target creature you
//	 control gains deathtouch until end of turn."
//
// Sneak is the shared alternative cost (sneak.go, CR 702.190). The trigger
// is Starling, Aerial Ally's keyword grant on Sun Titan's two conditions
// (WhenThisEntersOrAttacks). Sneaked in, he enters (the trigger fires) but
// was never declared as an attacker (CR 508.4), so it fires once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "59d96fc0-6580-43c2-8c6a-dcf983644561",
		Name:             "Shredder, Unrelenting",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"deathtouch"},
		AlternativeCosts: []game.AlternativeCost{effects.Sneak("{3}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEntersOrAttacks("Shredder, Unrelenting — another target creature you control gains deathtouch until end of turn",
				effects.EbTargetCreatureGainsUntilEOT("deathtouch", "Shredder, Unrelenting — deathtouch until end of turn")),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
		},
	})
}
