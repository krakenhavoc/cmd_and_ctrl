package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodroot Apothecary — Creature — Squirrel Druid {2}{G}, 3/3 (Bloomburrow
// Commander, #5368):
//
//	"Toxic 2 (Players dealt combat damage by this creature also get two
//	 poison counters. A player with ten or more poison counters loses the
//	 game.)
//	 When this creature enters, you and target opponent each create a
//	 Treasure token.
//	 Whenever an opponent sacrifices a noncreature token, that player gets
//	 two poison counters."
//
// Toxic is the engine's damage tail (infect_wither_toxic.go). The enters
// trigger targets an opponent when it goes on the stack; each of you gets
// an untapped Treasure as it resolves, and a target that is gone gives its
// controller's Treasure only. The sacrifice trigger reads the permanent as
// it stood before it was sacrificed (a Treasure sacrificed for mana is
// one), so a creature token, or a token the Apothecary's own controller
// sacrifices, gives nothing. Each sacrifice is its own trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "09ecb8df-2a17-4e5a-814c-474bce54794d",
		Name:            "Bloodroot Apothecary",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"toxic 2"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Bloodroot Apothecary — you and target opponent each create a Treasure",
					cmdr2S33YouAndTargetOpponentEachCreateATreasure),
				TargetPlayer("target opponent", Opponent())),
			On(game.EventSacrifice, cmdr2S33OpponentSacrificedANoncreatureToken,
				"Bloodroot Apothecary — that player gets two poison counters", cmdr2S33SacrificerGetsTwoPoison),
		},
	})
}
