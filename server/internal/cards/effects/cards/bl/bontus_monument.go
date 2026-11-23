package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bontu's Monument — Legendary Artifact {3}:
//
//	"Black creature spells you cast cost {1} less to cast.
//	 Whenever you cast a creature spell, each opponent loses 1 life
//	 and you gain 1 life."
//
// The drain trigger fires on EVERY creature spell you cast, not just
// black ones — the cost clause and the trigger name different sets on
// purpose, exactly as printed. The drain is
// Do(EachOpponentLosesLife, GainLife), a footprinted pair (#3013).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "940ba435-7abc-40f8-a5af-c1c653b3284e",
		Name:         "Bontu's Monument",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Black creature spells you cast cost {1} less to cast.",
				effects.YourSpell(), effects.CreatureSpell(), effects.ColoredSpell("B")),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Creature(),
				"Bontu's Monument — each opponent loses 1 life and you gain 1 life",
				effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1})),
		},
	})
}
