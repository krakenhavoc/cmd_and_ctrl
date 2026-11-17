package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invasion of Kaladesh — Battle — Siege, defense 4, for {U}{R} (March of
// the Machine, #5477):
//
//	"(As a Siege enters, choose an opponent to protect it. You and
//	  others can attack it. When it's defeated, exile it, then cast it
//	  transformed.)
//	 When this Siege enters, create a 1/1 colorless Thopter artifact
//	 creature token with flying."
//
// The reminder text is engine behaviour keyed on the card type
// (game/battle.go). This Spec carries the Thopter and the defeated
// trigger that hands over Aetherwing, Golden-Scale Flagship.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     invasionOfKaladeshOracleID,
		Name:         "Invasion of Kaladesh",
		Completeness: CompletenessFull,
		Battle: &BattleSpec{
			Defense: 4,
			Subtype: BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Invasion of Kaladesh — create a 1/1 colorless Thopter artifact creature token with flying",
				Do(CreateToken{Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: 1})),
			DefeatedTrigger("Invasion of Kaladesh — defeated: exile it, then cast Aetherwing, Golden-Scale Flagship", SiegeDefeated()),
		},
	})
}

// invasionOfKaladeshOracleID is shared with the back face's spec, which
// registers under this ID plus "#1".
const invasionOfKaladeshOracleID = "bc6f4f73-e10c-44ae-926d-9e5f6ddc14be"
