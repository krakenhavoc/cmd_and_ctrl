package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Invasion of Moag // Bloomwielder Dryads — March of the Machine (#5477).
//
// Front, Battle — Siege 5, {2}{G}{W}: "When this Siege enters, put a +1/+1
// counter on each creature you control." Back, Bloomwielder Dryads, Creature —
// Dryad 3/3: "Ward {2}. At the beginning of your end step, put a +1/+1 counter
// on target creature you control."
//
// No simplification.
const InvasionOfMoagOracleID = "a9855326-8958-4b03-b81b-1f9e4d1335b1"

func init() {
	effects.Register(effects.Spec{
		OracleID:     InvasionOfMoagOracleID,
		Name:         "Invasion of Moag",
		Completeness: effects.CompletenessFull,
		Battle:       &effects.BattleSpec{Defense: 5, Subtype: effects.BattleSubtypeSiege},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Invasion of Moag — a +1/+1 counter on each creature you control", effects.B11PutCounterOnEachCreatureYouControl),
			effects.DefeatedTrigger("Invasion of Moag — defeated: exile it, then cast Bloomwielder Dryads", effects.SiegeDefeated()),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     InvasionOfMoagOracleID + "#1",
		Name:         "Bloomwielder Dryads",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Bloomwielder Dryads — ward {2}"),
			effects.Targeting(effects.AtYourEndStep("Bloomwielder Dryads — put a +1/+1 counter on target creature you control", effects.PlusOneCounterOnTheTarget),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
