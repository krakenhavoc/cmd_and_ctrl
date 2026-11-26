package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Screeching Bat // Stalking Vampire — {2}{B} Creature — Bat 2/2 // Creature —
// Vampire 5/5 (Innistrad, S82 OLD b445, #5572):
//
//	Front: "Flying
//	        At the beginning of your upkeep, you may pay {2}{B}{B}. If you do,
//	        transform this creature."
//	Back:  "At the beginning of your upkeep, you may pay {2}{B}{B}. If you do,
//	        transform this creature."
//
// Each face prints the same upkeep trigger. The payment is asked as the
// trigger resolves, and the turn over is ignored if the creature has already
// turned over since (CR 701.27f).
//
// No simplification.
func init() {
	const oracle = "209de43d-0cdd-4075-9f41-88e3d2ff90df"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Screeching Bat",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{screechingBatUpkeep("Screeching Bat")},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Stalking Vampire",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{screechingBatUpkeep("Stalking Vampire")},
	})
}

func screechingBatUpkeep(name string) game.TriggeredAbility {
	return effects.AtYourUpkeep(name+" — you may pay {2}{B}{B}; if you do, transform this creature",
		func(g *game.Game, item *game.StackItem) error {
			return effects.MayPay{
				Chooser:  item.Controller,
				Cost:     "{2}{B}{B}",
				Question: name + " — pay {2}{B}{B} to transform it?",
				Then:     screechingBatPaid,
			}.Apply(effects.NewContext(g, item))
		})
}

// screechingBatPaid is the paid half: transform it. An on-disk key: never
// renamed, never reused.
var screechingBatPaid = effects.PayUnlessBranches("pay-unless/screeching-bat-transform",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error { return effects.TransformThis{}.Apply(ctx) }, nil)
