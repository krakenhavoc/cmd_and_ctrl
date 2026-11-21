package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Baleful Ammit — Creature — Crocodile Demon {2}{B}, 4/3
// (Amonkhet, #5535):
//
//	"Lifelink
//	 When this creature enters, put a -1/-1 counter on target creature you control."
//
// Channeler Initiate's enters trigger with one counter; the target may
// be this creature itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1cfa9f7d-ef1f-4907-aeb1-15c0df33fccc",
		Name:            "Baleful Ammit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Baleful Ammit — put a -1/-1 counter on target creature you control", effects.Cmdr2B25CountersOnTarget(1)),
				effects.TargetCreature("target creature you control", effects.YouControl()),
			),
		},
	})
}
