package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invasion of Segovia — Battle — Siege, defense 4, for {2}{U} (March of
// the Machine, #5647):
//
//	"(As a Siege enters, choose an opponent to protect it. You and
//	  others can attack it. When it's defeated, exile it, then cast it
//	  transformed.)
//	 When this Siege enters, create two 1/1 blue Kraken creature tokens
//	 with trample."
//
// The reminder text is engine behaviour keyed on the card type
// (game/battle.go). This Spec carries the tokens and the defeated
// trigger that hands over Caetus, Sea Tyrant of Segovia.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     invasionOfSegoviaOracleID,
		Name:         "Invasion of Segovia",
		Completeness: CompletenessFull,
		Battle: &BattleSpec{
			Defense: 4,
			Subtype: BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Invasion of Segovia — create two 1/1 blue Kraken tokens with trample",
				Do(CreateToken{Template: TokenCard("1/1 blue Kraken with trample"), N: 2})),
			DefeatedTrigger("Invasion of Segovia — defeated: exile it, then cast Caetus, Sea Tyrant of Segovia", SiegeDefeated()),
		},
	})
}

// invasionOfSegoviaOracleID is shared with the back face's spec, which
// registers under this ID plus "#1".
const invasionOfSegoviaOracleID = "7b3c0d89-55ed-43d7-a00a-7b060857ec4e"
