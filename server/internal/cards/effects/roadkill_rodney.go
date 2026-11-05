package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Roadkill Rodney — {2} Artifact Creature — Robot 2/1 (Teenage Mutant
// Ninja Turtles Commander, #5011):
//
//	"Squad {3} (As an additional cost to cast this spell, you may pay {3}
//	 any number of times. When this creature enters, create that many
//	 tokens that are copies of it.)
//	 Deathtouch
//	 Whenever this creature deals combat damage to a player, create a
//	 Mutagen token."
//
// Squad is CR 702.157a's optional cost paid any number of times and its
// enters trigger (SquadMana, Squad: tmc_b4_helpers.go). The cap of 20
// payments on one announcement is the engine's, not the card's, the
// multikicker reason: sixty extra mana is beyond any real board. The
// copies were never cast, so they paid no squad cost and make no copies
// of their own, but each has deathtouch and the Mutagen trigger. The
// Mutagen token is the catalog's CR 111.10v template (MutagenToken).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cd8761ef-b57b-456b-aa12-374ac825087e",
		Name:            "Roadkill Rodney",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		OptionalCosts:   []game.AdditionalCost{SquadMana("{3}", 20)},
		Triggered: []game.TriggeredAbility{
			Squad("Roadkill Rodney"),
			WheneverThisDealsCombatDamageToAPlayer("Roadkill Rodney — create a Mutagen token",
				Do(CreateToken{Template: MutagenToken(), N: 1})),
		},
	})
}
