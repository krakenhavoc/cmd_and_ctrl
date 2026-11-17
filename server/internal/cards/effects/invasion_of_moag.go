package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invasion of Moag // Bloomwielder Dryads — March of the Machine (#5477).
//
// Front, Battle — Siege 5, {2}{G}{W}: "When this Siege enters, put a +1/+1
// counter on each creature you control." Back, Bloomwielder Dryads, Creature —
// Dryad 3/3: "Ward {2}. At the beginning of your end step, put a +1/+1 counter
// on target creature you control."
//
// No simplification.
const invasionOfMoagOracleID = "a9855326-8958-4b03-b81b-1f9e4d1335b1"

func init() {
	Register(Spec{
		OracleID:     invasionOfMoagOracleID,
		Name:         "Invasion of Moag",
		Completeness: CompletenessFull,
		Battle:       &BattleSpec{Defense: 5, Subtype: BattleSubtypeSiege},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Invasion of Moag — a +1/+1 counter on each creature you control", b11PutCounterOnEachCreatureYouControl),
			DefeatedTrigger("Invasion of Moag — defeated: exile it, then cast Bloomwielder Dryads", SiegeDefeated()),
		},
	})
	Register(Spec{
		OracleID:     invasionOfMoagOracleID + "#1",
		Name:         "Bloomwielder Dryads",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Bloomwielder Dryads — ward {2}"),
			Targeting(AtYourEndStep("Bloomwielder Dryads — put a +1/+1 counter on target creature you control", PlusOneCounterOnTheTarget),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
