package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invasion of Ravnica // Guildpact Paragon — Battle — Siege, defense 4, for
// {5} (March of the Machine, S82 batch 23):
//
//	"(As a Siege enters, choose an opponent to protect it. You and others
//	  can attack it. When it's defeated, exile it, then cast it
//	  transformed.)
//	 When this Siege enters, exile target nonland permanent an opponent
//	 controls that isn't exactly two colors."
//
// The target is chosen as the trigger goes on the stack (CR 603.3d) and is
// re-checked as it resolves (CR 608.2b): colourless, monocoloured and
// three-or-more-colour permanents qualify, read through the layers, and a
// permanent that became exactly two colours in response is no longer a
// legal target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     invasionOfRavnicaOracleID,
		Name:         "Invasion of Ravnica",
		Completeness: CompletenessFull,
		Battle: &BattleSpec{
			Defense: 4,
			Subtype: BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Invasion of Ravnica — exile target nonland permanent an opponent controls that isn't exactly two colors",
				socB37ExileTheLegalTarget),
				TargetPermanent("target nonland permanent an opponent controls that isn't exactly two colors",
					Nonland(), OpponentControls(), Not(ExactlyTwoColors()))),
			DefeatedTrigger("Invasion of Ravnica — defeated: exile it, then cast Guildpact Paragon", SiegeDefeated()),
		},
	})
}

// invasionOfRavnicaOracleID is shared with the back face's spec, which
// registers under this ID plus "#1".
const invasionOfRavnicaOracleID = "78da7a38-0f08-4371-bda3-38f079f7724d"
