package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wasteland Raider — Creature — Human Mercenary {2}{B}{B}, 4/3
// (Fallout, #5469):
//
//	"Squad {2} (As an additional cost to cast this spell, you may pay
//	 {2} any number of times. When this creature enters, create that
//	 many tokens that are copies of it.)
//	 When this creature enters, each player sacrifices a creature of
//	 their choice."
//
// Imperial Hullbreachers' squad (SquadMana, Squad: CR 702.157a, the
// engine's cap of 20 payments on one announcement, multikicker's
// reason) and Fleshbag Marauder's edict, the controller included. Each
// token copy has the enters trigger too, so each makes every player
// sacrifice again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "29439669-c9be-46f3-a62f-16de51fefe88",
		Name:          "Wasteland Raider",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{SquadMana("{2}", 20)},
		Triggered: []game.TriggeredAbility{
			Squad("Wasteland Raider"),
			TriggerEdict(WhenThisEnters("Wasteland Raider — each player sacrifices a creature", Do(EachPlayerSacrifices{
				Match: Creature(),
				Label: "a creature",
			}))),
		},
	})
}
