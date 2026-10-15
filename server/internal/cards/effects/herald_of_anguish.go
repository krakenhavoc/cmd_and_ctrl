package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "97b559fe-b09d-4883-9509-c45ace2d8487",
		Name:            "Herald of Anguish",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		TapCost:         Improvise(),
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Herald of Anguish — each opponent discards a card", eachOpponentDiscardsOne),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice an artifact: Target creature gets -2/-2 until end of turn.",
			Cost:    Plus(ManaCost("{1}{B}"), SacrificeN(1, "an artifact", Artifact())),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return vsPumpTheTarget(NewContext(g, item), -2, -2, "Herald of Anguish — -2/-2")
			},
		}},
	})
}
