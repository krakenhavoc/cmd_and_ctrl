package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Borborygmos — Legendary Creature — Cyclops {3}{R}{R}{G}{G}, 6/7
// (Guildpact, #5599):
//
//	"Trample
//	 Whenever Borborygmos deals combat damage to a player, put a +1/+1
//	 counter on each creature you control."
//
// Abandoned Air Temple's body (b11PutCounterOnEachCreatureYouControl) on
// the combat-damage trigger: the set is read as the trigger resolves, so
// Borborygmos gets one too if it is still there, and a creature that
// arrives mid-placement gets none.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9ab0820b-4f5f-47b3-8fa9-43f802aed275",
		Name:            "Borborygmos",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Borborygmos — put a +1/+1 counter on each creature you control",
				b11PutCounterOnEachCreatureYouControl),
		},
	})
}
