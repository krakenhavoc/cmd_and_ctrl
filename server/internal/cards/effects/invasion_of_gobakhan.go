package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invasion of Gobakhan // Lightshield Array — Battle — Siege, defense 3, for
// {1}{W} (March of the Machine, S82 batch 23):
//
//	"(As a Siege enters, choose an opponent to protect it. You and others
//	  can attack it. When it's defeated, exile it, then cast it
//	  transformed.)
//	 When this Siege enters, look at target opponent's hand. You may exile a
//	 nonland card from it. For as long as that card remains exiled, its owner
//	 may play it. A spell cast this way costs {2} more to cast."
//
// Zero, Incorporeal Sensate's clause word for word, so the body is shared
// (lookAtOpponentsHandExileTaxed): a looked-at hand, one optional nonland
// pick, an exile whose play permission belongs to the card's owner while it
// stays exiled and carries the {2} increase (CR 601.2f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     invasionOfGobakhanOracleID,
		Name:         "Invasion of Gobakhan",
		Completeness: CompletenessFull,
		Battle: &BattleSpec{
			Defense: 3,
			Subtype: BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Invasion of Gobakhan — look at target opponent's hand; you may exile a nonland card from it",
				lookAtOpponentsHandExileTaxed),
				TargetPlayer("target opponent", Opponent())),
			DefeatedTrigger("Invasion of Gobakhan — defeated: exile it, then cast Lightshield Array", SiegeDefeated()),
		},
	})
}

// invasionOfGobakhanOracleID is shared with the back face's spec, which
// registers under this ID plus "#1".
const invasionOfGobakhanOracleID = "86c32a56-3272-4738-b0b3-a8aa49a17ce4"
