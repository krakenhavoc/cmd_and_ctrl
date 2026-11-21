package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emberheart Challenger — Creature — Mouse Warrior {1}{R}, 2/2
// (Bloomburrow, #5230):
//
//	"Haste
//	 Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)
//	 Valiant — Whenever this creature becomes the target of a spell or
//	 ability you control for the first time each turn, exile the top card
//	 of your library. Until end of turn, you may play that card."
//
// Haste and prowess are printed keywords; prowess is the engine's keyword
// trigger (game/prowess.go). Valiant is the shared trigger (valiant.go)
// with Outpost Siege's impulse exile: the top card of your library, which
// you may play (a land included) until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5b4b21ed-c24e-4038-83fb-bd7f3c3426cd",
		Name:            "Emberheart Challenger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste", "prowess"},
		Triggered: []game.TriggeredAbility{
			effects.Valiant("Emberheart Challenger — Valiant: exile the top card of your library; until end of turn, you may play it",
				effects.OutpostSiegeImpulse),
		},
	})
}
