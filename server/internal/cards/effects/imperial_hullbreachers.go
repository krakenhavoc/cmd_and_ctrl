package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Imperial Hullbreachers — Creature — Klingon Warrior {2}{W}, 2/2 (Star
// Trek Commander, #4974):
//
//	"Squad {2} (As an additional cost to cast this spell, you may pay {2}
//	 any number of times. When this creature enters, create that many
//	 tokens that are copies of it.)
//	 Flying
//	 Battle cry (Whenever this creature attacks, each other attacking
//	 creature gets +1/+0 until end of turn.)"
//
// Squad is CR 702.157a's optional cost paid any number of times and its
// enters trigger (SquadMana, Squad: tmc_b4_helpers.go). The cap of 20
// payments on one announcement is the engine's, not the card's, the
// multikicker reason: forty extra mana is beyond any real board. Battle
// cry is CR 702.91a's trigger row (BattleCry, battle_cry.go). The token
// copies were never cast, so they make no copies of their own, but each
// flies and has battle cry.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "14c52244-4d09-49b9-b7c2-17ef1ffa2561",
		Name:            "Imperial Hullbreachers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		OptionalCosts:   []game.AdditionalCost{SquadMana("{2}", 20)},
		Triggered: []game.TriggeredAbility{
			Squad("Imperial Hullbreachers"),
			BattleCry("Imperial Hullbreachers"),
		},
	})
}
