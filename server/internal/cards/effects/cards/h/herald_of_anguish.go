package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Herald of Anguish — Creature — Demon {5}{B}{B}, 5/5:
//
//	"Improvise ... Flying
//	 At the beginning of your end step, each opponent discards a card.
//	 {1}{B}, Sacrifice an artifact: Target creature gets -2/-2 until end
//	 of turn."
//
// Improvise is Spec.TapCost; the end-step discard is Burglar Rat's
// per-opponent prompt; the sacrifice cost is Sai's SacrificeN. No
// simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "97b559fe-b09d-4883-9509-c45ace2d8487",
		Name:            "Herald of Anguish",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		TapCost:         effects.Improvise(),
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Herald of Anguish — each opponent discards a card", effects.EachOpponentDiscardsOne),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice an artifact: Target creature gets -2/-2 until end of turn.",
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -2, -2)),
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.SacrificeN(1, "an artifact", effects.Artifact())),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.VsPumpTheTarget(effects.NewContext(g, item), -2, -2, "Herald of Anguish — -2/-2")
			},
		}},
	})
}
